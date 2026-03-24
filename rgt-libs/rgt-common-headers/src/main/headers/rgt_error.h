#ifndef _RGT_ERROR_H_
#define _RGT_ERROR_H_

#include "rgt_types.h"
#include "rgt_error.ch"

#include "cfl_str.h"

#define RGT_EMPTY_ERROR { NULL, RGT_ERROR_NONE, RGT_ERROR_TYPE_NO_ERROR, CFL_FALSE }

struct _RGT_ERROR {
   CFL_STRP   message;
   CFL_UINT16 code;
   CFL_UINT8  type;
   CFL_BOOL   locked;
};

extern void rgt_error_finalize(void);
extern void rgt_error_clear(void);
extern CFL_BOOL rgt_error_hasError(void);
extern void rgt_error_set(CFL_UINT8 errorType, CFL_UINT16 errorCode, const char * message, ...);
extern RGT_ERRORP rgt_error_getLast(void);
extern char * rgt_error_getLastMessage(void);
extern CFL_UINT8 rgt_error_getLastType(void);
extern CFL_INT16 rgt_error_getLastCode(void);
extern HB_ERRCODE rgt_error_launch(CFL_UINT16 uiGenCode, CFL_UINT16 uiSubCode, CFL_UINT16 uiFlags,
                            const char *description, const char *operation, PHB_ITEM *pErrorPtr);
extern HB_ERRCODE rgt_error_launchFromRGTError(CFL_UINT16 uiFlags, const char *operation, PHB_ITEM *pErrorPtr);

#endif
