#include <stdio.h>

#include <hbapiitm.h>
#include <hbvm.h>

#include "hbapi.h"
#include "hbapierr.h"
#include "hbapifs.h"
#include "hbapigt.h"
#include "hbdate.h"

#include "hbgtinfo.ch"

#include "cfl_str.h"

#include "rgt_app_connection.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_screen.h"
#include "rgt_thread.h"

#define SESSION_MODE_NORMAL "0"
#define SESSION_MODE_TRANSACTION "1"

static RGT_APP_CONNECTIONP s_connection = NULL;

static CFL_INT64 strToInt64(char *str) {
   CFL_INT64 val = 0;
   while (*str >= '0' && *str <= '9') {
      val = (val * 10) + *str - '0';
      ++str;
   }
   return val;
}

CFL_BOOL rgt_app_openConnection(char *server, char *port, char *strSessionId) {
   CFL_INT64 sessionId;

   RGT_LOG_ENTER("rgt_app_openConnection", (NULL));
   RGT_LOG_INFO(("rgt_app_openConnection(server=%s, port=%s, session=%s)", server, port, strSessionId));
   /* App not launched by RGT Server or launched as standalone app. */
   if (server == NULL && port == NULL && strSessionId == NULL) {
      RGT_LOG_EXIT("rgt_app_openConnection", (NULL));
      return CFL_FALSE;
   }
   if (server == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_ENV_VAR_NOT, "Server address not found.");
      RGT_LOG_EXIT("rgt_app_openConnection", (NULL));
      return CFL_FALSE;
   }
   if (port == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_ENV_VAR_NOT, "Server port not found.");
      RGT_LOG_EXIT("rgt_app_openConnection", (NULL));
      return CFL_FALSE;
   }
   if (strSessionId == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_ENV_VAR_NOT, "Authentication token not found.");
      RGT_LOG_EXIT("rgt_app_openConnection", (NULL));
      return CFL_FALSE;
   }
   sessionId = strToInt64(strSessionId);
   s_connection = rgt_app_conn_new(server, (CFL_UINT16)atoi(port), sessionId);
   RGT_LOG_EXIT("rgt_app_openConnection", (NULL));
   return s_connection != NULL ? CFL_TRUE : CFL_FALSE;
}

void rgt_app_closeConnection(void) {
   RGT_LOG_ENTER("rgt_app_closeConnection", (NULL));
   RGT_LOG_INFO(("rgt_app_closeConnection()"));
   if (s_connection != NULL) {
      RGT_APP_CONNECTIONP conn = s_connection;
      s_connection = NULL;
      rgt_app_conn_close(conn, "");
      rgt_app_conn_free(conn);
   }
   RGT_LOG_EXIT("rgt_app_closeConnection", (NULL));
}

RGT_APP_CONNECTIONP rgt_app_getConnection(void) {
   return s_connection;
}

CFL_BOOL rgt_app_isConnected(void) {
   return rgt_app_conn_isActive(s_connection);
}

CFL_BOOL rgt_app_inTransactionMode(void) {
   return s_connection != NULL && IN_TRANSACTION_MODE(s_connection);
}

static void executeRemoteFunction(void) {
   PHB_ITEM pFuncName = hb_param(1, HB_IT_STRING);
   int argsCount = hb_pcount() - 1;
   PHB_ITEM pResult;

   RGT_LOG_ENTER("executeRemoteFunction", (NULL));

   if (pFuncName == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_INVALID_ARG, "Remote procedure name not informed");
      hb_ret();
      RGT_LOG_EXIT("executeRemoteFunction", (NULL));
      return;
   }
   RGT_LOG_DEBUG(("executeRemoteFunction(name='%s', args=%d)", hb_itemGetCPtr(pFuncName), argsCount));
   if (argsCount > 0) {
      int iParam;
      PHB_ITEM *pArgs = RGT_HB_ALLOC(argsCount * sizeof(PHB_ITEM));
      for (iParam = 0; iParam < argsCount; iParam++) {
         pArgs[iParam] = hb_param(iParam + 2, HB_IT_ANY);
      }
      pResult = rgt_app_conn_execRemoteFunction(s_connection, hb_itemGetCPtr(pFuncName), argsCount, pArgs);
      RGT_HB_FREE(pArgs);
   } else {
      pResult = rgt_app_conn_execRemoteFunction(s_connection, hb_itemGetCPtr(pFuncName), 0, NULL);
   }
   if (pResult) {
      hb_itemReturnRelease(pResult);
   } else {
      hb_ret();
   }
   RGT_LOG_EXIT("executeRemoteFunction", (NULL));
}

static void executeLocalFunction(void) {
   PHB_ITEM pFuncName = hb_param(1, HB_IT_STRING);
   int argsCount = hb_pcount();
   int iParam;
   PHB_DYNS funSymbol;

   RGT_LOG_ENTER("executeLocalFunction", (NULL));
   if (pFuncName == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_INVALID_ARG, "Procedure name not informed");
      hb_ret();
      RGT_LOG_EXIT("executeLocalFunction", (NULL));
      return;
   }
   funSymbol = hb_dynsymFindName(hb_itemGetCPtr(pFuncName));
   if (funSymbol == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_INVALID_ARG, "Procedure %s not found", hb_itemGetCPtr(pFuncName));
      hb_ret();
      RGT_LOG_EXIT("executeLocalFunction", (NULL));
      return;
   }
   RGT_LOG_DEBUG(("executeLocalFunction(name='%s', args=%d)", hb_itemGetCPtr(pFuncName), argsCount - 1));
   hb_vmPushSymbol(hb_dynsymSymbol(funSymbol));
   hb_vmPushNil();
   for (iParam = 2; iParam <= argsCount; iParam++) {
      hb_vmPush(hb_param(iParam, HB_IT_ANY));
   }
   hb_vmDo(argsCount - 1);
   RGT_LOG_EXIT("executeLocalFunction", (NULL));
}

CFL_UINT32 rgt_app_getUpdateInterval(void) {
   if (s_connection != NULL) {
      return s_connection->updateTerminalInterval;
   } else {
      RGT_LOG_ERROR(("rgt_app_getUpdateInterval(): s_connection is NULL"));
      return 0;
   }
}

void rgt_app_setUpdateInterval(CFL_UINT32 newValue) {
   if (s_connection != NULL) {
      s_connection->updateTerminalInterval = newValue;
   }
}

CFL_UINT32 rgt_app_getRPCTimeout(void) {
   if (s_connection != NULL) {
      return s_connection->rpcTimeout;
   } else {
      RGT_LOG_ERROR(("rgt_app_getRPCTimeout(): s_connection is NULL"));
      return 0;
   }
}

void rgt_app_setRPCTimeout(CFL_UINT32 newValue) {
   if (s_connection != NULL) {
      s_connection->rpcTimeout = newValue;
   } else {
      RGT_LOG_ERROR(("rgt_app_setRPCTimeout(): s_connection is NULL"));
   }
}

CFL_UINT32 rgt_app_getKeepAliveInterval(void) {
   if (s_connection != NULL) {
      return s_connection->keepAliveInterval;
   } else {
      RGT_LOG_ERROR(("rgt_app_getKeepAliveInterval(): s_connection is NULL"));
      return 0;
   }
}

void rgt_app_setKeepAliveInterval(CFL_UINT32 newValue) {
   if (s_connection != NULL) {
      s_connection->keepAliveInterval = newValue;
   } else {
      RGT_LOG_ERROR(("rgt_app_setKeepAliveInterval(): s_connection is NULL"));
   }
}

CFL_STRP rgt_app_getServerAddress(void) {
   if (s_connection != NULL) {
      return s_connection->serverAddress;
   } else {
      RGT_LOG_ERROR(("rgt_app_getServerAddress(): s_connection is NULL"));
   }
   return NULL;
}

void rgt_app_handleLastError(const char *operation) {
   if (rgt_app_inTransactionMode()) {
      return;
   }
   if (rgt_error_getLastCode() != RGT_ERROR_SOCKET) {
      rgt_error_launchFromRGTError(EF_NONE, operation, NULL);
      return;
   }
   hb_vmQuit();
   exit(1);
}

void rgt_app_setTitle(const char *title) {
   HB_GT_INFO gtInfo;
   PHB_ITEM pTitle = hb_itemPutC(NULL, title);

   memset(&gtInfo, 0, sizeof(gtInfo));
   gtInfo.pNewVal = pTitle;
   hb_gtInfo(HB_GTI_WINTITLE, &gtInfo);
   if (gtInfo.pResult) {
      hb_itemRelease(gtInfo.pResult);
   }
   hb_itemRelease(pTitle);
}

void rgt_app_initEnv(void) {
   RGT_LOG_ENTER("rgt_app_initEnv", (NULL));
   rgt_log_initEnv();
   rgt_thread_initEnv();
   hb_gtSetMode(25, 80);
   RGT_LOG_EXIT("rgt_app_initEnv", (NULL));
}

/**************************************************************************************************
 *                                       CLIPPER API
 **************************************************************************************************/
HB_FUNC(RGT_EXECREMOTE) {
   RGT_LOG_ENTER("HB_FUN_RGT_EXECREMOTE",
                 ("rpc-local=%s, conn-active=%s", (RPC_LOCAL_EXEC_LOST_CONNECTION(s_connection) ? "true" : "false"),
                  (rgt_app_conn_isActive(s_connection) ? "true" : "false")));
   if (s_connection == NULL) {
      executeLocalFunction();
   } else if (RPC_LOCAL_EXEC_LOST_CONNECTION(s_connection) && !rgt_app_conn_isActive(s_connection)) {
      executeLocalFunction();
      rgt_error_clear();
   } else {
      executeRemoteFunction();
   }
   if (rgt_error_hasError()) {
      rgt_app_handleLastError("RGT_EXECREMOTE");
   }
   RGT_LOG_EXIT("HB_FUN_RGT_EXECREMOTE", (NULL));
}

HB_FUNC(RGT_PUTFILE) {
   PHB_ITEM pLocalFile = hb_param(1, HB_IT_STRING);
   PHB_ITEM pRemoteFile = hb_param(2, HB_IT_STRING);
   int iChunkSize = hb_parni(3);

   RGT_LOG_ENTER("HB_FUN_RGT_PUTFILE", (NULL));
   if (pLocalFile == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_INVALID_ARG, "Local file path and name to transfer not informed");
      hb_retni(RGT_ERROR_APP);
      RGT_LOG_EXIT("HB_FUN_RGT_PUTFILE", (NULL));
      return;
   }
   if (pRemoteFile == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_INVALID_ARG, "Target file path and name not informed");
      hb_retni(RGT_ERROR_APP);
      RGT_LOG_EXIT("HB_FUN_RGT_PUTFILE", (NULL));
      return;
   }
   if (iChunkSize < 0) {
      iChunkSize = 0;
   }
   RGT_LOG_DEBUG(("RGT_PUT_FILE(local='%s', remote='%s')", hb_itemGetCPtr(pLocalFile), hb_itemGetCPtr(pRemoteFile)));
   if (s_connection != NULL) {
      hb_retni(rgt_app_conn_putFile(s_connection, hb_itemGetCPtr(pLocalFile), hb_itemGetCPtr(pRemoteFile), (CFL_UINT32)iChunkSize));
   } else if (hb_fileCopy(hb_itemGetCPtr(pLocalFile), hb_itemGetCPtr(pRemoteFile))) {
      hb_retni(0);
   } else {
      hb_retni(hb_fsError());
   }
   if (rgt_error_hasError()) {
      rgt_app_handleLastError("RGT_PUTFILE");
   }
   RGT_LOG_EXIT("HB_FUN_RGT_PUTFILE", (NULL));
}

HB_FUNC(RGT_GETFILE) {
   PHB_ITEM pRemoteFile = hb_param(1, HB_IT_STRING);
   PHB_ITEM pLocalFile = hb_param(2, HB_IT_STRING);
   int iChunkSize = hb_parni(3);

   RGT_LOG_ENTER("HB_FUN_RGT_GETFILE", (NULL));
   if (pLocalFile == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_INVALID_ARG, "Local file path and name to transfer not informed");
      hb_retni(RGT_ERROR_APP);
      RGT_LOG_EXIT("HB_FUN_RGT_GETFILE", (NULL));
      return;
   }
   if (pRemoteFile == NULL) {
      rgt_error_set(RGT_APP, RGT_ERROR_INVALID_ARG, "Target file path and name not informed");
      hb_retni(RGT_ERROR_APP);
      RGT_LOG_EXIT("HB_FUN_RGT_GETFILE", (NULL));
      return;
   }
   if (iChunkSize < 0) {
      iChunkSize = 0;
   }
   RGT_LOG_DEBUG(("RGT_GET_FILE(remote='%s', local='%s')", hb_itemGetCPtr(pRemoteFile), hb_itemGetCPtr(pLocalFile)));
   if (s_connection != NULL) {
      hb_retni(
          (rgt_app_conn_getFile(s_connection, hb_itemGetCPtr(pRemoteFile), hb_itemGetCPtr(pLocalFile), (CFL_UINT32)iChunkSize)));
   } else if (hb_fileCopy(hb_itemGetCPtr(pRemoteFile), hb_itemGetCPtr(pLocalFile))) {
      hb_retni(0);
   } else {
      hb_retni(hb_fsError());
   }
   if (rgt_error_hasError()) {
      rgt_app_handleLastError("RGT_GETFILE");
   }
   RGT_LOG_EXIT("HB_FUN_RGT_GETFILE", (NULL));
}

HB_FUNC(RGT_REMOTETERMINAL) {
   hb_retl(s_connection != NULL);
}

HB_FUNC(RGT_UPDATETERMINAL) {
   RGT_LOG_ENTER("RGT_UPDATETERMINAL", (NULL));
   if (rgt_app_isConnected()) {
      RGT_LOG_DEBUG(("RGT_UPDATETERMINAL()"));
      rgt_screen_fullUpdated(s_connection->screen);
      rgt_app_conn_updateTerminal(s_connection);
      if (rgt_error_hasError()) {
         rgt_app_handleLastError("RGT_UPDATETERMINAL");
      }
   } else {
      RGT_LOG_DEBUG(("RGT_UPDATETERMINAL(): not connected"));
   }
   RGT_LOG_EXIT("RGT_UPDATETERMINAL", (NULL));
}

/**
 * Interval to update screen
 * @param
 * @return
 */
HB_FUNC(RGT_UPDATEINTERVAL) {
   PHB_ITEM pNewValue;
   RGT_LOG_ENTER("RGT_UPDATEINTERVAL", (NULL));
   pNewValue = hb_param(1, HB_IT_NUMERIC);
   hb_retni(rgt_app_getUpdateInterval());
   if (pNewValue) {
      rgt_app_setUpdateInterval(hb_itemGetNI(pNewValue));
   }
   RGT_LOG_EXIT("RGT_UPDATEINTERVAL", (NULL));
}

/**
 * timeout for remote procedure calls
 * @param [nNewTimeout] optional value for new RPC timeout.
 *
 * @return current timeout value
 */
HB_FUNC(RGT_RPCTIMEOUT) {
   PHB_ITEM pNewValue;
   RGT_LOG_ENTER("RGT_RPCTIMEOUT", (NULL));
   pNewValue = hb_param(1, HB_IT_NUMERIC);
   hb_retni(rgt_app_getRPCTimeout());
   if (pNewValue && hb_itemGetNI(pNewValue) >= 0) {
      rgt_app_setRPCTimeout((CFL_UINT32)hb_itemGetNI(pNewValue));
   }
   RGT_LOG_EXIT("RGT_RPCTIMEOUT", (NULL));
}

/**
 * Communication timeout
 * @param newTimeout optional parameter indicating the communication timeout.
 * @return current communication timeout
 */
HB_FUNC(RGT_TIMEOUT) {
   RGT_LOG_ENTER("RGT_TIMEOUT", (NULL));
   if (s_connection != NULL) {
      PHB_ITEM pNewValue = hb_param(1, HB_IT_NUMERIC);
      hb_retni(rgt_app_connGetTimeout(s_connection));
      if (pNewValue && hb_itemGetNI(pNewValue) >= 0) {
         rgt_app_connSetTimeout(s_connection, (CFL_UINT32)hb_itemGetNI(pNewValue));
      }
   } else {
      hb_retni(0);
      RGT_LOG_ERROR(("RGT_TIMEOUT(): not connected"));
   }
   RGT_LOG_EXIT("RGT_TIMEOUT", (NULL));
}

/**
 * interval to send keed alive message to TE
 * @param
 * @return
 */
HB_FUNC(RGT_KEEPALIVE) {
   PHB_ITEM pNewValue;
   RGT_LOG_ENTER("RGT_KEEPALIVE", (NULL));
   pNewValue = hb_param(1, HB_IT_NUMERIC);
   hb_retni(rgt_app_getKeepAliveInterval());
   if (pNewValue && hb_itemGetNI(pNewValue) >= 0) {
      rgt_app_setKeepAliveInterval((CFL_UINT32)hb_itemGetNI(pNewValue));
   }
   RGT_LOG_EXIT("RGT_KEEPALIVE", (NULL));
}

/**
 * set/get options of application session in server
 *
 */
HB_FUNC(RGT_SETSESSIONOPTION) {
   PHB_ITEM pOption = hb_param(1, HB_IT_STRING);
   PHB_ITEM pValue = hb_param(2, HB_IT_ANY);
   HB_BOOL bSuccess = HB_FALSE;

   RGT_LOG_ENTER("RGT_SETSESSIONOPTION", (NULL));
   if (s_connection == NULL || pOption == NULL || pValue == NULL) {
      hb_retl(HB_FALSE);
      RGT_LOG_EXIT("RGT_SETSESSIONOPTION", (NULL));
      return;
   }
   switch (hb_itemType(pValue)) {
   case HB_IT_DATE: {
      int iYear, iMonth, iDay;
      char buffer[9];
      hb_dateDecode(hb_itemGetDL(pValue), &iYear, &iMonth, &iDay);
      snprintf(buffer, sizeof(buffer), "%04d%02d%02d", iYear, iMonth, iDay);
      bSuccess = (HB_BOOL)rgt_app_setSessionOption(s_connection, hb_itemGetCPtr(pOption), buffer);
      RGT_LOG_DEBUG(("RGT_SETSESSIONOPTION(option='%s', date value=%s) => %s", hb_itemGetCPtr(pOption), buffer,
                     (bSuccess ? "true" : "false")));
   } break;

   case HB_IT_INTEGER:
   case HB_IT_LONG:
   case HB_IT_DOUBLE: {
      char *strValue = hb_itemStr(pValue, NULL, NULL);
      if (strValue != NULL) {
         bSuccess = (HB_BOOL)rgt_app_setSessionOption(s_connection, hb_itemGetCPtr(pOption), strValue);
         RGT_HB_FREE(strValue);
         RGT_LOG_DEBUG(("RGT_SETSESSIONOPTION(option='%s', number value=%s) => %s", hb_itemGetCPtr(pOption), strValue,
                        (bSuccess ? "true" : "false")));
      } else {
         RGT_LOG_ERROR(("RGT_SETSESSIONOPTION(option='%s'): invalid number"));
      }
   } break;

   case HB_IT_LOGICAL:
      bSuccess = (HB_BOOL)rgt_app_setSessionOption(s_connection, hb_itemGetCPtr(pOption), hb_itemGetL(pValue) ? "true" : "false");
      RGT_LOG_DEBUG(("RGT_SETSESSIONOPTION(option='%s', boolean value=%s) => %s", hb_itemGetCPtr(pOption),
                     hb_itemGetL(pValue) ? "true" : "false", (bSuccess ? "true" : "false")));
      break;

   case HB_IT_MEMO:
   case HB_IT_STRING:
      bSuccess = (HB_BOOL)rgt_app_setSessionOption(s_connection, hb_itemGetCPtr(pOption), hb_itemGetCPtr(pValue));
      RGT_LOG_DEBUG(("RGT_SETSESSIONOPTION(option='%s', str value='%s') => %s", hb_itemGetCPtr(pOption), hb_itemGetCPtr(pValue),
                     (bSuccess ? "true" : "false")));
      break;

   default:
      RGT_LOG_ERROR(("RGT_SETSESSIONOPTION(option='%s') => unsupported datatype", hb_itemGetCPtr(pOption)));
      bSuccess = HB_FALSE;
   }
   hb_retl(bSuccess);
   RGT_LOG_EXIT("RGT_SETSESSIONOPTION", (NULL));
}

HB_FUNC(RGT_SERVERADDRESS) {
   CFL_STRP addr;
   RGT_LOG_ENTER("RGT_SERVERADDRESS", (NULL));
   addr = rgt_app_getServerAddress();
   if (addr != NULL) {
      hb_retclen(cfl_str_getPtr(addr), cfl_str_getLength(addr));
   } else {
      hb_retclen_const("", 0);
   }
   RGT_LOG_EXIT("RGT_SERVERADDRESS", (NULL));
}

HB_FUNC(RGT_FILETRANSFERCHUNKSIZE) {
   PHB_ITEM pNewValue = hb_param(1, HB_IT_NUMERIC);

   RGT_LOG_ENTER("RGT_FILETRANSFERCHUNKSIZE", (NULL));
   if (s_connection == NULL) {
      hb_retni(0);
      RGT_LOG_EXIT("RGT_FILETRANSFERCHUNKSIZE", (NULL));
      return;
   }
   hb_retni((int)rgt_app_conn_getFileTransferChunkSize(s_connection));
   if (pNewValue) {
      rgt_app_conn_setFileTransferChunkSize(s_connection, (CFL_UINT32)hb_itemGetNI(pNewValue));
   }
   RGT_LOG_EXIT("RGT_FILETRANSFERCHUNKSIZE", (NULL));
}

HB_FUNC(RGT_ISCONNECTED) {
   hb_retl((HB_BOOL)rgt_app_isConnected());
}

HB_FUNC(RGT_INTRANSACTIONMODE) {
   hb_retl((HB_BOOL)rgt_app_inTransactionMode());
}

HB_FUNC(RGT_BEGINTRANSACTION) {
   RGT_LOG_ENTER("RGT_BEGINTRANSACTION", (NULL));
   if (s_connection == NULL) {
      hb_retl(HB_FALSE);
      RGT_LOG_EXIT("RGT_BEGINTRANSACTION", (NULL));
      return;
   }
   if (IN_TRANSACTION_MODE(s_connection)) {
      hb_retl(HB_TRUE);
      RGT_LOG_EXIT("RGT_BEGINTRANSACTION", (NULL));
      return;
   }
   if (rgt_app_setSessionOption(s_connection, "MODE", SESSION_MODE_TRANSACTION)) {
      s_connection->sessionMode = RGT_SESS_MODE_TRANSACTION;
      hb_retl(HB_TRUE);
      RGT_LOG_DEBUG(("RGT_BEGINTRANSACTION(): success"));
   } else {
      hb_retl(HB_FALSE);
      RGT_LOG_ERROR(("RGT_BEGINTRANSACTION(): failed"));
   }
   if (rgt_error_hasError()) {
      rgt_app_handleLastError("RGT_BEGINTRANSACTION");
   }
   RGT_LOG_EXIT("RGT_BEGINTRANSACTION", (NULL));
}

HB_FUNC(RGT_ENDTRANSACTION) {
   RGT_LOG_ENTER("RGT_ENDTRANSACTION", (NULL));
   if (s_connection != NULL && IN_TRANSACTION_MODE(s_connection)) {
      s_connection->sessionMode = RGT_SESS_MODE_NORMAL;
      rgt_app_setSessionOption(s_connection, "MODE", SESSION_MODE_NORMAL);
      hb_retl(HB_TRUE);
      RGT_LOG_DEBUG(("RGT_ENDTRANSACTION(): success"));
   } else {
      hb_retl(HB_FALSE);
      RGT_LOG_ERROR(("RGT_ENDTRANSACTION(): failed"));
   }
   RGT_LOG_EXIT("RGT_ENDTRANSACTION", (NULL));
}

HB_FUNC(RGT_RPCEXECLOCALLOSTCONNECTION) {
   RGT_LOG_ENTER("RGT_RPCEXECLOCALLOSTCONNECTION", (NULL));
   if (s_connection != NULL) {
      PHB_ITEM pNewValue = hb_param(1, HB_IT_LOGICAL);
      HB_BOOL execLocal = (HB_BOOL)rgt_app_conn_isRpcExecuteLocalLostConnection(s_connection);
      if (pNewValue != NULL) {
         rgt_app_conn_setRpcExecuteLocalLostConnection(s_connection, (CFL_BOOL)hb_itemGetL(pNewValue));
      }
      hb_retl(execLocal);
   } else {
      hb_retl(HB_FALSE);
   }
   RGT_LOG_EXIT("RGT_RPCEXECLOCALLOSTCONNECTION", (NULL));
}

HB_FUNC(RGT_SESSIONID) {
   RGT_LOG_ENTER("RGT_SESSIONID", (NULL));
   if (s_connection != NULL) {
      hb_retnll((HB_LONGLONG)s_connection->sessionId);
   } else {
      hb_retni(0);
   }
   RGT_LOG_EXIT("RGT_SESSIONID", (NULL));
}
