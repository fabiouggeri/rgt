#ifndef _RGT_ERROR_CH

#define _RGT_ERROR_CH


#define RGT_RESP_SUCCESS         0

#define RGT_ERROR_NONE            0
#define RGT_ERROR_UNKNOWN         1
#define RGT_ERROR_SOCKET          2
#define RGT_ERROR_PROTOCOL        3
#define RGT_ERROR_RESPONSE        4
#define RGT_ERROR_UNKNOWN_CMD     5
#define RGT_ERROR_AUTHENTICATION  6
#define RGT_ERROR_CONNECTION_LOST 7
#define RGT_ERROR_UNKNOWN_RESP    8
#define RGT_ERROR_SESSION_CLOSED  9

/* Terminal Emulator errors */
#define RGT_ERROR_AUTH_ERROR             10
#define RGT_ERROR_APP_LAUNCH_ERROR       11
#define RGT_ERROR_INVALID_ARG            12

/* Application errors */
#define RGT_ERROR_APP_CONNECTION_ERROR   20
#define RGT_ERROR_APP_INIT_GT            21
#define RGT_ERROR_APP_SCREEN_BUSY        22

/* File errors */
#define RGT_ERROR_FILE_SYSTEM            100
#define RGT_ERROR_CREATING_FILE          101
#define RGT_ERROR_READING_FILE           102
#define RGT_ERROR_WRITING_FILE           103
#define RGT_ERROR_OPENING_FILE           104

/* Remote Procedure Call errors */
#define RGT_ERROR_RPC_INVALID_DATA_TYPE  200
#define RGT_ERROR_RPC_INVALID_PAR_TYPE   201
#define RGT_ERROR_RPC_UNDEFINED_FUNCTION 202
#define RGT_ERROR_RPC_DATA_CORRUPTION    203

/* Environment error */
#define RGT_ERROR_ENV_VAR_NOT            300

/* Timeout error */
#define RGT_ERROR_TIMEOUT                301

#define RGT_ERROR_INVALID_SESSION_OPTION 400

/* Resouce allocation error */
#define RGT_ERROR_ALLOC_RESOURCE         500


#define RGT_ERROR_TYPE_NO_ERROR ' '
#define RGT_ERROR_TYPE_APP      'A'
#define RGT_ERROR_TYPE_TRM      'T'
#define RGT_ERROR_TYPE_SRV      'S'

#define RGT_ERROR_TERMINAL 30001
#define RGT_ERROR_APP      30002

#endif
