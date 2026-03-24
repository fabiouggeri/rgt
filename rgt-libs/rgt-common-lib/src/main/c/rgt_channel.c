#include "hbapiitm.h"

#include "rgt_channel.h"
#include "rgt_channel_unidirectional.h"
#include "rgt_channel_bidirectional.h"
#include "rgt_channel_queue.h"
#include "rgt_log.h"

static CFL_UINT8 s_channelType = RGT_CHANNEL_BIDIRECTIONAL;

CFL_UINT8 rgt_channel_type(void) {
   return s_channelType;
}

void rgt_channel_setType(CFL_UINT8 newChannelType) {
   RGT_LOG_ENTER("rgt_channel_setType", ("type=%d", newChannelType));
   if (newChannelType <= RGT_CHANNEL_MAX) {
      s_channelType = newChannelType;
   }
   RGT_LOG_EXIT("rgt_channel_setType", (NULL));
}

void rgt_channel_setTypeByName(const char *channelTypeName) {
   RGT_LOG_ENTER("rgt_channel_setTypeByName", ("type=%s", channelTypeName));
   if (hb_stricmp(channelTypeName, RGT_CHANNEL_TYPE_QUEUE) == 0) {
      s_channelType = RGT_CHANNEL_QUEUE;
   } else if (hb_stricmp(channelTypeName, RGT_CHANNEL_TYPE_UNIDIRECTIONAL) == 0) {
      s_channelType = RGT_CHANNEL_UNIDIRECTIONAL;
   } else {
      s_channelType = RGT_CHANNEL_BIDIRECTIONAL;
   }
   RGT_LOG_EXIT("rgt_channel_setTypeByName", (NULL));
}

RGT_CHANNELP rgt_channel_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   RGT_CHANNELP channel;
   RGT_LOG_ENTER("rgt_channel_open", (NULL));
   switch (s_channelType) {
      case RGT_CHANNEL_QUEUE:
         channel = rgt_channel_queue_open(connectionType, server, port);
         break;
      case RGT_CHANNEL_BIDIRECTIONAL:
         channel = rgt_channel_bidirectional_open(connectionType, server, port);
         break;
      default:
         channel = rgt_channel_unidirectional_open(connectionType, server, port);
         break;
   }
   RGT_LOG_INFO(("rgt_channel_open() type=%d result=%s", (int)s_channelType, (channel != NULL ? "success" : "failed")));
   RGT_LOG_EXIT("rgt_channel_open", (NULL));
   return channel;
}

void rgt_channel_close(RGT_CHANNELP channel) {
   RGT_LOG_ENTER("rgt_channel_close", (NULL));
   RGT_LOG_INFO(("rgt_channel_close()"));
   if (channel != NULL) {
      channel->close(channel);
   } else {
      RGT_LOG_DEBUG(("rgt_channel_close(). channel already closed."));
   }
   RGT_LOG_EXIT("rgt_channel_close", (NULL));
}

void rgt_channel_free(RGT_CHANNELP channel) {
   RGT_LOG_ENTER("rgt_channel_free", (NULL));
   if (channel != NULL) {
      channel->free(channel);
   } else {
      RGT_LOG_ERROR(("rgt_channel_free(). channel already released."));
   }
   RGT_LOG_EXIT("rgt_channel_free", (NULL));
}

CFL_BOOL rgt_channel_isOpen(RGT_CHANNELP channel) {
   return channel != NULL && channel->isOpen(channel);
}

CFL_BOOL rgt_channel_waitData(RGT_CHANNELP channel, CFL_UINT32 timeout, CFL_BOOL *bError) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_waitData", (NULL));
   success = rgt_channel_isOpen(channel) && channel->waitData(channel, timeout, bError);
   RGT_LOG_EXIT("rgt_channel_waitData", (NULL));
   return success;
}

CFL_BOOL rgt_channel_hasData(RGT_CHANNELP channel) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_hasData", (NULL));
   success = rgt_channel_isOpen(channel) && channel->hasData(channel);
   RGT_LOG_EXIT("rgt_channel_hasData", (NULL));
   return success;
}

CFL_BOOL rgt_channel_tryRead(RGT_CHANNELP channel, CFL_BUFFERP buffer) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_tryRead", (NULL));
   success = rgt_channel_isOpen(channel) && channel->tryRead(channel, buffer);
   RGT_LOG_EXIT("rgt_channel_tryRead", (NULL));
   return success;
}

CFL_BOOL rgt_channel_read(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_read", (NULL));
   success = rgt_channel_isOpen(channel) && channel->read(channel, buffer, timeout);
   RGT_LOG_EXIT("rgt_channel_read", (NULL));
   return success;
}

CFL_BOOL rgt_channel_write(RGT_CHANNELP channel, CFL_BUFFERP buffer) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_write", (NULL));
   success = rgt_channel_isOpen(channel) && channel->write(channel, buffer);
   RGT_LOG_EXIT("rgt_channel_write", (NULL));
   return success;
}

CFL_BOOL rgt_channel_writeAndRead(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_writeAndRead", (NULL));
   success = rgt_channel_isOpen(channel) && channel->writeAndRead(channel, buffer, timeout);
   RGT_LOG_EXIT("rgt_channel_writeAndRead", (NULL));
   return success;
}

CFL_BOOL rgt_channel_writeAndReadFirstCommand(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_writeAndReadFirstCommand", (NULL));
   success = rgt_channel_isOpen(channel) && channel->writeAndReadFirstCommand(channel, buffer, timeout);
   RGT_LOG_EXIT("rgt_channel_writeAndReadFirstCommand", (NULL));
   return success;
}

CFL_UINT64 rgt_channel_lastRead(RGT_CHANNELP channel) {
   RGT_LOG_ENTER("rgt_channel_lastRead", (NULL));
   if (channel != NULL) {
      RGT_LOG_EXIT("rgt_channel_lastRead", (NULL));
      return channel->lastRead;
   } else {
      RGT_LOG_ERROR(("rgt_channel_lastRead(). channel already released."));
      RGT_LOG_EXIT("rgt_channel_lastRead", (NULL));
      return 0;
   }
}

CFL_UINT64 rgt_channel_lastWrite(RGT_CHANNELP channel) {
   RGT_LOG_ENTER("rgt_channel_lastWrite", (NULL));
   if (channel != NULL) {
      RGT_LOG_EXIT("rgt_channel_lastWrite", (NULL));
      return channel->lastWrite;
   } else {
      RGT_LOG_ERROR(("rgt_channel_lastWrite(). channel already released."));
      RGT_LOG_EXIT("rgt_channel_lastWrite", (NULL));
      return 0;
   }
}

HB_FUNC(RGT_CHANNELTYPE) {
   PHB_ITEM pNewType;
   
   RGT_LOG_ENTER("RGT_CHANNELTYPE", (NULL));
   pNewType = hb_param(1, HB_IT_NUMERIC);
   hb_retni(rgt_channel_type());
   if (pNewType) {
      rgt_channel_setType(hb_itemGetNI(pNewType));
   }
   RGT_LOG_EXIT("RGT_CHANNELTYPE", (NULL));
}
