#ifndef _RGT_SCREEN_H

#define _RGT_SCREEN_H

#include "cfl_buffer.h"
#include "cfl_lock.h"
#include "rgt_types.h"

#include "hbgtcore.h"

struct _RGT_SCREEN {
      CFL_INT8 screenType;
      CFL_UINT16 height;
      CFL_UINT16 width;
      CFL_UINT16 topUpdate;
      CFL_UINT16 leftUpdate;
      CFL_UINT16 bottomUpdate;
      CFL_UINT16 rightUpdate;
      CFL_INT16 cursorRow;
      CFL_INT16 cursorCol;
      CFL_INT16 cursorStyle;
      CFL_INT16 lastCursorRow;
      CFL_INT16 lastCursorCol;
      CFL_INT16 lastCursorStyle;
      CFL_BOOL changed;
      RGT_LOCK locked;
      PHB_GT pGT;
};

extern RGT_SCREENP rgt_screen_new(CFL_UINT16 rows, CFL_UINT16 cols, CFL_INT8 screenType);
extern void rgt_screen_reset(RGT_SCREENP screen, CFL_UINT16 rows, CFL_UINT16 cols, CFL_INT8 screenType);
extern void rgt_screen_free(RGT_SCREENP screen);
extern void rgt_screen_rectUpdated(RGT_SCREENP screen, CFL_UINT16 rowIni, CFL_UINT16 colIni, CFL_UINT16 rowEnd, CFL_UINT16 colEnd);

extern void rgt_screen_clear(RGT_SCREENP screen);
extern CFL_BOOL rgt_screen_toBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer, CFL_BOOL force);
extern void rgt_screen_fullToBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer);
extern void rgt_screen_fromBuffer(RGT_SCREENP screen, CFL_BUFFERP buffer);
extern void rgt_screen_fullUpdated(RGT_SCREENP screen);
extern CFL_INT8 rgt_screen_type(void);
extern void rgt_screen_setType(CFL_INT8 newScreenType);
extern void rgt_screen_setCursorPos(RGT_SCREENP screen, CFL_INT16 row, CFL_INT16 col);
extern void rgt_screen_setCursorStyle(RGT_SCREENP screen, CFL_INT16 style);
extern CFL_INT16 rgt_screen_getCursorRow(RGT_SCREENP screen);
extern CFL_INT16 rgt_screen_getCursorCol(RGT_SCREENP screen);
extern CFL_INT16 rgt_screen_getCursorStyle(RGT_SCREENP screen);
extern CFL_BOOL rgt_screen_isChanged(RGT_SCREENP screen);

#endif
