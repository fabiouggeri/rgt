#include "hbset.h"

#ifdef __linux__
   #include <sys/time.h>
#else
   #include "windows.h"
#endif

#include "rgt_util.h"

#define EPOCH ((CFL_UINT64) 116444736000000000ULL)
#define MAX_INT32 4294967295ULL

#ifdef __XHB__

static BOOL set_logical(PHB_ITEM pItem, BOOL bDefault) {
   HB_BOOL bLogical = bDefault;

   if (HB_IS_LOGICAL(pItem)) {
      bLogical = hb_itemGetL(pItem);
   } else if (HB_IS_STRING(pItem)) {
      char * szString = pItem->item.asString.value;
      HB_ULONG ulLen = pItem->item.asString.length;

      if (ulLen >= 2 && toupper(szString[ 0 ]) == 'O' && toupper(szString[ 1 ]) == 'N') {
         bLogical = HB_TRUE;
      } else if (ulLen >= 3 && toupper(szString[ 0 ]) == 'O' && toupper(szString[ 1 ]) == 'F' && toupper(szString[ 2 ]) == 'F') {
         bLogical = HB_FALSE;
      }
   }

   return bLogical;
}

static int set_number(PHB_ITEM pItem, int iOldValue) {
   return HB_IS_NUMERIC(pItem) ? hb_itemGetNI(pItem) : iOldValue;
}

int hb_setGetTypeAhead(void) {
   return hb_set.HB_SET_TYPEAHEAD;
}

HB_BOOL hb_setGetCancel(void) {
   return hb_set.HB_SET_CANCEL;
}

HB_BOOL hb_setGetDebug(void) {
   return hb_set.HB_SET_DEBUG;
}

PHB_ITEM hb_setGetItem(int set_specifier, PHB_ITEM pResult, PHB_ITEM pArg1, PHB_ITEM pArg2) {
   HB_SYMBOL_UNUSED(pArg2);
   switch (set_specifier) {
      case HB_SET_TYPEAHEAD:
         if (pResult) {
            hb_itemPutNI(pResult, hb_set.HB_SET_TYPEAHEAD);
         }

         if (pArg1) {
            /* Set the value and limit the range */
            int old = hb_set.HB_SET_TYPEAHEAD;
            hb_set.HB_SET_TYPEAHEAD = set_number(pArg1, old);
            if (hb_set.HB_SET_TYPEAHEAD == 0) {
               /* Do nothing */;
            } else if (hb_set.HB_SET_TYPEAHEAD < 16) {
               hb_set.HB_SET_TYPEAHEAD = 16;
            } else if (hb_set.HB_SET_TYPEAHEAD > 4096) {
               hb_set.HB_SET_TYPEAHEAD = 4096;
            }

            /* Always reset the buffer, but only reallocate if the size changed */
            hb_inkeyReset(old == hb_set.HB_SET_TYPEAHEAD ? FALSE : TRUE);
         }
         break;
      case HB_SET_CANCEL:
         if (pResult) {
            hb_itemPutL(pResult, hb_set.HB_SET_CANCEL);
         }
         if (pArg1) {
            hb_set.HB_SET_CANCEL = set_logical(pArg1, hb_set.HB_SET_CANCEL);
         }
         break;
      case HB_SET_DEBUG:
         if (pResult) {
            hb_itemPutL(pResult, hb_set.HB_SET_DEBUG);
         }
         if (pArg1) {
            hb_set.HB_SET_DEBUG = set_logical(pArg1, hb_set.HB_SET_DEBUG);
         }
         break;
   }
   return pResult;
}

PHB_SYMB hb_dynsymFindSymbol( const char * szName ) {
   PHB_DYNS pDynSym = hb_dynsymFind( szName );
   return pDynSym ? pDynSym->pSymbol : NULL;
}

PHB_ITEM hb_itemPutSymbol( PHB_ITEM pItem, PHB_SYMB pSym ) {
   if( pItem ) {
      if( HB_IS_COMPLEX( pItem ) )
         hb_itemClear( pItem );
   } else {
      pItem = hb_itemNew( NULL );
   }
   pItem->type = HB_IT_SYMBOL;
   pItem->item.asSymbol.value = pSym;
   pItem->item.asSymbol.uiSuperClass = 0;
   pItem->item.asSymbol.paramcnt = 0;
   return pItem;
}

#endif

#ifdef _WIN32

/*
 * timezone information is stored outside the kernel so tzp isn't used anymore.
 */
int gettimeofday(struct timeval * tp, struct timezone * tzp) {
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
      tp->tv_sec = (long) (time / 1000000UL);
      tp->tv_usec = (long) (time % 1000000UL);
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
         return (CFL_UINT32) MAX_INT32;
      } else {
         return (CFL_UINT32) ellapsed;
      }
   } else {
      return (CFL_UINT32) 0;
   }
}
