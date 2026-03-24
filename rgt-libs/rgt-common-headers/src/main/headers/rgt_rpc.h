#ifndef _RGT_RPC_H_
#define _RGT_RPC_H_

#include "cfl_types.h"
#include "cfl_buffer.h"

#include "hbstack.h"
#include "hbvm.h"
#include "hbapi.h"
#include "hbapiitm.h"
#include "hbset.h"
#include "hbdate.h"

#include "rgt_types.h"

#ifdef __HBR__
   #define RGT_RPC_DECIMALS hb_setGetDecimals()
#else
   #define RGT_RPC_DECIMALS hb_set.HB_SET_DECIMALS
#endif


#define RGT_RPC_PAR_ZERO        0x00
#define RGT_RPC_PAR_ONE         0x01
#define RGT_RPC_PAR_TWO         0x02
#define RGT_RPC_PAR_THREE       0x03
#define RGT_RPC_PAR_FOUR        0x04
#define RGT_RPC_PAR_FIVE        0x05
#define RGT_RPC_PAR_SIX         0x06
#define RGT_RPC_PAR_SEVEN       0x07
#define RGT_RPC_PAR_EIGTH       0x08
#define RGT_RPC_PAR_NINE        0x09
#define RGT_RPC_PAR_BLOCK       0x0A
#define RGT_RPC_PAR_SYMBOL      0x0B
#define RGT_RPC_PAR_SHORT       0x21
#define RGT_RPC_PAR_STR32       0x22
#define RGT_RPC_PAR_DATE        0x23
#define RGT_RPC_PAR_TRUE        0x24
#define RGT_RPC_PAR_FALSE       0x25
#define RGT_RPC_PAR_BYTE        0x26
#define RGT_RPC_PAR_INT         0x28
#define RGT_RPC_PAR_LONG        0x29
#define RGT_RPC_PAR_BIG_NUM     0x2A
#define RGT_RPC_PAR_FLOAT       0x2B
#define RGT_RPC_PAR_DOUBLE      0x2D
#define RGT_RPC_PAR_NULL        0x2F
#define RGT_RPC_PAR_CHAR_ZERO   0x30 // '0'
#define RGT_RPC_PAR_CHAR_ONE    0x31 // '1'
#define RGT_RPC_PAR_CHAR_TWO    0x32 // '2'
#define RGT_RPC_PAR_CHAR_THREE  0x33 // '3'
#define RGT_RPC_PAR_CHAR_FOUR   0x34 // '4'
#define RGT_RPC_PAR_CHAR_FIVE   0x35 // '5'
#define RGT_RPC_PAR_CHAR_SIX    0x36 // '6'
#define RGT_RPC_PAR_CHAR_SEVEN  0x37 // '7'
#define RGT_RPC_PAR_CHAR_EIGTH  0x38 // '8'
#define RGT_RPC_PAR_CHAR_NINE   0x39 // '9'
#define RGT_RPC_PAR_STR_EMPTY   0x3A
#define RGT_RPC_PAR_ARRAY32     0x3B
#define RGT_RPC_PAR_EMPTY_ARRAY 0x3C
#define RGT_RPC_PAR_TIMESTAMP   0x3D
#define RGT_RPC_PAR_EMPTY_HASH  0x3E
#define RGT_RPC_PAR_HASH        0x3F
#define RGT_RPC_PAR_PTR32       0x40
#define RGT_RPC_PAR_UPPER_A     0x41 // 'A'
#define RGT_RPC_PAR_UPPER_B     0x42 // 'B'
#define RGT_RPC_PAR_UPPER_C     0x43 // 'C'
#define RGT_RPC_PAR_UPPER_D     0x44 // 'D'
#define RGT_RPC_PAR_UPPER_E     0x45 // 'E'
#define RGT_RPC_PAR_UPPER_F     0x46 // 'F'
#define RGT_RPC_PAR_UPPER_G     0x47 // 'G'
#define RGT_RPC_PAR_UPPER_H     0x48 // 'H'
#define RGT_RPC_PAR_UPPER_I     0x49 // 'I'
#define RGT_RPC_PAR_UPPER_J     0x4A // 'J'
#define RGT_RPC_PAR_UPPER_K     0x4B // 'K'
#define RGT_RPC_PAR_UPPER_L     0x4C // 'L'
#define RGT_RPC_PAR_UPPER_M     0x4D // 'M'
#define RGT_RPC_PAR_UPPER_N     0x4E // 'N'
#define RGT_RPC_PAR_UPPER_O     0x4F // 'O'
#define RGT_RPC_PAR_UPPER_P     0x50 // 'P'
#define RGT_RPC_PAR_UPPER_Q     0x51 // 'Q'
#define RGT_RPC_PAR_UPPER_R     0x52 // 'R'
#define RGT_RPC_PAR_UPPER_S     0x53 // 'S'
#define RGT_RPC_PAR_UPPER_T     0x54 // 'T'
#define RGT_RPC_PAR_UPPER_U     0x55 // 'U'
#define RGT_RPC_PAR_UPPER_V     0x56 // 'V'
#define RGT_RPC_PAR_UPPER_W     0x57 // 'W'
#define RGT_RPC_PAR_UPPER_X     0x58 // 'X'
#define RGT_RPC_PAR_UPPER_Y     0x59 // 'Y'
#define RGT_RPC_PAR_UPPER_Z     0x5A // 'Z'
#define RGT_RPC_PAR_PTR64       0x5B
#define RGT_RPC_PAR_LOWER_A     0x61 // 'a'
#define RGT_RPC_PAR_LOWER_B     0x62 // 'b'
#define RGT_RPC_PAR_LOWER_C     0x63 // 'c'
#define RGT_RPC_PAR_LOWER_D     0x64 // 'd'
#define RGT_RPC_PAR_LOWER_E     0x65 // 'e'
#define RGT_RPC_PAR_LOWER_F     0x66 // 'f'
#define RGT_RPC_PAR_LOWER_G     0x67 // 'g'
#define RGT_RPC_PAR_LOWER_H     0x68 // 'h'
#define RGT_RPC_PAR_LOWER_I     0x69 // 'i'
#define RGT_RPC_PAR_LOWER_J     0x6A // 'j'
#define RGT_RPC_PAR_LOWER_K     0x6B // 'k'
#define RGT_RPC_PAR_LOWER_L     0x6C // 'l'
#define RGT_RPC_PAR_LOWER_M     0x6D // 'm'
#define RGT_RPC_PAR_LOWER_N     0x6E // 'n'
#define RGT_RPC_PAR_LOWER_O     0x6F // 'o'
#define RGT_RPC_PAR_LOWER_P     0x70 // 'p'
#define RGT_RPC_PAR_LOWER_Q     0x71 // 'q'
#define RGT_RPC_PAR_LOWER_R     0x72 // 'r'
#define RGT_RPC_PAR_LOWER_S     0x73 // 's'
#define RGT_RPC_PAR_LOWER_T     0x74 // 't'
#define RGT_RPC_PAR_LOWER_U     0x75 // 'u'
#define RGT_RPC_PAR_LOWER_V     0x76 // 'v'
#define RGT_RPC_PAR_LOWER_W     0x77 // 'w'
#define RGT_RPC_PAR_LOWER_X     0x78 // 'x'
#define RGT_RPC_PAR_LOWER_Y     0x79 // 'y'
#define RGT_RPC_PAR_LOWER_Z     0x7A // 'z'
#define RGT_RPC_PAR_ARRAY8      0x7B
#define RGT_RPC_PAR_ARRAY16     0x7C
#define RGT_RPC_PAR_OBJECT      0x7D
#define RGT_RPC_PAR_STR16       0x7E
#define RGT_RPC_PAR_STR8        0x7F
#define RGT_RPC_PAR_STR1        0x80
#define RGT_RPC_PAR_STR32_PAD32 0x81
#define RGT_RPC_PAR_STR32_PAD16 0x82
#define RGT_RPC_PAR_STR32_PAD8  0x83
#define RGT_RPC_PAR_STR16_PAD32 0x84
#define RGT_RPC_PAR_STR16_PAD16 0x85
#define RGT_RPC_PAR_STR16_PAD8  0x86
#define RGT_RPC_PAR_STR8_PAD32  0x87
#define RGT_RPC_PAR_STR8_PAD16  0x88
#define RGT_RPC_PAR_STR8_PAD8   0x89
#define RGT_RPC_PAR_STR0_PAD32  0x8A
#define RGT_RPC_PAR_STR0_PAD16  0x8B
#define RGT_RPC_PAR_STR0_PAD8   0x8C
#define RGT_RPC_PAR_END         0x2E // '.'

extern HB_BOOL rgt_rpc_putItem(CFL_BUFFERP buffer, PHB_ITEM pItem);
extern void rgt_rpc_executeFunction(CFL_BUFFERP buffer);
extern PHB_ITEM rgt_rpc_getItem(CFL_BUFFERP buffer, CFL_UINT8 parType, PHB_ITEM pItem);

#endif
