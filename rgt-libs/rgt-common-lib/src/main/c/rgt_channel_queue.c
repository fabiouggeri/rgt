#include "cfl_buffer.h"
#include "cfl_socket.h"
#include "cfl_str.h"
#include "cfl_sync_queue.h"

#include "rgt_channel.h"
#include "rgt_channel_queue.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_thread.h"
#include "rgt_types.h"
#include "rgt_util.h"

#define DEFAULT_READ_QUEUE_SIZE 64
#define DEFAULT_WRITE_QUEUE_SIZE 64

#define RESOURCE_ALLOCATION_ERROR(ct) rgt_error_set(ct, RGT_ERROR_ALLOC_RESOURCE, "Error allocating resource")

#define WAIT_QUEUE_EMPTY (5 * SECOND)
#define WAIT_THREAD_TIMEOUT (2 * SECOND)

#define QUEUE_CHANNEL(c) ((RGT_QUEUE_CHANNELP)c)

typedef struct _RGT_QUEUE_CHANNEL {
      RGT_CHANNEL channel;
      CFL_SOCKET socket;
      RGT_THREADP readThread;
      CFL_SYNC_QUEUEP readQueue;
      RGT_THREADP writeThread;
      CFL_SYNC_QUEUEP writeQueue;
      CFL_BOOL isActive;
} RGT_QUEUE_CHANNEL, *RGT_QUEUE_CHANNELP;

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

static void channel_closeSocket(RGT_QUEUE_CHANNELP channel) {
   CFL_SOCKET socket = channel->socket;
   RGT_LOG_ENTER("rgt_channel_queue.channel_closeSocket", (NULL));
   channel->socket = CFL_INVALID_SOCKET;
   if (socket != CFL_INVALID_SOCKET) {
      cfl_socket_shutdown(socket, CFL_TRUE, CFL_TRUE);
      cfl_socket_close(socket);
   } else {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_closeSocket(). socket already closed."));
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_closeSocket", (NULL));
}

static CFL_BOOL channel_isOpen(RGT_CHANNELP c) {
   return c != NULL && QUEUE_CHANNEL(c)->isActive && QUEUE_CHANNEL(c)->socket != CFL_INVALID_SOCKET;
}

static CFL_BUFFERP channel_readAll(RGT_QUEUE_CHANNELP channel) {
   int retVal;
   RGT_PACKET_LEN_TYPE packetLen;
   CFL_BUFFERP buffer;

   RGT_LOG_ENTER("rgt_channel_queue.channel_readAll", (NULL));
   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_readAll: connection is closed"));
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", (NULL));
      return NULL;
   }

   // read header
   retVal = cfl_socket_receiveAll(channel->socket, (char *)&packetLen, RGT_PACKET_LEN_FIELD_SIZE);
   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_readAll: error reading header. channel closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", ("error reading header. channel closed"));
      return NULL;
   } else if (retVal == CFL_SOCKET_ERROR) {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                    cfl_socket_lastErrorCode());
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", ("error reading header. socket error: %ld", cfl_socket_lastErrorCode()));
      return NULL;
   } else if (retVal == 0) {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_readAll(): error reading header. socket closed"));
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", ("error reading header. socket closed"));
      return NULL;
   }

   // read body
   RGT_LOG_DEBUG(("rgt_channel_queue.channel_readAll: Data read. Packet Len(%d)=%#0*X", RGT_PACKET_LEN_FIELD_SIZE,
                  2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen));
   if (packetLen <= 0) {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_PROTOCOL, "Zero length packet found. Header: %#0*X.",
                    2 + RGT_PACKET_LEN_FIELD_SIZE * 2, packetLen);
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", ("zero length packet"));
      return NULL;
   }

   buffer = cfl_buffer_newCapacity(packetLen);
   if (buffer == NULL) {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_ALLOC_RESOURCE, "Error allocating resource");
      RGT_LOG_EXIT("rgt_channel_queue33.channel_readAll", ("error allocating resource"));
      return NULL;
   }
   retVal = cfl_socket_receiveAll(channel->socket, (char *)cfl_buffer_getDataPtr(buffer), packetLen);
   if (!channel_isOpen((RGT_CHANNELP)channel)) {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_readAll: error reading body. channel closed"));
      cfl_buffer_free(buffer);
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", ("error reading body. channel closed"));
      return NULL;
   } else if (retVal == CFL_SOCKET_ERROR) {
      rgt_error_set(channel->channel.connectionType, RGT_ERROR_SOCKET, "Error reading data from socket: %ld",
                    cfl_socket_lastErrorCode());
      cfl_buffer_free(buffer);
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", ("error reading body. socket error: %ld", cfl_socket_lastErrorCode()));
      return NULL;
   } else if (retVal == 0) {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_readAll: error reading body. socket closed"));
      cfl_buffer_free(buffer);
      channel->isActive = CFL_FALSE;
      RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", ("error reading body. socket closed"));
      return NULL;
   }
   cfl_buffer_setPosition(buffer, packetLen);
   cfl_buffer_flip(buffer);
   ((RGT_CHANNELP)channel)->lastRead = CURRENT_TIME;
   RGT_LOG_EXIT("rgt_channel_queue.channel_readAll", (NULL));
   return buffer;
}

static void channel_readData(void *param) {
   RGT_QUEUE_CHANNELP channel = (RGT_QUEUE_CHANNELP)param;
   CFL_BOOL timesUp;

   RGT_LOG_ENTER("rgt_channel_queue.channel_readData", (NULL));
   while (channel_isOpen((RGT_CHANNELP)channel)) {
      if (cfl_socket_selectRead(channel->socket, 5 * SECOND) > 0) {
         CFL_BUFFERP buffer = channel_readAll(channel);
         if (buffer != NULL) {
            cfl_sync_queue_put(channel->readQueue, buffer);
         } else {
            RGT_LOG_DEBUG(("rgt_channel_queue.channel_readData(): %s",
                           (channel_isOpen((RGT_CHANNELP)channel) ? "null buffer" : "socket closed")));
            break;
         }
      }
   }
   cfl_sync_queue_waitEmptyTimeout(channel->readQueue, WAIT_QUEUE_EMPTY, &timesUp);
   cfl_sync_queue_cancel(channel->readQueue);
   RGT_LOG_EXIT("rgt_channel_queue.channel_readData", (NULL));
}

static void channel_writeData(void *param) {
   RGT_QUEUE_CHANNELP channel = (RGT_QUEUE_CHANNELP)param;
   CFL_BOOL running;
   CFL_BOOL timesUp;

   RGT_LOG_ENTER("rgt_channel_queue.channel_writeData", (NULL));

   running = channel_isOpen((RGT_CHANNELP)channel);
   while (running) {
      CFL_BUFFERP buffer = cfl_sync_queue_getTimeout(channel->writeQueue, 3 * SECOND, &timesUp);
      if (buffer != NULL) {
         if (cfl_socket_sendAllBuffer(channel->socket, buffer)) {
            ((RGT_CHANNELP)channel)->lastWrite = CURRENT_TIME;
            RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("rgt_channel_queue.channel_writeData().", buffer, RGT_PACKET_LEN_FIELD_SIZE));
            running = channel_isOpen((RGT_CHANNELP)channel) && !cfl_sync_queue_canceled(channel->writeQueue);
         } else {
            RGT_LOG_DEBUG(("rgt_channel_queue.channel_writeData(): error writing data in socket"));
            running = CFL_FALSE;
         }
         cfl_buffer_free(buffer);
      } else if (!timesUp) {
         RGT_LOG_DEBUG(("rgt_channel_queue.channel_writeData(): %s",
                        (cfl_sync_queue_canceled(channel->writeQueue) ? "queue canceled" : "error waiting data from queue")));
         running = CFL_FALSE;
      } else {
         running = channel_isOpen((RGT_CHANNELP)channel) && !cfl_sync_queue_canceled(channel->writeQueue);
      }
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_writeData", (NULL));
}

static void channel_close(RGT_CHANNELP c) {
   RGT_QUEUE_CHANNELP channel;
   CFL_BOOL timesUp;
   RGT_LOG_ENTER("rgt_channel_queue.channel_close", (NULL));
   if (c == NULL) {
      RGT_LOG_ERROR(("rgt_channel_queue.channel_close(). channel is NULL."));
      RGT_LOG_EXIT("rgt_channel_queue.channel_close", (NULL));
      return;
   }
   channel = QUEUE_CHANNEL(c);
   channel->isActive = CFL_FALSE;
   cfl_sync_queue_waitEmptyTimeout(channel->writeQueue, WAIT_QUEUE_EMPTY, &timesUp);
   cfl_sync_queue_cancel(channel->writeQueue);
   channel_closeSocket(channel);
   rgt_thread_waitTimeout(channel->writeThread, WAIT_THREAD_TIMEOUT);
   if (rgt_thread_isRunning(channel->writeThread)) {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_close. Write thread killed after timeout."));
      rgt_thread_kill(channel->writeThread);
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_close", (NULL));
}

static void channel_free(RGT_CHANNELP c) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   RGT_LOG_ENTER("rgt_channel_queue.channel_free", (NULL));
   if (channel != NULL) {
      rgt_thread_free(channel->readThread);
      rgt_thread_free(channel->writeThread);
      cfl_sync_queue_free(channel->readQueue);
      cfl_sync_queue_free(channel->writeQueue);
      RGT_HB_FREE(channel);
   } else {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_free. Channel already free."));
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_free", (NULL));
}

static RGT_QUEUE_CHANNELP channel_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   RGT_QUEUE_CHANNELP channel;

   RGT_LOG_ENTER("rgt_channel_queue.channel_open", (NULL));
   channel = RGT_HB_ALLOC(sizeof(RGT_QUEUE_CHANNEL));
   if (channel == NULL) {
      RESOURCE_ALLOCATION_ERROR(connectionType);
      RGT_LOG_EXIT("rgt_channel_queue.channel_open", (NULL));
      return NULL;
   }

   channel->readQueue = cfl_sync_queue_new(DEFAULT_READ_QUEUE_SIZE);
   if (channel->readQueue == NULL) {
      channel_free((RGT_CHANNELP)channel);
      RESOURCE_ALLOCATION_ERROR(connectionType);
      RGT_LOG_EXIT("rgt_channel_queue.channel_open", (NULL));
      return NULL;
   }

   channel->isActive = CFL_TRUE;
   channel->writeQueue = cfl_sync_queue_new(DEFAULT_WRITE_QUEUE_SIZE);
   if (channel->writeQueue == NULL) {
      channel_free((RGT_CHANNELP)channel);
      RESOURCE_ALLOCATION_ERROR(connectionType);
      RGT_LOG_EXIT("rgt_channel_queue.channel_open", (NULL));
      return NULL;
   }

   channel->socket = cfl_socket_open(server, port);
   if (channel->socket == CFL_INVALID_SOCKET) {
      rgt_error_set(connectionType, RGT_ERROR_SOCKET, "Error connecting to application server. server=%s port= %u error code=%ld",
                    server, port, cfl_socket_lastErrorCode());
      channel_free((RGT_CHANNELP)channel);
      RGT_LOG_EXIT("rgt_channel_queue.channel_open", (NULL));
      return NULL;
   }
   channel->channel.connectionType = connectionType;
   cfl_socket_setNoDelay(channel->socket, CFL_TRUE);

   channel->writeThread = rgt_thread_start(channel_writeData, channel, "RGT Write Queue");
   if (channel->writeThread == NULL) {
      channel_free((RGT_CHANNELP)channel);
      RESOURCE_ALLOCATION_ERROR(connectionType);
      RGT_LOG_EXIT("rgt_channel_queue.channel_open", (NULL));
      return NULL;
   }

   channel->readThread = rgt_thread_start(channel_readData, channel, "RGT Read Queue");
   if (channel->readThread == NULL) {
      channel_free((RGT_CHANNELP)channel);
      RESOURCE_ALLOCATION_ERROR(connectionType);
      RGT_LOG_EXIT("rgt_channel_queue.channel_open", (NULL));
      return NULL;
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_open", (NULL));
   return channel;
}

static CFL_BOOL channel_waitData(RGT_CHANNELP c, CFL_UINT32 timeout, CFL_BOOL *error) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   CFL_UINT32 itemsCount;
   CFL_BOOL timesUp;
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_queue.channel_waitData", (NULL));
   if (!channel_isOpen(c)) {
      RGT_LOG_DEBUG(("rgt_channel_queue.waitData(). connection is closed"));
      *error = CFL_TRUE;
      RGT_LOG_EXIT("rgt_channel_queue.channel_waitData", (NULL));
      return CFL_FALSE;
   }

   itemsCount = cfl_sync_queue_waitNotEmptyTimeout(channel->readQueue, timeout, &timesUp);
   bSuccess = itemsCount > 0 ? CFL_TRUE : CFL_FALSE;
   if (bSuccess) {
      *error = CFL_FALSE;
   } else {
      *error = timesUp ? CFL_FALSE : CFL_TRUE;
   }
   RGT_LOG_DEBUG(("rgt_channel_queue.channel_waitData(). success=%s", bSuccess ? "true" : "false"));
   RGT_LOG_EXIT("rgt_channel_queue.channel_waitData", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_hasData(RGT_CHANNELP c) {
   CFL_BOOL bSuccess;
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   RGT_LOG_ENTER("rgt_channel_queue.channel_hasData", (NULL));
   bSuccess = !cfl_sync_queue_isEmpty(channel->readQueue);
   RGT_LOG_DEBUG(("rgt_channel_queue.channel_hasData(). success=%s", bSuccess ? "true" : "false"));
   RGT_LOG_EXIT("rgt_channel_queue.channel_hasData", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_read(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   CFL_BOOL bSuccess;
   CFL_BOOL timesUp;
   CFL_BUFFERP bufferRead;

   RGT_LOG_ENTER("rgt_channel_queue.channel_read", ("timeout=%u", timeout));
   bufferRead = cfl_sync_queue_getTimeout(channel->readQueue, timeout, &timesUp);
   if (bufferRead != NULL) {
      cfl_buffer_reset(buffer);
      cfl_buffer_putBuffer(buffer, bufferRead);
      cfl_buffer_flip(buffer);
      cfl_buffer_free(bufferRead);
      RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
      bSuccess = CFL_TRUE;
   } else {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_read() no data read."));
      bSuccess = CFL_FALSE;
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_read", (NULL));
   return bSuccess;
}

static CFL_BUFFERP channel_readBuffer(RGT_CHANNELP c, CFL_UINT32 timeout) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   CFL_BOOL timesUp;
   CFL_BUFFERP buffer;

   RGT_LOG_ENTER("rgt_channel_queue.channel_read", ("timeout=%u", timeout));
   buffer = cfl_sync_queue_getTimeout(channel->readQueue, timeout, &timesUp);
   if (buffer != NULL) {
      RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
   } else {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_read() no data read."));
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_read", (NULL));
   return buffer;
}

static CFL_BOOL channel_tryRead(RGT_CHANNELP c, CFL_BUFFERP buffer) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   CFL_BOOL bSuccess;
   CFL_BUFFERP bufferRead;

   RGT_LOG_ENTER("rgt_channel_queue.channel_tryRead", (NULL));
   bufferRead = cfl_sync_queue_tryGet(channel->readQueue, &bSuccess);
   if (bSuccess) {
      cfl_buffer_reset(buffer);
      cfl_buffer_putBuffer(buffer, bufferRead);
      cfl_buffer_flip(buffer);
      cfl_buffer_free(bufferRead);
      RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
   } else {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_tryRead() no data found."));
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_tryRead", (NULL));
   return bSuccess;
}

static CFL_BUFFERP channel_tryReadBuffer(RGT_CHANNELP c) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   CFL_BUFFERP buffer;
   CFL_BOOL bSuccess;

   RGT_LOG_ENTER("rgt_channel_queue.channel_tryRead", (NULL));
   buffer = cfl_sync_queue_tryGet(channel->readQueue, &bSuccess);
   if (buffer != NULL) {
      RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data read.", buffer, 0));
   } else {
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_tryRead() no data found."));
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_tryRead", (NULL));
   return buffer;
}

static CFL_BOOL channel_write(RGT_CHANNELP c, CFL_BUFFERP buffer) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   CFL_BOOL bSuccess;
   CFL_BUFFERP writeBuffer;

   RGT_LOG_ENTER("rgt_channel_queue.channel_write", (NULL));
   if (!channel->isActive) {
      RGT_LOG_EXIT("rgt_channel_queue.channel_write", (NULL));
      return CFL_FALSE;
   }

   cfl_buffer_flip(buffer);
   RGT_PUT_PACKET_LEN(buffer, cfl_buffer_length(buffer) - RGT_PACKET_LEN_FIELD_SIZE);
   cfl_buffer_rewind(buffer);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data to write.", buffer, RGT_PACKET_LEN_FIELD_SIZE));

   writeBuffer = cfl_buffer_clone(buffer);
   if (cfl_sync_queue_put(channel->writeQueue, writeBuffer)) {
      bSuccess = CFL_TRUE;
      RGT_LOG_DEBUG(("rgt_channel_queue.channel_write() data sent to queue."));
   } else {
      bSuccess = CFL_FALSE;
      rgt_error_set(c->connectionType, RGT_ERROR_SOCKET, "Error sending data: %ld", cfl_socket_lastErrorCode());
   }
   cfl_buffer_reset(buffer);
   RGT_LOG_EXIT("rgt_channel_queue.channel_write", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_writeAndRead(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   CFL_BOOL bSuccess;
   RGT_LOG_ENTER("rgt_channel_queue.channel_writeAndRead", (NULL));
   bSuccess = channel_write(c, buffer) && channel_read(c, buffer, timeout);
   RGT_LOG_EXIT("rgt_channel_queue.channel_writeAndRead", (NULL));
   return bSuccess;
}

static CFL_BOOL channel_writeAndReadFirstCommand(RGT_CHANNELP c, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   RGT_QUEUE_CHANNELP channel = QUEUE_CHANNEL(c);
   CFL_BOOL bSuccess;
   CFL_BUFFERP writeBuffer;

   RGT_LOG_ENTER("rgt_channel_queue.channel_writeAndReadFirstCommand", (NULL));
   if (!channel->isActive) {
      RGT_LOG_EXIT("rgt_channel_queue.channel_writeAndReadFirstCommand", (NULL));
      return CFL_FALSE;
   }
   cfl_buffer_flip(buffer);
   cfl_buffer_skip(buffer, RGT_MAGIC_NUMBER_SIZE);
   RGT_PUT_PACKET_LEN(buffer, cfl_buffer_length(buffer) - RGT_PACKET_LEN_FIELD_SIZE - RGT_MAGIC_NUMBER_SIZE);
   cfl_buffer_rewind(buffer);
   RGT_LOG(RGT_LOG_LEVEL_DEBUG, bufferToHex("Data write.", buffer, 8));

   writeBuffer = cfl_buffer_clone(buffer);
   if (cfl_sync_queue_put(channel->writeQueue, writeBuffer)) {
      bSuccess = channel_read(c, buffer, timeout);
   } else {
      bSuccess = CFL_FALSE;
      rgt_error_set(c->connectionType, RGT_ERROR_SOCKET, "Error sending data to server: %ld", cfl_socket_lastErrorCode());
   }
   RGT_LOG_EXIT("rgt_channel_queue.channel_writeAndReadFirstCommand", (NULL));
   return bSuccess;
}

RGT_CHANNELP rgt_channel_queue_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   RGT_QUEUE_CHANNELP channel;

   RGT_LOG_ENTER("rgt_channel_queue_open", (NULL));
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
   RGT_LOG_EXIT("rgt_channel_queue_open", (NULL));
   return (RGT_CHANNELP)channel;
}
