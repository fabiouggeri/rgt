/*
 * Remote Terminal Emulator
 */

#ifdef __HBR__

#define HB_GT_NAME RGTTRM

#include <stdio.h>
#include <stdlib.h>


#include "hbapierr.h"
#include "hbapiitm.h"
#include "hbgtcore.h"
#include "hbinit.h"
#include "hbset.h"
#include "hbvm.h"


#include "rgt_app_api.h"
#include "rgt_app_connection.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_types.h"


#include "cfl_mem.h"

static int s_GtId;
static HB_GT_FUNCS SuperTable;

#define HB_GTSUPER (&SuperTable)
#define HB_GTID_PTR (&s_GtId)

#define RGT_GTTRM_GET(p) ((PRGT_GTTRM)HB_GTLOCAL(p))
#undef HB_GTSUPERTABLE
#define HB_GTSUPERTABLE(g) (&(RGT_GTTRM_GET(g)->SuperTable))

#define STR_IS_EMPTY(s) (s == NULL || s[0] == '\0')

typedef struct {
      PHB_GT pGT;
      HB_GT_FUNCS SuperTable;
} RGT_GTTRM, *PRGT_GTTRM;

static PRGT_GTTRM s_rgtGT = NULL;
static CFL_BOOL s_callRGTInit = CFL_FALSE;
static int s_LastGetTypeAhead = 0;

static int hb_gt_rgt_InkeyGet(PHB_GT pGT, HB_BOOL fWait, double dSeconds, int iEventMask) {
   int iKey;
   RGT_LOG_ENTER("hb_gt_rgt_InkeyGet", ("%s, %f, %d", fWait ? "true" : "false", dSeconds, iEventMask));
   iKey = HB_GTSUPER_INKEYGET(pGT, fWait, dSeconds, iEventMask);
   RGT_LOG_EXIT("hb_gt_rgt_InkeyGet", ("key=%d", iKey));
   return iKey;
}

static int hb_gt_rgt_ReadKey(PHB_GT pGT, int iEventMask) {
   int iKey;
   RGT_APP_CONNECTIONP conn = rgt_app_getConnection();

   RGT_LOG_ENTER("hb_gt_rgt_ReadKey", ("%d", iEventMask));
   if (rgt_app_conn_isActive(conn) && !rgt_error_hasError()) {
      if (!rgt_app_conn_isUpdateBackground(conn)) {
         rgt_app_conn_updateTerminal(conn);
      }
      iKey = rgt_app_conn_getKey(conn);
      RGT_LOG_DEBUG(("hb_gt_rgt_ReadKey(). Key: %d", iKey));
      if (rgt_error_hasError()) {
         RGT_LOG_EXIT("hb_gt_rgt_ReadKey", ("key=%d", iKey));
         rgt_app_handleLastError("GT_READKEY");
         return 0;
      }
   } else if (rgt_app_inTransactionMode()) {
      iKey = 0;
   } else {
      hb_vmRequestQuit();
      iKey = 0;
   }
   RGT_LOG_EXIT("hb_gt_rgt_ReadKey", ("key=%d", iKey));
   return iKey;
}

static const char *hb_gt_rgt_Version(PHB_GT pGT, int iType) {
   if (iType == 0) {
      return HB_GT_DRVNAME(HB_GT_NAME);
   }

   return "RGT - Remote Graphical Terminal";
}

static HB_BOOL hb_gt_rgt_SetMode(PHB_GT pGT, int iRows, int iCols) {
   RGT_LOG_ENTER("hb_gt_rgt_SetMode", ("%d, %d", iRows, iCols));

   if (HB_GTSUPER_SETMODE(pGT, iRows, iCols)) {
      if (rgt_app_isConnected() && !rgt_error_hasError()) {
         rgt_app_conn_prepareTerminal(rgt_app_getConnection());
         if (rgt_error_hasError()) {
            RGT_LOG_EXIT("hb_gt_rgt_SetMode", (NULL));
            rgt_app_handleLastError("GT_SETMODE");
            return HB_FALSE;
         }
      } else if (!rgt_app_inTransactionMode()) {
         hb_vmRequestQuit();
      }
      RGT_LOG_EXIT("hb_gt_rgt_SetMode", (NULL));
      return HB_TRUE;
   }

   RGT_LOG_EXIT("hb_gt_rgt_SetMode", (NULL));
   return HB_FALSE;
}

static void hb_gt_rgt_SetPos(PHB_GT pGT, int iRow, int iCol) {
   RGT_LOG_ENTER("hb_gt_rgt_SetPos", ("%d, %d", iRow, iCol));

   HB_GTSUPER_SETPOS(pGT, iRow, iCol);
   if (rgt_app_isConnected() && !rgt_error_hasError()) {
      rgt_app_conn_setCursorPos(rgt_app_getConnection(), iRow, iCol);
      if (rgt_error_hasError()) {
         RGT_LOG_EXIT("hb_gt_rgt_SetPos", (NULL));
         rgt_app_handleLastError("GT_SETPOS");
         return;
      }
   } else if (!rgt_app_inTransactionMode()) {
      hb_vmRequestQuit();
   }
   RGT_LOG_EXIT("hb_gt_rgt_SetPos", (NULL));
}

static void hb_gt_rgt_SetCursorStyle(PHB_GT pGT, int iStyle) {
   RGT_LOG_ENTER("hb_gt_rgt_SetCursorStyle", ("%d", iStyle));

   HB_GTSUPER_SETCURSORSTYLE(pGT, iStyle);
   if (rgt_app_isConnected() && !rgt_error_hasError()) {
      rgt_app_conn_setCursorStyle(rgt_app_getConnection(), iStyle);
      if (rgt_error_hasError()) {
         RGT_LOG_EXIT("hb_gt_rgt_SetCursorStyle", (NULL));
         rgt_app_handleLastError("GT_SETCURSORSTYLE");
         return;
      }
   } else if (!rgt_app_inTransactionMode()) {
      hb_vmRequestQuit();
   }
   RGT_LOG_EXIT("hb_gt_rgt_SetCursorStyle", (NULL));
}

static void hb_gt_rgt_Redraw(PHB_GT pGT, int iRow, int iCol, int iSize) {
   RGT_LOG_ENTER("hb_gt_rgt_Redraw", ("%d,%d,%d", iRow, iCol, iSize));
   if (rgt_app_isConnected() && !rgt_error_hasError()) {
      RGT_APP_CONNECTIONP conn = rgt_app_getConnection();
      int iRowIni = iRow;
      int iColIni = iCol;
      int iSizeIni = iSize;
      if (iSize > 0 && iRow < hb_gtMaxRow() + 1 && iCol < hb_gtMaxCol() + 1) {
         int iColor;
         HB_BYTE bAttr;

         while (iSize-- > 0) {
#if defined(UNICODE)
            HB_USHORT usChar;
            if (!HB_GTSELF_GETSCRCHAR(pGT, iRow, iCol, &iColor, &bAttr, &usChar)) {
               break;
            }
            rgt_app_conn_putChar(conn, iRow, iCol++, (CFL_UINT16)hb_cdpGetU16Ctrl(usChar), iColor, bAttr);
#else
            HB_UCHAR uc;
            if (!HB_GTSELF_GETSCRUC(pGT, iRow, iCol, &iColor, &bAttr, &uc, HB_TRUE)) {
               break;
            }
            rgt_app_conn_putChar(conn, iRow, iCol++, (CFL_UINT16)uc, iColor, bAttr);
#endif
         }
      }
      HB_GTSUPER_REDRAW(pGT, iRowIni, iColIni, iSizeIni);
      if (!rgt_app_conn_isUpdateBackground(conn)) {
         rgt_app_conn_updateTerminal(conn);
      }
      if (rgt_error_hasError()) {
         RGT_LOG_EXIT("hb_gt_rgt_Redraw", (NULL));
         rgt_app_handleLastError("GT_REDRAW");
         return;
      }
   } else if (!rgt_app_inTransactionMode()) {
      hb_vmRequestQuit();
   } else {
      HB_GTSUPER_REDRAW(pGT, iRow, iCol, iSize);
   }
   RGT_LOG_EXIT("hb_gt_rgt_Redraw", (NULL));
}

/* *********************************************************************** */

/* *********************************************************************** */
/* dDuration is in 'Ticks' (18.2 per second) */
static void hb_gt_rgt_Tone(PHB_GT pGT, double dFrequency, double dDuration) {
   RGT_APP_CONNECTIONP conn = rgt_app_getConnection();
   RGT_LOG_ENTER("hb_gt_rgt_Tone", ("%f,%f", dFrequency, dDuration));
   if (rgt_app_conn_isActive(conn) && !rgt_error_hasError()) {
      rgt_app_conn_tone(rgt_app_getConnection(), dFrequency, dDuration);
      if (!rgt_app_conn_isUpdateBackground(conn)) {
         rgt_app_conn_updateTerminal(conn);
      }
      if (rgt_error_hasError()) {
         RGT_LOG_EXIT("hb_gt_rgt_Tone", (NULL));
         rgt_app_handleLastError("GT_TONE");
         return;
      }
   } else if (!rgt_app_inTransactionMode()) {
      hb_vmRequestQuit();
   }
   HB_GTSUPER_TONE(pGT, dFrequency, dDuration);
   RGT_LOG_EXIT("hb_gt_rgt_Tone", (NULL));
}

static PRGT_GTTRM rgt_gt_base(void) {
   PHB_GT pGT;

   RGT_LOG_ENTER("hb_gt_base", (NULL));

   pGT = hb_gt_Base();
   if (pGT) {
      if (RGT_GTTRM_GET(pGT)) {
         RGT_LOG_EXIT("hb_gt_base", (NULL));
         return RGT_GTTRM_GET(pGT);
      } else {
         PRGT_GTTRM pRgtGT = (PRGT_GTTRM)RGT_HB_ALLOC(sizeof(RGT_GTTRM));

         memset(pRgtGT, 0, sizeof(RGT_GTTRM));
         HB_GTLOCAL(pGT) = pRgtGT;
         pRgtGT->pGT = pGT;

         if (hb_gtLoad(HB_GT_DRVNAME(HB_GT_NAME), pGT, HB_GTSUPERTABLE(pGT))) {
            RGT_LOG_EXIT("hb_gt_base", (NULL));
            return pRgtGT;
         }

         HB_GTLOCAL(pGT) = NULL;
         RGT_HB_FREE(pRgtGT);
         RGT_LOG_ERROR(("rgt_gt_base: error loading RGT driver."));
      }
      hb_gt_BaseFree(pGT);
   }

   RGT_LOG_EXIT("hb_gt_base", (NULL));
   return NULL;
}

static void hb_gt_rgt_Exit(PHB_GT pGT) {
   RGT_LOG_ENTER("hb_gt_rgt_Exit", ("%p", pGT));
   RGT_LOG_INFO(("hb_gt_rgt_Exit()"));

   HB_GTSUPER_EXIT(pGT);

   if (s_rgtGT != NULL) {
      RGT_HB_FREE(s_rgtGT);
      s_rgtGT = NULL;
   }
   RGT_LOG_EXIT("hb_gt_rgt_Exit", (NULL));
}

static HB_BOOL hb_gt_FuncInit(PHB_GT_FUNCS pFuncTable) {
   pFuncTable->Exit = hb_gt_rgt_Exit;
   pFuncTable->InkeyGet = hb_gt_rgt_InkeyGet;
   pFuncTable->ReadKey = hb_gt_rgt_ReadKey;
   pFuncTable->Version = hb_gt_rgt_Version;
   pFuncTable->SetMode = hb_gt_rgt_SetMode;
   pFuncTable->SetPos = hb_gt_rgt_SetPos;
   pFuncTable->SetCursorStyle = hb_gt_rgt_SetCursorStyle;
   pFuncTable->Redraw = hb_gt_rgt_Redraw;
   pFuncTable->Tone = hb_gt_rgt_Tone;

   return HB_TRUE;
}

static const HB_GT_INIT gtInit = {HB_GT_DRVNAME(HB_GT_NAME), hb_gt_FuncInit, HB_GTSUPER, HB_GTID_PTR};

HB_GT_ANNOUNCE(HB_GT_NAME)

static void rgtExit(void *cargo) {
   // desconectar do servidor
   rgt_app_closeConnection();
   rgt_error_finalize();
}

HB_FUNC(RGT_GT_INIT) {
   RGT_LOG_ENTER("RGT_GT_INIT", (NULL));
   RGT_LOG_INFO(("RGT_GT_INIT()"));
   if (!s_callRGTInit) {
      char title[1024];
      char *server = getenv(RGT_SERVER_ADDR_VAR);
      char *port = getenv(RGT_SERVER_PORT_VAR);
      char *strSessionId = getenv(RGT_AUTH_TOKEN_VAR);
      const char *appPathName;
      s_callRGTInit = CFL_TRUE;
      if (STR_IS_EMPTY(server) && STR_IS_EMPTY(port) && STR_IS_EMPTY(strSessionId)) {
         RGT_LOG_EXIT("RGT_GT_INIT", (NULL));
         return;
      }
      // cfl_mem_set((CFL_MALLOC_FUNC) hb_xgrab, (CFL_REALLOC_FUNC) hb_xrealloc, (CFL_FREE_FUNC) hb_xfree);
      hb_gtRegister(&gtInit);
      appPathName = hb_cmdargARGVN(0);
      snprintf(title, sizeof(title), "RGTAPP[%s][%s][%s]: %s", strSessionId, server, port, appPathName);
      rgt_app_setTitle(title);
      rgt_app_initEnv();
      rgt_error_clear();

      if (!rgt_app_openConnection(server, port, strSessionId)) {
         if (rgt_error_hasError()) {
            RGT_LOG_ERROR(("RGT_GT_INIT: %s", rgt_error_getLastMessage()));
         } else {
            RGT_LOG_ERROR(("RGT_GT_INIT: unkown error connecting to server"));
         }
         hb_vmRequestQuit();
         RGT_LOG_EXIT("RGT_GT_INIT", (NULL));
         return;
      }

      s_rgtGT = rgt_gt_base();
      if (s_rgtGT == NULL) {
         hb_vmRequestQuit();
      }
      hb_vmAtQuit(rgtExit, NULL);
   }
   RGT_LOG_EXIT("RGT_GT_INIT", (NULL));
}

/* *********************************************************************** */
// #include "hbgtreg.h"
/* *********************************************************************** */

#endif
