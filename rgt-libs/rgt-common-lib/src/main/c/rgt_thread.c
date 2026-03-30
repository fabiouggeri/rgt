#include "hbapiitm.h"
#include "hbvm.h"

#include "rgt_log.h"
#include "rgt_thread.h"

#ifdef __HBR__
HB_FUNC_EXTERN(HB_THREADQUITREQUEST);
HB_FUNC_EXTERN(HB_THREADWAIT);
extern HB_EXPORT void hb_threadReleaseCPU(void);
extern PHB_ITEM hb_threadStart(HB_ULONG ulAttr, PHB_CARGO_FUNC pFunc, void *cargo);
#endif

static CFL_UINT8 s_threadType = RGT_THREAD_CFL;

static void executeFunction(void *param) {
   RGT_THREADP thread;
   RGT_LOG_ENTER("threadExecuteFunction", (NULL));
   thread = (RGT_THREADP)param;
   thread->running = CFL_TRUE;
   thread->func(thread->param);
   thread->running = CFL_FALSE;
   RGT_LOG_EXIT("threadExecuteFunction", (NULL));
}

void rgt_thread_setType(CFL_UINT8 threadType) {
   if (threadType <= RGT_THREAD_TYPE_MAX) {
      s_threadType = threadType;
   }
}

CFL_UINT8 rgt_thread_getType(void) {
   return s_threadType;
}

RGT_THREADP rgt_thread_start(RGT_THREAD_FUNC func, void *param, const char *description) {
   RGT_THREADP thread;
   RGT_LOG_ENTER("rgt_thread_start", (NULL));
   thread = RGT_HB_ALLOC(sizeof(RGT_THREAD));
   thread->threadType = s_threadType;
   thread->running = CFL_FALSE;
   thread->func = func;
   thread->param = param;
#ifdef __HBR__
   if (s_threadType == RGT_THREAD_CFL) {
      thread->handle.cflThread = cfl_thread_newWithDescription(executeFunction, description);
      if (thread->handle.cflThread == NULL || !cfl_thread_start(thread->handle.cflThread, thread)) {
         RGT_HB_FREE(thread);
         thread = NULL;
      }
   } else {
      thread->handle.hbThread = hb_threadStart(0, executeFunction, thread);
      if (thread->handle.hbThread == NULL) {
         RGT_HB_FREE(thread);
         thread = NULL;
      }
   }
#else
   thread->handle.cflThread = cfl_thread_newWithDescription(executeFunction, description);
   if (thread->handle.cflThread == NULL || !cfl_thread_start(thread->handle.cflThread, thread)) {
      RGT_HB_FREE(thread);
      thread = NULL;
   }
#endif
   RGT_LOG_EXIT("rgt_thread_start", ("success=%s", thread != NULL ? "true" : "false"));
   return thread;
}

void rgt_thread_free(RGT_THREADP thread) {
   RGT_LOG_ENTER("rgt_thread_free", (NULL));
   if (thread == NULL) {
      RGT_LOG_EXIT("rgt_thread_free", (NULL));
      return;
   }
#ifdef __HBR__
   if (thread->threadType == RGT_THREAD_CFL) {
      cfl_thread_free(thread->handle.cflThread);
   } else {
      hb_itemRelease(thread->handle.hbThread);
   }
#else
   cfl_thread_free(thread->handle.cflThread);
#endif
   RGT_HB_FREE(thread);
   RGT_LOG_EXIT("rgt_thread_free", (NULL));
}

void rgt_thread_sleep(CFL_UINT32 time) {
   cfl_thread_sleep(time);
}

CFL_BOOL rgt_thread_isRunning(RGT_THREADP thread) {
   return thread->running;
}

void rgt_thread_kill(RGT_THREADP thread) {
   RGT_LOG_ENTER("rgt_thread_kill", (NULL));
#ifdef __HBR__
   if (thread->running) {
      if (thread->threadType == RGT_THREAD_CFL) {
         cfl_thread_kill(thread->handle.cflThread);
      } else {
         hb_itemDoC("HB_THREADQUITREQUEST", 1, thread->handle.hbThread);
      }
   }
#else
   cfl_thread_kill(thread->handle.cflThread);
#endif
   RGT_LOG_EXIT("rgt_thread_kill", (NULL));
}

void rgt_thread_waitTimeout(RGT_THREADP thread, CFL_INT32 timeout) {
   RGT_LOG_ENTER("rgt_thread_waitTimeout", (NULL));
#ifdef __HBR__
   if (thread->threadType == RGT_THREAD_CFL) {
      cfl_thread_waitTimeout(thread->handle.cflThread, timeout);
   } else {
      PHB_ITEM pTimeout = hb_itemPutNI(NULL, (int)timeout);
      hb_itemDoC("HB_THREADWAIT", 2, thread->handle.hbThread, pTimeout);
      hb_itemRelease(pTimeout);
   }
#else
   cfl_thread_waitTimeout(thread->handle.cflThread, timeout);
#endif
   RGT_LOG_EXIT("rgt_thread_waitTimeout", (NULL));
}

void rgt_thread_yield(RGT_THREADP thread) {
#ifdef __HBR__
   if (thread->threadType == RGT_THREAD_CFL) {
      cfl_thread_yield();
   } else {
      hb_threadReleaseCPU();
   }
#else
   HB_SYMBOL_UNUSED(thread);
   cfl_thread_yield();
#endif
}

void rgt_thread_initEnv(void) {
   char *threadType = getenv(RGT_THREAD_TYPE_VAR);
   if (threadType != NULL) {
      if (hb_stricmp(threadType, "HB") == 0) {
         s_threadType = RGT_THREAD_HB;
      } else if (hb_stricmp(threadType, "CFL") == 0) {
         s_threadType = RGT_THREAD_CFL;
      }
   }
}

char *rgt_thread_getDescription(RGT_THREADP thread) {
#ifdef __HBR__
   if (thread->running) {
      if (thread->threadType == RGT_THREAD_CFL) {
         return cfl_str_getPtr(cfl_thread_getDescription(thread->handle.cflThread));
      } else {
         return "";
      }
   }
#else
   return cfl_str_getPtr(cfl_thread_getDescription(thread->handle.cflThread));
#endif
}
