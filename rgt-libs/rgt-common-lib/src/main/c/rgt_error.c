#include "hbvm.h"
#include "hbapi.h"
#include "hbapiitm.h"
#include "hbapierr.h"

#include "rgt_error.h"
#include "rgt_log.h"
#include "cfl_str.h"
#include "cfl_atomic.h"

static RGT_ERROR s_error = RGT_EMPTY_ERROR;

void rgt_error_finalize(void) {
   RGT_LOG_ENTER("rgt_error_finalize", (NULL));
   if (! cfl_atomic_compareAndSetBoolean(&s_error.locked, CFL_FALSE, CFL_TRUE)) {
      if (s_error.message != NULL) {
         cfl_str_free(s_error.message);
         s_error.message = NULL;
      }
      cfl_atomic_setBoolean(&s_error.locked, CFL_FALSE);
   }
   RGT_LOG_EXIT("rgt_error_finalize", (NULL));
}

void rgt_error_clear() {
   RGT_LOG_ENTER("rgt_error_clear", (NULL));
   if (! cfl_atomic_compareAndSetBoolean(&s_error.locked, CFL_FALSE, CFL_TRUE)) {
      s_error.type = RGT_ERROR_TYPE_NO_ERROR;
      s_error.code = RGT_ERROR_NONE;
      if (s_error.message) {
         cfl_str_clear(s_error.message);
      }
      cfl_atomic_setBoolean(&s_error.locked, CFL_FALSE);
   }
   RGT_LOG_EXIT("rgt_error_clear", (NULL));
}

CFL_BOOL rgt_error_hasError(void) {
   CFL_BOOL error;
   if (! cfl_atomic_compareAndSetBoolean(&s_error.locked, CFL_FALSE, CFL_TRUE)) {
      error = s_error.type != RGT_ERROR_TYPE_NO_ERROR || s_error.code != RGT_ERROR_NONE;
      cfl_atomic_setBoolean(&s_error.locked, CFL_FALSE);
   } else {
      error = CFL_FALSE;
   }
   return error;
}

static char * errorType(CFL_UINT8 errorType) {
   switch (errorType) {
      case RGT_TERMINAL:
         return "TE";

      case RGT_SERVER:
         return "SRV";

      case RGT_APP:
         return "APP";

      default:
         return "UNKNOWN";
   }
}

void rgt_error_set(CFL_UINT8 errType, CFL_UINT16 errCode, const char * message, ...) {
   va_list pArgs;

   RGT_LOG_ENTER("rgt_error_set", (NULL));
   if (! cfl_atomic_compareAndSetBoolean(&s_error.locked, CFL_FALSE, CFL_TRUE)) {
      s_error.type = errType;
      s_error.code = errCode;

      if (message != NULL) {
         va_start(pArgs, message);
         s_error.message = cfl_str_setFormatArgs(s_error.message, message, pArgs);
         va_end(pArgs);
         RGT_LOG_DEBUG(("%s(%d): %s", errorType(errType), errCode, cfl_str_getPtr(s_error.message)));
      } else {
         RGT_LOG_DEBUG(("%s(%d)", errorType(errType), errCode));
         if (s_error.message) {
            cfl_str_free(s_error.message);
            s_error.message = NULL;
         }
      }
      cfl_atomic_setBoolean(&s_error.locked, CFL_FALSE);
   }
   RGT_LOG_EXIT("rgt_error_set", (NULL));
}

RGT_ERRORP rgt_error_getLast(void) {
   return &s_error;
}

char * rgt_error_getLastMessage(void) {
   char *msg;
   RGT_LOG_ENTER("rgt_error_getLastMessage", (NULL));
   if (! cfl_atomic_compareAndSetBoolean(&s_error.locked, CFL_FALSE, CFL_TRUE)) {
      if (s_error.message != NULL) {
         msg = cfl_str_getPtr(s_error.message);
      } else {
         msg = "";
      }
      cfl_atomic_setBoolean(&s_error.locked, CFL_FALSE);
   } else {
      msg = "";
   }
   RGT_LOG_EXIT("rgt_error_getLastMessage", (NULL));
   return msg;
}

CFL_UINT8 rgt_error_getLastType(void) {
   return s_error.type;
}

CFL_INT16 rgt_error_getLastCode(void) {
   return s_error.code;
}

HB_ERRCODE rgt_error_launch(CFL_UINT16 uiGenCode, CFL_UINT16 uiSubCode, CFL_UINT16 uiFlags,
                            const char *description, const char *operation, PHB_ITEM *pErrorPtr) {
   HB_ERRCODE errCode = HB_FAILURE;

   RGT_LOG_ENTER("rgt_error_launch", (NULL));
   if( hb_vmRequestQuery() == 0 ) {
       PHB_ITEM pError;
      if (pErrorPtr) {
         if (! *pErrorPtr) {
            *pErrorPtr = hb_errNew();
         }
         pError = *pErrorPtr;
      } else {
         pError = hb_errNew();
      }
      hb_errPutGenCode(pError, uiGenCode);
      hb_errPutSubCode(pError, uiSubCode);
      if (description) {
         hb_errPutDescription(pError, description);
      } else {
         hb_errPutDescription(pError, "Internal error.");
      }
      if (operation) {
         hb_errPutOperation(pError, operation);
      }

      if (uiFlags) {
         hb_errPutFlags(pError, uiFlags);
      }

      hb_errPutSeverity(pError, ES_ERROR);
      hb_errPutSubSystem(pError, "RGT");
      errCode = hb_errLaunch(pError);

      if (! pErrorPtr) {
         hb_itemRelease(pError);
      }
   }
   RGT_LOG_EXIT("rgt_error_launch", (NULL));
   return errCode;
}

HB_ERRCODE rgt_error_launchFromRGTError(CFL_UINT16 uiFlags, const char *operation, PHB_ITEM *pErrorPtr) {
   HB_ERRCODE errcode;
   RGT_LOG_ENTER("rgt_error_launchFromRGTError", (NULL));
   if (! cfl_atomic_compareAndSetBoolean(&s_error.locked, CFL_FALSE, CFL_TRUE)) {
      errcode = rgt_error_launch(s_error.type == RGT_TERMINAL ? RGT_ERROR_TERMINAL : RGT_ERROR_APP,
                                 s_error.code,
                                 uiFlags,
                                 cfl_str_getPtr(s_error.message),
                                 operation,
                                 pErrorPtr);
      cfl_atomic_setBoolean(&s_error.locked, CFL_FALSE);
   } else {
      errcode = HB_FAILURE;
   }
   RGT_LOG_EXIT("rgt_error_launchFromRGTError", (NULL));
   return errcode;
}
