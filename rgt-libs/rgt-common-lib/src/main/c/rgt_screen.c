#include <stdlib.h>

#include "cfl_types.h"

#include "hbapicdp.h"
#include "hbapigt.h"
#include "hbapiitm.h"
#include "hbgtcore.h"
#include "hbstack.h"

#include "cfl_buffer.h"
#include "cfl_lock.h"
#include "rgt_log.h"
#include "rgt_screen.h"

extern HB_ERRCODE rgt_error_launch(CFL_UINT16 uiGenCode, CFL_UINT16 uiSubCode, CFL_UINT16 uiFlags, const char *description,
                                   const char *operation, PHB_ITEM *pErrorPtr);

RGT_SCREENP rgt_screen_new(CFL_UINT16 rows, CFL_UINT16 cols, CFL_INT8 screenType) {
   RGT_SCREENP screen;

   RGT_LOG_ENTER("rgt_screen_new",
                 ("rows=%u col=%u type=%s", rows, cols, (screenType == RGT_SCREEN_TYPE_SIMPLE ? "simple" : "extended")));
   screen = (RGT_SCREENP)RGT_HB_ALLOC(sizeof(RGT_SCREEN));
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_new(): error allocating RGT_SCREEN"));
      RGT_LOG_EXIT("rgt_screen_new", (NULL));
      return NULL;
   }
   screen->changed = CFL_FALSE;
   screen->height = rows;
   screen->width = cols;
   screen->topUpdate = rows;
   screen->leftUpdate = cols;
   screen->bottomUpdate = 0;
   screen->rightUpdate = 0;
   screen->screenType = screenType;
   screen->cursorRow = 0;
   screen->cursorCol = 0;
   screen->cursorStyle = 0;
   screen->lastCursorRow = 0;
   screen->lastCursorCol = 0;
   screen->lastCursorStyle = 0;
   RGT_LOCK_INIT(screen->locked);
   screen->pGT = hb_stackGetGT();
   RGT_LOG_EXIT("rgt_screen_new", (NULL));
   return screen;
}

void rgt_screen_reset(RGT_SCREENP screen, CFL_UINT16 rows, CFL_UINT16 cols, CFL_INT8 screenType) {
   RGT_LOG_ENTER("rgt_screen_reset", (NULL));
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_reset(): screen argument is null"));
      RGT_LOG_EXIT("rgt_screen_reset", (NULL));
      return;
   }
   RGT_LOCK_ACQUIRE(screen->locked);
   if (screen->height != rows || screen->width != cols || screen->screenType != screenType) {
      screen->height = rows;
      screen->width = cols;
      screen->screenType = screenType;
   }
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_reset", (NULL));
}

void rgt_screen_free(RGT_SCREENP screen) {
   if (screen != NULL) {
      RGT_LOCK_FREE(screen->locked);
      RGT_HB_FREE(screen);
   } else {
      RGT_LOG_ERROR(("rgt_screen_free(): screen argument is null"));
   }
}

void rgt_screen_rectUpdated(RGT_SCREENP screen, CFL_UINT16 rowIni, CFL_UINT16 colIni, CFL_UINT16 rowEnd, CFL_UINT16 colEnd) {
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_rectUpdated(): screen argument is null"));
   } else if (rowIni <= rowEnd && colIni <= colEnd && rowIni < screen->height && colIni < screen->width &&
              rowEnd < screen->height && colEnd < screen->width) {
      RGT_LOCK_ACQUIRE(screen->locked);
      if (rowIni < screen->topUpdate) {
         screen->topUpdate = rowIni;
      }
      if (rowEnd > screen->bottomUpdate) {
         screen->bottomUpdate = rowEnd;
      }
      if (colIni < screen->leftUpdate) {
         screen->leftUpdate = colIni;
      }
      if (colEnd > screen->rightUpdate) {
         screen->rightUpdate = colEnd;
      }
      screen->changed = CFL_TRUE;
      RGT_LOCK_RELEASE(screen->locked);
   }
}

void rgt_screen_clear(RGT_SCREENP screen) {
   RGT_LOG_ENTER("rgt_screen_clear", (NULL));
   if (screen != NULL) {
      RGT_LOCK_ACQUIRE(screen->locked);
      screen->topUpdate = screen->height - 1;
      screen->leftUpdate = screen->width - 1;
      screen->bottomUpdate = 0;
      screen->rightUpdate = 0;
      screen->cursorCol = screen->lastCursorCol;
      screen->cursorRow = screen->lastCursorRow;
      screen->cursorStyle = screen->lastCursorStyle;
      screen->changed = CFL_FALSE;
      RGT_LOCK_RELEASE(screen->locked);
   } else {
      RGT_LOG_ERROR(("rgt_screen_clear(): screen argument is null"));
   }
   RGT_LOG_EXIT("rgt_screen_clear", (NULL));
}

static void charToBuffer(RGT_SCREENP screen, PHB_CODEPAGE cdp, CFL_UINT16 iRow, CFL_UINT16 iCol, CFL_BUFFERP buffer) {
   int iColor;
   HB_BYTE attr;
   HB_USHORT character;

   if (HB_GTSELF_GETCHAR(screen->pGT, iRow, iCol, &iColor, &attr, &character)) {
      cfl_buffer_putUInt8(buffer, screen->pGT->fVgaCell ? (CFL_UINT8)hb_cdpGetChar(cdp, character) : (CFL_UINT8)(0xFF & character));
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)iColor);
   } else {
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)HB_GTSELF_GETCLEARCHAR(screen->pGT));
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)HB_GTSELF_GETCLEARCOLOR(screen->pGT));
   }
}

static void char16ToBuffer(RGT_SCREENP screen, CFL_UINT16 iRow, CFL_UINT16 iCol, CFL_BUFFERP buffer) {
   int iColor;
   HB_BYTE attr;
   HB_USHORT character;
   if (HB_GTSELF_GETCHAR(screen->pGT, iRow, iCol, &iColor, &attr, &character)) {
      cfl_buffer_putUInt16(buffer, (CFL_UINT16)character);
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)iColor);
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)attr);
   } else {
      cfl_buffer_putUInt16(buffer, (CFL_UINT16)HB_GTSELF_GETCLEARCHAR(screen->pGT));
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)HB_GTSELF_GETCLEARCOLOR(screen->pGT));
      cfl_buffer_putUInt8(buffer, 0x00);
   }
}

CFL_BOOL rgt_screen_toBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer, CFL_BOOL force) {
   CFL_UINT16 iRow;
   CFL_UINT16 iCol;
   CFL_UINT16 topUpdate;
   CFL_UINT16 leftUpdate;
   CFL_UINT16 bottomUpdate;
   CFL_UINT16 rightUpdate;

   RGT_LOG_ENTER("rgt_screen_toBuffer", (NULL));
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_toBuffer(): screen argument is null"));
      RGT_LOG_EXIT("rgt_screen_toBuffer", (NULL));
      return CFL_FALSE;
   }
   RGT_LOCK_ACQUIRE(screen->locked);
   if (!screen->changed && !force) {
      RGT_LOCK_RELEASE(screen->locked);
      RGT_LOG_EXIT("rgt_screen_toBuffer", (NULL));
      return CFL_FALSE;
   }

   topUpdate = screen->topUpdate;
   leftUpdate = screen->leftUpdate;
   bottomUpdate = screen->bottomUpdate;
   rightUpdate = screen->rightUpdate;
   screen->topUpdate = screen->height - 1;
   screen->leftUpdate = screen->width - 1;
   screen->bottomUpdate = 0;
   screen->rightUpdate = 0;
   screen->lastCursorCol = screen->cursorCol;
   screen->lastCursorRow = screen->cursorRow;
   screen->lastCursorStyle = screen->cursorStyle;
   screen->changed = CFL_FALSE;
   cfl_buffer_putInt16(buffer, screen->cursorStyle);
   cfl_buffer_putInt16(buffer, screen->cursorRow);
   cfl_buffer_putInt16(buffer, screen->cursorCol);
   cfl_buffer_putUInt16(buffer, topUpdate);
   cfl_buffer_putUInt16(buffer, leftUpdate);
   cfl_buffer_putUInt16(buffer, bottomUpdate);
   cfl_buffer_putUInt16(buffer, rightUpdate);
   if (screen->screenType == RGT_SCREEN_TYPE_SIMPLE) {
      PHB_CODEPAGE cdp = HB_GTSELF_HOSTCP(screen->pGT);
      for (iRow = topUpdate; iRow <= bottomUpdate; iRow++) {
         for (iCol = leftUpdate; iCol <= rightUpdate; iCol++) {
            charToBuffer(screen, cdp, iRow, iCol, buffer);
         }
      }
   } else {
      for (iRow = topUpdate; iRow <= bottomUpdate; iRow++) {
         for (iCol = leftUpdate; iCol <= rightUpdate; iCol++) {
            char16ToBuffer(screen, iRow, iCol, buffer);
         }
      }
   }
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_toBuffer", (NULL));
   return CFL_TRUE;
}

void rgt_screen_fullToBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer) {
   CFL_UINT16 iRow;
   CFL_UINT16 iCol;

   RGT_LOG_ENTER("rgt_screen_fullToBuffer", (NULL));
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_fullToBuffer(): screen argument is null"));
      RGT_LOG_EXIT("rgt_screen_fullToBuffer", (NULL));
      return;
   }
   RGT_LOCK_ACQUIRE(screen->locked);
   cfl_buffer_putInt16(buffer, screen->height);
   cfl_buffer_putInt16(buffer, screen->width);
   cfl_buffer_putInt16(buffer, screen->cursorStyle);
   cfl_buffer_putInt16(buffer, screen->cursorRow);
   cfl_buffer_putInt16(buffer, screen->cursorCol);
   if (screen->screenType == RGT_SCREEN_TYPE_SIMPLE) {
      PHB_CODEPAGE cdp = HB_GTSELF_HOSTCP(screen->pGT);
      for (iRow = 0; iRow < screen->height; iRow++) {
         for (iCol = 0; iCol < screen->width; iCol++) {
            charToBuffer(screen, cdp, iRow, iCol, buffer);
         }
      }
   } else {
      for (iRow = 0; iRow < screen->height; iRow++) {
         for (iCol = 0; iCol < screen->width; iCol++) {
            char16ToBuffer(screen, iRow, iCol, buffer);
         }
      }
   }
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_fullToBuffer", (NULL));
}

void rgt_screen_fromBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer) {
   CFL_UINT16 iRow;
   CFL_UINT16 iCol;
   HB_USHORT character;
   int color;
   HB_BYTE attr;

   RGT_LOG_ENTER("rgt_screen_fromBuffer", (NULL));
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_fromBuffer(): screen argument is null"));
      RGT_LOG_EXIT("rgt_screen_fromBuffer", (NULL));
      return;
   }
   RGT_LOCK_ACQUIRE(screen->locked);
   screen->cursorStyle = cfl_buffer_getInt16(buffer);
   screen->cursorRow = cfl_buffer_getInt16(buffer);
   screen->cursorCol = cfl_buffer_getInt16(buffer);
   screen->topUpdate = cfl_buffer_getUInt16(buffer);
   screen->leftUpdate = cfl_buffer_getUInt16(buffer);
   screen->bottomUpdate = cfl_buffer_getUInt16(buffer);
   screen->rightUpdate = cfl_buffer_getUInt16(buffer);
   screen->changed = CFL_TRUE;
   if (!HB_GTSELF_LOCK(screen->pGT)) {
      RGT_LOCK_RELEASE(screen->locked);
      RGT_LOG_EXIT("rgt_screen_fromBuffer", (NULL));
      return;
   }
   if (screen->screenType == RGT_SCREEN_TYPE_SIMPLE) {
      PHB_CODEPAGE cdp = HB_GTSELF_HOSTCP(screen->pGT);
      for (iRow = screen->topUpdate; iRow <= screen->bottomUpdate; iRow++) {
         for (iCol = screen->leftUpdate; iCol <= screen->rightUpdate; iCol++) {
            character = (HB_USHORT)cfl_buffer_getUInt8(buffer);
            color = (int)cfl_buffer_getUInt8(buffer);
            HB_GTSELF_PUTCHAR(screen->pGT, iRow, iCol, color, 0, hb_cdpGetU16(cdp, (HB_UCHAR)character));
         }
      }
   } else {
      for (iRow = screen->topUpdate; iRow <= screen->bottomUpdate; iRow++) {
         for (iCol = screen->leftUpdate; iCol <= screen->rightUpdate; iCol++) {
            character = (HB_USHORT)cfl_buffer_getUInt16(buffer);
            color = (int)cfl_buffer_getUInt8(buffer);
            attr = (HB_BYTE)cfl_buffer_getUInt8(buffer);
            HB_GTSELF_PUTCHAR(screen->pGT, iRow, iCol, color, 0, character);
         }
      }
   }
   HB_GTSELF_SETCURSORSTYLE(screen->pGT, (int)screen->cursorStyle);
   HB_GTSELF_SETPOS(screen->pGT, (int)screen->cursorRow, (int)screen->cursorCol);
   HB_GTSELF_REFRESH(screen->pGT);
   HB_GTSELF_UNLOCK(screen->pGT);
   screen->topUpdate = screen->height - 1;
   screen->leftUpdate = screen->width - 1;
   screen->bottomUpdate = 0;
   screen->rightUpdate = 0;
   screen->changed = CFL_FALSE;
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_fromBuffer", (NULL));
}

void rgt_screen_fullUpdated(RGT_SCREENP screen) {
   int iRow;
   int iCol;
   int iCursor;

   RGT_LOG_ENTER("rgt_screen_fullUpdated", (" rows=%d, cols=%d", screen->height, screen->width));

   RGT_LOCK_ACQUIRE(screen->locked);
   screen->topUpdate = 0;
   screen->leftUpdate = 0;
   screen->bottomUpdate = screen->height - 1;
   screen->rightUpdate = screen->width - 1;
   screen->changed = CFL_TRUE;
   hb_gtGetPos(&iRow, &iCol);
   screen->cursorRow = (CFL_INT16)iRow;
   screen->cursorCol = (CFL_INT16)iCol;
   hb_gtGetCursor(&iCursor);
   screen->cursorStyle = (CFL_INT16)iCursor;
   RGT_LOCK_RELEASE(screen->locked);

   RGT_LOG_EXIT("rgt_screen_fullUpdated", (NULL));
}

CFL_INT8 rgt_screen_type(void) {
   HB_BOOL simpleScreenBuffer;
   HB_GT_INFO gtInfo;

   memset(&gtInfo, 0, sizeof(gtInfo));
   hb_gtInfo(HB_GTI_COMPATBUFFER, &gtInfo);

   if (gtInfo.pResult) {
      simpleScreenBuffer = hb_itemGetL(gtInfo.pResult);
      hb_itemRelease(gtInfo.pResult);
   } else {
      simpleScreenBuffer = HB_FALSE;
   }
   return simpleScreenBuffer ? RGT_SCREEN_TYPE_SIMPLE : RGT_SCREEN_TYPE_EXTENDED;
}

void rgt_screen_setType(CFL_INT8 newScreenType) {
   HB_GT_INFO gtInfo;
   memset(&gtInfo, 0, sizeof(gtInfo));
   gtInfo.pNewVal = hb_itemPutL(NULL, newScreenType == RGT_SCREEN_TYPE_SIMPLE ? HB_TRUE : HB_FALSE);
   hb_gtInfo(HB_GTI_COMPATBUFFER, &gtInfo);

   hb_itemRelease(gtInfo.pNewVal);
   if (gtInfo.pResult) {
      hb_itemRelease(gtInfo.pResult);
   }
}

void rgt_screen_setCursorPos(RGT_SCREENP screen, CFL_INT16 row, CFL_INT16 col) {
   if (screen != NULL) {
      RGT_LOCK_ACQUIRE(screen->locked);
      if (screen->cursorRow != row || screen->cursorCol != col) {
         screen->cursorRow = row;
         screen->cursorCol = col;
         screen->changed = CFL_TRUE;
      }
      RGT_LOCK_RELEASE(screen->locked);
   } else {
      RGT_LOG_ERROR(("rgt_screen_setCursorPos(): screen argument is null"));
   }
}

void rgt_screen_setCursorStyle(RGT_SCREENP screen, CFL_INT16 style) {
   if (screen != NULL) {
      RGT_LOCK_ACQUIRE(screen->locked);
      if (screen->cursorStyle != style) {
         screen->cursorStyle = style;
         screen->changed = CFL_TRUE;
      }
      RGT_LOCK_RELEASE(screen->locked);
   }
}

CFL_INT16 rgt_screen_getCursorRow(RGT_SCREENP screen) {
   if (screen != NULL) {
      CFL_INT16 row;
      RGT_LOCK_ACQUIRE(screen->locked);
      row = screen->cursorRow;
      RGT_LOCK_RELEASE(screen->locked);
      return row;
   } else {
      RGT_LOG_ERROR(("rgt_screen_getCursorRow(): screen argument is null"));
      return 0;
   }
}

CFL_INT16 rgt_screen_getCursorCol(RGT_SCREENP screen) {
   if (screen != NULL) {
      CFL_INT16 col;
      RGT_LOCK_ACQUIRE(screen->locked);
      col = screen->cursorCol;
      RGT_LOCK_RELEASE(screen->locked);
      return col;
   } else {
      RGT_LOG_ERROR(("rgt_screen_getCursorCol(): screen argument is null"));
      return 0;
   }
}

CFL_INT16 rgt_screen_getCursorStyle(RGT_SCREENP screen) {
   if (screen != NULL) {
      CFL_INT16 style;
      RGT_LOCK_ACQUIRE(screen->locked);
      style = screen->cursorStyle;
      RGT_LOCK_RELEASE(screen->locked);
      return style;
   } else {
      RGT_LOG_ERROR(("rgt_screen_getCursorStyle(): screen argument is null"));
      return 0;
   }
}

CFL_BOOL rgt_screen_isChanged(RGT_SCREENP screen) {
   if (screen != NULL) {
      CFL_BOOL changed;
      RGT_LOCK_ACQUIRE(screen->locked);
      changed = screen->changed;
      RGT_LOCK_RELEASE(screen->locked);
      // RGT_LOG_DEBUG(("rgt_screen_isChanged(): changed? %s", (changed ? "YES" : "NO")));
      return changed;
   } else {
      RGT_LOG_ERROR(("rgt_screen_isChanged(): screen argument is null"));
      return CFL_FALSE;
   }
}
