#include "hbset.h"

#ifdef __linux__
#include <sys/time.h>
#else
#include "windows.h"
#endif

#include "rgt_util.h"

#define EPOCH ((CFL_UINT64)116444736000000000ULL)
#define MAX_INT32 4294967295ULL

#ifdef _WIN32

/*
 * timezone information is stored outside the kernel so tzp isn't used anymore.
 */
int gettimeofday(struct timeval *tp, struct timezone *tzp) {
   FILETIME file_time;
   CFL_UINT64 time = 0ull;

   HB_SYMBOL_UNUSED(tzp);
   if (tp != NULL) {
      GetSystemTimeAsFileTime(&file_time);

      time |= file_time.dwHighDateTime;
      time <<= 32;
      time |= file_time.dwLowDateTime;

      /*converting file time to unix epoch*/
      time -= EPOCH;
      time /= 10; /*convert into microseconds*/
      tp->tv_sec = (long)(time / 1000000UL);
      tp->tv_usec = (long)(time % 1000000UL);
   }
   return 0;
}

#endif

CFL_UINT64 rgt_current_timemillis(void) {
   struct timeval tp;
   gettimeofday(&tp, NULL);
   return (((CFL_UINT64)tp.tv_sec) * MILLIS_PER_SEC) + (((CFL_UINT64)tp.tv_usec) / MILLIS_PER_SEC);
}

CFL_UINT32 rgt_elapsed_timemillis(CFL_UINT64 t1, CFL_UINT64 t2) {
   if (t2 > t1) {
      const CFL_UINT64 ellapsed = t2 - t1;
      if (ellapsed > MAX_INT32) {
         return (CFL_UINT32)MAX_INT32;
      } else {
         return (CFL_UINT32)ellapsed;
      }
   } else {
      return (CFL_UINT32)0;
   }
}
