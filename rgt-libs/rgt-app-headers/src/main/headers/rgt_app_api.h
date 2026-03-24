#ifndef _RGT_APP_API_H

#define _RGT_APP_API_H

#include "rgt_types.h"

extern CFL_BOOL rgt_app_openConnection(char *server, char *port, char *strSessionId);
extern void rgt_app_closeConnection(void);
extern RGT_APP_CONNECTIONP rgt_app_getConnection(void);
extern CFL_BOOL rgt_app_isConnected(void);
extern CFL_BOOL rgt_app_inTransactionMode(void);
extern CFL_UINT32 rgt_app_getKeepAliveInterval(void);
extern void rgt_app_setKeepAliveInterval(CFL_UINT32 newValue);
extern void rgt_app_handleLastError(const char *operation);
extern void rgt_app_setTitle(const char *title);
extern void rgt_app_initEnv(void);

#endif