#ifndef _RGT_CHANNEL_H_
#define _RGT_CHANNEL_H_

#include "cfl_buffer.h"
#include "rgt_types.h"

typedef void (*RGT_CHANNEL_CLOSE)(RGT_CHANNELP channel);
typedef void (*RGT_CHANNEL_FREE)(RGT_CHANNELP channel);
typedef CFL_BOOL (*RGT_CHANNEL_WAIT_DATA)(RGT_CHANNELP channel, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
typedef CFL_BOOL (*RGT_CHANNEL_HAS_DATA)(RGT_CHANNELP channel);
typedef CFL_BOOL (*RGT_CHANNEL_TRY_READ)(RGT_CHANNELP channel, CFL_BUFFERP buffer);
typedef CFL_BUFFERP (*RGT_CHANNEL_TRY_READ_BUFFER)(RGT_CHANNELP channel);
typedef CFL_BOOL (*RGT_CHANNEL_READ)(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
typedef CFL_BUFFERP (*RGT_CHANNEL_READ_BUFFER)(RGT_CHANNELP channel, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
typedef CFL_BOOL (*RGT_CHANNEL_WRITE)(RGT_CHANNELP channel, CFL_BUFFERP buffer);
typedef CFL_BOOL (*RGT_CHANNEL_WRITE_READ)(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
typedef CFL_BOOL (*RGT_CHANNEL_WRITE_READ_FIRST)(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
typedef CFL_BOOL (*RGT_CHANNEL_IS_OPEN)(RGT_CHANNELP channel);

struct _RGT_CHANNEL {
      RGT_CHANNEL_CLOSE close;
      RGT_CHANNEL_FREE free;
      RGT_CHANNEL_WAIT_DATA waitData;
      RGT_CHANNEL_HAS_DATA hasData;
      RGT_CHANNEL_TRY_READ tryRead;
      RGT_CHANNEL_TRY_READ_BUFFER tryReadBuffer;
      RGT_CHANNEL_READ read;
      RGT_CHANNEL_READ_BUFFER readBuffer;
      RGT_CHANNEL_WRITE write;
      RGT_CHANNEL_WRITE_READ writeAndRead;
      RGT_CHANNEL_WRITE_READ_FIRST writeAndReadFirstCommand;
      RGT_CHANNEL_IS_OPEN isOpen;
      CFL_UINT64 lastRead;
      CFL_UINT64 lastWrite;
      CFL_UINT8 connectionType;
};

extern CFL_UINT8 rgt_channel_type(void);
extern void rgt_channel_setType(CFL_UINT8 newChannelType);
extern void rgt_channel_setTypeByName(const char *channelTypeName);
extern RGT_CHANNELP rgt_channel_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port);
extern void rgt_channel_close(RGT_CHANNELP channel);
extern void rgt_channel_free(RGT_CHANNELP channel);
extern CFL_BOOL rgt_channel_waitData(RGT_CHANNELP channel, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
extern CFL_BOOL rgt_channel_hasData(RGT_CHANNELP channel);
extern CFL_BUFFERP rgt_channel_tryReadBuffer(RGT_CHANNELP channel);
extern CFL_BUFFERP rgt_channel_readBuffer(RGT_CHANNELP channel, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
extern CFL_BOOL rgt_channel_tryRead(RGT_CHANNELP channel, CFL_BUFFERP buffer);
extern CFL_BOOL rgt_channel_read(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
extern CFL_BOOL rgt_channel_write(RGT_CHANNELP channel, CFL_BUFFERP buffer);
extern CFL_BOOL rgt_channel_writeAndRead(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout, CFL_BOOL *bTimeout);
extern CFL_BOOL rgt_channel_writeAndReadFirstCommand(RGT_CHANNELP channel, CFL_BUFFERP buffer, CFL_UINT32 timeout,
                                                     CFL_BOOL *bTimeout);
extern CFL_BOOL rgt_channel_isOpen(RGT_CHANNELP channel);
extern CFL_UINT64 rgt_channel_lastRead(RGT_CHANNELP channel);
extern CFL_UINT64 rgt_channel_lastWrite(RGT_CHANNELP channel);

#endif
