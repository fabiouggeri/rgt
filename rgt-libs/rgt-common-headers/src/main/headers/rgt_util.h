#ifndef _RGT_UTIL_H_

#define _RGT_UTIL_H_

#include <time.h>

#include "hbapi.h"
#include "rgt_types.h"


#define MILLISEC 1
#define SECOND (1000 * MILLISEC)
#define MINUTE (60 * SECOND)
#define HOUR (60 * MINUTE)
#define DAY (24 * HOUR)
#define MILLIS_PER_SEC 1000
#define TIMEMILLIS_ELAPSED(t1, t2) rgt_elapsed_timemillis(t1, t2)
#define CURRENT_TIME rgt_current_timemillis()

#ifdef _MSC_VER
extern int gettimeofday(struct timeval *tp, struct timezone *tzp);
#endif

extern CFL_UINT64 rgt_current_timemillis(void);
extern CFL_UINT32 rgt_elapsed_timemillis(CFL_UINT64 t1, CFL_UINT64 t2);

#endif
