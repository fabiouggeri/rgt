package terminal

import (
	"fmt"
	"rgt-server/protocol"
)

const (
	/* Unknown */
	TRM_UNKNOWN         protocol.OperationCode = 0x00
	TRM_TE_APP_RESPONSE protocol.OperationCode = 0xE0

	/* Terminal operations */
	TRM_TE_LOGIN       protocol.OperationCode = 0x0A
	TRM_TE_LOGOUT      protocol.OperationCode = 0x0B
	TRM_TE_RECONNECT   protocol.OperationCode = 0x0C
	TRM_TE_MIN_OP_CODE protocol.OperationCode = TRM_TE_LOGIN
	TRM_TE_MAX_OP_CODE protocol.OperationCode = TRM_TE_RECONNECT

	/* App operations */
	TRM_APP_LOGIN   protocol.OperationCode = 0x32
	TRM_APP_LOGOUT  protocol.OperationCode = 0x33
	TRM_APP_SET_ENV protocol.OperationCode = 0x34
	TRM_APP_UPDATE  protocol.OperationCode = 0x35
	// TRM_APP_READ_KEY       protocol.OperationCode = 0x36 --> Deprecated
	TRM_APP_RPC            protocol.OperationCode = 0x37
	TRM_APP_PUT_FILE       protocol.OperationCode = 0x38
	TRM_APP_GET_FILE       protocol.OperationCode = 0x39
	TRM_APP_KEY_BUFFER_LEN protocol.OperationCode = 0x3A
	TRM_APP_RECONNECT      protocol.OperationCode = 0x3B
	TRM_APP_KEEP_ALIVE     protocol.OperationCode = 0x3C
	TRM_APP_SESSION_CONFIG protocol.OperationCode = 0x3D
	TRM_APP_MIN_OP_CODE    protocol.OperationCode = TRM_APP_LOGIN
	TRM_APP_MAX_OP_CODE    protocol.OperationCode = TRM_APP_SESSION_CONFIG

	/* Standalone app operations */
	TRM_STANDALONE_APP_EXEC        protocol.OperationCode = 0x64
	TRM_STANDALONE_APP_SEND_OUTPUT protocol.OperationCode = 0x65
	TRM_STANDALONE_APP_SEND_STATUS protocol.OperationCode = 0x66
	TRM_STANDALONE_APP_MIN_OP_CODE protocol.OperationCode = TRM_STANDALONE_APP_EXEC
	TRM_STANDALONE_APP_MAX_OP_CODE protocol.OperationCode = TRM_STANDALONE_APP_SEND_STATUS

	/* Server operations */
	TRM_SRV_GET_SCREEN  protocol.OperationCode = 0x80
	TRM_SRV_MIN_OP_CODE protocol.OperationCode = TRM_SRV_GET_SCREEN
	TRM_SRV_MAX_OP_CODE protocol.OperationCode = TRM_SRV_GET_SCREEN

	SUCCESS                protocol.ResponseCode = 0
	UNKNOWN_ERROR          protocol.ResponseCode = 1
	SOCKET_ERROR           protocol.ResponseCode = 2
	PROTOCOL_ERROR         protocol.ResponseCode = 3
	RESPONSE_ERROR         protocol.ResponseCode = 4
	UNKNOWN_COMMAND_ERROR  protocol.ResponseCode = 5
	AUTHENTICATOR_ERROR    protocol.ResponseCode = 6
	CONNECTION_LOST_ERROR  protocol.ResponseCode = 7
	UNKNOWN_RESPONSE_ERROR protocol.ResponseCode = 8
	SESSION_CLOSED_ERROR   protocol.ResponseCode = 9

	TE_AUTH_ERROR        protocol.ResponseCode = 10
	TE_APP_LAUNCH_ERROR  protocol.ResponseCode = 11
	TE_INVALID_ARG_ERROR protocol.ResponseCode = 12

	APP_CONNECT_ERROR     protocol.ResponseCode = 20
	APP_INIT_GT_ERROR     protocol.ResponseCode = 21
	APP_SCREEN_BUSY_ERROR protocol.ResponseCode = 22

	FILESYSTEM_ERROR protocol.ResponseCode = 100
	CREATING_ERROR   protocol.ResponseCode = 101
	READING_ERROR    protocol.ResponseCode = 102
	WRITING_ERROR    protocol.ResponseCode = 103
	OPENING_ERROR    protocol.ResponseCode = 104

	INVALID_DATA_TYPE_ERROR  protocol.ResponseCode = 200
	INVALID_PAR_TYPE_ERROR   protocol.ResponseCode = 201
	UNDEFINED_FUNCTION_ERROR protocol.ResponseCode = 202
	DATA_CORRUPTION_ERROR    protocol.ResponseCode = 203

	/* Environment error */
	ENV_VAR_NOT_FOUND_ERROR protocol.ResponseCode = 300

	/* Timeout error */
	TIMEOUT_ERROR protocol.ResponseCode = 301

	INVALID_SESSION_OPTION_ERROR protocol.ResponseCode = 400
	ADMIN_CLIENT_NOT_FOUND_ERROR protocol.ResponseCode = 500
	UNKNOWN_HOST_ERROR           protocol.ResponseCode = 10000
	MAX_CODE_ERROR               protocol.ResponseCode = 31999 // 0x7CFF

	/* Response codes for server requests */
	TERMINAL_SEND_SCREEN protocol.ResponseCode = 0x7E00
	MIN_ADMIN_REQUEST    protocol.ResponseCode = TERMINAL_SEND_SCREEN
	MAX_ADMIN_REQUEST    protocol.ResponseCode = TERMINAL_SEND_SCREEN

	/* Data send to app without a request */
	APP_NOT_REQUESTED_RESP protocol.OperationCode = 0x7F
	TERMINAL_KEEP_ALIVE    protocol.ResponseCode  = 0x7FFE
	TERMINAL_KEY_UPDATE    protocol.ResponseCode  = 0x7FFF

	/* Admin Response code */
	ADMIN_REQUEST_RESP_OP_CODE protocol.OperationCode = 0xFF
	ADMIN_REQUEST_RESPONSE     protocol.ResponseCode  = 0xFF00

	SERVER_PROTOCOL_VERSION             int16 = 5
	TRM_STANDALONE_APP_PROTOCOL_VERSION int16 = 1
)

type TerminalError struct {
	message   string
	errorCode protocol.ResponseCode
}

var operationCodes = map[protocol.OperationCode]string{
	TRM_UNKNOWN:         "Unknown operation",
	TRM_TE_APP_RESPONSE: "TE/APP response",

	/* Terminal operations */
	TRM_TE_LOGIN:     "TE login",
	TRM_TE_LOGOUT:    "TE logout",
	TRM_TE_RECONNECT: "TE reconnect",

	/* App operations */
	TRM_APP_LOGIN:          "APP login",
	TRM_APP_LOGOUT:         "APP logout",
	TRM_APP_SET_ENV:        "APP set TE environment",
	TRM_APP_UPDATE:         "APP update TE screen",
	TRM_APP_RPC:            "APP remote procedure call",
	TRM_APP_PUT_FILE:       "APP put file",
	TRM_APP_GET_FILE:       "APP get file",
	TRM_APP_KEY_BUFFER_LEN: "APP set keyboard buffer len",
	TRM_APP_RECONNECT:      "APP reconnect",
	TRM_APP_KEEP_ALIVE:     "APP send keep alive",
	TRM_APP_SESSION_CONFIG: "APP session config",

	/* Standalone app operations */
	TRM_STANDALONE_APP_EXEC:        "Standalone APP exec",
	TRM_STANDALONE_APP_SEND_OUTPUT: "Standalone send output",
	TRM_STANDALONE_APP_SEND_STATUS: "Standalone send status",

	/* Server operations */
	TRM_SRV_GET_SCREEN: "ADMIN get TE screen",
}

var reponseCodes = map[protocol.ResponseCode]string{
	SUCCESS:                "Success",
	UNKNOWN_ERROR:          "Unknown error",
	SOCKET_ERROR:           "Socket error",
	PROTOCOL_ERROR:         "Protocol error",
	RESPONSE_ERROR:         "Response error",
	UNKNOWN_COMMAND_ERROR:  "Unknown command",
	AUTHENTICATOR_ERROR:    "Authenticator error",
	CONNECTION_LOST_ERROR:  "Connection lost",
	UNKNOWN_RESPONSE_ERROR: "Unknown response",
	SESSION_CLOSED_ERROR:   "Session closed",

	TE_AUTH_ERROR:        "Terminal Emulator authentication error",
	TE_APP_LAUNCH_ERROR:  "Error launching application",
	TE_INVALID_ARG_ERROR: "Invalid argument",

	APP_CONNECT_ERROR:     "Application connection error",
	APP_INIT_GT_ERROR:     "Error initializing GT",
	APP_SCREEN_BUSY_ERROR: "Screen operation not finished",

	CREATING_ERROR: "Error creating file",
	READING_ERROR:  "Error reading file",
	WRITING_ERROR:  "Error writing file",
	OPENING_ERROR:  "Error opening file",

	INVALID_DATA_TYPE_ERROR:  "Invalid data type",
	INVALID_PAR_TYPE_ERROR:   "invalida parameter type",
	UNDEFINED_FUNCTION_ERROR: "undefined function",
	DATA_CORRUPTION_ERROR:    "Data corruption",

	ENV_VAR_NOT_FOUND_ERROR: "Environment variable not found",
	TIMEOUT_ERROR:           "Timeout waiting response",

	INVALID_SESSION_OPTION_ERROR: "Invalid session option",
	UNKNOWN_HOST_ERROR:           "Unknown host",
	ADMIN_CLIENT_NOT_FOUND_ERROR: "admin client not found",
}

var EOFError = NewError(SOCKET_ERROR, "EOF")

func NewError(respCode protocol.ResponseCode, message ...any) *TerminalError {
	return &TerminalError{errorCode: respCode,
		message: fmt.Sprint(message...)}
}

func (e *TerminalError) GetResponseCode() protocol.ResponseCode {
	return e.errorCode
}

func (e *TerminalError) Error() string {
	if len(e.message) != 0 {
		return e.message
	} else {
		msg, found := reponseCodes[e.errorCode]
		if found {
			return msg
		} else {
			return fmt.Sprint("unknown error: ", e.errorCode)
		}
	}
}

func IsValidResponseCode(code protocol.ResponseCode) bool {
	_, found := reponseCodes[code]
	return found
}

func GetResponseCodeDescription(code protocol.ResponseCode) string {
	desc, found := reponseCodes[code]
	if found {
		return desc
	} else {
		return "unknown error"
	}
}

func GetOperationCodeDescription(opCode protocol.OperationCode) string {
	desc, found := operationCodes[opCode]
	if found {
		return desc
	} else {
		return fmt.Sprint("unknown operation: ", opCode)
	}
}
