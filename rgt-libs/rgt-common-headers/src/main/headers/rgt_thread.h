#ifndef _RGT_THREAD_H_
#define _RGT_THREAD_H_

#include "hbapi.h"

#include "cfl_thread.h"

#include "rgt_types.h"

#define RGT_THREAD_CFL 0
#define RGT_THREAD_HB 1
#define RGT_THREAD_TYPE_MIN RGT_THREAD_CFL
#define RGT_THREAD_TYPE_MAX RGT_THREAD_HB

#define RGT_THREAD_TYPE_VAR "RGT_THREAD_TYPE"

typedef void (*RGT_THREAD_FUNC)(void *param);

struct _RGT_THREAD {
      union {
            PHB_ITEM hbThread;
            CFL_THREADP cflThread;
      } handle;
      RGT_THREAD_FUNC func;
      void *param;
      CFL_UINT8 threadType;
      CFL_BOOL running;
};

extern RGT_THREADP rgt_thread_start(RGT_THREAD_FUNC func, void *param, const char *description);
extern void rgt_thread_free(RGT_THREADP thread);
extern void rgt_thread_sleep(CFL_UINT32 time);
extern CFL_BOOL rgt_thread_isRunning(RGT_THREADP thread);
extern void rgt_thread_kill(RGT_THREADP thread);
extern void rgt_thread_waitTimeout(RGT_THREADP thread, CFL_INT32 timeout);
extern void rgt_thread_yield(RGT_THREADP thread);
extern void rgt_thread_setType(CFL_UINT8 threadType);
extern CFL_UINT8 rgt_thread_getType(void);
extern void rgt_thread_initEnv(void);
extern char *rgt_thread_getDescription(RGT_THREADP thread);

#endif