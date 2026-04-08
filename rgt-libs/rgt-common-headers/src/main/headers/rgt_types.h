#ifndef _RGT_TYPES_H_

#define _RGT_TYPES_H_

// #ifdef __linux__
//    #include <sys/types.h>
//    #include <unistd.h>
//    #include <sys/syscall.h>
// #endif

#include "hbapigt.h"

#include "cfl_types.h"
// #ifdef __SPIN_LOCK__
// #include "cfl_atomic.h"
// #endif

#ifdef CFL_ARCH_64
#define RGT_64BITS
#else
#define RGT_32BITS
#endif

#ifdef _MSC_VER
#define snprintf _snprintf
#endif

#define hb_vmQuitRequest() (hb_vmRequestQuery() == HB_QUIT_REQUESTED)

#define RGT_CMD_UNKNOWN 0x00

/* Terminal command type */
#define RGT_TRM_CMD_LOGIN 0x0A
#define RGT_TRM_CMD_LOGOUT 0x0B
#define RGT_TRM_CMD_RECONNECT 0x0C

/* Application command type */
#define RGT_APP_CMD_LOGIN 0x32
#define RGT_APP_CMD_LOGOUT 0x33
#define RGT_APP_CMD_SET_ENV 0x34
#define RGT_APP_CMD_UPDATE 0x35
// #define RGT_APP_CMD_READ_KEY           0x36 --> Deprecated
#define RGT_APP_CMD_RPC 0x37
#define RGT_APP_CMD_PUT_FILE 0x38
#define RGT_APP_CMD_GET_FILE 0x39
#define RGT_APP_CMD_KEY_BUF_LEN 0x3A
#define RGT_APP_CMD_RECONNECT 0x3B
#define RGT_APP_CMD_KEEP_ALIVE 0x3C
#define RGT_APP_SESSION_CONFIG 0x3D

#define RGT_STANDALONE_APP_EXEC 0x64
#define RGT_STANDALONE_APP_SEND_OUTPUT 0x65
#define RGT_STANDALONE_APP_SEND_STATUS 0x66

/* Server command type */
#define RGT_ADMIN_GET_SCREEN 0x80

#define RGT_ADMIN_RESPONSE 0xFF00 // 255

/* Terminal background data sent to App */
#define RGT_RESP_TRM_KEEP_ALIVE 0x7FFE
#define RGT_RESP_TRM_KEY_UPDATE 0x7FFF

#define RGT_TERMINAL 'T'
#define RGT_SERVER 'S'
#define RGT_APP 'A'

#define RGT_AUTH_TOKEN_VAR "RGT_AUTH_TOKEN"
#define RGT_SERVER_ADDR_VAR "RGT_SERVER_ADDR"
#define RGT_SERVER_PORT_VAR "RGT_SERVER_PORT"
#define RGT_STANDALONE_APP_VAR "RGT_STANDALONE_APP"

#define RGT_PACKET_LEN_TYPE CFL_INT32
#define RGT_PACKET_LEN_FIELD_SIZE sizeof(RGT_PACKET_LEN_TYPE)
#define RGT_REQUEST_OP_FIELD_SIZE sizeof(CFL_UINT8)
#define RGT_RESPONSE_CODE_FIELD_SIZE sizeof(CFL_UINT16)

#define RGT_HEADER_MAGIC_NUMBER 0x5CDBA4EA
#define RGT_MAGIC_NUMBER_SIZE 4
#define RGT_REQUEST_HEADER_SIZE (RGT_PACKET_LEN_FIELD_SIZE + RGT_REQUEST_OP_FIELD_SIZE)
#define RGT_RESPONSE_HEADER_SIZE (RGT_PACKET_LEN_FIELD_SIZE + RGT_RESPONSE_CODE_FIELD_SIZE)

#define RGT_PUT_PACKET_LEN(buffer, len) cfl_buffer_putInt32(buffer, len)
#define RGT_GET_PACKET_LEN(buffer) cfl_buffer_getInt32(buffer)

#define RGT_SERVER_PROTOCOL_VERSION 5
#define RGT_TRM_APP_PROTOCOL_VERSION 2
#define RGT_ADMIN_PROTOCOL_VERSION 6

#define RGT_BUFFER_LEN_FIELD_TYPE CFL_INT32

#define RGT_TE_IO_BUFFER_SIZE 8192
#define RGT_APP_IO_BUFFER_SIZE 8192
#define RGT_KEY_BUFFER_SIZE 256
#define RGT_TONE_BUFFER_SIZE 128

// #ifdef __SPIN_LOCK__
//    #define RGT_LOCK            CFL_BOOL
//    #define RGT_LOCK_INIT(l)    l = CFL_FALSE
//    #define RGT_LOCK_FREE(l)
//    #define RGT_LOCK_ACQUIRE(l) while (! cfl_atomic_compareAndSetBoolean(&(l), CFL_FALSE, CFL_TRUE)) rgt_thread_yield()
//    #define RGT_LOCK_RELEASE(l) while (cfl_atomic_compareAndSetBoolean(&(l), CFL_TRUE, CFL_FALSE)) rgt_thread_yield()
// #else
#define RGT_LOCK CFL_LOCK
#define RGT_LOCK_INIT(l) cfl_lock_init(&l)
#define RGT_LOCK_FREE(l) cfl_lock_free(&l)
#define RGT_LOCK_ACQUIRE(l) cfl_lock_acquire(&l)
#define RGT_LOCK_RELEASE(l) cfl_lock_release(&l)
// #endif

#define RGT_SCREEN_TYPE_SIMPLE 0x00
#define RGT_SCREEN_TYPE_EXTENDED 0x01

#define RGT_CLP_COMP CFL_UINT8
#define RGT_CLP_COMP_UNKNOWN 0x00
#define RGT_CLP_COMP_XHARBOUR 0x01
#define RGT_CLP_COMP_HARBOUR 0x02

#define RGT_CHANNEL_UNIDIRECTIONAL 0
#define RGT_CHANNEL_BIDIRECTIONAL 1
#define RGT_CHANNEL_QUEUE 2
#define RGT_CHANNEL_MIN RGT_CHANNEL_UNIDIRECTIONAL
#define RGT_CHANNEL_MAX RGT_CHANNEL_QUEUE

#define RGT_CHANNEL_TYPE_VAR "RGT_CHANNEL_TYPE"
#define RGT_CHANNEL_TYPE_UNIDIRECTIONAL "UNIDIRECTIONAL"
#define RGT_CHANNEL_TYPE_BIDIRECTIONAL "BIDIRECTIONAL"
#define RGT_CHANNEL_TYPE_QUEUE "QUEUE"

#define RGT_UPDATE_MODE_VAR "RGT_UPDATE_MODE"
#define RGT_UPDATE_MODE_BACKGROUND "BACKGROUND"
#define RGT_UPDATE_MODE_MAIN_THREAD "MAIN_THREAD"

#define RGT_SESS_MODE_NORMAL 0
#define RGT_SESS_MODE_TRANSACTION 1

#define RGT_HB_ALLOC(s) hb_xgrab(s)
#define RGT_HB_REALLOC(m, s) hb_xrealloc(m, s)
#define RGT_HB_FREE(m) hb_xfree(m)
// #define RGT_HB_ALLOC(s)      malloc(s)
// #define RGT_HB_REALLOC(m, s) realloc(m, s)
// #define RGT_HB_FREE(m)       free(m)

struct _RGT_SCREEN;
typedef struct _RGT_SCREEN RGT_SCREEN;
typedef struct _RGT_SCREEN *RGT_SCREENP;

struct _RGT_INT_BUFFER;
typedef struct _RGT_INT_BUFFER RGT_INT_BUFFER;
typedef struct _RGT_INT_BUFFER *RGT_INT_BUFFERP;

struct _RGT_APP_CONNECTION;
typedef struct _RGT_APP_CONNECTION RGT_APP_CONNECTION;
typedef struct _RGT_APP_CONNECTION *RGT_APP_CONNECTIONP;

struct _RGT_ERROR;
typedef struct _RGT_ERROR RGT_ERROR;
typedef struct _RGT_ERROR *RGT_ERRORP;

struct _RGT_TRM_CONNECTION;
typedef struct _RGT_TRM_CONNECTION RGT_TRM_CONNECTION;
typedef struct _RGT_TRM_CONNECTION *RGT_TRM_CONNECTIONP;

struct _RGT_CHANNEL;
typedef struct _RGT_CHANNEL RGT_CHANNEL;
typedef struct _RGT_CHANNEL *RGT_CHANNELP;

struct _RGT_THREAD;
typedef struct _RGT_THREAD RGT_THREAD;
typedef struct _RGT_THREAD *RGT_THREADP;

#endif
