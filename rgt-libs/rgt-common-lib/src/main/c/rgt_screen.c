#include <stdlib.h>

#include "cfl_types.h"

#ifdef __XHB__
#include "hbapigt.h"
#include "hbapiitm.h"
#else
#include "hbapicdp.h"
#include "hbapigt.h"
#include "hbapiitm.h"
#include "hbgtcore.h"
#include "hbstack.h"

#endif

#include "cfl_buffer.h"
#include "cfl_lock.h"
#include "rgt_log.h"
#include "rgt_screen.h"


extern HB_ERRCODE rgt_error_launch(CFL_UINT16 uiGenCode, CFL_UINT16 uiSubCode, CFL_UINT16 uiFlags, const char *description,
                                   const char *operation, PHB_ITEM *pErrorPtr);

RGT_SCREENP rgt_screen_new(CFL_UINT16 rows, CFL_UINT16 cols, CFL_INT8 screenType) {
   RGT_SCREENP screen;
   HB_SIZE bufferLen = (rows + 1) * (cols + 1) * sizeof(RGT_SCREEN_CELL);
   RGT_LOG_ENTER("rgt_screen_new",
                 ("rows=%u col=%u type=%s", rows, cols, (screenType == RGT_SCREEN_TYPE_SIMPLE ? "simple" : "extended")));
   screen = (RGT_SCREENP)RGT_HB_ALLOC(sizeof(RGT_SCREEN));
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_new(): error allocating RGT_SCREEN"));
      RGT_LOG_EXIT("rgt_screen_new", (NULL));
      return NULL;
   }
   screen->buffer = (RGT_SCREEN_CELLP)RGT_HB_ALLOC(bufferLen);
   if (screen->buffer == NULL) {
      RGT_LOG_ERROR(("rgt_screen_new(): error allocating screen buffer"));
      RGT_HB_FREE(screen);
      RGT_LOG_EXIT("rgt_screen_new", (NULL));
      return NULL;
   }
   memset(screen->buffer, 0, bufferLen);
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
#ifdef __HBR__
   screen->pGT = hb_stackGetGT();
#endif
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
      HB_SIZE bufferLen = (rows + 1) * (cols + 1) * sizeof(RGT_SCREEN_CELL);
      screen->height = rows;
      screen->width = cols;
      screen->screenType = screenType;
      screen->buffer = (RGT_SCREEN_CELLP)RGT_HB_REALLOC(screen->buffer, bufferLen);
      if (screen->buffer != NULL) {
         memset(screen->buffer, 0, bufferLen);
      } else {
         RGT_LOG_ERROR(("rgt_screen_reset(): error allocating screen buffer"));
      }
   }
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_reset", (NULL));
}

void rgt_screen_free(RGT_SCREENP screen) {
   if (screen != NULL) {
      if (screen->buffer != NULL) {
         RGT_HB_FREE(screen->buffer);
      }
      RGT_LOCK_FREE(screen->locked);
      RGT_HB_FREE(screen);
   } else {
      RGT_LOG_ERROR(("rgt_screen_free(): screen argument is null"));
   }
}

void rgt_screen_putChar(RGT_SCREENP screen, CFL_UINT16 iRow, CFL_UINT16 iCol, CFL_UINT16 character, CFL_UINT8 color,
                        CFL_UINT8 attr) {
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_putChar(): screen argument is null"));
   } else if (iRow < screen->height && iCol < screen->width) {
      CFL_UINT32 iIndex;
      iIndex = iRow * screen->width + iCol;
      if (screen->buffer[iIndex].character != character || screen->buffer[iIndex].color != color ||
          screen->buffer[iIndex].attr != attr) {
         RGT_LOCK_ACQUIRE(screen->locked);
         screen->buffer[iIndex].character = character;
         screen->buffer[iIndex].color = color;
         screen->buffer[iIndex].attr = attr;
         if (iRow < screen->topUpdate) {
            screen->topUpdate = iRow;
         }
         if (iRow > screen->bottomUpdate) {
            screen->bottomUpdate = iRow;
         }
         if (iCol < screen->leftUpdate) {
            screen->leftUpdate = iCol;
         }
         if (iCol > screen->rightUpdate) {
            screen->rightUpdate = iCol;
         }
         screen->changed = CFL_TRUE;
         RGT_LOCK_RELEASE(screen->locked);
      }
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
      memset(screen->buffer, 0, (screen->height + 1) * (screen->width + 1) * sizeof(RGT_SCREEN_CELL));
      RGT_LOCK_RELEASE(screen->locked);
   } else {
      RGT_LOG_ERROR(("rgt_screen_clear(): screen argument is null"));
   }
   RGT_LOG_EXIT("rgt_screen_clear", (NULL));
}

CFL_BOOL rgt_screen_toBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer, CFL_BOOL force) {
   CFL_UINT32 iIndex;
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
      for (iRow = topUpdate; iRow <= bottomUpdate; iRow++) {
         for (iCol = leftUpdate; iCol <= rightUpdate; iCol++) {
            iIndex = iRow * screen->width + iCol;
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)(0xFF & screen->buffer[iIndex].character));
            cfl_buffer_putUInt8(buffer, screen->buffer[iIndex].color);
         }
      }
   } else {
      for (iRow = topUpdate; iRow <= bottomUpdate; iRow++) {
         for (iCol = leftUpdate; iCol <= rightUpdate; iCol++) {
            iIndex = iRow * screen->width + iCol;
            cfl_buffer_putUInt16(buffer, screen->buffer[iIndex].character);
            cfl_buffer_putUInt8(buffer, screen->buffer[iIndex].color);
            cfl_buffer_putUInt8(buffer, screen->buffer[iIndex].attr);
         }
      }
   }
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_toBuffer", (NULL));
   return CFL_TRUE;
}

void rgt_screen_fullToBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer) {
   CFL_UINT32 iIndex;
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
      for (iRow = 0; iRow < screen->height; iRow++) {
         for (iCol = 0; iCol < screen->width; iCol++) {
            iIndex = iRow * screen->width + iCol;
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)(0xFF & screen->buffer[iIndex].character));
            cfl_buffer_putUInt8(buffer, screen->buffer[iIndex].color);
         }
      }
   } else {
      for (iRow = 0; iRow < screen->height; iRow++) {
         for (iCol = 0; iCol < screen->width; iCol++) {
            iIndex = iRow * screen->width + iCol;
            cfl_buffer_putUInt16(buffer, screen->buffer[iIndex].character);
            cfl_buffer_putUInt8(buffer, screen->buffer[iIndex].color);
            cfl_buffer_putUInt8(buffer, screen->buffer[iIndex].attr);
         }
      }
   }
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_fullToBuffer", (NULL));
}

void rgt_screen_fromBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer) {
   CFL_UINT32 iIndex;
   CFL_UINT16 iRow;
   CFL_UINT16 iCol;

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
   if (screen->screenType == RGT_SCREEN_TYPE_SIMPLE) {
      for (iRow = screen->topUpdate; iRow <= screen->bottomUpdate; iRow++) {
         for (iCol = screen->leftUpdate; iCol <= screen->rightUpdate; iCol++) {
            iIndex = iRow * screen->width + iCol;
            screen->buffer[iIndex].character = (CFL_UINT16)cfl_buffer_getUInt8(buffer);
            screen->buffer[iIndex].color = cfl_buffer_getUInt8(buffer);
            screen->buffer[iIndex].attr = 0;
         }
      }
   } else {
      for (iRow = screen->topUpdate; iRow <= screen->bottomUpdate; iRow++) {
         for (iCol = screen->leftUpdate; iCol <= screen->rightUpdate; iCol++) {
            iIndex = iRow * screen->width + iCol;
            screen->buffer[iIndex].character = cfl_buffer_getUInt16(buffer);
            screen->buffer[iIndex].color = cfl_buffer_getUInt8(buffer);
            screen->buffer[iIndex].attr = cfl_buffer_getUInt8(buffer);
         }
      }
   }
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_fromBuffer", (NULL));
}

void rgt_screen_capture(RGT_SCREENP screen) {
#ifdef __HBR__
   int iRow;
   int iCol;
   int iCursor;
   PHB_GT pGT = screen->pGT;
   PHB_CODEPAGE cdp;

   RGT_LOG_ENTER("rgt_screen_capture", (" rows=%d, cols=%d", screen->height, screen->width));

   if (pGT == NULL) {
      rgt_error_launch(30001, 21, 0, "Graphics Terminal not detected", "rgt_screen_capture", NULL);
      RGT_LOG_EXIT("rgt_screen_capture", (NULL));
      return;
   }

   cdp = pGT->fVgaCell ? HB_GTSELF_HOSTCP(pGT) : NULL;

   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_capture(): screen argument is null"));
      RGT_LOG_EXIT("rgt_screen_capture", (NULL));
      return;
   }

   RGT_LOCK_ACQUIRE(screen->locked);
   for (iRow = 0; iRow < screen->height; iRow++) {
      for (iCol = 0; iCol < screen->width; iCol++) {
         CFL_UINT32 iIndex = iRow * screen->width + iCol;
         int iColor;
         HB_BYTE attr;
         HB_USHORT character;
         if (HB_GTSELF_GETCHAR(pGT, iRow, iCol, &iColor, &attr, &character)) {
            if (pGT->fVgaCell) {
               screen->buffer[iIndex].character = hb_cdpGetChar(cdp, character);
               screen->buffer[iIndex].color = (CFL_UINT8)iColor;
            } else {
               screen->buffer[iIndex].character = character;
               screen->buffer[iIndex].color = (CFL_UINT8)iColor;
               screen->buffer[iIndex].attr = (CFL_UINT8)attr;
            }
         }
      }
   }
#else
   SHORT iRow;
   SHORT iCol;
   USHORT iCursor;
   int iPos;
   UINT iSize;
   HB_BYTE *buffer;

   RGT_LOG_ENTER("rgt_screen_capture", (" rows=%d, cols=%d", screen->height, screen->width));

   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_capture(): screen argument is null"));
      RGT_LOG_EXIT("rgt_screen_capture", (NULL));
      return;
   }

   RGT_LOCK_ACQUIRE(screen->locked);
   hb_gtRectSize(0, 0, screen->height - 1, screen->width - 1, &iSize);
   buffer = RGT_HB_ALLOC(iSize + 1);
   hb_gtSave(0, 0, screen->height - 1, screen->width - 1, (void *)buffer);
   iPos = 0;
   if (screen->screenType == RGT_SCREEN_TYPE_SIMPLE) {
      for (iRow = 0; iRow < screen->height; iRow++) {
         for (iCol = 0; iCol < screen->width; iCol++) {
            CFL_UINT32 iIndex = iRow * screen->width + iCol;
            screen->buffer[iIndex].character = (CFL_UINT16)buffer[iPos++];
            screen->buffer[iIndex].color = (CFL_UINT8)buffer[iPos++];
            screen->buffer[iIndex].attr = 0;
         }
      }
   } else {
      for (iRow = 0; iRow < screen->height; iRow++) {
         for (iCol = 0; iCol < screen->width; iCol++) {
            CFL_UINT32 iIndex = iRow * screen->width + iCol;
            screen->buffer[iIndex].character = *((CFL_UINT16 *)&buffer[iPos]);
            iPos += 2;
            screen->buffer[iIndex].color = (CFL_UINT8)buffer[iPos++];
            screen->buffer[iIndex].attr = (CFL_UINT8)buffer[iPos++];
         }
      }
   }
   RGT_HB_FREE(buffer);
#endif
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
   RGT_LOG_EXIT("rgt_screen_capture", (NULL));
}

void rgt_screen_draw(RGT_SCREENP screen) {
   int dispCount;
   CFL_UINT32 iIndex;
   CFL_UINT16 iRow;
   CFL_UINT16 iCol;
   CFL_UINT16 topUpdate;
   CFL_UINT16 leftUpdate;
   CFL_UINT16 bottomUpdate;
   CFL_UINT16 rightUpdate;
#ifdef __HBR__
   PHB_CODEPAGE cdp = hb_gtHostCP();
   PHB_GT pGT;
#else
   BYTE *buffer;
   BYTE *aux;
   int bufferLen;
#endif

   RGT_LOG_ENTER("rgt_screen_draw", (NULL));
   if (screen == NULL) {
      RGT_LOG_ERROR(("rgt_screen_draw(): screen argument is null"));
      RGT_LOG_EXIT("rgt_screen_draw", (NULL));
      return;
   }
   RGT_LOCK_ACQUIRE(screen->locked);
   topUpdate = screen->topUpdate;
   leftUpdate = screen->leftUpdate;
   bottomUpdate = screen->bottomUpdate;
   rightUpdate = screen->rightUpdate;
   screen->topUpdate = screen->height - 1;
   screen->leftUpdate = screen->width - 1;
   screen->bottomUpdate = 0;
   screen->rightUpdate = 0;
   screen->changed = CFL_FALSE;

#ifdef __HBR__
   pGT = (PHB_GT)screen->pGT;
   if (pGT == NULL) {
      RGT_LOCK_RELEASE(screen->locked);
      rgt_error_launch(30001, 21, 0, "Graphics Terminal not detected", "rgt_screen_draw", NULL);
      RGT_LOG_EXIT("rgt_screen_draw", (NULL));
      return;
   }

   if (!HB_GTSELF_LOCK(pGT)) {
      RGT_LOCK_RELEASE(screen->locked);
      RGT_LOG_EXIT("rgt_screen_draw", (NULL));
      return;
   }
   dispCount = HB_GTSELF_DISPCOUNT(pGT);
   HB_GTSELF_DISPBEGIN(pGT);

   for (iRow = topUpdate; iRow <= bottomUpdate; iRow++) {
      for (iCol = leftUpdate; iCol <= rightUpdate; iCol++) {
         iIndex = iRow * screen->width + iCol;
         if (screen->buffer[iIndex].character != 0) {
            HB_GTSELF_PUTCHAR(pGT, iRow, iCol, (int)screen->buffer[iIndex].color, (HB_BYTE)screen->buffer[iIndex].attr,
                              hb_cdpGetU16(cdp, (HB_UCHAR)screen->buffer[iIndex].character));
         }
      }
   }
   HB_GTSELF_SETCURSORSTYLE(pGT, (int)screen->cursorStyle);
   HB_GTSELF_SETPOS(pGT, (int)screen->cursorRow, (int)screen->cursorCol);
   while (HB_GTSELF_DISPCOUNT(pGT) > 0) {
      HB_GTSELF_DISPEND(pGT);
   }
   while (dispCount-- > 0) {
      HB_GTSELF_DISPBEGIN(pGT);
   }
   HB_GTSELF_FLUSH(pGT);
   HB_GTSELF_UNLOCK(pGT);
#else
   dispCount = hb_gtDispCount();
   hb_gtDispBegin();
   bufferLen = (bottomUpdate - topUpdate + 1) * (rightUpdate - leftUpdate + 1) * 2;
   if (bufferLen > 0) {
      buffer = (BYTE *)RGT_HB_ALLOC(bufferLen * sizeof(BYTE));
      hb_gt_GetText(topUpdate, leftUpdate, bottomUpdate, rightUpdate, buffer);
      aux = buffer;
      for (iRow = topUpdate; iRow <= bottomUpdate; iRow++) {
         for (iCol = leftUpdate; iCol <= rightUpdate; iCol++) {
            iIndex = iRow * screen->width + iCol;
            if (screen->buffer[iIndex].character != 0) {
               *aux++ = (BYTE)screen->buffer[iIndex].character;
               *aux++ = (BYTE)screen->buffer[iIndex].color;
            } else {
               aux++;
               aux++;
            }
         }
      }
      hb_gt_PutText(topUpdate, leftUpdate, bottomUpdate, rightUpdate, buffer);
      RGT_HB_FREE(buffer);
   }
   hb_gtSetCursor((int)screen->cursorStyle);
   hb_gtSetPos((int)screen->cursorRow, (int)screen->cursorCol);
   while (hb_gtDispCount() > 0) {
      hb_gtDispEnd();
   }
   while (dispCount-- > 0) {
      hb_gtDispBegin();
   }
#endif
   RGT_LOCK_RELEASE(screen->locked);
   RGT_LOG_EXIT("rgt_screen_draw", (NULL));
}

CFL_INT8 rgt_screen_type(void) {
#ifdef __HBR__
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
#else
   return RGT_SCREEN_TYPE_SIMPLE;
#endif
}

void rgt_screen_setType(CFL_INT8 newScreenType) {
#ifdef __HBR__
   HB_GT_INFO gtInfo;
   memset(&gtInfo, 0, sizeof(gtInfo));
   gtInfo.pNewVal = hb_itemPutL(NULL, newScreenType == RGT_SCREEN_TYPE_SIMPLE ? HB_TRUE : HB_FALSE);
   hb_gtInfo(HB_GTI_COMPATBUFFER, &gtInfo);

   hb_itemRelease(gtInfo.pNewVal);
   if (gtInfo.pResult) {
      hb_itemRelease(gtInfo.pResult);
   }
#else
   HB_SYMBOL_UNUSED(newScreenType);
#endif
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
