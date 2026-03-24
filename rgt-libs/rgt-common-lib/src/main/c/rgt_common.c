#include <stdio.h>
#include <stdlib.h>

#include "cfl_socket.h"
#include "cfl_buffer.h"
#include "cfl_str.h"
#include "cfl_mem.h"

#include "rgt_common.h"
#include "rgt_error.h"
#include "rgt_log.h"

void rgt_common_prepareCommand(CFL_BUFFERP buffer, CFL_UINT8 cmdId) {
   cfl_buffer_reset(buffer);
   cfl_buffer_putInt32(buffer, 0);
   cfl_buffer_putUInt8(buffer, cmdId);
}

void rgt_common_prepareFirstCommand(CFL_BUFFERP buffer, CFL_UINT8 cmdId) {
   cfl_buffer_reset(buffer);
   cfl_buffer_putInt32(buffer, RGT_HEADER_MAGIC_NUMBER);
   cfl_buffer_putInt32(buffer, 0);
   cfl_buffer_putUInt8(buffer, cmdId);
}

void rgt_common_prepareResponse(CFL_BUFFERP buffer, CFL_UINT16 errCode) {
   cfl_buffer_reset(buffer);
   cfl_buffer_putInt32(buffer, 0);
   cfl_buffer_putUInt16(buffer, errCode);
}

void rgt_common_prepareAdminResponse(CFL_BUFFERP buffer, CFL_UINT64 adminClientId, CFL_UINT16 errCode) {
   cfl_buffer_reset(buffer);
   cfl_buffer_putInt32(buffer, 0);
   cfl_buffer_putUInt16(buffer, RGT_ADMIN_RESPONSE);
   cfl_buffer_putUInt64(buffer, adminClientId);
   cfl_buffer_putUInt16(buffer, errCode);
}

void rgt_common_errorFromServer(CFL_UINT8 errorType, CFL_UINT16 errorCode, CFL_BUFFERP buffer, char *message) {
   char * msgError = cfl_buffer_getCharArray(buffer);
   if (message) {
      rgt_error_set(errorType, errorCode, "Error %d - %s: %s", errorCode, message, msgError);
   } else {
      rgt_error_set(errorType, errorCode, "Error %d - %s", errorCode, msgError);
   }
   CFL_MEM_FREE(msgError);
}

