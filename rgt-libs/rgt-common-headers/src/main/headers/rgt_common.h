#ifndef _RGT_COMMON_H_
#define _RGT_COMMON_H_

#include "rgt_types.h"

extern void rgt_common_prepareCommand(CFL_BUFFERP buffer, CFL_UINT8 cmdId);
extern void rgt_common_prepareFirstCommand(CFL_BUFFERP buffer, CFL_UINT8 cmdId);
extern void rgt_common_prepareResponse(CFL_BUFFERP buffer, CFL_UINT16 errCode);
extern void rgt_common_prepareAdminResponse(CFL_BUFFERP buffer, CFL_UINT64 adminClientId, CFL_UINT16 errCode);
extern void rgt_common_errorFromServer(CFL_UINT8 errorType, CFL_UINT16 errorCode, CFL_BUFFERP buffer, char *message);

#endif
