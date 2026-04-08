#include "cfl_types.h"

#include "hbapicls.h"
#include "hbdate.h"

#include "hbmacro.ch"

#include "cfl_buffer.h"
#include "cfl_mem.h"
#include "cfl_str.h"
#include "rgt_common.h"
#include "rgt_error.h"
#include "rgt_log.h"
#include "rgt_rpc.h"
#include "rgt_util.h"

static PHB_ITEM evalMacro(PHB_ITEM pItem, const char *szExpr, HB_SIZE exprLen) {
   const char *type;

   pItem = hb_itemPutCL(pItem, szExpr, exprLen);
   type = hb_macroGetType(pItem);
   if (strcmp(type, "U") != 0 && strcmp(type, "UE") != 0) {
      hb_vmPushString(szExpr, exprLen);
      hb_macroGetValue(hb_stackItemFromTop(-1), 0, HB_SM_RT_MACRO);
      hb_itemMove(pItem, hb_stackItemFromTop(-1));
      hb_stackPop();
   } else {
      hb_itemClear(pItem);
   }

   return pItem;
}

static HB_BOOL classNameToBuffer(CFL_BUFFERP buffer, PHB_ITEM pItem) {
   const char *szClass;
   const char *szFunc;
   HB_USHORT uiClass = hb_objGetClass(pItem);
   if (uiClass == 0) {
      return HB_FALSE;
   }
   szClass = (char *)hb_clsName(uiClass);
   szFunc = (char *)hb_clsFuncName(uiClass);
   if (szClass && szFunc) {
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_OBJECT);
      cfl_buffer_putCharArray(buffer, szClass);
      cfl_buffer_putCharArray(buffer, szFunc);
      return HB_TRUE;
   }
   return HB_FALSE;
}

static HB_SIZE countRightSpaces(const char *szValue, HB_SIZE len) {
   HB_SIZE end = len;
   if (szValue == NULL) {
      return 0;
   }
   while (end > 0 && szValue[end - 1] == ' ') {
      --end;
   }
   return len - end;
}

static void stringToBuffer(CFL_BUFFERP buffer, PHB_ITEM pItem) {
   HB_SIZE len = hb_itemGetCLen(pItem);
   if (len > 4) {
      const char *str = (const char *)hb_itemGetCPtr(pItem);
      CFL_UINT32 rightSpaces = (CFL_UINT32)countRightSpaces(str, len);
      CFL_UINT32 filledLen = (CFL_UINT32)len - rightSpaces;
      if (filledLen > 65535) {
         if (rightSpaces > 65535) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR32_PAD32);
            cfl_buffer_putUInt32(buffer, filledLen);
            cfl_buffer_putUInt32(buffer, rightSpaces);
         } else if (rightSpaces > 255) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR32_PAD16);
            cfl_buffer_putUInt32(buffer, filledLen);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)rightSpaces);
         } else if (rightSpaces > 0) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR32_PAD8);
            cfl_buffer_putUInt32(buffer, filledLen);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)rightSpaces);
         } else {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR32);
            cfl_buffer_putUInt32(buffer, filledLen);
         }
         cfl_buffer_put(buffer, (void *)str, filledLen);
      } else if (filledLen > 255) {
         if (rightSpaces > 65535) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR16_PAD32);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)filledLen);
            cfl_buffer_putUInt32(buffer, rightSpaces);
         } else if (rightSpaces > 255) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR16_PAD16);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)filledLen);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)rightSpaces);
         } else if (rightSpaces > 0) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR16_PAD8);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)filledLen);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)rightSpaces);
         } else {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR16);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)filledLen);
         }
         cfl_buffer_put(buffer, (void *)str, filledLen);
      } else if (filledLen > 0) {
         if (rightSpaces > 65535) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR8_PAD32);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)filledLen);
            cfl_buffer_putUInt32(buffer, rightSpaces);
         } else if (rightSpaces > 255) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR8_PAD16);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)filledLen);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)rightSpaces);
         } else if (rightSpaces > 0) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR8_PAD8);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)filledLen);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)rightSpaces);
         } else {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR8);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)filledLen);
         }
         cfl_buffer_put(buffer, (void *)str, filledLen);
      } else if (rightSpaces > 65535) {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR0_PAD32);
         cfl_buffer_putUInt32(buffer, rightSpaces);
      } else if (rightSpaces > 255) {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR0_PAD16);
         cfl_buffer_putUInt16(buffer, (CFL_UINT16)rightSpaces);
      } else if (rightSpaces > 0) {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR0_PAD8);
         cfl_buffer_putUInt8(buffer, (CFL_UINT8)rightSpaces);
      } else {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR_EMPTY);
      }

      // 2..4
   } else if (len > 1) {
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR8);
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)len);
      cfl_buffer_put(buffer, (void *)hb_itemGetCPtr(pItem), (CFL_UINT32)len);
      // 1
   } else if (len == 1) {
      const char *aux = hb_itemGetCPtr(pItem);
      if ((aux[0] >= 'A' && aux[0] <= 'Z') || (aux[0] >= '0' && aux[0] <= '9') || (aux[0] >= 'a' && aux[0] <= 'z')) {
         cfl_buffer_putUInt8(buffer, (CFL_UINT8)aux[0]);
      } else {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR1);
         cfl_buffer_putUInt8(buffer, (CFL_UINT8)aux[0]);
      }
      // 0
   } else {
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_STR_EMPTY);
   }
}

HB_BOOL rgt_rpc_putItem(CFL_BUFFERP buffer, PHB_ITEM pItem) {
   HB_MAXINT val;
   HB_SIZE i;
   HB_SIZE len;

   RGT_LOG_ENTER("rgt_rpc_putItem", (NULL));

   switch (hb_itemType(pItem)) {
   case HB_IT_DATE: {
      int iYear, iMonth, iDay;
      hb_dateDecode(hb_itemGetDL(pItem), &iYear, &iMonth, &iDay);
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_DATE);
      cfl_buffer_putUInt16(buffer, (CFL_UINT16)iYear);
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)iMonth);
      cfl_buffer_putUInt8(buffer, (CFL_UINT8)iDay);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=date"));
   } break;

   case HB_IT_LOGICAL:
      cfl_buffer_putUInt8(buffer, hb_itemGetL(pItem) ? RGT_RPC_PAR_TRUE : RGT_RPC_PAR_FALSE);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=logical"));
      break;

   case HB_IT_INTEGER:
   case HB_IT_LONG:
      val = hb_itemGetNInt(pItem);
      if (HB_LIM_INT8(val)) {
         if (val >= 0 && val <= 9) {
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)val);
            RGT_LOG_TRACE(("rgt_rpc_putItem(). type=digit"));
         } else {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_BYTE);
            cfl_buffer_putInt8(buffer, (CFL_INT8)hb_itemGetNI(pItem));
            RGT_LOG_TRACE(("rgt_rpc_putItem(). type=int8"));
         }
      } else if (HB_LIM_INT16(val)) {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_SHORT);
         cfl_buffer_putInt16(buffer, (CFL_INT16)hb_itemGetNI(pItem));
         RGT_LOG_TRACE(("rgt_rpc_putItem(). type=int16"));
      } else if (HB_LIM_INT32(val)) {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_INT);
         cfl_buffer_putInt32(buffer, (CFL_INT32)hb_itemGetNL(pItem));
         RGT_LOG_TRACE(("rgt_rpc_putItem(). type=int32"));
      } else {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_LONG);
         cfl_buffer_putInt64(buffer, (CFL_INT64)hb_itemGetNLL(pItem));
         RGT_LOG_TRACE(("rgt_rpc_putItem(). type=int64"));
      }
      break;

   case HB_IT_DOUBLE:
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_DOUBLE);
      cfl_buffer_putDouble(buffer, hb_itemGetND(pItem));
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=double"));
      break;

   case HB_IT_MEMO:
      stringToBuffer(buffer, pItem);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=memo"));
      break;

   case HB_IT_STRING:
      stringToBuffer(buffer, pItem);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=string"));
      break;

   case HB_IT_NIL:
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_NULL);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=nil"));
      break;

   case HB_IT_POINTER:
#ifdef RGT_32BITS
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_PTR32);
      cfl_buffer_putInt32(buffer, (CFL_INT32)hb_itemGetPtr(pItem));
#else
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_PTR64);
      cfl_buffer_putInt64(buffer, (CFL_INT64)hb_itemGetPtr(pItem));
#endif
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=pointer"));
      break;

   case HB_IT_ARRAY: {
      classNameToBuffer(buffer, pItem);
      len = hb_arrayLen(pItem);
      if (len == 0) {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_EMPTY_ARRAY);
      } else {
         if (len <= 255) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_ARRAY8);
            cfl_buffer_putUInt8(buffer, (CFL_UINT8)len);
         } else if (len <= 65535) {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_ARRAY16);
            cfl_buffer_putUInt16(buffer, (CFL_UINT16)len);
         } else {
            cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_ARRAY32);
            cfl_buffer_putUInt32(buffer, (CFL_UINT32)len);
         }
         for (i = 1; i <= len; i++) {
            if (!rgt_rpc_putItem(buffer, hb_arrayGetItemPtr(pItem, i))) {
               return HB_FALSE;
            }
         }
      }
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=array len=%d", len));
   } break;
   case HB_IT_TIMESTAMP: {
      int iYear, iMonth, iDay, iHour, iMin, iSec, iMSec;
      hb_timeStampUnpack(hb_itemGetTD(pItem), &iYear, &iMonth, &iDay, &iHour, &iMin, &iSec, &iMSec);
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_TIMESTAMP);
      cfl_buffer_putUInt16(buffer, (CFL_INT16)iYear);
      cfl_buffer_putUInt8(buffer, (CFL_INT8)iMonth);
      cfl_buffer_putUInt8(buffer, (CFL_INT8)iDay);
      cfl_buffer_putUInt8(buffer, (CFL_INT8)iHour);
      cfl_buffer_putUInt8(buffer, (CFL_INT8)iMin);
      cfl_buffer_putUInt8(buffer, (CFL_INT8)iSec);
      cfl_buffer_putUInt16(buffer, (CFL_INT16)iMSec);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=timestamp"));
   } break;

   case HB_IT_HASH:
      len = hb_hashLen(pItem);
      if (len == 0) {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_EMPTY_HASH);
      } else {
         cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_HASH);
         cfl_buffer_putUInt32(buffer, (CFL_UINT32)len);
         for (i = 1; i <= len; i++) {
            if (!rgt_rpc_putItem(buffer, hb_hashGetKeyAt(pItem, i))) {
               return HB_FALSE;
            }
            if (!rgt_rpc_putItem(buffer, hb_hashGetValueAt(pItem, i))) {
               return HB_FALSE;
            }
         }
      }
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=hash len=%d", len));
      break;

   case HB_IT_SYMBOL: {
      const PHB_SYMB sym = hb_itemGetSymbol(pItem);
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_SYMBOL);
      cfl_buffer_putCharArray(buffer, sym->szName);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=symbol"));
   } break;

   case HB_IT_BLOCK:
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_BLOCK);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=codeblock"));
      break;

   default:
      cfl_buffer_putUInt8(buffer, RGT_RPC_PAR_NULL);
      RGT_LOG_TRACE(("rgt_rpc_putItem(). type=unknown"));
      break;
   }
   RGT_LOG_EXIT("rgt_rpc_putItem", (NULL));
   return HB_TRUE;
}

static PHB_ITEM strBufferToItem(CFL_BUFFERP buffer, CFL_UINT32 len, CFL_UINT32 pad, PHB_ITEM pItem) {
   char *strVal = RGT_HB_ALLOC(len + pad + 1);
   if (len > 0) {
      cfl_buffer_copy(buffer, (CFL_UINT8 *)strVal, len);
   }
   if (pad > 0) {
      memset(strVal + len, ' ', pad);
   }
   strVal[len + pad] = '\0';
   return hb_itemPutCLPtr(pItem, strVal, len + pad);
}

static PHB_ITEM createObject(const char *szClass, char *szFunc, PHB_ITEM pArray) {
   if (hb_clsFindClass(szClass, szFunc) == 0) {
      if (szFunc != NULL && strlen(szFunc) > 0) {
         hb_itemRelease(hb_itemDoC(szFunc, 0));
      } else {
         hb_itemRelease(hb_itemDoC(szClass, 0));
      }
   }
   if (hb_objSetClass(pArray, szClass, szFunc) == 0) {
      RGT_LOG_ERROR(("Unknown class '%s'. Class function: '%s'", szClass, szFunc));
      return NULL;
   }
   return pArray;
}

PHB_ITEM rgt_rpc_getItem(CFL_BUFFERP buffer, CFL_UINT8 parType, PHB_ITEM pItem) {
   CFL_UINT32 i;
   CFL_UINT8 itemType;
   CFL_UINT32 len;
   CFL_UINT32 pad;
   double dValue;
   int iWidth;
   int iDec;

   RGT_LOG_ENTER("rgt_rpc_getItem", (NULL));
   switch (parType) {
   case RGT_RPC_PAR_BIG_NUM: {
      char *strVal;
      HB_BOOL fDbl;
      HB_MAXINT lValue;
      len = cfl_buffer_getCharArrayLength(buffer);
      strVal = cfl_buffer_getCharArray(buffer);
      fDbl = hb_valStrnToNum(strVal, len, &lValue, &dValue, &iDec, &iWidth);
      if (fDbl) {
         pItem = hb_itemPutNDLen(pItem, dValue, iWidth, iDec);
      } else {
         pItem = hb_itemPutNIntLen(pItem, lValue, iWidth);
      }
      CFL_MEM_FREE(strVal);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=big_num"));
   } break;

   case RGT_RPC_PAR_STR_EMPTY:
      pItem = hb_itemPutCL(pItem, "", 0);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str_empty"));
      break;

   case RGT_RPC_PAR_STR1: {
      CFL_UINT8 aux = cfl_buffer_getUInt8(buffer);
      pItem = hb_itemPutCL(pItem, (char *)&aux, 1);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str1"));
   } break;

   case RGT_RPC_PAR_STR8:
      len = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pItem = strBufferToItem(buffer, len, 0, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str8 len=%u", len));
      break;

   case RGT_RPC_PAR_STR16:
      len = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pItem = strBufferToItem(buffer, len, 0, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str16 len=%u", len));
      break;

   case RGT_RPC_PAR_STR32:
      len = cfl_buffer_getUInt32(buffer);
      pItem = strBufferToItem(buffer, len, 0, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str32 len=%u", len));
      break;

   case RGT_RPC_PAR_STR32_PAD32:
      len = cfl_buffer_getUInt32(buffer);
      pad = cfl_buffer_getUInt32(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str32_pad32 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR32_PAD16:
      len = cfl_buffer_getUInt32(buffer);
      pad = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str32_pad16 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR32_PAD8:
      len = cfl_buffer_getUInt32(buffer);
      pad = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str32_pad8 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR16_PAD32:
      len = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pad = cfl_buffer_getUInt32(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str16_pad32 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR16_PAD16:
      len = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pad = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str16_pad16 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR16_PAD8:
      len = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pad = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str16_pad8 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR8_PAD32:
      len = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pad = cfl_buffer_getUInt32(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str8_pad32 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR8_PAD16:
      len = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pad = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str8_pad16 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR8_PAD8:
      len = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pad = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pItem = strBufferToItem(buffer, len, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str8_pad8 len=%u pad=%u", len, pad));
      break;

   case RGT_RPC_PAR_STR0_PAD32:
      pad = cfl_buffer_getUInt32(buffer);
      pItem = strBufferToItem(buffer, 0, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str0_pad32 pad=%u", pad));
      break;

   case RGT_RPC_PAR_STR0_PAD16:
      pad = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      pItem = strBufferToItem(buffer, 0, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str0_pad16 pad=%u", pad));
      break;

   case RGT_RPC_PAR_STR0_PAD8:
      pad = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      pItem = strBufferToItem(buffer, 0, pad, pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=str0_pad8 pad=%u", pad));
      break;

   case RGT_RPC_PAR_UPPER_A:
   case RGT_RPC_PAR_UPPER_B:
   case RGT_RPC_PAR_UPPER_C:
   case RGT_RPC_PAR_UPPER_D:
   case RGT_RPC_PAR_UPPER_E:
   case RGT_RPC_PAR_UPPER_F:
   case RGT_RPC_PAR_UPPER_G:
   case RGT_RPC_PAR_UPPER_H:
   case RGT_RPC_PAR_UPPER_I:
   case RGT_RPC_PAR_UPPER_J:
   case RGT_RPC_PAR_UPPER_K:
   case RGT_RPC_PAR_UPPER_L:
   case RGT_RPC_PAR_UPPER_M:
   case RGT_RPC_PAR_UPPER_N:
   case RGT_RPC_PAR_UPPER_O:
   case RGT_RPC_PAR_UPPER_P:
   case RGT_RPC_PAR_UPPER_Q:
   case RGT_RPC_PAR_UPPER_R:
   case RGT_RPC_PAR_UPPER_S:
   case RGT_RPC_PAR_UPPER_T:
   case RGT_RPC_PAR_UPPER_U:
   case RGT_RPC_PAR_UPPER_V:
   case RGT_RPC_PAR_UPPER_W:
   case RGT_RPC_PAR_UPPER_X:
   case RGT_RPC_PAR_UPPER_Y:
   case RGT_RPC_PAR_UPPER_Z:
   case RGT_RPC_PAR_LOWER_A:
   case RGT_RPC_PAR_LOWER_B:
   case RGT_RPC_PAR_LOWER_C:
   case RGT_RPC_PAR_LOWER_D:
   case RGT_RPC_PAR_LOWER_E:
   case RGT_RPC_PAR_LOWER_F:
   case RGT_RPC_PAR_LOWER_G:
   case RGT_RPC_PAR_LOWER_H:
   case RGT_RPC_PAR_LOWER_I:
   case RGT_RPC_PAR_LOWER_J:
   case RGT_RPC_PAR_LOWER_K:
   case RGT_RPC_PAR_LOWER_L:
   case RGT_RPC_PAR_LOWER_M:
   case RGT_RPC_PAR_LOWER_N:
   case RGT_RPC_PAR_LOWER_O:
   case RGT_RPC_PAR_LOWER_P:
   case RGT_RPC_PAR_LOWER_Q:
   case RGT_RPC_PAR_LOWER_R:
   case RGT_RPC_PAR_LOWER_S:
   case RGT_RPC_PAR_LOWER_T:
   case RGT_RPC_PAR_LOWER_U:
   case RGT_RPC_PAR_LOWER_V:
   case RGT_RPC_PAR_LOWER_W:
   case RGT_RPC_PAR_LOWER_X:
   case RGT_RPC_PAR_LOWER_Y:
   case RGT_RPC_PAR_LOWER_Z:
   case RGT_RPC_PAR_CHAR_ZERO:
   case RGT_RPC_PAR_CHAR_ONE:
   case RGT_RPC_PAR_CHAR_TWO:
   case RGT_RPC_PAR_CHAR_THREE:
   case RGT_RPC_PAR_CHAR_FOUR:
   case RGT_RPC_PAR_CHAR_FIVE:
   case RGT_RPC_PAR_CHAR_SIX:
   case RGT_RPC_PAR_CHAR_SEVEN:
   case RGT_RPC_PAR_CHAR_EIGTH:
   case RGT_RPC_PAR_CHAR_NINE: {
      char caracter = (char)parType;
      pItem = hb_itemPutCL(pItem, &caracter, 1);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=char"));
   } break;

   case RGT_RPC_PAR_TRUE:
      pItem = hb_itemPutL(pItem, HB_TRUE);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=true"));
      break;

   case RGT_RPC_PAR_FALSE:
      pItem = hb_itemPutL(pItem, HB_FALSE);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=false"));
      break;

   case RGT_RPC_PAR_NULL:
      if (pItem == NULL) {
         pItem = hb_itemNew(NULL);
      } else {
         hb_itemClear(pItem);
      }
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=nil"));
      break;

   case RGT_RPC_PAR_ZERO:
   case RGT_RPC_PAR_ONE:
   case RGT_RPC_PAR_TWO:
   case RGT_RPC_PAR_THREE:
   case RGT_RPC_PAR_FOUR:
   case RGT_RPC_PAR_FIVE:
   case RGT_RPC_PAR_SIX:
   case RGT_RPC_PAR_SEVEN:
   case RGT_RPC_PAR_EIGTH:
   case RGT_RPC_PAR_NINE:
      pItem = hb_itemPutNI(pItem, (int)parType);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=digit"));
      break;

   case RGT_RPC_PAR_DATE: {
      int iYear = (int)cfl_buffer_getUInt16(buffer);
      int iMonth = (int)cfl_buffer_getUInt8(buffer);
      int iDay = (int)cfl_buffer_getUInt8(buffer);
      pItem = hb_itemPutD(pItem, iYear, iMonth, iDay);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=date"));
   } break;

   case RGT_RPC_PAR_BYTE: {
      int val = (int)cfl_buffer_getInt8(buffer);
      pItem = hb_itemPutNI(pItem, val);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=int8"));
   } break;

   case RGT_RPC_PAR_SHORT: {
      int val = (int)cfl_buffer_getInt16(buffer);
      pItem = hb_itemPutNI(pItem, val);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=int16"));
   } break;

   case RGT_RPC_PAR_INT: {
      long val = (long)cfl_buffer_getInt32(buffer);
      pItem = hb_itemPutNL(pItem, val);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=int32"));
   } break;

   case RGT_RPC_PAR_LONG: {
      HB_LONGLONG val = (HB_LONGLONG)cfl_buffer_getInt64(buffer);
      pItem = hb_itemPutNLL(pItem, val);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=int64"));
   } break;

   case RGT_RPC_PAR_FLOAT:
      dValue = (double)cfl_buffer_getFloat(buffer);
      pItem = hb_itemPutNDLen(pItem, dValue, HB_DBL_LENGTH(dValue), RGT_RPC_DECIMALS);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=float32"));
      break;

   case RGT_RPC_PAR_DOUBLE:
      dValue = cfl_buffer_getDouble(buffer);
      pItem = hb_itemPutNDLen(pItem, dValue, HB_DBL_LENGTH(dValue), RGT_RPC_DECIMALS);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=float64"));
      break;

   case RGT_RPC_PAR_PTR32:
      pItem = hb_itemPutPtr(pItem, (void *)(uintptr_t)cfl_buffer_getInt32(buffer));
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=pointer"));
      break;

   case RGT_RPC_PAR_PTR64:
      pItem = hb_itemPutPtr(pItem, (void *)(uintptr_t)cfl_buffer_getInt64(buffer));
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=pointer"));
      break;

   case RGT_RPC_PAR_OBJECT: {
      char *szClass;
      char *szFunc;
      CFL_UINT8 arrayType;
      szClass = cfl_buffer_getCharArray(buffer);
      if (szClass == NULL) {
         RGT_LOG_ERROR(("Failed to extract class name from buffer"));
         return NULL;
      }
      szFunc = cfl_buffer_getCharArray(buffer);
      if (szFunc == NULL) {
         RGT_LOG_ERROR(("Failed to extract function name from buffer"));
         CFL_MEM_FREE(szClass);
         return NULL;
      }
      arrayType = cfl_buffer_getUInt8(buffer);
      pItem = rgt_rpc_getItem(buffer, arrayType, pItem);
      pItem = createObject(szClass, szFunc, pItem);
      CFL_MEM_FREE(szClass);
      CFL_MEM_FREE(szFunc);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=object"));
      break;
   }

   case RGT_RPC_PAR_EMPTY_ARRAY:
      if (pItem == NULL) {
         pItem = hb_itemNew(NULL);
      }
      hb_arrayNew(pItem, 0);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=array0"));
      break;

   case RGT_RPC_PAR_ARRAY8:
      if (pItem == NULL) {
         pItem = hb_itemNew(NULL);
      }
      len = (CFL_UINT32)cfl_buffer_getUInt8(buffer);
      hb_arrayNew(pItem, len);
      for (i = 1; i <= len; i++) {
         itemType = cfl_buffer_getUInt8(buffer);
         rgt_rpc_getItem(buffer, itemType, hb_arrayGetItemPtr(pItem, i));
      }
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=array8 len=%u", len));
      break;

   case RGT_RPC_PAR_ARRAY16:
      if (pItem == NULL) {
         pItem = hb_itemNew(NULL);
      }
      len = (CFL_UINT32)cfl_buffer_getUInt16(buffer);
      hb_arrayNew(pItem, len);
      for (i = 1; i <= len; i++) {
         itemType = cfl_buffer_getUInt8(buffer);
         rgt_rpc_getItem(buffer, itemType, hb_arrayGetItemPtr(pItem, i));
      }
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=array16 len=%u", len));
      break;

   case RGT_RPC_PAR_ARRAY32:
      if (pItem == NULL) {
         pItem = hb_itemNew(NULL);
      }
      len = cfl_buffer_getUInt32(buffer);
      hb_arrayNew(pItem, len);
      for (i = 1; i <= len; i++) {
         itemType = cfl_buffer_getUInt8(buffer);
         rgt_rpc_getItem(buffer, itemType, hb_arrayGetItemPtr(pItem, i));
      }
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=array32 len=%u", len));
      break;

   case RGT_RPC_PAR_TIMESTAMP: {
      int iYear, iMonth, iDay, iHour, iMin, iSec, iMSec;
      iYear = (int)cfl_buffer_getUInt16(buffer);
      iMonth = (int)cfl_buffer_getUInt8(buffer);
      iDay = (int)cfl_buffer_getUInt8(buffer);
      iHour = (int)cfl_buffer_getUInt8(buffer);
      iMin = (int)cfl_buffer_getUInt8(buffer);
      iSec = (int)cfl_buffer_getUInt8(buffer);
      iMSec = (int)cfl_buffer_getUInt16(buffer);
      pItem = hb_itemPutTD(pItem, hb_timeStampPack(iYear, iMonth, iDay, iHour, iMin, iSec, iMSec));
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=timestamp"));
   } break;

   case RGT_RPC_PAR_EMPTY_HASH:
      pItem = hb_hashNew(pItem);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=hash0"));
      break;

   case RGT_RPC_PAR_HASH:
      len = cfl_buffer_getUInt32(buffer);
      pItem = hb_hashNew(pItem);
      hb_hashPreallocate(pItem, len);
      for (i = 0; i < len; i++) {
         PHB_ITEM pKey;
         PHB_ITEM pValue;
         itemType = cfl_buffer_getUInt8(buffer);
         pKey = rgt_rpc_getItem(buffer, itemType, NULL);
         itemType = cfl_buffer_getUInt8(buffer);
         pValue = rgt_rpc_getItem(buffer, itemType, NULL);
         hb_hashAdd(pItem, pKey, pValue);
         hb_itemRelease(pKey);
         hb_itemRelease(pValue);
      }
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=hash len=%u", len));
      break;

   case RGT_RPC_PAR_SYMBOL: {
      CFL_STRP symbolName = cfl_buffer_getString(buffer);
      PHB_SYMB pSymbol = hb_dynsymFindSymbol(cfl_str_getPtr(symbolName));
      cfl_str_free(symbolName);
      if (pSymbol != NULL) {
         pItem = hb_itemPutSymbol(pItem, pSymbol);
      } else {
         pItem = NULL;
      }
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=symbol"));
   } break;

   case RGT_RPC_PAR_BLOCK:
      pItem = evalMacro(pItem, "{|| .T. }", 9);
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=codeblock"));
      break;

   default:
      pItem = NULL;
      RGT_LOG_TRACE(("rgt_rpc_getItem(). type=unknown code=%x (%d)", parType, (int)parType));
      break;
   }
   RGT_LOG_EXIT("rgt_rpc_getItem", (NULL));
   return pItem;
}

void rgt_rpc_executeFunction(CFL_BUFFERP buffer) {
   PHB_DYNS funSymbol;
   char *functionName;

   RGT_LOG_ENTER("rgt_rpc_executeFunction", (NULL));
   // Extrai nome da funcao
   functionName = cfl_buffer_getCharArray(buffer);
   RGT_LOG_DEBUG(("RPC function: %s", functionName));
   funSymbol = hb_dynsymFindName(functionName);
   // Obtem parametros
   if (funSymbol) {
      int parCount = 0;
      HB_BOOL parTypeError = HB_FALSE;
      CFL_UINT8 parType;
      hb_vmPushSymbol(hb_dynsymSymbol(funSymbol));
      hb_vmPushNil();
      parType = cfl_buffer_getUInt8(buffer);
      while (parType != RGT_RPC_PAR_END && !parTypeError) {
         PHB_ITEM pItem = rgt_rpc_getItem(buffer, parType, NULL);
         if (pItem != NULL) {
            ++parCount;
            RGT_LOG_PARAM(RGT_LOG_LEVEL_DEBUG, parCount, pItem);
            hb_vmPush(pItem);
            hb_itemRelease(pItem);
            parType = cfl_buffer_getUInt8(buffer);
         } else {
            parTypeError = HB_TRUE;
         }
      }
      if (!parTypeError) {
         // Executa a funcao
         hb_vmDo(parCount);
         RGT_LOG_PARAM(RGT_LOG_LEVEL_DEBUG, 0, hb_stackReturnItem());
         rgt_common_prepareResponse(buffer, RGT_RESP_SUCCESS);
         if (!rgt_rpc_putItem(buffer, hb_stackReturnItem())) {
            rgt_common_prepareResponse(buffer, RGT_ERROR_RPC_INVALID_DATA_TYPE);
            cfl_buffer_putFormat(buffer, "Returned data type is not supported: %s", hb_itemTypeStr(hb_stackReturnItem()));
         }
         hb_itemClear(hb_stackReturnItem());
         hb_gcCollect();
      } else {
         rgt_common_prepareResponse(buffer, RGT_ERROR_RPC_INVALID_PAR_TYPE);
         cfl_buffer_putFormat(buffer, "Invalid data type for parameter %d", parCount + 1);
         // Remove os elementos da pilha
         while (parCount > 0) {
            hb_stackPop();
            parCount--;
         }
         hb_stackPop();
         hb_stackPop();
      }
      // Escreve no buffer o resultado
   } else {
      CFL_STRP error = cfl_str_setFormat(NULL, "Function %s not found in TE", functionName);
      rgt_common_prepareResponse(buffer, RGT_ERROR_RPC_UNDEFINED_FUNCTION);
      cfl_buffer_putCharArray(buffer, cfl_str_getPtr(error));
      cfl_str_free(error);
   }
   CFL_MEM_FREE(functionName);
   RGT_LOG_EXIT("rgt_rpc_executeFunction", (NULL));
}
