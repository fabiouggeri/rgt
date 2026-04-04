#include <stdio.h>
#include <stdlib.h>

#include "cfl_buffer.h"
#include "cfl_lock.h"
#include "cfl_socket.h"
#include "cfl_str.h"

#include "rgt_channel.h"
#include "rgt_channel_unidirectional.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_types.h"
#include "rgt_util.h"

#define SINGLE_CHANNEL(c) ((RGT_SINGLE_CHANNELP)c)

#define RESOURCE_ALLOCATION_ERROR(ct) rgt_error_set(ct, RGT_ERROR_ALLOC_RESOURCE, "Error allocating resource")

#define GET_UINT32(b) ((CFL_UINT32)((b[0] & 0xFF) | ((b[1] & 0xFF) << 8) | ((b[2] & 0xFF) << 16) | ((b[3] & 0xFF) << 24)));
#define GET_BYTE(b, i) ((CFL_INT8)(b[i] & 0xFF))

typedef struct _RGT_SINGLE_CHANNEL {
      RGT_CHANNEL channel;
      CFL_SOCKET socket;
      RGT_LOCK ioLock;
      CFL_BOOL isActive;
} RGT_SINGLE_CHANNEL, *RGT_SINGLE_CHANNELP;

static CFL_STRP bufferToHex(const char *label, CFL_BUFFERP buffer, CFL_UINT32 bodyStart) {
   CFL_UINT32 labelLen = (CFL_UINT32)strlen(label);
   CFL_UINT32 bufferLen = cfl_buffer_length(buffer);
   CFL_STRP pStr = cfl_str_new(labelLen + bufferLen * 2);
   CFL_UINT8 *data = cfl_buffer_getDataPtr(buffer);
   CFL_UINT32 i;

   cfl_str_appendLen(pStr, label, labelLen);
   if (bodyStart > 0) {
      cfl_str_appendFormat(pStr, " Packet Len(%u)=0x", bodyStart);
      for (i = 0; i < bodyStart; i++) {
         cfl_str_appendFormat(pStr, "%02X", data[i]);
      }
   }
   cfl_str_appendFormat(pStr, " Packet Data(%u)=0x", bufferLen - bodyStart);
   for (i = bodyStart; i < bufferLen; i++) {
      cfl_str_appendFormat(pStr, "%02X", data[i]);
   }
   return pStr;
}

static void channel_closeSocket(RGT_SINGLE_CHANNELP channel) {
   CFL_SOCKET socket;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_closeSocket", (NULL));
   socket = channel->socket;
   channel->socket = CFL_INVALID_SOCKET;
   if (socket != CFL_INVALID_SOCKET) {
      cfl_socket_shutdown(socket, CFL_TRUE, CFL_TRUE);
      cfl_socket_close(socket);
   } else {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_closeSocket(). socket already closed."));
   }
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_closeSocket", (NULL));
}

static CFL_BOOL channel_isOpen(RGT_CHANNELP c) {
   return c != NULL && SINGLE_CHANNEL(c)->isActive && SINGLE_CHANNEL(c)->socket != CFL_INVALID_SOCKET;
}

static void channel_close(RGT_CHANNELP c) {
   RGT_SINGLE_CHANNELP channel;
   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_close", (NULL));
   if (c == NULL) {
      RGT_LOG_ERROR(("rgt_channel_unidirectional.channel_close(). channel is NULL."));
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_free", (NULL));
      return;
   }
   channel = SINGLE_CHANNEL(c);
   RGT_LOCK_ACQUIRE(channel->ioLock);
   channel->isActive = CFL_FALSE;
   channel_closeSocket(channel);
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_close", (NULL));
}

static void channel_free(RGT_CHANNELP c) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_free", (NULL));
   if (channel != NULL) {
      RGT_LOCK_FREE(channel->ioLock);
      RGT_HB_FREE(channel);
   }
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_free", (NULL));
}

static CFL_BOOL channel_waitData(RGT_CHANNELP c, CFL_UINT32 timeout, CFL_BOOL *error) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   int retVal;
   CFL_BOOL bSuccess;
   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_waitData", (NULL));
   if (!channel_isOpen(c)) {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_waitData(): connection is closed"));
      *error = CFL_TRUE;
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_waitData", (NULL));
      return CFL_FALSE;
   }

   RGT_LOCK_ACQUIRE(channel->ioLock);
   retVal = cfl_socket_selectRead(channel->socket, timeout > 0 ? timeout : CFL_WAIT_FOREVER);
   RGT_LOCK_RELEASE(channel->ioLock);
   *error = retVal < 0 ? CFL_TRUE : CFL_FALSE;
   bSuccess = retVal > 0 ? CFL_TRUE : CFL_FALSE;
   RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_waitData() result=%s", bSuccess ? "true" : "false"));
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_waitData", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_hasData(RGT_CHANNELP c) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOCK_ACQUIRE(channel->ioLock);
   bSuccess = cfl_socket_selectRead(channel->socket, 0);
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_hasData() result=%s", bSuccess ? "true" : "false"));
   return bSuccess;
}

static CFL_BOOL channel_readAll(RGT_SINGLE_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   int retVal;
   RGT_PACKET_LEN_TYPE packetLen;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_readAll", (NULL));
   cfl_buffer_reset(buffer);

   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): channel is closed"));
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("channel is closed"));
      return CFL_FALSE;
   }

   if (timeout > 0) {
      retVal = cfl_socket_selectRead(channel->socket, timeout);
      if (!channel_isOpen((RGT_CHANNELP)channel)) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): error waiting header. channel is closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("error waiting header. channel is closed"));
         return CFL_FALSE;
      } else if (retVal == CFL_SOCKET_ERROR) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error waiting data from socket: %ld",
                       cfl_socket_lastErrorCode());
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll",
                      ("error waiting header. socket error: %ld", cfl_socket_lastErrorCode()));
         return CFL_FALSE;
      } else if (retVal == 0) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): timeout waiting data."));
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("timeout waiting header"));
         return CFL_FALSE;
      }
   }
   // data must be here, so do a normal cfl_socket_receive()
   retVal = cfl_socket_receiveAll(channel->socket, (char *)&packetLen, RGT_PACKET_LEN_FIELD_SIZE);
   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): error reading header. channel is closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("error reading header. channel is closed"));
      return CFL_FALSE;
   } else if (retVal == CFL_SOCKET_ERROR) {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                    cfl_socket_lastErrorCode());
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll",
                   ("error reading header. socket error: %ld", cfl_socket_lastErrorCode()));
      return CFL_FALSE;
   } else if (retVal == 0) {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): error reading header. socket closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("error reading header. socket closed"));
      return CFL_FALSE;
   }

   RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): Data read. Packet Len(%d)=%#0*X", RGT_PACKET_LEN_FIELD_SIZE,
                  2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen));
   if (packetLen > 0) {
      if ((CFL_UINT32)packetLen > cfl_buffer_capacity(buffer) && !cfl_buffer_setCapacity(buffer, packetLen)) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_ALLOC_RESOURCE, "Error allocating resource");
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("error allocating resource"));
         return CFL_FALSE;
      }
      retVal = cfl_socket_receiveAll(channel->socket, (char *)cfl_buffer_getDataPtr(buffer), packetLen);
      if (!channel_isOpen((RGT_CHANNELP)channel)) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): error reading body. channel closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("error reading body. channel closed"));
         return CFL_FALSE;
      } else if (retVal == CFL_SOCKET_ERROR) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                       cfl_socket_lastErrorCode());
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll",
                      ("error reading body. socket error: %ld", cfl_socket_lastErrorCode()));
         return CFL_FALSE;
      } else if (retVal == 0) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAll(): error reading body. socket closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("error reading body. socket closed"));
         return CFL_FALSE;
      }
      cfl_buffer_setPosition(buffer, packetLen);
      cfl_buffer_flip(buffer);
   } else {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_PROTOCOL, "Zero length packet found. Header: %#0*X.",
                    2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen);
      cfl_buffer_reset(buffer);
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", ("zero length packet"));
      return CFL_FALSE;
   }
   ((RGT_CHANNELP)channel)->lastRead = CURRENT_TIME;
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAll", (NULL));
   return CFL_TRUE;
}

static CFL_BUFFERP channel_readAllBuffer(RGT_SINGLE_CHANNELP channel, CFL_UINT32 timeout) {
   int retVal;
   RGT_PACKET_LEN_TYPE packetLen;
   CFL_BUFFERP buffer;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_readAllBuffer", (NULL));

   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): channel is closed"));
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("channel is closed"));
      return NULL;
   }

   if (timeout > 0) {
      retVal = cfl_socket_selectRead(channel->socket, timeout);
      if (!channel_isOpen((RGT_CHANNELP)channel)) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): error waiting header. channel is closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("error waiting header. channel is closed"));
         return NULL;
      } else if (retVal == CFL_SOCKET_ERROR) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error waiting data from socket: %ld",
                       cfl_socket_lastErrorCode());
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer",
                      ("error waiting header. socket error: %ld", cfl_socket_lastErrorCode()));
         return NULL;
      } else if (retVal == 0) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): timeout waiting data."));
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("timeout waiting header"));
         return NULL;
      }
   }
   // data must be here, so do a normal cfl_socket_receive()
   retVal = cfl_socket_receiveAll(channel->socket, (char *)&packetLen, RGT_PACKET_LEN_FIELD_SIZE);
   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): error reading header. channel is closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("error reading header. channel is closed"));
      return NULL;
   } else if (retVal == CFL_SOCKET_ERROR) {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                    cfl_socket_lastErrorCode());
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer",
                   ("error reading header. socket error: %ld", cfl_socket_lastErrorCode()));
      return NULL;
   } else if (retVal == 0) {
      RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): error reading header. socket closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("error reading header. socket closed"));
      return NULL;
   }

   RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): Data read. Packet Len(%d)=%#0*X", RGT_PACKET_LEN_FIELD_SIZE,
                  2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen));
   if (packetLen > 0) {
      buffer = cfl_buffer_newCapacity((CFL_UINT32)packetLen);
      if (buffer == NULL) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_ALLOC_RESOURCE, "Error allocating resource");
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("error allocating resource"));
         return NULL;
      }
      retVal = cfl_socket_receiveAll(channel->socket, (char *)cfl_buffer_getDataPtr(buffer), packetLen);
      if (!channel_isOpen((RGT_CHANNELP)channel)) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): error reading body. channel closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("error reading body. channel closed"));
         cfl_buffer_free(buffer);
         return NULL;
      } else if (retVal == CFL_SOCKET_ERROR) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                       cfl_socket_lastErrorCode());
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer",
                      ("error reading body. socket error: %ld", cfl_socket_lastErrorCode()));
         cfl_buffer_free(buffer);
         return NULL;
      } else if (retVal == 0) {
         RGT_LOG_DEBUG(("rgt_channel_unidirectional.channel_readAllBuffer(): error reading body. socket closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("error reading body. socket closed"));
         cfl_buffer_free(buffer);
         return NULL;
      }
      cfl_buffer_setPosition(buffer, packetLen);
      cfl_buffer_flip(buffer);
   } else {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_PROTOCOL, "Zero length packet found. Header: %#0*X.",
                    2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen);
      cfl_buffer_reset(buffer);
      RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", ("zero length packet"));
      return NULL;
   }
   ((RGT_CHANNELP)channel)->lastRead = CURRENT_TIME;
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_readAllBuffer", (NULL));
   return buffer;
}

static CFL_BOOL channel_read(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_read", ("timeout=%u", timeout));
   RGT_LOCK_ACQUIRE(channel->ioLock);
   bSuccess = channel_readAll(channel, buffer, timeout);
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_read", (NULL));
   return bSuccess;
}

static CFL_BUFFERP channel_readBuffer(RGT_CHANNELP c, CFL_UINT32 timeout) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BUFFERP buffer;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_read", ("timeout=%u", timeout));
   RGT_LOCK_ACQUIRE(channel->ioLock);
   buffer = channel_readAllBuffer(channel, timeout);
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_read", (NULL));
   return buffer;
}

static CFL_BOOL channel_tryRead(RGT_CHANNELP c, CFL_BUFFERP buffer) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_tryRead", (NULL));
   RGT_LOCK_ACQUIRE(channel->ioLock);
   if (cfl_socket_selectRead(channel->socket, 0)) {
      bSuccess = channel_readAll(channel, buffer, 0);
      RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
   } else {
      bSuccess = CFL_FALSE;
   }
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_tryRead", (NULL));
   return bSuccess;
}

static CFL_BUFFERP channel_tryReadBuffer(RGT_CHANNELP c) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BUFFERP buffer;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_tryRead", (NULL));
   RGT_LOCK_ACQUIRE(channel->ioLock);
   if (cfl_socket_selectRead(channel->socket, 0)) {
      buffer = channel_readAllBuffer(channel, 0);
      RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
   } else {
      buffer = NULL;
   }
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_tryRead", (NULL));
   return buffer;
}

static CFL_BOOL channel_write(RGT_CHANNELP c, CFL_BUFFERP buffer) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BOOL bSuccess = CFL_TRUE;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_write", (NULL));
   cfl_buffer_flip(buffer);
   RGT_PUT_PACKET_LEN(buffer, cfl_buffer_length(buffer) - RGT_PACKET_LEN_FIELD_SIZE);
   cfl_buffer_rewind(buffer);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data write.", buffer, RGT_PACKET_LEN_FIELD_SIZE));
   RGT_LOCK_ACQUIRE(channel->ioLock);
   if (cfl_socket_sendAllBuffer(channel->socket, buffer)) {
      c->lastWrite = CURRENT_TIME;
   } else {
      bSuccess = CFL_FALSE;
      channel->isActive = CFL_FALSE;
      rgt_error_set(c->connectionType, RGT_ERROR_SOCKET, "Error sending data: %ld", cfl_socket_lastErrorCode());
   }
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_write", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_writeAndRead(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BOOL bSuccess;
   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_writeAndRead", (NULL));

   cfl_buffer_flip(buffer);
   RGT_PUT_PACKET_LEN(buffer, cfl_buffer_length(buffer) - RGT_PACKET_LEN_FIELD_SIZE);
   cfl_buffer_rewind(buffer);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data write.", buffer, RGT_PACKET_LEN_FIELD_SIZE));

   RGT_LOCK_ACQUIRE(channel->ioLock);
   if (cfl_socket_sendAllBuffer(channel->socket, buffer)) {
      c->lastWrite = CURRENT_TIME;
      bSuccess = channel_readAll(channel, buffer, timeout);
   } else {
      bSuccess = CFL_FALSE;
      channel->isActive = CFL_FALSE;
      rgt_error_set(c->connectionType, RGT_ERROR_SOCKET, "Error sending data: %ld", cfl_socket_lastErrorCode());
   }
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_writeAndRead", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_writeAndReadFirstCommand(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   RGT_SINGLE_CHANNELP channel = SINGLE_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_unidirectional.channel_writeAndReadFirstCommand", (NULL));
   cfl_buffer_flip(buffer);
   cfl_buffer_skip(buffer, RGT_MAGIC_NUMBER_SIZE);
   RGT_PUT_PACKET_LEN(buffer, cfl_buffer_length(buffer) - RGT_PACKET_LEN_FIELD_SIZE - RGT_MAGIC_NUMBER_SIZE);
   cfl_buffer_rewind(buffer);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data write.", buffer, 8));
   RGT_LOCK_ACQUIRE(channel->ioLock);
   bSuccess = cfl_socket_sendAllBuffer(channel->socket, buffer);
   c->lastWrite = CURRENT_TIME;
   if (bSuccess) {
      bSuccess = channel_readAll(channel, buffer, timeout);
   } else {
      channel->isActive = CFL_FALSE;
      rgt_error_set(c->connectionType, RGT_ERROR_SOCKET, "Error sending data to server: %ld", cfl_socket_lastErrorCode());
   }
   RGT_LOCK_RELEASE(channel->ioLock);
   RGT_LOG_EXIT("rgt_channel_unidirectional.channel_writeAndReadFirstCommand", (NULL));
   return bSuccess;
}

static RGT_SINGLE_CHANNELP channel_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   CFL_SOCKET socket;
   RGT_SINGLE_CHANNELP channel;

   socket = cfl_socket_open(server, port);
   if (socket == CFL_INVALID_SOCKET) {
      rgt_error_set(connectionType, RGT_ERROR_SOCKET, "Error connecting to application server. server=%s port= %u error code=%ld",
                    server, port, cfl_socket_lastErrorCode());
      return NULL;
   }
   cfl_socket_setNoDelay(socket, CFL_TRUE);

   channel = RGT_HB_ALLOC(sizeof(RGT_SINGLE_CHANNEL));
   if (channel == NULL) {
      cfl_socket_close(socket);
      RESOURCE_ALLOCATION_ERROR(connectionType);
      return NULL;
   }
   channel->channel.connectionType = connectionType;
   channel->socket = socket;
   channel->isActive = CFL_TRUE;
   RGT_LOCK_INIT(channel->ioLock);
   return channel;
}

RGT_CHANNELP rgt_channel_unidirectional_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   RGT_SINGLE_CHANNELP channel;
   RGT_LOG_ENTER("rgt_channel_unidirectional_open", (NULL));
   channel = channel_open(connectionType, server, port);
   if (channel != NULL) {
      channel->channel.close = channel_close;
      channel->channel.free = channel_free;
      channel->channel.isOpen = channel_isOpen;
      channel->channel.waitData = channel_waitData;
      channel->channel.hasData = channel_hasData;
      channel->channel.read = channel_read;
      channel->channel.readBuffer = channel_readBuffer;
      channel->channel.tryRead = channel_tryRead;
      channel->channel.tryReadBuffer = channel_tryReadBuffer;
      channel->channel.write = channel_write;
      channel->channel.writeAndRead = channel_writeAndRead;
      channel->channel.writeAndReadFirstCommand = channel_writeAndReadFirstCommand;
      channel->channel.lastRead = CURRENT_TIME;
      channel->channel.lastWrite = CURRENT_TIME;
   }
   RGT_LOG_EXIT("rgt_channel_unidirectional_open", (NULL));
   return (RGT_CHANNELP)channel;
}
