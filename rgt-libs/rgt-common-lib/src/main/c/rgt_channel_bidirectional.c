#include <stdio.h>
#include <stdlib.h>

#include "cfl_buffer.h"
#include "cfl_lock.h"
#include "cfl_socket.h"
#include "cfl_str.h"

#include "rgt_channel.h"
#include "rgt_channel_bidirectional.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_types.h"
#include "rgt_util.h"

#define BI_CHANNEL(c) ((RGT_BI_CHANNELP)c)

#define RESOURCE_ALLOCATION_ERROR(ct) rgt_error_set(ct, RGT_ERROR_ALLOC_RESOURCE, "Error allocating resource")

#define GET_UINT32(b) ((CFL_UINT32)((b[0] & 0xFF) | ((b[1] & 0xFF) << 8) | ((b[2] & 0xFF) << 16) | ((b[3] & 0xFF) << 24)))
#define GET_BYTE(b, i) ((CFL_INT8)(b[i] & 0xFF))

typedef struct _RGT_BI_CHANNEL {
      RGT_CHANNEL channel;
      CFL_SOCKET socket;
      RGT_LOCK readLock;
      RGT_LOCK writeLock;
      CFL_BOOL isActive;
} RGT_BI_CHANNEL, *RGT_BI_CHANNELP;

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

static void channel_closeSocket(RGT_BI_CHANNELP channel) {
   CFL_SOCKET socket;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_closeSocket", (NULL));
   socket = channel->socket;
   channel->socket = CFL_INVALID_SOCKET;
   if (socket != CFL_INVALID_SOCKET) {
      cfl_socket_shutdown(socket, CFL_TRUE, CFL_TRUE);
      cfl_socket_close(socket);
   } else {
      RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_closeSocket(). socket already closed."));
   }
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_closeSocket", (NULL));
}

static CFL_BOOL channel_isOpen(RGT_CHANNELP c) {
   return c != NULL && BI_CHANNEL(c)->isActive && BI_CHANNEL(c)->socket != CFL_INVALID_SOCKET;
}

static void channel_close(RGT_CHANNELP c) {
   RGT_BI_CHANNELP channel;
   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_close", (NULL));
   if (c == NULL) {
      RGT_LOG_ERROR(("rgt_channel_bidirectional.channel_close(). channel is NULL."));
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_close", (NULL));
      return;
   }
   channel = BI_CHANNEL(c);
   channel->isActive = CFL_FALSE;
   channel_closeSocket(channel);
   RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_close(). channel closed."));
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_close", (NULL));
}

static void channel_free(RGT_CHANNELP c) {
   RGT_BI_CHANNELP channel = BI_CHANNEL(c);
   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_free", (NULL));
   if (channel != NULL) {
      RGT_LOCK_FREE(channel->readLock);
      RGT_LOCK_FREE(channel->writeLock);
      RGT_HB_FREE(channel);
      RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_free(). channel released."));
   } else {
      RGT_LOG_ERROR(("rgt_channel_bidirectional.channel_free(). channel already released."));
   }
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_free", (NULL));
}

static CFL_BOOL channel_waitData(RGT_CHANNELP c, CFL_UINT32 timeout, CFL_BOOL *error) {
   RGT_BI_CHANNELP channel = BI_CHANNEL(c);
   int retVal;
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_waitData", (NULL));
   if (channel_isOpen(c)) {
      RGT_LOCK_ACQUIRE(channel->readLock);
      retVal = cfl_socket_selectRead(channel->socket, timeout > 0 ? timeout : CFL_WAIT_FOREVER);
      RGT_LOCK_RELEASE(channel->readLock);
      *error = retVal < 0 ? CFL_TRUE : CFL_FALSE;
      bSuccess = retVal > 0 ? CFL_TRUE : CFL_FALSE;
   } else {
      RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_waitData(). connection is closed"));
      *error = CFL_TRUE;
      bSuccess = CFL_FALSE;
   }

   RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_waitData=%s", bSuccess ? "true" : "false"));
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_waitData", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_hasData(RGT_CHANNELP c) {
   RGT_BI_CHANNELP channel = BI_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_hasData", (NULL));
   RGT_LOCK_ACQUIRE(channel->readLock);
   bSuccess = cfl_socket_selectRead(channel->socket, 0);
   RGT_LOCK_RELEASE(channel->readLock);
   RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_hasData=%s", bSuccess ? "true" : "false"));
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_hasData", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_readAll(RGT_BI_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   int retVal;
   RGT_PACKET_LEN_TYPE packetLen;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_readAll", (NULL));
   cfl_buffer_reset(buffer);

   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll: channel is closed"));
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("channel is closed"));
      return CFL_FALSE;
   }

   if (timeout > 0) {
      retVal = cfl_socket_selectRead(channel->socket, timeout);
      if (!channel_isOpen((RGT_CHANNELP)channel)) {
         RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll: error waiting header. channel closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("error waiting header. channel closed"));
         return CFL_FALSE;
      } else if (retVal == CFL_SOCKET_ERROR) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error waiting data from socket: %ld",
                       cfl_socket_lastErrorCode());
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll",
                      ("error waiting header. socket error:%ld", cfl_socket_lastErrorCode()));
         return CFL_FALSE;
      } else if (retVal == 0) {
         RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll(): timeout waiting header"));
         RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("timeout waiting header"));
         return CFL_FALSE;
      }
   }

   // data must be here, so do a normal cfl_socket_receive()
   retVal = cfl_socket_receiveAll(channel->socket, (char *)&packetLen, RGT_PACKET_LEN_FIELD_SIZE);
   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll: error reading header. channel is closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("channel is closed"));
      return CFL_FALSE;
   } else if (retVal == CFL_SOCKET_ERROR) {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                    cfl_socket_lastErrorCode());
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll",
                   ("error reading header. socket error:%ld", cfl_socket_lastErrorCode()));
      return CFL_FALSE;
   } else if (retVal == 0) {
      RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll: error reading header. socket closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("error reading header. socket closed"));
      return CFL_FALSE;
   }

   RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll: Data read. Packet Len(%d)=%#0*X", RGT_PACKET_LEN_FIELD_SIZE,
                  2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen));
   if (packetLen > 0) {
      if ((CFL_UINT32)packetLen > cfl_buffer_capacity(buffer) && !cfl_buffer_setCapacity(buffer, packetLen)) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_ALLOC_RESOURCE, "Error allocating resource");
         RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("error allocating resource"));
         return CFL_FALSE;
      }
      retVal = cfl_socket_receiveAll(channel->socket, (char *)cfl_buffer_getDataPtr(buffer), packetLen);
      if (!channel_isOpen((RGT_CHANNELP)channel)) {
         RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll: error reading body. channel is closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("error reading body. channel is closed"));
         return CFL_FALSE;
      } else if (retVal == CFL_SOCKET_ERROR) {
         rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                       cfl_socket_lastErrorCode());
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll",
                      ("error reading body. socket error: %ld", cfl_socket_lastErrorCode()));
         return CFL_FALSE;
      } else if (retVal == 0) {
         RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_readAll: error reading body. socket closed"));
         channel->isActive = CFL_FALSE;
         RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("socket closed"));
         return CFL_FALSE;
      }
      cfl_buffer_setPosition(buffer, packetLen);
      cfl_buffer_flip(buffer);
   } else {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_PROTOCOL, "Zero length packet found. Header: %#0*X.",
                    2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen);
      cfl_buffer_reset(buffer);
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", ("zero length packet"));
      return CFL_FALSE;
   }
   ((RGT_CHANNELP)channel)->lastRead = CURRENT_TIME;
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_readAll", (NULL));
   return CFL_TRUE;
}

static CFL_BOOL channel_read(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   RGT_BI_CHANNELP channel = BI_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_read", ("timeout=%u", timeout));
   RGT_LOCK_ACQUIRE(channel->readLock);
   bSuccess = channel_readAll(channel, buffer, timeout);
   RGT_LOCK_RELEASE(channel->readLock);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("rgt_channel_bidirectional.channel_read(). Data read.", buffer, 0));
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_read", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_tryRead(RGT_CHANNELP c, CFL_BUFFERP buffer) {
   RGT_BI_CHANNELP channel = BI_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_tryRead", (NULL));
   RGT_LOCK_ACQUIRE(channel->readLock);
   if (cfl_socket_selectRead(channel->socket, 0)) {
      bSuccess = channel_readAll(channel, buffer, 0);
      RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("rgt_channel_bidirectional.channel_tryRead(). Data read.", buffer, 0));
   } else {
      RGT_LOG_DEBUG(("rgt_channel_bidirectional.channel_tryRead() no data found"));
      bSuccess = CFL_FALSE;
   }
   RGT_LOCK_RELEASE(channel->readLock);
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_tryRead", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_write(RGT_CHANNELP c, CFL_BUFFERP buffer) {
   RGT_BI_CHANNELP channel = BI_CHANNEL(c);
   CFL_BOOL bSuccess = CFL_TRUE;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_write", (NULL));
   cfl_buffer_flip(buffer);
   RGT_PUT_PACKET_LEN(buffer, cfl_buffer_length(buffer) - RGT_PACKET_LEN_FIELD_SIZE);
   cfl_buffer_rewind(buffer);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG,
           bufferToHex("rgt_channel_bidirectional.channel_write(). Data write.", buffer, RGT_PACKET_LEN_FIELD_SIZE));
   RGT_LOCK_ACQUIRE(channel->writeLock);
   if (cfl_socket_sendAllBuffer(channel->socket, buffer)) {
      c->lastWrite = CURRENT_TIME;
   } else {
      rgt_error_set(c->connectionType, RGT_ERROR_SOCKET, "Error sending data: %ld", cfl_socket_lastErrorCode());
      bSuccess = CFL_FALSE;
      channel->isActive = CFL_FALSE;
   }
   RGT_LOCK_RELEASE(channel->writeLock);
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_write", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_writeAndRead(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   CFL_BOOL bSuccess;
   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_writeAndRead", (NULL));
   bSuccess = channel_write(c, buffer) && channel_read(c, buffer, timeout);
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_writeAndRead", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_writeAndReadFirstCommand(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   RGT_BI_CHANNELP channel = BI_CHANNEL(c);
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_writeAndReadFirstCommand", (NULL));
   cfl_buffer_flip(buffer);
   cfl_buffer_skip(buffer, RGT_MAGIC_NUMBER_SIZE);
   RGT_PUT_PACKET_LEN(buffer, cfl_buffer_length(buffer) - RGT_PACKET_LEN_FIELD_SIZE - RGT_MAGIC_NUMBER_SIZE);
   cfl_buffer_rewind(buffer);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("rgt_channel_bidirectional.channel_writeAndReadFirstCommand.", buffer, 8));
   RGT_LOCK_ACQUIRE(channel->writeLock);
   bSuccess = cfl_socket_sendAllBuffer(channel->socket, buffer);
   c->lastWrite = CURRENT_TIME;
   RGT_LOCK_RELEASE(channel->writeLock);
   if (bSuccess) {
      bSuccess = channel_read(c, buffer, timeout);
   } else {
      rgt_error_set(c->connectionType, RGT_ERROR_SOCKET, "Error sending data to server: %ld", cfl_socket_lastErrorCode());
      channel->isActive = CFL_FALSE;
   }
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_writeAndReadFirstCommand", (NULL));
   return bSuccess;
}

static RGT_BI_CHANNELP channel_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   CFL_SOCKET socket;
   RGT_BI_CHANNELP channel;

   RGT_LOG_ENTER("rgt_channel_bidirectional.channel_open", (NULL));
   socket = cfl_socket_open(server, port);
   if (socket == CFL_INVALID_SOCKET) {
      rgt_error_set(connectionType, RGT_ERROR_SOCKET, "Error connecting to application server. server=%s port= %u error code=%ld",
                    server, port, cfl_socket_lastErrorCode());
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_open", (NULL));
      return NULL;
   }

   channel = RGT_HB_ALLOC(sizeof(RGT_BI_CHANNEL));
   if (channel == NULL) {
      cfl_socket_close(socket);
      RESOURCE_ALLOCATION_ERROR(connectionType);
      RGT_LOG_EXIT("rgt_channel_bidirectional.channel_open", (NULL));
      return NULL;
   }
   cfl_socket_setNoDelay(socket, CFL_TRUE);
   channel->channel.connectionType = connectionType;
   channel->socket = socket;
   channel->isActive = CFL_TRUE;
   RGT_LOCK_INIT(channel->readLock);
   RGT_LOCK_INIT(channel->writeLock);
   RGT_LOG_EXIT("rgt_channel_bidirectional.channel_open", (NULL));
   return channel;
}

RGT_CHANNELP rgt_channel_bidirectional_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   RGT_BI_CHANNELP channel;

   RGT_LOG_ENTER("rgt_channel_bidirectional_open", (NULL));
   channel = channel_open(connectionType, server, port);
   if (channel != NULL) {
      channel->channel.close = channel_close;
      channel->channel.free = channel_free;
      channel->channel.isOpen = channel_isOpen;
      channel->channel.waitData = channel_waitData;
      channel->channel.hasData = channel_hasData;
      channel->channel.read = channel_read;
      channel->channel.tryRead = channel_tryRead;
      channel->channel.write = channel_write;
      channel->channel.writeAndRead = channel_writeAndRead;
      channel->channel.writeAndReadFirstCommand = channel_writeAndReadFirstCommand;
      channel->channel.lastRead = CURRENT_TIME;
      channel->channel.lastWrite = CURRENT_TIME;
   }
   RGT_LOG_EXIT("rgt_channel_bidirectional_open", (NULL));
   return (RGT_CHANNELP)channel;
}
