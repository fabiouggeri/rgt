#include "hbapiitm.h"

#include "rgt_channel.h"
#include "rgt_channel_bidirectional.h"
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
   // if (hb_stricmp(channelTypeName, RGT_CHANNEL_XXXXX) == 0) {
   //    s_channelType = RGT_CHANNEL_XXXX;
   // } else {
   s_channelType = RGT_CHANNEL_BIDIRECTIONAL;
   // }
   RGT_LOG_EXIT("rgt_channel_setTypeByName", (NULL));
}

RGT_CHANNELP rgt_channel_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port) {
   RGT_CHANNELP channel;
   RGT_LOG_ENTER("rgt_channel_open", (NULL));
   // switch (s_channelType) {
   // case RGT_CHANNEL_XXXX:
   //    channel = rgt_channel_xxxx_open(connectionType, server, port);
   //    break;
   // default:
   channel = rgt_channel_bidirectional_open(connectionType, server, port);
   //    break;
   // }
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

CFL_BOOL rgt_channel_waitData(RGT_CHANNELP channel, CFL_UINT32 timeout, CFL_BOOL *bTimeout) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_waitData", (NULL));
   success = channel != NULL && channel->waitData(channel, timeout, bTimeout);
   RGT_LOG_EXIT("rgt_channel_waitData", (NULL));
   return success;
}

CFL_BOOL rgt_channel_hasData(RGT_CHANNELP channel) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_hasData", (NULL));
   success = channel != NULL && channel->hasData(channel);
   RGT_LOG_EXIT("rgt_channel_hasData", (NULL));
   return success;
}

CFL_BOOL rgt_channel_tryRead(RGT_CHANNELP channel, CFL_BUFFERP buffer) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_tryRead", (NULL));
   success = channel != NULL && channel->tryRead(channel, buffer);
   RGT_LOG_EXIT("rgt_channel_tryRead", (NULL));
   return success;
}

CFL_BUFFERP rgt_channel_tryReadBuffer(RGT_CHANNELP channel) {
   RGT_LOG_ENTER("rgt_channel_tryReadBuffer", (NULL));
   if (channel != NULL) {
      CFL_BUFFERP buffer = channel->tryReadBuffer(channel);
      RGT_LOG_EXIT("rgt_channel_tryReadBuffer", (NULL));
      return buffer;
   }
   RGT_LOG_EXIT("rgt_channel_tryReadBuffer", (NULL));
   return NULL;
}

CFL_BOOL rgt_channel_read(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_read", (NULL));
   success = channel != NULL && channel->read(channel, buffer, timeout, bTimeout);
   RGT_LOG_EXIT("rgt_channel_read", (NULL));
   return success;
}

CFL_BUFFERP rgt_channel_readBuffer(RGT_CHANNELP channel, CFL_UINT32 timeout, CFL_BOOL *bTimeout) {
   RGT_LOG_ENTER("rgt_channel_readBuffer", (NULL));
   if (channel != NULL) {
      CFL_BUFFERP buffer = channel->readBuffer(channel, timeout, bTimeout);
      RGT_LOG_EXIT("rgt_channel_readBuffer", (NULL));
      return buffer;
   }
   RGT_LOG_EXIT("rgt_channel_readBuffer", (NULL));
   return NULL;
}

CFL_BOOL rgt_channel_write(RGT_CHANNELP channel, CFL_BUFFERP buffer) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_write", (NULL));
   success = channel != NULL && channel->write(channel, buffer);
   RGT_LOG_EXIT("rgt_channel_write", (NULL));
   return success;
}

CFL_BOOL rgt_channel_writeAndRead(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_writeAndRead", (NULL));
   success = channel != NULL && channel->writeAndRead(channel, buffer, timeout, bTimeout);
   RGT_LOG_EXIT("rgt_channel_writeAndRead", (NULL));
   return success;
}

CFL_BOOL rgt_channel_writeAndReadFirstCommand(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout) {
   CFL_BOOL success;
   RGT_LOG_ENTER("rgt_channel_writeAndReadFirstCommand", (NULL));
   success = channel != NULL && channel->writeAndReadFirstCommand(channel, buffer, timeout, bTimeout);
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
