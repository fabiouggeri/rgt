#ifndef _RGT_UTIL_H_

#define _RGT_UTIL_H_

#include <time.h>

#include "rgt_types.h"
#include "hbapi.h"

#define MILLISEC                  1
#define SECOND                    (1000 * MILLISEC)
#define MINUTE                    (60 * SECOND)
#define HOUR                      (60 * MINUTE)
#define DAY                       (24 * HOUR)
#define MILLIS_PER_SEC            1000
#define TIMEMILLIS_ELAPSED(t1,t2) rgt_elapsed_timemillis(t1, t2)
#define CURRENT_TIME              rgt_current_timemillis()

#ifdef __XHB__

   extern int hb_setGetTypeAhead(void);
   extern HB_BOOL hb_setGetCancel(void);
   extern HB_BOOL hb_setGetDebug(void);
   extern PHB_ITEM hb_setGetItem(int set_specifier, PHB_ITEM pResult, PHB_ITEM pArg1, PHB_ITEM pArg2);
   extern PHB_SYMB hb_dynsymFindSymbol( const char * szName );
   extern PHB_ITEM hb_itemPutSymbol( PHB_ITEM pItem, PHB_SYMB pSym );

#endif

#ifdef _MSC_VER
   extern int gettimeofday(struct timeval * tp, struct timezone * tzp);
#endif

extern CFL_UINT64 rgt_current_timemillis(void);
extern CFL_UINT32 rgt_elapsed_timemillis(CFL_UINT64 t1, CFL_UINT64 t2);

#endif
