#ifndef RGT_TRM_CONNECTION_H_

#define RGT_TRM_CONNECTION_H_

#include <time.h>

#include "cfl_lock.h"
#include "rgt_types.h"

#include "hbgtcore.h"

#define RGT_TRM_IO_BUFFER_SIZE 8192

struct _RGT_TRM_CONNECTION {
      CFL_STRP server;
      CFL_STRP username;
      RGT_CHANNELP channel;
      CFL_BUFFERP keyBuffer;
      CFL_SYNC_QUEUEP appRequestsQueue;
      CFL_SYNC_QUEUEP serverRequestsQueue;
      CFL_SYNC_QUEUEP responseQueue;
      RGT_SCREENP screen;
      PHB_GT pGT;
      CFL_INT64 sessionId;
      CFL_UINT64 lastTimeReceivedAppData;
      CFL_UINT64 lastTimeSentDataToApp;
      CFL_UINT64 lastTimeSentKeysToApp;
      CFL_UINT32 timeoutAppCommunication;
      CFL_UINT32 sendKeysInterval;
      CFL_UINT32 timeout;
      CFL_UINT16 port;
      RGT_CLP_COMP appClpCompiler;
      CFL_BOOL isActive;
};

extern RGT_TRM_CONNECTIONP rgt_trm_login(const char *server, CFL_UINT16 port, const char *commandLine, const char *workDir,
                                         const char *username, const char *password, CFL_UINT16 argc, const char *argv[]);
extern void rgt_trm_logout(void);
extern void rgt_trm_listenConnection(RGT_TRM_CONNECTIONP conn);
extern void rgt_trm_connSetTimeout(RGT_TRM_CONNECTIONP conn, CFL_UINT32 timeout);
extern CFL_UINT32 rgt_trm_connGetTimeout(RGT_TRM_CONNECTIONP conn);
extern void rgt_trm_setMaxKeysInBuffer(CFL_UINT32 maxKeys);
extern CFL_UINT32 rgt_trm_getMaxKeysInBuffer(void);
extern CFL_INT64 rgt_trm_sessionId(void);

#endif
