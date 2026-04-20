#include <stdio.h>
#include <stdlib.h>

#include "hbapi.h"
#include "hbapierr.h"
#include "hbapifs.h"
#include "hbapigt.h"
#include "hbapiitm.h"
#include "hbset.h"
#include "hbvm.h"

#include "hbapicdp.h"

#include "cfl_buffer.h"
#include "cfl_mem.h"
#include "cfl_socket.h"
#include "cfl_str.h"
#include "cfl_sync_queue.h"

#include "rgt_channel.h"
#include "rgt_common.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_rpc.h"
#include "rgt_screen.h"
#include "rgt_thread.h"
#include "rgt_trm_connection.h"
#include "rgt_util.h"

#define DEFAULT_SENDKEYS_INTERVAL 300
#define DEFAULT_TIMEOUT (30 * MINUTE)
#define DEFAULT_BUFFER_SIZE 32 * 1024
#define DEFAULT_TIMEOUT_WITHOUT_APP_COMMUNICATION 0 // (5 * MINUTE)
#define WAIT_DATA_TIMEOUT (1 * SECOND)
#define WAIT_THREAD_TIMEOUT (10 * SECOND)
#define WAIT_KEY_TIMEOUT 20
#define DEFAULT_MAX_SEND_KEYS 3

#define DEFAULT_APP_QUEUE_SIZE 512
#define DEFAULT_RESPONSE_QUEUE_SIZE 512
#define DEFAULT_SERVER_QUEUE_SIZE 16

#define KEY_CODE(c, k) ((CFL_INT32)((c)->appClpCompiler == RGT_CLP_COMP_XHARBOUR ? hb_inkeyKeyXHB(k) : k))
#define READ_KEY(c) filterKey(HB_GTSELF_READKEY((c)->pGT, s_keyMask), s_keyMask)

#define KEY_SIZE sizeof(CFL_INT32)
#define KEYS_IN_BUFFER(b) ((cfl_buffer_length(b) - RGT_RESPONSE_HEADER_SIZE) / KEY_SIZE)
#define ENOUGH_KEYS_IN_BUFFER(b) (KEYS_IN_BUFFER(b) >= s_maxSendKeys)

#define IS_ALIVE(conn) ((conn)->isActive && hb_vmIsActive() && !hb_vmQuitRequest() && !rgt_error_hasError())

#define IS_TIMEOUT_APP_COMMUNICATION(c, t) ((c)->timeoutAppCommunication > 0 && (t) > (c)->timeoutAppCommunication)

static int s_keyMask = INKEY_KEYBOARD;
static RGT_TRM_CONNECTIONP s_connection = NULL;
static CFL_INT64 s_sessionId = 0;
static CFL_STRP s_logoutMessage = NULL;
static CFL_UINT32 s_defaultIntervalSendKeys = DEFAULT_SENDKEYS_INTERVAL;
static CFL_UINT32 s_maxSendKeys = DEFAULT_MAX_SEND_KEYS;

#define XHB_K_SH_LEFT 424  /* Shift-Left  */
#define XHB_K_SH_UP 425    /* Shift-Up    */
#define XHB_K_SH_RIGHT 426 /* Shift-Right */
#define XHB_K_SH_DOWN 427  /* Shift-Down  */

#define XHB_K_SH_INS 428  /* Shift-Ins   */
#define XHB_K_SH_DEL 429  /* Shift-Del   */
#define XHB_K_SH_HOME 430 /* Shift-Home  */
#define XHB_K_SH_END 431  /* Shift-End   */
#define XHB_K_SH_PGUP 432 /* Shift-PgUp  */
#define XHB_K_SH_PGDN 433 /* Shift-PgDn  */

#define XHB_K_SH_RETURN 434 /* Shift-Enter */
#define XHB_K_SH_ENTER 434  /* Shift-Enter */

static int hb_inkeyKeyXHB(int iKey) {
   if (HB_INKEY_ISEXT(iKey)) {
      int iFlags = HB_INKEY_FLAGS(iKey), iValue = HB_INKEY_VALUE(iKey);

      if (HB_INKEY_ISKEY(iKey)) {
         if ((iFlags & (HB_KF_SHIFT | HB_KF_CTRL | HB_KF_ALT)) == HB_KF_SHIFT && iValue >= 0 && iValue < 32) {
            switch (iValue) {
            case HB_KX_LEFT:
               return XHB_K_SH_LEFT;
            case HB_KX_UP:
               return XHB_K_SH_UP;
            case HB_KX_RIGHT:
               return XHB_K_SH_RIGHT;
            case HB_KX_DOWN:
               return XHB_K_SH_DOWN;
            case HB_KX_INS:
               return XHB_K_SH_INS;
            case HB_KX_DEL:
               return XHB_K_SH_DEL;
            case HB_KX_HOME:
               return XHB_K_SH_HOME;
            case HB_KX_END:
               return XHB_K_SH_END;
            case HB_KX_PGUP:
               return XHB_K_SH_PGUP;
            case HB_KX_PGDN:
               return XHB_K_SH_PGDN;
            case HB_KX_ENTER:
               return XHB_K_SH_ENTER;
            }
         }
      }
      if (HB_INKEY_ISKEY(iKey) || HB_INKEY_ISCHAR(iKey) || HB_INKEY_ISUNICODE(iKey)) {
         if ((iFlags & (HB_KF_CTRL | HB_KF_ALT)) == HB_KF_CTRL) {
            if (iValue >= 'A' && iValue <= 'Z')
               return 512 + (iValue - 'A') + 1;
            else if (iValue >= 'a' && iValue <= 'z')
               return 512 + (iValue - 'a') + 1;
         }
      }
   }
   return hb_inkeyKeyStd(iKey);
}

static int filterKey(int iKey, int iEventMask) {
   int iMask;

   if (HB_INKEY_ISEXT(iKey)) {
      if (HB_INKEY_ISEVENT(iKey))
         iMask = HB_INKEY_GTEVENT;
      else if (HB_INKEY_ISMOUSEPOS(iKey))
         iMask = INKEY_MOVE;
      else if (HB_INKEY_ISMOUSEKEY(iKey)) {
         switch (HB_INKEY_VALUE(iKey)) {
         case K_MOUSEMOVE:
         case K_MMLEFTDOWN:
         case K_MMRIGHTDOWN:
         case K_MMMIDDLEDOWN:
         case K_NCMOUSEMOVE:
            iMask = INKEY_MOVE;
            break;
         case K_LBUTTONDOWN:
         case K_LDBLCLK:
            iMask = INKEY_LDOWN;
            break;
         case K_LBUTTONUP:
            iMask = INKEY_LUP;
            break;
         case K_RBUTTONDOWN:
         case K_RDBLCLK:
            iMask = INKEY_RDOWN;
            break;
         case K_RBUTTONUP:
            iMask = INKEY_RUP;
            break;
         case K_MBUTTONDOWN:
         case K_MBUTTONUP:
         case K_MDBLCLK:
            iMask = INKEY_MMIDDLE;
            break;
         case K_MWFORWARD:
         case K_MWBACKWARD:
            iMask = INKEY_MWHEEL;
            break;
         default:
            iMask = INKEY_KEYBOARD;
         }
      } else
         iMask = INKEY_KEYBOARD;
   } else {
      switch (iKey) {
      case K_MOUSEMOVE:
      case K_MMLEFTDOWN:
      case K_MMRIGHTDOWN:
      case K_MMMIDDLEDOWN:
      case K_NCMOUSEMOVE:
         iMask = INKEY_MOVE;
         break;
      case K_LBUTTONDOWN:
      case K_LDBLCLK:
         iMask = INKEY_LDOWN;
         break;
      case K_LBUTTONUP:
         iMask = INKEY_LUP;
         break;
      case K_RBUTTONDOWN:
      case K_RDBLCLK:
         iMask = INKEY_RDOWN;
         break;
      case K_RBUTTONUP:
         iMask = INKEY_RUP;
         break;
      case K_MBUTTONDOWN:
      case K_MBUTTONUP:
      case K_MDBLCLK:
         iMask = INKEY_MMIDDLE;
         break;
      case K_MWFORWARD:
      case K_MWBACKWARD:
         iMask = INKEY_MWHEEL;
         break;
      case HB_K_RESIZE:
      case HB_K_CLOSE:
      case HB_K_GOTFOCUS:
      case HB_K_LOSTFOCUS:
      case HB_K_CONNECT:
      case HB_K_DISCONNECT:
      case HB_K_TERMINATE:
      case HB_K_MENU:
         iMask = HB_INKEY_GTEVENT;
         break;
      default:
         iMask = INKEY_KEYBOARD;
         break;
      }
   }

   if ((iMask & iEventMask) == 0)
      return 0;

   if (HB_INKEY_ISEXT(iKey) && (iEventMask & HB_INKEY_EXT) == 0)
      iKey = hb_inkeyKeyStd(iKey);

   return iKey;
}

static RGT_TRM_CONNECTIONP createConnection(const char *server, CFL_UINT16 port, const char *username, RGT_CHANNELP channel) {
   RGT_TRM_CONNECTIONP conn = (RGT_TRM_CONNECTIONP)RGT_HB_ALLOC(sizeof(RGT_TRM_CONNECTION));
   conn->server = cfl_str_newBuffer(server);
   conn->port = port;
   conn->username = cfl_str_newBuffer(username);
   conn->channel = channel;
   conn->appRequestsQueue = cfl_sync_queue_new(DEFAULT_APP_QUEUE_SIZE);
   conn->responseQueue = cfl_sync_queue_new(DEFAULT_RESPONSE_QUEUE_SIZE);
   conn->serverRequestsQueue = cfl_sync_queue_new(DEFAULT_SERVER_QUEUE_SIZE);
   conn->appClpCompiler = RGT_CLP_COMP_UNKNOWN;
   conn->screen = NULL;
   conn->timeout = DEFAULT_TIMEOUT;
   conn->timeoutAppCommunication = DEFAULT_TIMEOUT_WITHOUT_APP_COMMUNICATION;
   conn->lastTimeReceivedAppData = CURRENT_TIME;
   conn->lastTimeSentDataToApp = CURRENT_TIME;
   conn->lastTimeSentKeysToApp = CURRENT_TIME;
   conn->sendKeysInterval = s_defaultIntervalSendKeys;
   conn->isActive = CFL_TRUE;
   conn->pGT = hb_stackGetGT();
   conn->keyBuffer = cfl_buffer_newCapacity(RGT_KEY_BUFFER_SIZE);
   rgt_common_prepareResponse(conn->keyBuffer, RGT_RESP_TRM_KEY_UPDATE);
   return conn;
}

static void releaseConnection(RGT_TRM_CONNECTIONP conn) {
   if (conn) {
      cfl_str_free(conn->server);
      cfl_str_free(conn->username);
      cfl_sync_queue_free(conn->appRequestsQueue);
      cfl_sync_queue_free(conn->serverRequestsQueue);
      cfl_sync_queue_free(conn->responseQueue);
      if (conn->screen) {
         rgt_screen_free(conn->screen);
      }
      cfl_buffer_free(conn->keyBuffer);
      RGT_HB_FREE(conn);
   }
}

static void putLocalAddress(CFL_BUFFERP buffer) {
   char host[256];
   cfl_socket_hostAddress(host, sizeof(host));
   cfl_buffer_putCharArray(buffer, host);
}

static void keepAliveCommand(CFL_BUFFERP buffer) {
   RGT_LOG_ENTER("keepAliveCommand", (NULL));
   // cfl_buffer_getUInt8(buffer);
   cfl_buffer_free(buffer);
   RGT_LOG_EXIT("keepAliveCommand", (NULL));
}

static void updateScreenFromCommunicationBuffer(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   RGT_LOG_ENTER("updateScreenFromCommunicationBuffer", (NULL));
   rgt_screen_fromBuffer(conn->screen, buffer);
   RGT_LOG_EXIT("updateScreenFromCommunicationBuffer", (NULL));
}

static void playTonesFromCommunicationBuffer(CFL_BUFFERP buffer) {
   CFL_UINT16 tones;
   CFL_UINT16 i;
   RGT_LOG_ENTER("playTonesFromCommunicationBuffer", (NULL));
   tones = cfl_buffer_getUInt16(buffer);
   RGT_LOG_DEBUG(("playTonesFromCommunicationBuffer() tones=%d", tones));
   for (i = 0; i < tones; i++) {
      double dFrequency = cfl_buffer_getDouble(buffer);
      double dDuration = cfl_buffer_getDouble(buffer);
      hb_gtTone(dFrequency, dDuration);
   }
   RGT_LOG_EXIT("playTonesFromCommunicationBuffer", (NULL));
}

static void updateCommand(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   RGT_LOG_ENTER("updateCommand", (NULL));
   updateScreenFromCommunicationBuffer(conn, buffer);
   playTonesFromCommunicationBuffer(buffer);
   cfl_buffer_free(buffer);
   RGT_LOG_EXIT("updateCommand", (NULL));
}

static void sendResponse(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   if (buffer == NULL) {
      return;
   }
   cfl_sync_queue_put(conn->responseQueue, buffer);
}

static CFL_BUFFERP sendKeysInBuffer(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP keyBuffer) {
   RGT_LOG_DEBUG(("sendKeysInBuffer. keys=%u", KEYS_IN_BUFFER(keyBuffer)));
   cfl_buffer_putInt32(keyBuffer, 0);
   sendResponse(conn, keyBuffer);
   conn->lastTimeSentDataToApp = conn->lastTimeSentKeysToApp = CURRENT_TIME;
   keyBuffer = cfl_buffer_newCapacity(RGT_KEY_BUFFER_SIZE);
   rgt_common_prepareResponse(keyBuffer, RGT_RESP_TRM_KEY_UPDATE);
   return keyBuffer;
}

static CFL_BUFFERP sendKeysToApp(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP keyBuffer) {
   int iKey;

   RGT_LOG_ENTER("sendKeysToApp", (NULL));
   do {
      iKey = READ_KEY(conn);
      if (iKey == 0) {
         if (ENOUGH_KEYS_IN_BUFFER(keyBuffer) ||
             (KEYS_IN_BUFFER(keyBuffer) > 0 &&
              TIMEMILLIS_ELAPSED(conn->lastTimeSentKeysToApp, CURRENT_TIME) >= conn->sendKeysInterval)) {
            keyBuffer = sendKeysInBuffer(conn, keyBuffer);
         }
         break;
      } else if (iKey == HB_BREAK_FLAG && hb_setGetCancel()) {
         conn->isActive = CFL_FALSE;
         RGT_LOG_EXIT("sendKeysToApp", (NULL));
         return keyBuffer;
      } else {
         cfl_buffer_putInt32(keyBuffer, KEY_CODE(conn, iKey));
         if (ENOUGH_KEYS_IN_BUFFER(keyBuffer) ||
             (KEYS_IN_BUFFER(keyBuffer) > 0 &&
              TIMEMILLIS_ELAPSED(conn->lastTimeSentKeysToApp, CURRENT_TIME) >= conn->sendKeysInterval)) {
            keyBuffer = sendKeysInBuffer(conn, keyBuffer);
            break;
         }
      }
   } while (conn->isActive);
   RGT_LOG_EXIT("sendKeysToApp", (NULL));
   return keyBuffer;
}

static CFL_BUFFERP waitData(RGT_TRM_CONNECTIONP conn, CFL_BOOL fSetEnv) {
   CFL_BOOL isTimeout;
   CFL_BUFFERP buffer = NULL;
   CFL_UINT32 sleepTime;

   RGT_LOG_ENTER("waitData", (NULL));
   sleepTime = conn->sendKeysInterval / 2;
   if (sleepTime < WAIT_KEY_TIMEOUT) {
      sleepTime = WAIT_KEY_TIMEOUT;
   }
   while (buffer == NULL && conn->isActive) {
      if (fSetEnv) {
         conn->keyBuffer = sendKeysToApp(conn, conn->keyBuffer);
      }
      buffer = cfl_sync_queue_getTimeout(conn->appRequestsQueue, sleepTime, &isTimeout);
      if (buffer == NULL && IS_TIMEOUT_APP_COMMUNICATION(conn, TIMEMILLIS_ELAPSED(conn->lastTimeReceivedAppData, CURRENT_TIME))) {
         printf("\nTimeout without app communication");
         rgt_error_set(RGT_TERMINAL, RGT_ERROR_TIMEOUT, "Timeout without app communication");
         RGT_LOG_EXIT("waitData", (NULL));
         return NULL;
      }
   }
   RGT_LOG_EXIT("waitData", ("buffer=%p", buffer));
   return buffer;
}

static CFL_BUFFERP sendRespReceiveCmd(RGT_TRM_CONNECTIONP conn, CFL_INT32 timeout, CFL_INT8 cmdWaiting, CFL_BUFFERP buffer) {
   CFL_BOOL isTimeout;

   sendResponse(conn, buffer);
   conn->lastTimeSentDataToApp = CURRENT_TIME;
   buffer = (CFL_BUFFERP)cfl_sync_queue_getTimeout(conn->appRequestsQueue, timeout, &isTimeout);
   if (buffer != NULL) {
      do {
         CFL_UINT8 cmd = cfl_buffer_getUInt8(buffer);
         if (cmd == RGT_APP_CMD_UPDATE) {
            updateCommand(conn, buffer);
            buffer = (CFL_BUFFERP)cfl_sync_queue_getTimeout(conn->appRequestsQueue, timeout, &isTimeout);
         } else if (cmd == cmdWaiting) {
            return buffer;
         } else {
            cfl_buffer_free(buffer);
            rgt_common_prepareResponse(buffer, RGT_ERROR_PROTOCOL);
            cfl_buffer_putCharArray(buffer, "Unexpected data received");
            return NULL;
         }
      } while (buffer != NULL);
   }
   return NULL;
}

RGT_TRM_CONNECTIONP rgt_trm_login(const char *server, CFL_UINT16 port, const char *commandLine, const char *workDir,
                                  const char *username, const char *password, CFL_UINT16 argc, const char *argv[]) {
   RGT_TRM_CONNECTIONP conn;
   RGT_CHANNELP channel;
   CFL_UINT16 respCode;
   const char *osUser;
   int i;
   CFL_BUFFERP buffer;
   int logLevel;
   char *logPathName;
   CFL_BOOL bTimeout;

   RGT_LOG_TRACE(("rgt_trm_login server=%s, port=%d, command line=%s, work dir=%s", server, port, commandLine, workDir));
   // connect to server
   channel = rgt_channel_open(RGT_TERMINAL, server, port);

   if (channel == NULL) {
      rgt_error_set(RGT_TERMINAL, rgt_error_getLastCode(), "Unable to connect to Terminal Server.");
      return NULL;
   }
   conn = createConnection(server, port, username, channel);
   buffer = cfl_buffer_new();
   // send authentication and start up app infor
   rgt_common_prepareFirstCommand(buffer, RGT_TRM_CMD_LOGIN);
   cfl_buffer_putInt16(buffer, RGT_SERVER_PROTOCOL_VERSION);
   cfl_buffer_putCharArray(buffer, username);
   cfl_buffer_putCharArray(buffer, password);
   osUser = getenv("USERNAME");
   if (osUser != NULL) {
      cfl_buffer_putCharArray(buffer, osUser);
   } else {
      cfl_buffer_putCharArray(buffer, "");
   }
   putLocalAddress(buffer);
   cfl_buffer_putCharArray(buffer, workDir);
   cfl_buffer_putCharArray(buffer, commandLine);
   /* Put arguments in buffer */
   cfl_buffer_putInt8(buffer, (CFL_UINT8)argc);
   for (i = 0; i < argc; i++) {
      cfl_buffer_putCharArray(buffer, argv[i]);
   }
   if (!rgt_channel_writeAndReadFirstCommand(conn->channel, buffer, conn->timeout, &bTimeout)) {
      releaseConnection(conn);
      cfl_buffer_free(buffer);
      return NULL;
   }
   if (cfl_buffer_length(buffer) <= 1) {
      rgt_error_set(RGT_TERMINAL, RGT_ERROR_PROTOCOL, "Invalid response size from server.");
      releaseConnection(conn);
      cfl_buffer_free(buffer);
      return NULL;
   }
   respCode = cfl_buffer_getUInt16(buffer);
   if (respCode != RGT_RESP_SUCCESS) {
      rgt_common_errorFromServer(RGT_TERMINAL, respCode, buffer, NULL);
      releaseConnection(conn);
      cfl_buffer_free(buffer);
      return NULL;
   }
   conn->sessionId = cfl_buffer_getInt64(buffer);
   s_sessionId = conn->sessionId;
   logLevel = (int)cfl_buffer_getInt8(buffer);
   if (!rgt_log_envLevel()) {
      rgt_log_setLevel(logLevel);
   }
   logPathName = cfl_buffer_getCharArray(buffer);
   if (!rgt_log_envFile()) {
      rgt_log_setPathName(logPathName);
   }
   CFL_MEM_FREE(logPathName);
   cfl_buffer_free(buffer);
   return conn;
}

void rgt_trm_logout(void) {
   RGT_LOG_ENTER("rgt_trm_logout", (NULL));
   s_connection->isActive = CFL_FALSE;
   rgt_channel_close(s_connection->channel);
   rgt_channel_free(s_connection->channel);
   releaseConnection(s_connection);
   s_connection = NULL;
   if (s_logoutMessage != NULL) {
      if (!cfl_str_isEmpty(s_logoutMessage)) {
         rgt_error_set(RGT_TERMINAL, RGT_ERROR_UNKNOWN, cfl_str_getPtr(s_logoutMessage));
      }
      cfl_str_free(s_logoutMessage);
      s_logoutMessage = NULL;
   }
   RGT_LOG_EXIT("rgt_trm_logout", (NULL));
}

static CFL_BOOL receiveFileContent(RGT_TRM_CONNECTIONP conn, CFL_STRP localPathName, CFL_INT32 timeout, CFL_BUFFERP buffer) {
   HB_FHANDLE fileHandle;
   CFL_UINT32 fileSize;

   RGT_LOG_ENTER("receiveFileContent", (" local file: %s", cfl_str_getPtr(localPathName)));
   fileHandle = hb_fsCreate(cfl_str_getPtr(localPathName), FC_NORMAL);
   if (fileHandle == FS_ERROR) {
      rgt_common_prepareResponse(buffer, RGT_ERROR_FILE_SYSTEM);
      cfl_buffer_putInt32(buffer, hb_fsGetFError());
      RGT_LOG_EXIT("receiveFileContent", ("Error: %d", hb_fsGetFError()));
      return CFL_FALSE;
   }
   /* File size */
   fileSize = cfl_buffer_getUInt32(buffer);
   while (fileSize > 0) {
      CFL_UINT32 chunkSize = cfl_buffer_getUInt32(buffer);
      CFL_UINT32 writeCount = (CFL_UINT32)hb_fsWrite(fileHandle, cfl_buffer_positionPtr(buffer), (HB_USHORT)chunkSize);
      if (writeCount != chunkSize) {
         rgt_common_prepareResponse(buffer, RGT_ERROR_FILE_SYSTEM);
         cfl_buffer_putInt32(buffer, hb_fsGetFError());
         hb_fsClose(fileHandle);
         RGT_LOG_EXIT("receiveFileContent", (NULL));
         return CFL_FALSE;
      }
      fileSize -= chunkSize;
      if (fileSize > 0) {
         rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
         buffer = sendRespReceiveCmd(conn, timeout, RGT_APP_CMD_PUT_FILE, buffer);
         if (buffer == NULL) {
            hb_fsClose(fileHandle);
            RGT_LOG_EXIT("receiveFileContent", (NULL));
            return CFL_FALSE;
         }
      }
   }
   hb_fsClose(fileHandle);
   rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
   RGT_LOG_EXIT("receiveFileContent", (NULL));
   return CFL_TRUE;
}

static void putFileCommand(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   CFL_STRP localPathName;

   RGT_LOG_ENTER("putFileCommand", (NULL));
   localPathName = cfl_buffer_getString(buffer);
   receiveFileContent(conn, localPathName, conn->timeout, buffer);
   cfl_str_free(localPathName);
   RGT_LOG_EXIT("putFileCommand", (NULL));
}

static CFL_BOOL sendFileContent(RGT_TRM_CONNECTIONP conn, CFL_STRP localPathName, CFL_UINT32 chunkSize, CFL_INT32 timeout,
                                CFL_BUFFERP buffer) {
   HB_FHANDLE fileHandle;
   HB_SIZE fileSize;
   char *readBuffer;

   RGT_LOG_ENTER("sendFileContent", (NULL));
   fileHandle = hb_fsOpen(cfl_str_getPtr(localPathName), FO_READ);
   if (fileHandle == FS_ERROR) {
      rgt_common_prepareResponse(buffer, RGT_ERROR_FILE_SYSTEM);
      cfl_buffer_putInt32(buffer, hb_fsGetFError());
      RGT_LOG_EXIT("sendFileContent", (NULL));
      return CFL_FALSE;
   }
   hb_fsSeek(fileHandle, 0L, FS_END);
   fileSize = (HB_SIZE)hb_fsTell(fileHandle);
   hb_fsSeek(fileHandle, 0L, FS_SET);
   rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
   /* File size */
   cfl_buffer_putUInt32(buffer, (CFL_UINT32)fileSize);
   readBuffer = (char *)RGT_HB_ALLOC(chunkSize * sizeof(char));
   while (fileSize > 0) {
      HB_SIZE nextReadSize;
      HB_SIZE readCount;
      /* Send data block of N Kbytes */
      if (chunkSize < fileSize) {
         nextReadSize = (HB_SIZE)chunkSize;
      } else {
         nextReadSize = fileSize;
      }
      readCount = hb_fsReadLarge(fileHandle, readBuffer, nextReadSize);
      if (readCount != nextReadSize) {
         rgt_common_prepareResponse(buffer, RGT_ERROR_FILE_SYSTEM);
         cfl_buffer_putInt32(buffer, hb_fsGetFError());
         RGT_HB_FREE(readBuffer);
         hb_fsClose(fileHandle);
         RGT_LOG_EXIT("sendFileContent", (NULL));
         return CFL_FALSE;
      }
      cfl_buffer_putUInt32(buffer, (CFL_UINT32)nextReadSize);
      cfl_buffer_put(buffer, readBuffer, (CFL_UINT32)nextReadSize);
      fileSize -= nextReadSize;
      if (fileSize > 0) {
         buffer = sendRespReceiveCmd(conn, timeout, RGT_APP_CMD_GET_FILE, buffer);
         if (buffer != NULL) {
            rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
         } else {
            RGT_HB_FREE(readBuffer);
            hb_fsClose(fileHandle);
            RGT_LOG_EXIT("sendFileContent", (NULL));
            return CFL_FALSE;
         }
      }
   }
   RGT_HB_FREE(readBuffer);
   hb_fsClose(fileHandle);
   RGT_LOG_EXIT("sendFileContent", (NULL));
   return CFL_TRUE;
}

static void getFileCommand(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   CFL_STRP fileToSent;
   CFL_UINT32 chunkSize;

   RGT_LOG_ENTER("getFileCommand", (NULL));
   fileToSent = cfl_buffer_getString(buffer);
   chunkSize = cfl_buffer_getUInt32(buffer);
   sendFileContent(conn, fileToSent, chunkSize, conn->timeout, buffer);
   cfl_str_free(fileToSent);
   RGT_LOG_EXIT("getFileCommand", (NULL));
}

static void setKeyBufferLenCommand(CFL_BUFFERP buffer) {
   PHB_ITEM pResult = hb_itemNew(NULL);
   PHB_ITEM pBufferSize;

   RGT_LOG_ENTER("setKeyBufferLenCommand", (NULL));
   pBufferSize = hb_itemPutNI(NULL, cfl_buffer_getInt16(buffer));
   hb_setGetItem(HB_SET_TYPEAHEAD, pResult, pBufferSize, NULL);
   hb_itemRelease(pResult);
   hb_itemRelease(pBufferSize);
   rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
   RGT_LOG_EXIT("setKeyBufferLenCommand", (NULL));
}

static CFL_BOOL isSupportedProtocolVersion(CFL_INT16 protocolVersion) {
   return protocolVersion == RGT_TRM_APP_PROTOCOL_VERSION;
}

static void setEnvironmentCommand(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   CFL_INT16 protocolVersion;

   RGT_LOG_ENTER("setEnvironmentCommand", (NULL));
   protocolVersion = cfl_buffer_getInt16(buffer);
   if (isSupportedProtocolVersion(protocolVersion)) {
      PHB_ITEM pResult = hb_itemNew(NULL);
      char *szColorStr;
      CFL_INT16 iCols;
      CFL_INT16 iRows;
      CFL_INT8 screenType;
      PHB_ITEM pBufferSize;
      CFL_INT8 flags;

      flags = cfl_buffer_getInt8(buffer);
      screenType = flags & 0x0F;
      conn->appClpCompiler = (flags & 0xF0) >> 4;
      iRows = cfl_buffer_getInt16(buffer);
      iCols = cfl_buffer_getInt16(buffer);
      conn->screen = rgt_screen_new(iRows, iCols, screenType);
      hb_gtSetMode(iRows, iCols);
      rgt_screen_fromBuffer(conn->screen, buffer);
      szColorStr = cfl_buffer_getCharArray(buffer);
      hb_gtSetColorStr(szColorStr);
      CFL_MEM_FREE(szColorStr);
      pBufferSize = hb_itemPutNI(NULL, cfl_buffer_getInt16(buffer));
      hb_setGetItem(HB_SET_TYPEAHEAD, pResult, pBufferSize, NULL);
      hb_itemRelease(pResult);
      hb_itemRelease(pBufferSize);
      rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
   } else {
      rgt_common_prepareResponse(buffer, RGT_ERROR_PROTOCOL);
      cfl_buffer_putFormat(buffer, "Terminal protocol version %d is incompatible with application version %d",
                           RGT_TRM_APP_PROTOCOL_VERSION, protocolVersion);
      rgt_error_set(RGT_TERMINAL, RGT_ERROR_PROTOCOL, "Terminal protocol version %d is incompatible with application version %d",
                    RGT_TRM_APP_PROTOCOL_VERSION, protocolVersion);
   }
   RGT_LOG_EXIT("setEnvironmentCommand", (NULL));
}

static void executeFunctionCommand(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   RGT_LOG_ENTER("executeFunctionCommand", (NULL));
   if (cfl_buffer_getBoolean(buffer)) {
      updateScreenFromCommunicationBuffer(conn, buffer);
   }
   playTonesFromCommunicationBuffer(buffer);
   rgt_rpc_executeFunction(buffer);
   RGT_LOG_EXIT("executeFunctionCommand", (NULL));
}

static void logoutCommand(RGT_TRM_CONNECTIONP conn, CFL_BUFFERP buffer) {
   CFL_BOOL updateScreen;

   RGT_LOG_ENTER("logoutCommand", (NULL));
   updateScreen = cfl_buffer_getBoolean(buffer);
   if (updateScreen) {
      updateScreenFromCommunicationBuffer(conn, buffer);
   }
   playTonesFromCommunicationBuffer(buffer);
   s_logoutMessage = cfl_buffer_getString(buffer);
   if (cfl_str_isEmpty(s_logoutMessage)) {
      cfl_str_free(s_logoutMessage);
      s_logoutMessage = NULL;
   }
   rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
   RGT_LOG_EXIT("logoutCommand", (NULL));
}

static RGT_SCREENP sendScreen(RGT_SCREENP screen, RGT_TRM_CONNECTIONP conn, CFL_UINT64 clientId, CFL_BUFFERP buffer) {
   RGT_LOG_ENTER("sendScreen", (NULL));

   if (screen == NULL) {
      screen = rgt_screen_new(conn->screen->height, conn->screen->width, RGT_SCREEN_TYPE_EXTENDED);
      screen->pGT = conn->screen->pGT;
   } else if (screen->height != conn->screen->height || screen->width != conn->screen->width) {
      rgt_screen_reset(screen, conn->screen->height, conn->screen->width, RGT_SCREEN_TYPE_EXTENDED);
   }
   rgt_common_prepareAdminResponse(buffer, clientId, RGT_RESP_SUCCESS);
   rgt_screen_fullUpdated(screen);
   cfl_buffer_putInt8(buffer, screen->screenType);
   rgt_screen_fullToBuffer(screen, buffer);
   sendResponse(conn, buffer);
   RGT_LOG_EXIT("sendScreen", (NULL));
   return screen;
}

static void handleServerRequests(void *param) {
   RGT_TRM_CONNECTIONP conn = (RGT_TRM_CONNECTIONP)param;
   CFL_BUFFERP buffer;
   RGT_SCREENP screen = NULL;

   RGT_LOG_ENTER("handleServerRequests", (NULL));

   while (conn->isActive) {
      buffer = cfl_sync_queue_get(conn->serverRequestsQueue);
      if (buffer != NULL) {
         CFL_UINT8 op = cfl_buffer_getUInt8(buffer);
         CFL_UINT64 clientId = cfl_buffer_getUInt64(buffer);
         switch (op) {
         case RGT_ADMIN_GET_SCREEN:
            screen = sendScreen(screen, conn, clientId, buffer);
            break;
         }
      }
   }
   conn->isActive = CFL_FALSE;
   if (screen != NULL) {
      rgt_screen_free(screen);
   }

   RGT_LOG_EXIT("handleServerRequests", (NULL));
}

#define RGT_GET_BUFFER(buffer, size)                                                                                               \
   buffer = cfl_buffer_newCapacity(size);                                                                                          \
   if (buffer == NULL) {                                                                                                           \
      rgt_error_set(RGT_TERMINAL, RGT_ERROR_ALLOC_RESOURCE, "Error allocating buffer");                                            \
      return;                                                                                                                      \
   }

static void receiveRequests(void *param) {
   RGT_TRM_CONNECTIONP conn = (RGT_TRM_CONNECTIONP)param;
   CFL_BUFFERP buffer;
   CFL_BOOL bTimeout;

   RGT_LOG_ENTER("receiveRequests", (NULL));
   while (conn->isActive) {
      buffer = rgt_channel_readBuffer(conn->channel, WAIT_DATA_TIMEOUT, &bTimeout);
      if (buffer != NULL) {
         CFL_UINT8 op = cfl_buffer_getUInt8(buffer);
         RGT_LOG_DEBUG(("Request received: %#04X", op));
         cfl_buffer_rewind(buffer);
         switch (op) {
         case RGT_ADMIN_GET_SCREEN:
            cfl_sync_queue_put(conn->serverRequestsQueue, buffer);
            break;
         case RGT_APP_CMD_KEEP_ALIVE:
            conn->lastTimeReceivedAppData = CURRENT_TIME;
            keepAliveCommand(buffer);
            break;
         default:
            conn->lastTimeReceivedAppData = CURRENT_TIME;
            cfl_sync_queue_put(conn->appRequestsQueue, buffer);
            break;
         }
      } else if (!bTimeout) {
         break;
      }
   }
   conn->isActive = CFL_FALSE;
   cfl_sync_queue_cancel(conn->appRequestsQueue);
   cfl_sync_queue_cancel(conn->serverRequestsQueue);
   RGT_LOG_EXIT("receiveRequests", (NULL));
}

static void sendResponses(void *param) {
   RGT_TRM_CONNECTIONP conn = (RGT_TRM_CONNECTIONP)param;
   CFL_BUFFERP buffer;

   RGT_LOG_ENTER("sendResponses", (NULL));
   buffer = (CFL_BUFFERP)cfl_sync_queue_get(conn->responseQueue);
   while (buffer != NULL && rgt_channel_isOpen(conn->channel)) {
      rgt_channel_write(conn->channel, buffer);
      cfl_buffer_free(buffer);
      buffer = (CFL_BUFFERP)cfl_sync_queue_get(conn->responseQueue);
   }
   conn->isActive = CFL_FALSE;
   RGT_LOG_EXIT("sendResponses", (NULL));
}

static void waitThread(RGT_THREADP thread) {
   if (thread == NULL) {
      return;
   }
   RGT_LOG_ENTER("waitThread", ("%s", rgt_thread_getDescription(thread)));
   rgt_thread_waitTimeout(thread, WAIT_THREAD_TIMEOUT);
   if (rgt_thread_isRunning(thread)) {
      rgt_thread_kill(thread);
   }
   rgt_thread_free(thread);
   RGT_LOG_EXIT("waitThread", (NULL));
}

void rgt_trm_listenConnection(RGT_TRM_CONNECTIONP conn) {
   RGT_THREADP receiveRequestsThread;
   RGT_THREADP sendResponseThread;
   RGT_THREADP serverRequestsThread = NULL;
   CFL_BOOL fContinue = CFL_TRUE;
   CFL_BOOL fSetEnv = CFL_FALSE;
   CFL_BOOL fLogout = CFL_FALSE;
   HB_BOOL fDebug = hb_setGetDebug();
   PHB_ITEM pFlag = hb_itemPutL(NULL, HB_FALSE);
   const char *dbgEnvVar = getenv("RGT_DEBUG");
   CFL_UINT8 lastUnknownOp = RGT_CMD_UNKNOWN;

   RGT_LOG_ENTER("rgt_trm_listenConnection", ("interval update: %ld", conn->sendKeysInterval));

   if (dbgEnvVar != NULL && hb_stricmp(dbgEnvVar, "true") == 0) {
      hb_itemPutL(pFlag, HB_TRUE);
   }
   hb_setGetItem(HB_SET_DEBUG, NULL, pFlag, NULL);

   receiveRequestsThread = rgt_thread_start(receiveRequests, conn, "RGT Read Thread");
   sendResponseThread = rgt_thread_start(sendResponses, conn, "RGT Write APP Thread");

   while (fContinue) {
      CFL_BUFFERP buffer;
      buffer = waitData(conn, fSetEnv);
      if (buffer != NULL) {
         CFL_UINT8 op = cfl_buffer_getUInt8(buffer);
         switch (op) {
         case RGT_APP_CMD_UPDATE:
            updateCommand(conn, buffer);
            break;

         case RGT_APP_CMD_PUT_FILE:
            putFileCommand(conn, buffer);
            sendResponse(conn, buffer);
            conn->lastTimeSentDataToApp = CURRENT_TIME;
            break;

         case RGT_APP_CMD_GET_FILE:
            getFileCommand(conn, buffer);
            sendResponse(conn, buffer);
            conn->lastTimeSentDataToApp = CURRENT_TIME;
            break;

         case RGT_APP_CMD_KEY_BUF_LEN:
            setKeyBufferLenCommand(buffer);
            sendResponse(conn, buffer);
            conn->lastTimeSentDataToApp = CURRENT_TIME;
            break;

         case RGT_APP_CMD_SET_ENV:
            setEnvironmentCommand(conn, buffer);
            fContinue = rgt_channel_write(conn->channel, buffer);
            conn->lastTimeSentDataToApp = CURRENT_TIME;
            if (!fSetEnv && fContinue) {
               serverRequestsThread = rgt_thread_start(handleServerRequests, conn, "RGT Handle Requests");
            }
            fSetEnv = fContinue;
            cfl_buffer_free(buffer);
            break;

         case RGT_APP_CMD_RPC:
            executeFunctionCommand(conn, buffer);
            sendResponse(conn, buffer);
            conn->lastTimeSentDataToApp = CURRENT_TIME;
            break;

         case RGT_APP_CMD_LOGOUT:
            logoutCommand(conn, buffer);
            sendResponse(conn, buffer);
            conn->isActive = CFL_FALSE;
            conn->lastTimeSentDataToApp = CURRENT_TIME;
            fContinue = CFL_FALSE;
            fLogout = CFL_TRUE;
            break;

         default:
            rgt_common_prepareResponse(buffer, RGT_ERROR_PROTOCOL);
            cfl_buffer_putFormat(buffer, "Internal error. Unknown command %d", op);
            sendResponse(conn, buffer);
            conn->lastTimeSentDataToApp = CURRENT_TIME;
            if (lastUnknownOp != op) {
               lastUnknownOp = op;
               RGT_LOG_ERROR(("rgt_trm_listenConnection(): Internal error. Unknown command %d", op));
            }
            break;
         }
      } else {
         fContinue = CFL_FALSE;
      }
      fContinue = fContinue && IS_ALIVE(conn);
   }
   conn->isActive = CFL_FALSE;
   waitThread(receiveRequestsThread);
   cfl_sync_queue_cancel(conn->responseQueue);
   waitThread(sendResponseThread);
   waitThread(serverRequestsThread);
   hb_itemPutL(pFlag, fDebug);
   hb_setGetItem(HB_SET_DEBUG, NULL, pFlag, NULL);
   hb_itemRelease(pFlag);
   if (fLogout) {
      rgt_error_clear();
   }
   RGT_LOG_EXIT("rgt_trm_listenConnection", (NULL));
}

void rgt_trm_connSetTimeout(RGT_TRM_CONNECTIONP conn, CFL_UINT32 timeout) {
   conn->timeout = timeout;
}

CFL_UINT32 rgt_trm_connGetTimeout(RGT_TRM_CONNECTIONP conn) {
   return conn->timeout;
}

void rgt_trm_setMaxKeysInBuffer(CFL_UINT32 maxKeys) {
   s_maxSendKeys = maxKeys;
}

CFL_UINT32 rgt_trm_getMaxKeysInBuffer(void) {
   return s_maxSendKeys;
}

CFL_INT64 rgt_trm_sessionId(void) {
   return s_sessionId;
}

static void showInitScreen(void) {
   hb_gtSetMode(25, 80);
   hb_gtScroll(0, 0, 24, 79, 25, 80);
   hb_gtWriteAt(0, 0, "                                                                                ", 80);
   hb_gtWriteAt(1, 0, "           ,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,                 ", 80);
   hb_gtWriteAt(2, 0, "          ******************************************************#               ", 80);
   hb_gtWriteAt(3, 0, "          ***                                                %**#               ", 80);
   hb_gtWriteAt(4, 0, "          ***                                                %**#               ", 80);
   hb_gtWriteAt(5, 0, "          ***        ********    ********  *********         %**#               ", 80);
   hb_gtWriteAt(6, 0, "          ***        ********.   ********  *********         %**#               ", 80);
   hb_gtWriteAt(7, 0, "          ***        ***   ***  ***           ***            %**#               ", 80);
   hb_gtWriteAt(8, 0, "          ***        *******    ***  *****    ***            %**#               ", 80);
   hb_gtWriteAt(9, 0, "          ***        ***  .***  .***  ****    ***            %*                 ", 80);
   hb_gtWriteAt(10, 0, "          ***        ***   ***   *********    ***             &**#              ", 80);
   hb_gtWriteAt(11, 0, "          ***                                               ********            ", 80);
   hb_gtWriteAt(12, 0, "          ***                                            &*********             ", 80);
   hb_gtWriteAt(13, 0, "          ***                            (****         *********(               ", 80);
   hb_gtWriteAt(14, 0, "          ***                         .*********      %*********                ", 80);
   hb_gtWriteAt(15, 0, "          ***                           **********(      *********(             ", 80);
   hb_gtWriteAt(16, 0, "          *******************************( /********/      %********/           ", 80);
   hb_gtWriteAt(17, 0, "           &&&&&&&&&&&&&&&&&&&&&&&&&&&&&&  *********%         ****.             ", 80);
   hb_gtWriteAt(18, 0, "                                        %*********                              ", 80);
   hb_gtWriteAt(19, 0, "                                       ********&                                ", 80);
   hb_gtWriteAt(20, 0, "                                         %***                                   ", 80);
   hb_gtWriteAt(21, 0, "                                                                                ", 80);
   hb_gtWriteAt(22, 0, "                RGT - Remote Graphical Terminal for (x)Harbour                  ", 80);
   hb_gtWriteAt(23, 0, "                    http://gitlab.com/fabiouggeri/remote-gt                     ", 80);
   hb_gtWriteAt(24, 0, "                                                                                ", 80);
}

/**************************************************************************************************
 *                                       CLIPPER API
 **************************************************************************************************/

/**
 * Connect to RGT Server
 *
 * @param cServerAddress     RGT Server address
 * @param nServerPort        RGT Server listen port
 * @param cCommandLine       Command line to be executed by RGT Server
 * @param cRemoteWorkingDir  Remote working directory for application
 * @param cRGTServerUsername Username to connect into RGT Server
 * @param cRGTServerPswd     Password of RGT Server user
 * @param aAppArgs           Arguments to be passed on command line
 */
HB_FUNC(RGT_TRMLOGIN) {
   PHB_ITEM pServer = hb_param(1, HB_IT_STRING);
   PHB_ITEM pPort = hb_param(2, HB_IT_NUMERIC);
   PHB_ITEM pCommandLine = hb_param(3, HB_IT_STRING);
   PHB_ITEM pWorkDir = hb_param(4, HB_IT_STRING);
   PHB_ITEM pUsername = hb_param(5, HB_IT_STRING);
   PHB_ITEM pPassword = hb_param(6, HB_IT_STRING);
   PHB_ITEM pArgs = hb_param(7, HB_IT_ARRAY);
   HB_SIZE argsCount;
   const char **argsValues = NULL;

   // cfl_mem_set((CFL_MALLOC_FUNC) hb_xgrab, (CFL_REALLOC_FUNC) hb_xrealloc, (CFL_FREE_FUNC) hb_xfree);

   rgt_log_setLabel("[TE] ");
   rgt_log_initEnv();

   RGT_LOG_ENTER("HB_FUN_RGT_TRMLOGIN", (NULL));
   rgt_error_clear();
   if (pServer == NULL) {
      rgt_error_launch(RGT_ERROR_TERMINAL, RGT_ERROR_INVALID_ARG, EF_NONE, "Server address not informed or invalid", "RGT_TRMLOGIN",
                       NULL);
      RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGIN", (NULL));
      return;
   }

   if (pPort == NULL || hb_itemGetNI(pPort) < 0 || hb_itemGetNI(pPort) > 65535) {
      rgt_error_launch(RGT_ERROR_TERMINAL, RGT_ERROR_INVALID_ARG, EF_NONE, "Server port number not informed or invalid",
                       "RGT_TRMLOGIN", NULL);
      RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGIN", (NULL));
      return;
   }

   if (pCommandLine == NULL) {
      rgt_error_launch(RGT_ERROR_TERMINAL, RGT_ERROR_INVALID_ARG, EF_NONE, "Command line not informed or invalid", "RGT_TRMLOGIN",
                       NULL);
      RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGIN", (NULL));
      return;
   }

   if (pWorkDir == NULL) {
      rgt_error_launch(RGT_ERROR_TERMINAL, RGT_ERROR_INVALID_ARG, EF_NONE, "Working dir not informed or invalid", "RGT_TRMLOGIN",
                       NULL);
      RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGIN", (NULL));
      return;
   }

   if (pUsername == NULL) {
      rgt_error_launch(RGT_ERROR_TERMINAL, RGT_ERROR_INVALID_ARG, EF_NONE, "Username not informed or invalid", "RGT_TRMLOGIN",
                       NULL);
      RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGIN", (NULL));
      return;
   }

   if (pPassword == NULL) {
      rgt_error_launch(RGT_ERROR_TERMINAL, RGT_ERROR_INVALID_ARG, EF_NONE, "Password not informed or invalid", "RGT_TRMLOGIN",
                       NULL);
      RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGIN", (NULL));
      return;
   }

   if (pArgs != NULL) {
      argsCount = hb_arrayLen(pArgs);
      if (argsCount > 0) {
         HB_SIZE iPar;
         argsValues = (const char **)RGT_HB_ALLOC(argsCount * sizeof(char *));
         for (iPar = 0; iPar < argsCount; iPar++) {
            argsValues[iPar] = hb_arrayGetCPtr(pArgs, iPar + 1);
         }
      }
   } else {
      argsCount = 0;
      argsValues = NULL;
   }

   showInitScreen();
   s_connection =
       rgt_trm_login(hb_itemGetCPtr(pServer), hb_itemGetNI(pPort), hb_itemGetCPtr(pCommandLine), hb_itemGetCPtr(pWorkDir),
                     hb_itemGetCPtr(pUsername), hb_itemGetCPtr(pPassword), (CFL_UINT16)argsCount, argsValues);
   if (argsCount > 0) {
      RGT_HB_FREE((void *)argsValues);
   }
   if (rgt_error_hasError()) {
      if (s_connection != NULL) {
         releaseConnection(s_connection);
         s_connection = NULL;
      }
      rgt_error_launchFromRGTError(EF_NONE, "RGT_TRMLOGIN", NULL);
   }
   RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGIN", (NULL));
}

/**
 * Starts communication between terminal and application
 */
HB_FUNC(RGT_TRMLISTEN) {
   RGT_LOG_ENTER("HB_FUN_RGT_TRMLISTEN", (NULL));
   if (s_connection) {
      rgt_trm_listenConnection(s_connection);
      if (rgt_error_hasError()) {
         rgt_trm_logout();
         rgt_error_launchFromRGTError(EF_NONE, "RGT_TRMLISTEN", NULL);
      }
   } else {
      rgt_error_launch(RGT_ERROR_TERMINAL, RGT_ERROR_INVALID_ARG, EF_NONE, "Not connected to server", "RGT_TRMLISTEN", NULL);
   }
   RGT_LOG_EXIT("HB_FUN_RGT_TRMLISTEN", (NULL));
}

/**
 * Disconnect from RGT Server
 */
HB_FUNC(RGT_TRMLOGOUT) {
   RGT_LOG_ENTER("HB_FUN_RGT_TRMLOGOUT", (NULL));
   if (s_connection) {
      rgt_trm_logout();
      if (rgt_error_hasError()) {
         rgt_error_launchFromRGTError(EF_NONE, "RGT_TRMLOGOUT", NULL);
      }
   } else {
      RGT_LOG_ERROR(("RGT_TRMLOGOUT(): Not connected to server"));
   }
   RGT_LOG_EXIT("HB_FUN_RGT_TRMLOGOUT", (NULL));
}

/**
 * Timeout waiting response from server
 * @param nNewTimeout Optional parameter indicating new connection timeout
 * @return current connection timeout
 */
HB_FUNC(RGT_TRMTIMEOUT) {
   RGT_LOG_ENTER("HB_FUN_RGT_TRMTIMEOUT", (NULL));
   if (s_connection) {
      PHB_ITEM pNewValue = hb_param(1, HB_IT_NUMERIC);
      hb_retni(s_connection->timeout);
      if (pNewValue && hb_itemGetNI(pNewValue) > 0) {
         s_connection->timeout = (CFL_UINT32)hb_itemGetNI(pNewValue);
      }
   } else {
      hb_retni(0);
   }
   RGT_LOG_EXIT("HB_FUN_RGT_TRMTIMEOUT", (NULL));
}

/**
 * Interval to send keys to app
 *
 * @param nNewUpdateInterval optional parameter setting new interval update
 * @return current interval update
 */
HB_FUNC(RGT_TRMUPDATEINTERVAL) {
   RGT_LOG_ENTER("HB_FUN_RGT_TRMUPDATEINTERVAL", (NULL));
   if (s_connection) {
      PHB_ITEM pNewValue = hb_param(1, HB_IT_NUMERIC);
      hb_retni(s_connection->sendKeysInterval);
      if (pNewValue && hb_itemGetNI(pNewValue) > 0) {
         s_connection->sendKeysInterval = (CFL_UINT32)hb_itemGetNI(pNewValue);
         if (s_connection->sendKeysInterval < 2) {
            s_connection->sendKeysInterval = 2;
         }
      }
   } else {
      hb_retni(0);
   }
   RGT_LOG_EXIT("HB_FUN_RGT_TRMUPDATEINTERVAL", (NULL));
}

/**
 * Default interval to send keys to app
 *
 * @param nNewUpdateInterval optional parameter setting new interval update
 * @return current interval update
 */
HB_FUNC(RGT_TRMDEFAULTUPDATEINTERVAL) {
   PHB_ITEM pNewValue;
   RGT_LOG_ENTER("HB_FUN_RGT_TRMDEFAULTUPDATEINTERVAL", (NULL));
   pNewValue = hb_param(1, HB_IT_NUMERIC);
   hb_retni(s_defaultIntervalSendKeys);
   if (pNewValue) {
      s_defaultIntervalSendKeys = (CFL_UINT32)hb_itemGetNI(pNewValue);
      if (s_defaultIntervalSendKeys < 2) {
         s_defaultIntervalSendKeys = 2;
      }
   }
   RGT_LOG_EXIT("HB_FUN_RGT_TRMDEFAULTUPDATEINTERVAL", (NULL));
}

/**
 * Max interval that terminal must receive a keep alive or update command
 *
 * @param nNewUpdateInterval optional parameter setting new interval update
 * @return current interval update
 */
HB_FUNC(RGT_TRMLACKCOMMUNICATIONTIMEOUT) {
   RGT_LOG_ENTER("HB_FUN_RGT_TRMLACKCOMMUNICATIONTIMEOUT", (NULL));
   if (s_connection) {
      PHB_ITEM pNewValue = hb_param(1, HB_IT_NUMERIC);
      hb_retni(s_connection->timeoutAppCommunication);
      if (pNewValue && hb_itemGetNI(pNewValue) > 0) {
         s_connection->timeoutAppCommunication = (CFL_UINT32)hb_itemGetNI(pNewValue);
      }
   } else {
      hb_retni(0);
   }
   RGT_LOG_EXIT("HB_FUN_RGT_TRMLACKCOMMUNICATIONTIMEOUT", (NULL));
}

HB_FUNC(RGT_SERVERADDRESS) {
   if (s_connection != NULL) {
      hb_retclen(cfl_str_getPtr(s_connection->server), cfl_str_getLength(s_connection->server));
   } else {
      hb_retclen_const("", 0);
   }
}

HB_FUNC(RGT_TRMMAXKEYSBUFFER) {
   PHB_ITEM pNewValue = hb_param(1, HB_IT_NUMERIC);
   hb_retni(rgt_trm_getMaxKeysInBuffer());
   if (pNewValue && hb_itemGetNI(pNewValue) > 0) {
      rgt_trm_setMaxKeysInBuffer((CFL_UINT32)hb_itemGetNI(pNewValue));
   }
}

HB_FUNC(RGT_SESSIONID) {
   hb_retnll((HB_LONGLONG)rgt_trm_sessionId());
}
