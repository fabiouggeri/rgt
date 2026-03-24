#ifndef _RGT_DOUBLE_CHANNEL_H_
#define _RGT_DOUBLE_CHANNEL_H_

#include "cfl_types.h"
#include "rgt_channel.h"

extern RGT_CHANNELP rgt_channel_bidirectional_open(CFL_UINT8 connectionType, const char *server, CFL_UINT16 port);

#endif

