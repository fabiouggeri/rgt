#ifndef _RGT_LOG_H
#define _RGT_LOG_H

#include <time.h>

#include "cfl_buffer.h"
#include "cfl_list.h"
#include "cfl_str.h"
#include "hbapi.h"

extern clock_t _rgtTraceLastClock;

extern unsigned int __rgtLoglevel;

#define RGT_LOG_LEVEL_OFF 0
#define RGT_LOG_LEVEL_ERROR 1
#define RGT_LOG_LEVEL_WARN 2
#define RGT_LOG_LEVEL_INFO 3
#define RGT_LOG_LEVEL_DEBUG 4
#define RGT_LOG_LEVEL_TRACE 5

#define RGT_LOG_LEVEL_VAR "RGT_LOG_LEVEL"
#define RGT_LOG_PATH_NAME_VAR "RGT_LOG_PATH_NAME"

#define LOG_OUT_FILE(l, x) rgt_log_write(l, rgt_log_format x)

#define __DEBUG(x)                                                                                                                 \
   printf x;                                                                                                                       \
   getch()
#define __TRACE(f, x)                                                                                                              \
   {                                                                                                                               \
      FILE *v;                                                                                                                     \
      v = fopen(f, "a");                                                                                                           \
      fprintf(v, x);                                                                                                               \
      fclose(v);                                                                                                                   \
   }

// #define ACTIVE_TRACE

#ifdef ACTIVE_TRACE
#undef RGT_LOG_LEVEL_TRACE
#define RGT_LOG_LEVEL_TRACE 0
#endif

#define RGT_LOG(l, x)                                                                                                              \
   if (__rgtLoglevel >= l)                                                                                                         \
   rgt_log_write(l, x)
#define RGT_LOG_ERROR(x)                                                                                                           \
   if (__rgtLoglevel >= RGT_LOG_LEVEL_ERROR)                                                                                       \
   LOG_OUT_FILE(RGT_LOG_LEVEL_ERROR, x)
#define RGT_LOG_WARN(x)                                                                                                            \
   if (__rgtLoglevel >= RGT_LOG_LEVEL_WARN)                                                                                        \
   LOG_OUT_FILE(RGT_LOG_LEVEL_WARN, x)
#define RGT_LOG_INFO(x)                                                                                                            \
   if (__rgtLoglevel >= RGT_LOG_LEVEL_INFO)                                                                                        \
   LOG_OUT_FILE(RGT_LOG_LEVEL_INFO, x)
#define RGT_LOG_DEBUG(x)                                                                                                           \
   if (__rgtLoglevel >= RGT_LOG_LEVEL_DEBUG)                                                                                       \
   LOG_OUT_FILE(RGT_LOG_LEVEL_DEBUG, x)
#define RGT_LOG_TRACE(x)                                                                                                           \
   if (__rgtLoglevel >= RGT_LOG_LEVEL_TRACE)                                                                                       \
   LOG_OUT_FILE(RGT_LOG_LEVEL_TRACE, x)
#if !defined(RGT_TRACE_FUNCTIONS_OFF)
#define RGT_LOG_ENTER(f, x)                                                                                                        \
   if (__rgtLoglevel >= RGT_LOG_LEVEL_TRACE)                                                                                       \
   rgt_log_writeEnter(f, __LINE__, rgt_log_format x)
#define RGT_LOG_EXIT(f, x)                                                                                                         \
   if (__rgtLoglevel >= RGT_LOG_LEVEL_TRACE)                                                                                       \
   rgt_log_writeExit(f, __LINE__, rgt_log_format x)
#else
#define RGT_LOG_ENTER(f, x)
#define RGT_LOG_EXIT(f, x)
#endif
#define RGT_LOG_PARAM(l, p, i)                                                                                                     \
   if (__rgtLoglevel >= l)                                                                                                         \
   rgt_log_param(l, p, i)
#define RGT_LOG_ITEM(l, n, i)                                                                                                      \
   if (__rgtLoglevel >= l)                                                                                                         \
   rgt_log_item(l, n, i)

#define RGT_LOG_IS_LEVEL(l) (__rgtLoglevel >= l)

extern CFL_STRP rgt_log_format(const char *format, ...);
extern void rgt_log_setLabel(const char *label);
extern int rgt_log_getLevel(void);
extern void rgt_log_setLevel(int level);
extern void rgt_log_write(int level, CFL_STRP out);
extern void rgt_log_writeEnter(const char *funName, CFL_UINT32 line, CFL_STRP out);
extern void rgt_log_writeExit(const char *funName, CFL_UINT32 line, CFL_STRP out);
extern struct tm *rgt_log_localtime(void);
extern void rgt_log_setPathName(const char *logPathName);
extern char *rgt_log_getPathName(void);
extern void rgt_log_param(unsigned int level, CFL_INT16 iPos, PHB_ITEM pItem);
extern void rgt_log_item(unsigned int level, char *itemName, PHB_ITEM pItem);
extern CFL_BOOL rgt_log_envLevel(void);
extern CFL_BOOL rgt_log_envFile(void);
extern void rgt_log_initEnv(void);

#endif
