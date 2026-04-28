#ifndef _RGT_APP_CONNECTION_H

#define _RGT_APP_CONNECTION_H

#include <time.h>

#include "rgt_screen.h"
#include "rgt_types.h"

#include "cfl_str.h"
#include "cfl_sync_queue.h"

#define rgt_app_conn_screenRectUpdated(conn, rowIni, colIni, rowEnd, colEnd)                                                       \
   rgt_screen_rectUpdated((conn)->screen, rowIni, colIni, rowEnd, colEnd)

#define IN_TRANSACTION_MODE(c) ((c)->sessionMode == RGT_SESS_MODE_TRANSACTION)
#define RPC_LOCAL_EXEC_LOST_CONNECTION(c) ((c)->rpcExecuteLocalLostConnection)

struct _RGT_APP_CONNECTION {
      CFL_STRP serverAddress;
      CFL_INT64 sessionId;
      RGT_CHANNELP channel;
      CFL_SYNC_QUEUEP receivedResponses;
      CFL_BUFFERP sendBuffer;
      CFL_BUFFERP recvBuffer;
      RGT_SCREENP screen;
      CFL_BUFFERP keyBuffer;
      RGT_LOCK keyBufferLocked;
      CFL_BUFFERP toneBuffer;
      RGT_LOCK toneBufferLocked;
      RGT_THREADP updateTerminalThread;
      CFL_UINT64 lastTerminalUpdate;
      CFL_UINT32 updateTerminalInterval;
      CFL_UINT32 fileTransferChunkSize;
      CFL_UINT32 timeout;
      CFL_UINT32 rpcTimeout;
      CFL_UINT32 tonesCount;
      CFL_UINT32 keepAliveInterval;
      CFL_UINT16 port;
      CFL_BOOL active;
      CFL_UINT8 sessionMode;
      CFL_BOOL rpcExecuteLocalLostConnection;
};

extern RGT_APP_CONNECTIONP rgt_app_conn_new(const char *server, CFL_UINT16 port, CFL_INT64 sessionId);
extern void rgt_app_connSetTimeout(RGT_APP_CONNECTIONP conn, CFL_UINT32 timeout);
extern CFL_UINT32 rgt_app_connGetTimeout(RGT_APP_CONNECTIONP conn);
extern void rgt_app_setDefaultTimeout(CFL_UINT32 defaultTimeout);
extern CFL_UINT32 rgt_app_getDefaultTimeout(void);
extern void rgt_app_conn_close(RGT_APP_CONNECTIONP conn, const char *message);
extern void rgt_app_conn_free(RGT_APP_CONNECTIONP conn);
extern CFL_BOOL rgt_app_conn_isActive(RGT_APP_CONNECTIONP conn);
extern CFL_BOOL rgt_app_conn_isUpdateBackground(RGT_APP_CONNECTIONP conn);
extern void rgt_app_conn_setCursorPos(RGT_APP_CONNECTIONP conn, int iRow, int iCol);
extern void rgt_app_conn_setCursorStyle(RGT_APP_CONNECTIONP conn, int iStyle);
extern void rgt_app_conn_updateTerminal(RGT_APP_CONNECTIONP conn);
extern int rgt_app_conn_getKey(RGT_APP_CONNECTIONP conn);
extern CFL_BOOL rgt_app_conn_prepareTerminal(RGT_APP_CONNECTIONP conn);
extern CFL_INT32 rgt_app_conn_putFile(RGT_APP_CONNECTIONP conn, const char *fileToSent, const char *remotePahName,
                                      CFL_UINT32 chunkSize);
extern CFL_INT32 rgt_app_conn_getFile(RGT_APP_CONNECTIONP conn, const char *fileToReceive, const char *localPahName,
                                      CFL_UINT32 chunkSize);
extern PHB_ITEM rgt_app_conn_execRemoteFunction(RGT_APP_CONNECTIONP conn, const char *functionName, CFL_INT16 argsCount,
                                                PHB_ITEM *pArgs);
extern void rgt_app_conn_tone(RGT_APP_CONNECTIONP conn, double dFrequency, double dDuration);
extern CFL_BOOL rgt_app_setSessionOption(RGT_APP_CONNECTIONP conn, const char *optionName, const char *optionValue);
extern void rgt_app_conn_setFileTransferChunkSize(RGT_APP_CONNECTIONP conn, CFL_UINT32 chunkSize);
extern CFL_UINT32 rgt_app_conn_getFileTransferChunkSize(RGT_APP_CONNECTIONP conn);
extern CFL_BOOL rgt_app_conn_isRpcExecuteLocalLostConnection(RGT_APP_CONNECTIONP conn);
extern void rgt_app_conn_setRpcExecuteLocalLostConnection(RGT_APP_CONNECTIONP conn, CFL_BOOL execLocal);

#endif
