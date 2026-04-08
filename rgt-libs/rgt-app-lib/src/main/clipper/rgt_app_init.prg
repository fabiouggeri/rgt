#include "hbgtinfo.ch"
#define GT_NAME() HB_GTINFO(HB_GTI_VERSION)

/**
 * RGT_Init
 *
 * @author fabiouggeri
 * @version 1.6.1
 * @since 01/12/2021
 */
Procedure RGT_Init()
   RGT_GT_Init()
Return

/**
 * _RgtSetIniMode
 *
 * @author fabiouggeri
 * @version 1.0.0
 * @since 21/05/2019
 */
Init Procedure _RgtSetIniMode()
   If ! Upper( GT_NAME() ) $ "_GTSTD_GTNUL_"
      RGT_GT_Init()
   EndIf
Return
