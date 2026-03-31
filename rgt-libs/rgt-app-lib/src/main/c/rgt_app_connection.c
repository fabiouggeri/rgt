#include <stdio.h>
#include <stdlib.h>

#include "hbapi.h"
#include "hbapierr.h"
#include "hbapifs.h"
#include "hbapigt.h"
#include "hbapiitm.h"
#include "hbset.h"
#include "hbvm.h"

#include "cfl_buffer.h"
#include "cfl_event.h"
#include "cfl_lock.h"
#include "cfl_mem.h"
#include "cfl_process.h"
#include "cfl_str.h"
#include "rgt_thread.h"

#include "rgt_app_connection.h"
#include "rgt_channel.h"
#include "rgt_common.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_rpc.h"
#include "rgt_screen.h"
#include "rgt_util.h"

#define DEFAULT_KEEP_ALIVE_INTERVAL (1 * MINUTE)
#define DEFAULT_RPC_TIMEOUT (0 * MINUTE)
#define DEFAULT_UPDATE_INTERVAL 50

#define IS_SEND_KEEP_ALIVE(c, t) ((c)->keepAliveInterval > 0 && (t) > (c)->keepAliveInterval)

#ifdef __HBR__
#define CLP_COMPILER RGT_CLP_COMP_HARBOUR
#else
#define CLP_COMPILER RGT_CLP_COMP_XHARBOUR
#endif

#define IS_UPDATE_TERMINAL(c) (rgt_screen_isChanged((c)->screen) || (c)->tonesCount > 0)

static CFL_UINT32 s_defaultTimeout = 30 * SECOND;

static CFL_UINT16 readBackgroundCommands(RGT_APP_CONNECTIONP conn, CFL_UINT16 errCode, CFL_UINT32 timeout);

static CFL_BOOL transactionMode(RGT_APP_CONNECTIONP conn) {
   conn->active = CFL_FALSE;
   return IN_TRANSACTION_MODE(conn) ? CFL_TRUE : CFL_FALSE;
}

static CFL_BOOL sendKeepAlive(RGT_APP_CONNECTIONP conn, CFL_BUFFERP buffer) {
   RGT_LOG_ENTER("sendKeepAlive", (NULL));
   RGT_LOG_DEBUG(("sendKeepAlive()"));
   rgt_common_prepareCommand(buffer, RGT_APP_CMD_KEEP_ALIVE);
   cfl_buffer_putInt8(buffer, (CFL_INT8)0);
   if (!rgt_channel_write(conn->channel, buffer) && !transactionMode(conn)) {
      RGT_LOG_ERROR(("sendKeepAlive(): %s", rgt_error_getLastMessage()));
      RGT_LOG_EXIT("sendKeepAlive", (NULL));
      return CFL_FALSE;
   }
   RGT_LOG_EXIT("sendKeepAlive", (NULL));
   return CFL_TRUE;
}

static void copyTonesToBuffer(RGT_APP_CONNECTIONP conn, CFL_BUFFERP buffer) {
   RGT_LOG_ENTER("copyTonesToBuffer", ("tones=%u", conn->tonesCount));
   RGT_LOCK_ACQUIRE(conn->toneBufferLocked);
   cfl_buffer_putUInt16(buffer, (CFL_UINT16)conn->tonesCount);
   cfl_buffer_flip(conn->toneBuffer);
   cfl_buffer_putBufferSize(buffer, conn->toneBuffer, (conn->tonesCount * sizeof(double) * 2));
   cfl_buffer_reset(conn->toneBuffer);
   conn->tonesCount = 0;
   RGT_LOCK_RELEASE(conn->toneBufferLocked);
   RGT_LOG_EXIT("copyTonesToBuffer", (NULL));
}

static CFL_BOOL updateTerminal(RGT_APP_CONNECTIONP conn, CFL_BUFFERP buffer) {
   CFL_BOOL screenChanged;
   RGT_LOG_ENTER("updateTerminal", (NULL));
   if (!rgt_app_conn_isActive(conn)) {
      RGT_LOG_EXIT("updateTerminal", (NULL));
      return CFL_FALSE;
   }
   screenChanged = rgt_screen_isChanged(conn->screen);
   RGT_LOG_DEBUG(("updateTerminal(): screen=%s tone=%u row=%d, col=%d", (screenChanged ? "true" : "false"), conn->tonesCount,
                  rgt_screen_getCursorRow(conn->screen), rgt_screen_getCursorCol(conn->screen)));
   if (screenChanged || conn->tonesCount > 0) {
      rgt_common_prepareCommand(buffer, RGT_APP_CMD_UPDATE);
      rgt_screen_toBuffer(conn->screen, buffer, CFL_TRUE);
      copyTonesToBuffer(conn, buffer);
      if (!rgt_channel_write(conn->channel, buffer) && !transactionMode(conn)) {
         RGT_LOG_ERROR(("updateTerminal(): %s", rgt_error_getLastMessage()));
         RGT_LOG_EXIT("updateTerminal", (NULL));
         return CFL_FALSE;
      }
      conn->lastTerminalUpdate = CURRENT_TIME;
   }
   RGT_LOG_EXIT("updateTerminal", (NULL));
   return CFL_TRUE;
}

static void backgroundTasks(void *param) {
   RGT_APP_CONNECTIONP conn;
   CFL_BUFFERP buffer = cfl_buffer_newCapacity(RGT_APP_IO_BUFFER_SIZE);
   CFL_BOOL running = CFL_TRUE;
   RGT_LOG_ENTER("backgroundTasks", (NULL));
   RGT_LOG_INFO(("rgt_app_connection.backgroundTasks(). started."));
   conn = (RGT_APP_CONNECTIONP)param;
   while (running && rgt_app_conn_isActive(conn) && hb_vmIsActive()) {
      CFL_UINT64 currTime = CURRENT_TIME;
      if (TIMEMILLIS_ELAPSED(conn->lastTerminalUpdate, currTime) >= conn->updateTerminalInterval && IS_UPDATE_TERMINAL(conn)) {
         running = updateTerminal(conn, buffer);
      } else if (IS_SEND_KEEP_ALIVE(conn, TIMEMILLIS_ELAPSED(rgt_channel_lastWrite(conn->channel), currTime))) {
         running = sendKeepAlive(conn, buffer);
      }
      rgt_thread_sleep(conn->updateTerminalInterval);
   }
   cfl_buffer_free(buffer);
   RGT_LOG_INFO(("rgt_app_connection.backgroundTasks(). finished."));
   RGT_LOG_EXIT("backgroundTasks", (NULL));
}

CFL_BOOL rgt_app_conn_prepareTerminal(RGT_APP_CONNECTIONP conn) {
   CFL_INT16 iRows = (CFL_INT16)hb_gtMaxRow() + 1;
   CFL_INT16 iCols = (CFL_INT16)hb_gtMaxCol() + 1;
   CFL_INT16 keyBufSize = hb_setGetTypeAhead();
   char szColorStr[HB_CLRSTR_LEN];

   RGT_LOG_ENTER("rgt_app_conn_prepareTerminal", (NULL));
   if (conn->screen == NULL) {
      conn->screen = rgt_screen_new(iRows, iCols, rgt_screen_type());
   } else {
      rgt_screen_reset(conn->screen, iRows, iCols, rgt_screen_type());
   }
   cfl_buffer_reset(conn->availableKeysBuffer);
   rgt_screen_capture(conn->screen);
   hb_gtGetColorStr(szColorStr);
   rgt_common_prepareCommand(conn->buffer, RGT_APP_CMD_SET_ENV);
   /* Terminal/App protocol version */
   cfl_buffer_putInt16(conn->buffer, RGT_TRM_APP_PROTOCOL_VERSION);
   /* Compiler and Screen type */
   cfl_buffer_putInt8(conn->buffer, (CLP_COMPILER << 4) | conn->screen->screenType);
   /* rows count */
   cfl_buffer_putInt16(conn->buffer, iRows);
   /* columns count */
   cfl_buffer_putInt16(conn->buffer, iCols);
   /* screen content */
   rgt_screen_toBuffer(conn->screen, conn->buffer, CFL_TRUE);
   /* color string */
   cfl_buffer_putCharArray(conn->buffer, szColorStr);
   /* keyboard buffer size */
   cfl_buffer_putInt16(conn->buffer, keyBufSize);
   if (rgt_app_conn_isActive(conn) && rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->timeout)) {
      CFL_INT16 errCode = readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->timeout);
      if (errCode != RGT_RESP_SUCCESS) {
         rgt_common_errorFromServer(RGT_APP, errCode, conn->buffer, "Error preparing terminal");
         RGT_LOG_EXIT("rgt_app_conn_prepareTerminal", (NULL));
         return CFL_FALSE;
      }
   } else {
      RGT_LOG_DEBUG(("rgt_app_conn_prepareTerminal(): error communicating with TE"));
      RGT_LOG_EXIT("rgt_app_conn_prepareTerminal", (NULL));
      return CFL_FALSE;
   }
   RGT_LOG_EXIT("rgt_app_conn_prepareTerminal", (NULL));
   return CFL_TRUE;
}

static CFL_BOOL bufferToAvailableKeys(RGT_APP_CONNECTIONP conn) {
   CFL_INT32 iKey;
   CFL_BOOL receivedKeys = CFL_FALSE;

   RGT_LOG_ENTER("bufferToAvailableKeys", (NULL));
   iKey = cfl_buffer_getInt32(conn->buffer);
   cfl_buffer_compact(conn->availableKeysBuffer);
   cfl_buffer_setPosition(conn->availableKeysBuffer, cfl_buffer_length(conn->availableKeysBuffer));
   while (iKey != 0) {
      if (iKey != HB_BREAK_FLAG || !IN_TRANSACTION_MODE(conn)) {
         cfl_buffer_putInt32(conn->availableKeysBuffer, iKey);
         receivedKeys = CFL_TRUE;
      } else {
         RGT_LOG_TRACE(("HB_BREAK_FLAG key discarded. App in transaction mode."));
      }
      iKey = cfl_buffer_getInt32(conn->buffer);
   }
   cfl_buffer_flip(conn->availableKeysBuffer);
   RGT_LOG_EXIT("bufferToAvailableKeys", (NULL));
   return receivedKeys;
}

static CFL_UINT16 readNextResponse(RGT_APP_CONNECTIONP conn, CFL_UINT32 timeout) {
   CFL_UINT16 errCode;
   RGT_LOG_ENTER("readNextResponse", (NULL));
   if (!rgt_app_conn_isActive(conn)) {
      cfl_buffer_reset(conn->buffer);
      cfl_buffer_putCharArray(conn->buffer, "Lost connection");
      cfl_buffer_flip(conn->buffer);
      errCode = RGT_ERROR_SOCKET;
   } else if (rgt_channel_read(conn->channel, conn->buffer, timeout)) {
      errCode = cfl_buffer_getUInt16(conn->buffer);
   } else {
      cfl_buffer_reset(conn->buffer);
      cfl_buffer_putCharArray(conn->buffer, rgt_error_getLastMessage());
      cfl_buffer_flip(conn->buffer);
      errCode = rgt_error_getLastCode();
   }
   RGT_LOG_EXIT("readNextResponse", (NULL));
   return errCode;
}

static CFL_UINT16 readBackgroundCommands(RGT_APP_CONNECTIONP conn, CFL_UINT16 errCode, CFL_UINT32 timeout) {
   CFL_BOOL bContinue = CFL_TRUE;
   RGT_LOG_ENTER("readBackgroundCommands", ("errCode=%hu", errCode));
   while (bContinue) {
      switch (errCode) {
      case RGT_RESP_TRM_KEY_UPDATE:
         bufferToAvailableKeys(conn);
         errCode = readNextResponse(conn, timeout);
         break;
      case RGT_RESP_TRM_KEEP_ALIVE:
         cfl_buffer_getUInt8(conn->buffer);
         errCode = readNextResponse(conn, timeout);
         break;
      default:
         bContinue = CFL_FALSE;
         break;
      }
   }
   RGT_LOG_EXIT("readBackgroundCommands", ("errCode=%hu", errCode));
   return errCode;
}

static CFL_BOOL startUpdateTerminalThread(RGT_APP_CONNECTIONP conn) {
   char *updateMode;
   RGT_LOG_ENTER("startUpdateTerminalThread", (NULL));
   updateMode = getenv(RGT_UPDATE_MODE_VAR);
   if (updateMode != NULL && hb_stricmp(updateMode, RGT_UPDATE_MODE_MAIN_THREAD) == 0) {
      RGT_LOG_DEBUG(("startUpdateTerminalThread(): main thread update"));
      RGT_LOG_EXIT("startUpdateTerminalThread", (NULL));
      return CFL_TRUE;
   }
   conn->updateTerminalThread = rgt_thread_start(backgroundTasks, conn, "RGT Background Tasks");
   RGT_LOG_DEBUG(("startUpdateTerminalThread(): background thread update"));
   RGT_LOG_EXIT("startUpdateTerminalThread", (NULL));
   return conn->updateTerminalThread != NULL;
}

static void finishUpdateTerminalThread(RGT_APP_CONNECTIONP conn) {
   RGT_THREADP thread;
   RGT_LOG_ENTER("finishUpdateTerminalThread", (NULL));
   if (conn->updateTerminalThread == NULL) {
      RGT_LOG_EXIT("finishUpdateTerminalThread", (NULL));
      return;
   }
   thread = conn->updateTerminalThread;
   conn->updateTerminalThread = NULL;
   if (rgt_thread_isRunning(thread)) {
      rgt_thread_waitTimeout(thread, 5000);
      if (rgt_thread_isRunning(thread)) {
         rgt_thread_kill(thread);
      }
   }
   rgt_thread_free(thread);
   RGT_LOG_EXIT("finishUpdateTerminalThread", (NULL));
}

static void selectChannelType(void) {
   const char *channelType;
   RGT_LOG_ENTER("selectChannelType", (NULL));
   channelType = getenv(RGT_CHANNEL_TYPE_VAR);
   if (channelType != NULL) {
      rgt_channel_setTypeByName(channelType);
   }
   RGT_LOG_EXIT("selectChannelType", (NULL));
}

static CFL_BOOL connectToServer(RGT_APP_CONNECTIONP conn) {
   RGT_CHANNELP channel;
   CFL_INT16 errCode;
   CFL_STRP logPathName;
   int logLevel;

   RGT_LOG_ENTER("connectToServer", (NULL));
   selectChannelType();
   channel = rgt_channel_open(RGT_APP, cfl_str_getPtr(conn->serverAddress), conn->port);
   if (channel == NULL) {
      rgt_error_set(RGT_APP, rgt_error_getLastCode(), "Error connecting to application server: %s", rgt_error_getLastMessage());
      RGT_LOG_EXIT("connectToServer", (NULL));
      return CFL_FALSE;
   }
   rgt_common_prepareFirstCommand(conn->buffer, RGT_APP_CMD_LOGIN);
   cfl_buffer_putInt16(conn->buffer, RGT_SERVER_PROTOCOL_VERSION);
   cfl_buffer_putInt64(conn->buffer, conn->sessionId);
   cfl_buffer_putInt64(conn->buffer, cfl_process_getId());
   if (!rgt_channel_writeAndReadFirstCommand(channel, conn->buffer, conn->timeout)) {
      rgt_channel_close(channel);
      rgt_channel_free(channel);
      RGT_LOG_EXIT("connectToServer", (NULL));
      return CFL_FALSE;
   }
   errCode = cfl_buffer_getUInt16(conn->buffer);
   if (errCode != RGT_RESP_SUCCESS) {
      rgt_channel_close(channel);
      rgt_channel_free(channel);
      rgt_common_errorFromServer(RGT_APP, errCode, conn->buffer, "Error connecting to server");
      RGT_LOG_EXIT("connectToServer", (NULL));
      return CFL_FALSE;
   }
   conn->active = CFL_TRUE;
   logLevel = (int)cfl_buffer_getInt8(conn->buffer);
   if (!rgt_log_envLevel()) {
      rgt_log_setLevel(logLevel);
   }
   logPathName = cfl_buffer_getString(conn->buffer);
   if (!rgt_log_envFile()) {
      rgt_log_setPathName(cfl_str_getPtr(logPathName));
   }
   cfl_str_free(logPathName);
   conn->channel = channel;
   RGT_LOG_INFO(("rgt_app_connection.connectToServer(). server=%s, port=%u, session=%lld", cfl_str_getPtr(conn->serverAddress),
                 conn->port, conn->sessionId));
   if (!rgt_app_conn_prepareTerminal(conn)) {
      conn->channel = NULL;
      rgt_channel_close(channel);
      rgt_channel_free(channel);
      RGT_LOG_EXIT("connectToServer", (NULL));
      return CFL_FALSE;
   }
   if (!startUpdateTerminalThread(conn)) {
      conn->channel = NULL;
      rgt_channel_close(channel);
      rgt_channel_free(channel);
      RGT_LOG_EXIT("connectToServer", (NULL));
      return CFL_FALSE;
   }
   RGT_LOG_EXIT("connectToServer", (NULL));
   return CFL_TRUE;
}

RGT_APP_CONNECTIONP rgt_app_conn_new(const char *server, CFL_UINT16 port, CFL_INT64 sessionId) {
   RGT_APP_CONNECTIONP conn;
   RGT_LOG_ENTER("rgt_app_conn_new", (NULL));
   conn = (RGT_APP_CONNECTIONP)RGT_HB_ALLOC(sizeof(RGT_APP_CONNECTION));
   conn->serverAddress = cfl_str_newBuffer(server);
   conn->port = port;
   conn->channel = NULL;
   conn->timeout = s_defaultTimeout;
   conn->sessionId = sessionId;
   conn->buffer = cfl_buffer_newCapacity(RGT_APP_IO_BUFFER_SIZE);
   conn->screen = NULL;
   conn->availableKeysBuffer = cfl_buffer_newCapacity(RGT_KEY_BUFFER_SIZE);
   conn->toneBuffer = cfl_buffer_newCapacity(RGT_TONE_BUFFER_SIZE);
   conn->fileTransferChunkSize = 32 * 1024;
   conn->lastTerminalUpdate = CURRENT_TIME;
   conn->updateTerminalInterval = DEFAULT_UPDATE_INTERVAL;
   conn->rpcTimeout = DEFAULT_RPC_TIMEOUT;
   conn->keepAliveInterval = DEFAULT_KEEP_ALIVE_INTERVAL;
   conn->active = CFL_FALSE;
   conn->updateTerminalThread = NULL;
   conn->sessionMode = RGT_SESS_MODE_NORMAL;
   conn->rpcExecuteLocalLostConnection = CFL_FALSE;
   RGT_LOCK_INIT(conn->toneBufferLocked);
   conn->tonesCount = 0;
   if (!connectToServer(conn)) {
      rgt_app_conn_free(conn);
      conn = NULL;
      RGT_LOG_DEBUG(("rgt_app_conn_new(): failed to connect"));
   }
   RGT_LOG_EXIT("rgt_app_conn_new", (NULL));
   return conn;
}

void rgt_app_connSetTimeout(RGT_APP_CONNECTIONP conn, CFL_UINT32 timeout) {
   RGT_LOG_ENTER("rgt_app_connSetTimeout", (NULL));
   conn->timeout = timeout;
   RGT_LOG_EXIT("rgt_app_connSetTimeout", (NULL));
}

CFL_UINT32 rgt_app_connGetTimeout(RGT_APP_CONNECTIONP conn) {
   CFL_UINT32 timeout;
   RGT_LOG_ENTER("rgt_app_connGetTimeout", (NULL));
   timeout = conn->timeout;
   RGT_LOG_EXIT("rgt_app_connGetTimeout", (NULL));
   return timeout;
}

void rgt_app_setDefaultTimeout(CFL_UINT32 defaultTimeout) {
   RGT_LOG_ENTER("rgt_app_setDefaultTimeout", (NULL));
   s_defaultTimeout = defaultTimeout;
   RGT_LOG_EXIT("rgt_app_setDefaultTimeout", (NULL));
}

CFL_UINT32 rgt_app_getDefaultTimeout(void) {
   CFL_UINT32 timeout;
   RGT_LOG_ENTER("rgt_app_getDefaultTimeout", (NULL));
   timeout = s_defaultTimeout;
   RGT_LOG_EXIT("rgt_app_getDefaultTimeout", (NULL));
   return timeout;
}

void rgt_app_conn_close(RGT_APP_CONNECTIONP conn, const char *message) {
   RGT_LOG_ENTER("rgt_app_conn_close", (NULL));
   if (conn == NULL) {
      RGT_LOG_EXIT("rgt_app_conn_close", (NULL));
      return;
   }
   conn->active = CFL_FALSE;
   finishUpdateTerminalThread(conn);
   if (rgt_channel_isOpen(conn->channel)) {
      rgt_common_prepareCommand(conn->buffer, RGT_APP_CMD_LOGOUT);
      cfl_buffer_putBoolean(conn->buffer, CFL_TRUE);
      if (!rgt_screen_toBuffer(conn->screen, conn->buffer, CFL_FALSE)) {
         cfl_buffer_skip(conn->buffer, -1);
         cfl_buffer_putBoolean(conn->buffer, CFL_FALSE);
      }
      copyTonesToBuffer(conn, conn->buffer);
      cfl_buffer_putCharArray(conn->buffer, message);
      if (rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->timeout)) {
         readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->timeout);
      } else {
         RGT_LOG_ERROR(("rgt_app_conn_close(): Error sending logout to terminal"));
      }
      rgt_channel_close(conn->channel);
      RGT_LOG_DEBUG(("rgt_app_conn_close(message='%s'): success", message));
   } else {
      RGT_LOG_DEBUG(("rgt_app_conn_close(): connection already closed"));
   }
   RGT_LOG_EXIT("rgt_app_conn_close", (NULL));
}

void rgt_app_conn_free(RGT_APP_CONNECTIONP conn) {
   RGT_LOG_ENTER("rgt_app_conn_free", (NULL));
   if (conn != NULL) {
      cfl_str_free(conn->serverAddress);
      conn->serverAddress = NULL;
      if (conn->buffer != NULL) {
         CFL_BUFFERP buffer = conn->buffer;
         conn->buffer = NULL;
         cfl_buffer_free(buffer);
      }
      if (conn->screen != NULL) {
         RGT_SCREENP screen = conn->screen;
         conn->screen = NULL;
         rgt_screen_free(screen);
      }
      if (conn->availableKeysBuffer != NULL) {
         CFL_BUFFERP buffer = conn->availableKeysBuffer;
         conn->availableKeysBuffer = NULL;
         cfl_buffer_free(buffer);
      }
      if (conn->toneBuffer != NULL) {
         CFL_BUFFERP buffer = conn->toneBuffer;
         conn->toneBuffer = NULL;
         cfl_buffer_free(buffer);
      }
      if (conn->channel != NULL) {
         RGT_CHANNELP channel = conn->channel;
         conn->channel = NULL;
         rgt_channel_free(channel);
      }
      RGT_LOCK_FREE(conn->toneBufferLocked);
      RGT_HB_FREE(conn);
      RGT_LOG_DEBUG(("rgt_app_conn_free(): connection released"));
   } else {
      RGT_LOG_ERROR(("rgt_app_conn_free(): NULL connection"));
   }
   RGT_LOG_EXIT("rgt_app_conn_free", (NULL));
}

CFL_BOOL rgt_app_conn_isActive(RGT_APP_CONNECTIONP conn) {
   return conn != NULL && conn->active && rgt_channel_isOpen(conn->channel);
}

CFL_BOOL rgt_app_conn_isUpdateBackground(RGT_APP_CONNECTIONP conn) {
   return conn != NULL && conn->updateTerminalThread != NULL;
}

void rgt_app_conn_setCursorPos(RGT_APP_CONNECTIONP conn, int iRow, int iCol) {
   RGT_LOG_ENTER("rgt_app_conn_setCursorPos", (NULL));
   if (conn != NULL) {
      rgt_screen_setCursorPos(conn->screen, iRow, iCol);
   } else {
      RGT_LOG_ERROR(("rgt_app_conn_setCursorPos(): connection argument is null"));
   }
   RGT_LOG_EXIT("rgt_app_conn_setCursorPos", (NULL));
}

void rgt_app_conn_setCursorStyle(RGT_APP_CONNECTIONP conn, int iStyle) {
   RGT_LOG_ENTER("rgt_app_conn_setCursorStyle", (NULL));
   if (conn != NULL) {
      rgt_screen_setCursorStyle(conn->screen, iStyle);
   } else {
      RGT_LOG_ERROR(("rgt_app_conn_setCursorStyle(): connection argument is null"));
   }
   RGT_LOG_EXIT("rgt_app_conn_setCursorStyle", (NULL));
}

void rgt_app_conn_updateTerminal(RGT_APP_CONNECTIONP conn) {
   RGT_LOG_ENTER("rgt_app_conn_updateTerminal", (NULL));
   if (conn != NULL) {
      updateTerminal(conn, conn->buffer);
   } else {
      RGT_LOG_ERROR(("rgt_app_conn_updateTerminal(): connection argument is null"));
   }
   RGT_LOG_EXIT("rgt_app_conn_updateTerminal", (NULL));
}

static CFL_BOOL readPendingKeys(RGT_APP_CONNECTIONP conn) {
   CFL_BOOL readKeys = CFL_FALSE;

   RGT_LOG_ENTER("readPendingKeys", (NULL));
   while (rgt_app_conn_isActive(conn) && rgt_channel_tryRead(conn->channel, conn->buffer)) {
      CFL_UINT16 errCode = cfl_buffer_getUInt16(conn->buffer);
      switch (errCode) {
      case RGT_RESP_TRM_KEY_UPDATE:
         readKeys = bufferToAvailableKeys(conn) || readKeys;
         break;
      case RGT_RESP_TRM_KEEP_ALIVE:
         cfl_buffer_getUInt8(conn->buffer);
         break;
      default: {
         char *msgError = cfl_buffer_getCharArray(conn->buffer);
         rgt_error_set(RGT_APP, RGT_ERROR_UNKNOWN_CMD, "Error reading keys %d - %s", errCode, msgError);
         CFL_MEM_FREE(msgError);
         RGT_LOG_EXIT("readPendingKeys", (NULL));
         return readKeys;
      }
      }
   }
   RGT_LOG_EXIT("readPendingKeys", (NULL));
   return readKeys;
}

int rgt_app_conn_getKey(RGT_APP_CONNECTIONP conn) {
   RGT_LOG_ENTER("rgt_app_conn_getKey", (NULL));
   if (conn == NULL) {
      RGT_LOG_ERROR(("rgt_app_conn_getKey(): connection argument is null"));
      RGT_LOG_EXIT("rgt_app_conn_getKey", (NULL));
      return 0;
   }
   if (cfl_buffer_remaining(conn->availableKeysBuffer) > 0 || readPendingKeys(conn)) {
      int key = (int)cfl_buffer_getInt32(conn->availableKeysBuffer);
      RGT_LOG_EXIT("rgt_app_conn_getKey", (NULL));
      return key;
   }
   RGT_LOG_EXIT("rgt_app_conn_getKey", (NULL));
   return 0;
}

static CFL_INT32 sendFileContent(RGT_APP_CONNECTIONP conn, const char *fileToSent, const char *remotePahName,
                                 CFL_UINT32 chunkSize) {
   CFL_UINT16 errCode;
   CFL_INT32 result = 0;
   HB_FHANDLE fileHandle;
   HB_SIZE fileSize;

   RGT_LOG_ENTER("sendFileContent", (NULL));
   if (conn == NULL) {
      RGT_LOG_ERROR(("sendFileContent(): connection argument is null"));
      RGT_LOG_EXIT("sendFileContent", (NULL));
      return 0;
   }

   fileHandle = hb_fsOpen(fileToSent, FO_READ);
   if (fileHandle == FS_ERROR) {
      RGT_LOG_EXIT("sendFileContent", (" Error: %d", hb_fsGetFError()));
      return hb_fsGetFError();
   }
   hb_fsSeek(fileHandle, 0L, FS_END);
   fileSize = (HB_SIZE)hb_fsTell(fileHandle);
   hb_fsSeek(fileHandle, 0L, FS_SET);
   /* File size */
   RGT_LOG_DEBUG(("sendFileContent. file size=%lu block size=%lu", fileSize, chunkSize));
   rgt_common_prepareCommand(conn->buffer, RGT_APP_CMD_PUT_FILE);
   /* File path and name */
   cfl_buffer_putCharArray(conn->buffer, remotePahName);
   cfl_buffer_putUInt32(conn->buffer, (CFL_UINT32)fileSize);
   if (fileSize > 0) {
      char *readBuffer = (char *)RGT_HB_ALLOC(sizeof(char) * chunkSize);
      HB_SIZE nextReadSize = (HB_SIZE)chunkSize;
      while (fileSize > 0 && result == 0) {
         HB_SIZE readCount;
         /* Send data block of N Kbytes */
         if (nextReadSize > fileSize) {
            nextReadSize = fileSize;
         }
         readCount = hb_fsReadLarge(fileHandle, readBuffer, nextReadSize);
         if (readCount == nextReadSize) {
            fileSize -= readCount;
            cfl_buffer_putUInt32(conn->buffer, (CFL_UINT32)nextReadSize);
            cfl_buffer_put(conn->buffer, readBuffer, (CFL_UINT32)nextReadSize);
            if (rgt_app_conn_isActive(conn) && rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->timeout)) {
               errCode = readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->timeout);
               if (errCode != RGT_RESP_SUCCESS) {
                  if (errCode == RGT_ERROR_FILE_SYSTEM) {
                     result = 1000 + cfl_buffer_getInt32(conn->buffer);
                  } else {
                     rgt_common_errorFromServer(RGT_APP, errCode, conn->buffer, "Error sending file");
                     result = RGT_ERROR_APP;
                  }
               }
            } else {
               RGT_LOG_DEBUG(("sendFileContent(): error communicating with TE"));
               result = RGT_ERROR_APP;
            }
            rgt_common_prepareCommand(conn->buffer, RGT_APP_CMD_PUT_FILE);
         } else {
            result = hb_fsGetFError();
         }
      }
      RGT_HB_FREE(readBuffer);

      /* Empty file */
   } else if (rgt_app_conn_isActive(conn) && rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->timeout)) {
      errCode = readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->timeout);
      if (errCode != RGT_RESP_SUCCESS) {
         if (errCode == RGT_ERROR_FILE_SYSTEM) {
            result = 1000 + cfl_buffer_getInt32(conn->buffer);
         } else {
            rgt_common_errorFromServer(RGT_APP, errCode, conn->buffer, "Error sending file");
            result = RGT_ERROR_APP;
         }
      }
   } else {
      RGT_LOG_DEBUG(("sendFileContent(): error communicating with TE"));
      result = RGT_ERROR_APP;
   }
   hb_fsClose(fileHandle);
   RGT_LOG_EXIT("sendFileContent", (NULL));
   return result;
}

static CFL_INT32 receiveFileContent(RGT_APP_CONNECTIONP conn, const char *localPathName) {
   CFL_UINT16 errCode;
   HB_FHANDLE fileHandle;
   CFL_UINT32 fileSize;
   CFL_UINT32 readBufferSize = 0;
   CFL_UINT8 *readBuffer = NULL;

   RGT_LOG_ENTER("receiveFileContent", (NULL));
   if (conn == NULL) {
      RGT_LOG_ERROR(("receiveFileContent(): connection argument is null"));
      RGT_LOG_EXIT("receiveFileContent", (NULL));
      return 0;
   }

   fileHandle = hb_fsCreate(localPathName, FC_NORMAL);
   if (fileHandle == FS_ERROR) {
      RGT_LOG_EXIT("receiveFileContent", (NULL));
      return hb_fsGetFError();
   }
   /* Tamanho do arquivo*/
   fileSize = cfl_buffer_getUInt32(conn->buffer);
   RGT_LOG_DEBUG(("receiveFileContent. file size=%lu", fileSize));
   while (fileSize > 0) {
      CFL_UINT32 writeCount;
      CFL_UINT32 nextReadSize = cfl_buffer_getUInt32(conn->buffer);
      if (nextReadSize > readBufferSize) {
         if (readBuffer == NULL) {
            readBuffer = (CFL_UINT8 *)RGT_HB_ALLOC(nextReadSize * sizeof(CFL_UINT8));
         } else {
            readBuffer = (CFL_UINT8 *)RGT_HB_REALLOC(readBuffer, nextReadSize * sizeof(CFL_UINT8));
         }
         readBufferSize = nextReadSize;
      }
      cfl_buffer_copy(conn->buffer, readBuffer, nextReadSize);
      fileSize -= nextReadSize;
      writeCount = (CFL_UINT32)hb_fsWrite(fileHandle, readBuffer, nextReadSize);
      if (writeCount != nextReadSize) {
         RGT_HB_FREE(readBuffer);
         hb_fsClose(fileHandle);
         RGT_LOG_EXIT("receiveFileContent", (NULL));
         return hb_fsGetFError();
      }
      if (fileSize > 0) {
         rgt_common_prepareCommand(conn->buffer, RGT_APP_CMD_GET_FILE);
         if (rgt_app_conn_isActive(conn) && rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->timeout)) {
            errCode = readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->timeout);
            if (errCode != RGT_RESP_SUCCESS) {
               RGT_HB_FREE(readBuffer);
               hb_fsClose(fileHandle);
               if (errCode == RGT_ERROR_FILE_SYSTEM) {
                  RGT_LOG_EXIT("receiveFileContent", (NULL));
                  return 1000 + cfl_buffer_getInt32(conn->buffer);
               } else {
                  rgt_common_errorFromServer(RGT_APP, errCode, conn->buffer, "Error receiving file");
                  RGT_LOG_EXIT("receiveFileContent", (NULL));
                  return RGT_ERROR_APP;
               }
            }
         } else {
            RGT_HB_FREE(readBuffer);
            hb_fsClose(fileHandle);
            RGT_LOG_DEBUG(("receiveFileContent(): error communicating with TE"));
            RGT_LOG_EXIT("receiveFileContent", (NULL));
            return RGT_ERROR_APP;
         }
      }
   }
   if (readBuffer != NULL) {
      RGT_HB_FREE(readBuffer);
   }
   hb_fsClose(fileHandle);
   RGT_LOG_EXIT("receiveFileContent", (NULL));
   return 0;
}

CFL_INT32 rgt_app_conn_putFile(RGT_APP_CONNECTIONP conn, const char *fileToSent, const char *remotePahName, CFL_UINT32 chunkSize) {
   CFL_INT32 result;

   RGT_LOG_ENTER("rgt_app_conn_putFile", ("%s -> %s.", fileToSent, remotePahName));
   result = sendFileContent(conn, fileToSent, remotePahName, chunkSize > 0 ? chunkSize : conn->fileTransferChunkSize);
   RGT_LOG_EXIT("rgt_app_conn_putFile", (NULL));
   return result;
}

CFL_INT32 rgt_app_conn_getFile(RGT_APP_CONNECTIONP conn, const char *fileToReceive, const char *localPathName,
                               CFL_UINT32 chunkSize) {
   CFL_INT32 result;

   RGT_LOG_ENTER("rgt_app_conn_getFile", ("%s <- %s.", localPathName, fileToReceive));
   if (conn == NULL) {
      RGT_LOG_ERROR(("rgt_app_conn_getFile(): connection argument is null"));
      RGT_LOG_EXIT("rgt_app_conn_getFile", (NULL));
      return 0;
   }
   rgt_common_prepareCommand(conn->buffer, RGT_APP_CMD_GET_FILE);
   /* File path and name */
   cfl_buffer_putCharArray(conn->buffer, fileToReceive);
   cfl_buffer_putUInt32(conn->buffer, chunkSize > 0 ? chunkSize : conn->fileTransferChunkSize);
   if (rgt_app_conn_isActive(conn) && rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->timeout)) {
      CFL_UINT16 errCode = readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->timeout);
      if (errCode == RGT_RESP_SUCCESS) {
         result = receiveFileContent(conn, localPathName);
      } else if (errCode == RGT_ERROR_FILE_SYSTEM) {
         result = 1000 + cfl_buffer_getInt32(conn->buffer);
      } else {
         rgt_common_errorFromServer(RGT_APP, errCode, conn->buffer, "Error receiving file");
         result = RGT_ERROR_APP;
      }
   } else {
      RGT_LOG_DEBUG(("rgt_app_conn_getFile(): error communicating with TE"));
      result = RGT_ERROR_APP;
   }
   RGT_LOG_EXIT("rgt_app_conn_getFile", (NULL));
   return result;
}

PHB_ITEM rgt_app_conn_execRemoteFunction(RGT_APP_CONNECTIONP conn, const char *functionName, CFL_INT16 argsCount, PHB_ITEM *pArgs) {
   CFL_INT16 iParam;
   PHB_ITEM pReturn = NULL;
   CFL_BOOL bSuccess = CFL_TRUE;

   RGT_LOG_ENTER("rgt_app_conn_execRemoteFunction", ("function=%s args=%d", functionName, argsCount));
   if (conn == NULL) {
      RGT_LOG_ERROR(("rgt_app_conn_execRemoteFunction(): connection argument is null"));
      RGT_LOG_EXIT("rgt_app_conn_execRemoteFunction", (NULL));
      return NULL;
   }
   rgt_common_prepareCommand(conn->buffer, RGT_APP_CMD_RPC);
   cfl_buffer_putBoolean(conn->buffer, CFL_TRUE);
   if (!rgt_screen_toBuffer(conn->screen, conn->buffer, CFL_FALSE)) {
      cfl_buffer_skip(conn->buffer, -1);
      cfl_buffer_putBoolean(conn->buffer, CFL_FALSE);
   }
   copyTonesToBuffer(conn, conn->buffer);
   cfl_buffer_putCharArray(conn->buffer, functionName);
   for (iParam = 0; iParam < argsCount && bSuccess; iParam++) {
      RGT_LOG_PARAM(RGT_LOG_LEVEL_TRACE, iParam + 1, pArgs[iParam]);
      if (!rgt_rpc_putItem(conn->buffer, pArgs[iParam])) {
         rgt_error_set(RGT_APP, RGT_ERROR_RPC_INVALID_PAR_TYPE, "Invalid or unsupported data type for parameter %d", iParam + 1);
         bSuccess = CFL_FALSE;
      }
   }
   if (bSuccess) {
      cfl_buffer_putUInt8(conn->buffer, RGT_RPC_PAR_END);
      if (rgt_app_conn_isActive(conn) && rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->rpcTimeout)) {
         CFL_UINT16 errCode;
         conn->lastTerminalUpdate = CURRENT_TIME;
         errCode = readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->rpcTimeout);
         if (errCode == RGT_RESP_SUCCESS) {
            CFL_UINT8 parType = cfl_buffer_getUInt8(conn->buffer);
            pReturn = rgt_rpc_getItem(conn->buffer, parType, NULL);
            RGT_LOG_PARAM(RGT_LOG_LEVEL_TRACE, 0, pReturn);
            if (pReturn == NULL) {
               rgt_error_set(RGT_APP, RGT_ERROR_RPC_INVALID_DATA_TYPE, "Invalid or unsupported data type returned");
            }
         } else {
            rgt_common_errorFromServer(RGT_APP, errCode, conn->buffer, "Error executing function");
         }
      } else {
         RGT_LOG_DEBUG(("rgt_app_conn_execRemoteFunction(): error communicating with TE"));
      }
   }
   RGT_LOG_EXIT("rgt_app_conn_execRemoteFunction", ("function=%s args=%d", functionName, argsCount));
   return pReturn;
}

void rgt_app_conn_tone(RGT_APP_CONNECTIONP conn, double dFrequency, double dDuration) {
   RGT_LOG_ENTER("rgt_app_conn_tone", (NULL));
   if (conn != NULL) {
      RGT_LOCK_ACQUIRE(conn->toneBufferLocked);
      cfl_buffer_putDouble(conn->toneBuffer, dFrequency);
      cfl_buffer_putDouble(conn->toneBuffer, dDuration);
      conn->tonesCount++;
      RGT_LOCK_RELEASE(conn->toneBufferLocked);
   } else {
      RGT_LOG_ERROR(("rgt_app_conn_tone(): connection argument is null"));
   }
   RGT_LOG_EXIT("rgt_app_conn_tone", (NULL));
}

CFL_BOOL rgt_app_setSessionOption(RGT_APP_CONNECTIONP conn, const char *optionName, const char *optionValue) {
   CFL_BOOL bSuccess = CFL_FALSE;
   RGT_LOG_ENTER("rgt_app_setSessionOption", ("option=%s, value=%s", optionName, optionValue));
   if (rgt_app_conn_isActive(conn)) {
      rgt_common_prepareCommand(conn->buffer, RGT_APP_SESSION_CONFIG);
      cfl_buffer_putInt32(conn->buffer, 1);
      cfl_buffer_putCharArray(conn->buffer, optionName);
      cfl_buffer_putCharArray(conn->buffer, optionValue);
      if (rgt_channel_writeAndRead(conn->channel, conn->buffer, conn->timeout)) {
         CFL_UINT16 errCode = readBackgroundCommands(conn, cfl_buffer_getUInt16(conn->buffer), conn->timeout);
         bSuccess = (errCode == RGT_RESP_SUCCESS ? CFL_TRUE : CFL_FALSE);
      } else {
         RGT_LOG_DEBUG(("rgt_app_setSessionOption(): Error communicating with TE."));
      }
   }
   RGT_LOG_EXIT("rgt_app_setSessionOption", (NULL));
   return bSuccess;
}

void rgt_app_conn_setFileTransferChunkSize(RGT_APP_CONNECTIONP conn, CFL_UINT32 chunkSize) {
   RGT_LOG_ENTER("rgt_app_conn_setFileTransferChunkSize", (NULL));
   if (chunkSize > 0) {
      conn->fileTransferChunkSize = chunkSize;
   }
   RGT_LOG_EXIT("rgt_app_conn_setFileTransferChunkSize", (NULL));
}

CFL_UINT32 rgt_app_conn_getFileTransferChunkSize(RGT_APP_CONNECTIONP conn) {
   CFL_UINT32 chunkSize;
   RGT_LOG_ENTER("rgt_app_conn_getFileTransferChunkSize", (NULL));
   chunkSize = conn->fileTransferChunkSize;
   RGT_LOG_EXIT("rgt_app_conn_getFileTransferChunkSize", (NULL));
   return chunkSize;
}

CFL_BOOL rgt_app_conn_isRpcExecuteLocalLostConnection(RGT_APP_CONNECTIONP conn) {
   CFL_BOOL execLocal;
   RGT_LOG_ENTER("rgt_app_conn_isRpcExecuteLocalLostConnection", (NULL));
   execLocal = conn->rpcExecuteLocalLostConnection;
   RGT_LOG_EXIT("rgt_app_conn_isRpcExecuteLocalLostConnection", (NULL));
   return execLocal;
}

void rgt_app_conn_setRpcExecuteLocalLostConnection(RGT_APP_CONNECTIONP conn, CFL_BOOL execLocal) {
   RGT_LOG_ENTER("rgt_app_conn_setRpcExecuteLocalLostConnection", (NULL));
   conn->rpcExecuteLocalLostConnection = execLocal;
   RGT_LOG_EXIT("rgt_app_conn_setRpcExecuteLocalLostConnection", (NULL));
}
