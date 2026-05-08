package admin

import (
	"rgt-server/protocol"
)

const (
	ADM_LOGIN                 protocol.OperationCode = 1
	ADM_LOGOFF                protocol.OperationCode = 2
	ADM_STOP_SERVICE          protocol.OperationCode = 3
	ADM_START_SERVICE         protocol.OperationCode = 4
	ADM_GET_SESSIONS          protocol.OperationCode = 5
	ADM_GET_STATUS            protocol.OperationCode = 6
	ADM_KILL_SESSION          protocol.OperationCode = 7
	ADM_KILL_ALL_SESSIONS     protocol.OperationCode = 8
	ADM_SET_CONFIG            protocol.OperationCode = 9
	ADM_GET_CONFIG            protocol.OperationCode = 10
	ADM_SAVE_CONFIG           protocol.OperationCode = 11
	ADM_LOAD_CONFIG           protocol.OperationCode = 12
	ADM_SET_LOG_LEVEL         protocol.OperationCode = 13
	ADM_GET_USERS             protocol.OperationCode = 14
	ADM_SET_USERS             protocol.OperationCode = 15
	ADM_SAVE_USERS            protocol.OperationCode = 16
	ADM_LOAD_USERS            protocol.OperationCode = 17
	ADM_ADD_USER              protocol.OperationCode = 18
	ADM_REMOVE_USER           protocol.OperationCode = 19
	ADM_KILL_ADMIN_SESSIONS   protocol.OperationCode = 20
	ADM_LIST_FILES            protocol.OperationCode = 21
	ADM_GET_FILE              protocol.OperationCode = 22
	ADM_PUT_FILE              protocol.OperationCode = 23
	ADM_REMOVE_FILE           protocol.OperationCode = 24
	ADM_SEND_TERMINAL_REQUEST protocol.OperationCode = 25
	ADM_GET_STATS             protocol.OperationCode = 26
	ADM_GET_SESSION_STATS     protocol.OperationCode = 27
	ADM_MIN_OP_CODE           protocol.OperationCode = ADM_LOGIN
	ADM_MAX_OP_CODE           protocol.OperationCode = ADM_GET_SESSION_STATS
	ADM_CANCEL                protocol.OperationCode = 126
	ADM_UNKNOWN               protocol.OperationCode = 0xFF

	SUCCESS                     protocol.ResponseCode = 0
	SERVER_ERROR                protocol.ResponseCode = 10
	INVALID_STATUS              protocol.ResponseCode = 11
	SESSION_NOT_FOUND           protocol.ResponseCode = 12
	ADMIN_SESSION_ALREADY_OPEN  protocol.ResponseCode = 13
	NOT_LOGGED                  protocol.ResponseCode = 14
	UNKNOWN_COMMAND             protocol.ResponseCode = 15
	ERROR_KILLING_ADMIN_SESSION protocol.ResponseCode = 16
	INVALID_CREDENTIAL          protocol.ResponseCode = 17
	PROTOCOL_ERROR              protocol.ResponseCode = 18
	SOCKET                      protocol.ResponseCode = 19
	CONNECTION_LOST             protocol.ResponseCode = 20
	NOT_ALLOWED_OPERATION       protocol.ResponseCode = 21
	FILE_READING_ERROR          protocol.ResponseCode = 22
	FILE_WRITING_ERROR          protocol.ResponseCode = 23
	AUTHENTICATOR_ERROR         protocol.ResponseCode = 24
	UNKNOWN_ERROR               protocol.ResponseCode = 0x7CFF // 32767

	ADMIN_PROTOCOL_VERSION int16 = 7
)

var (
	operationCodes = map[protocol.OperationCode]string{
		ADM_LOGIN:                 "Admin login",
		ADM_LOGOFF:                "Admin logoff",
		ADM_STOP_SERVICE:          "Stop service",
		ADM_START_SERVICE:         "Start service",
		ADM_GET_SESSIONS:          "Get sessions",
		ADM_GET_STATUS:            "Get status",
		ADM_KILL_SESSION:          "Kill session",
		ADM_KILL_ALL_SESSIONS:     "Kill all sessions",
		ADM_SET_CONFIG:            "Set config",
		ADM_GET_CONFIG:            "Get config",
		ADM_SAVE_CONFIG:           "Save config",
		ADM_LOAD_CONFIG:           "Load config",
		ADM_SET_LOG_LEVEL:         "Set log level",
		ADM_GET_USERS:             "Get users",
		ADM_SET_USERS:             "Set users",
		ADM_SAVE_USERS:            "Save users",
		ADM_LOAD_USERS:            "Load users",
		ADM_ADD_USER:              "Add user",
		ADM_REMOVE_USER:           "Remove user",
		ADM_KILL_ADMIN_SESSIONS:   "Kill admin sessions",
		ADM_LIST_FILES:            "List files",
		ADM_GET_FILE:              "Get file",
		ADM_PUT_FILE:              "Put file",
		ADM_REMOVE_FILE:           "Remove file",
		ADM_SEND_TERMINAL_REQUEST: "Send terminal request",
		ADM_GET_STATS:             "Get stats",
		ADM_GET_SESSION_STATS:     "Get session stats",
		ADM_CANCEL:                "Cancel operation",
		ADM_UNKNOWN:               "Unknown",
	}
	adminProtocol = protocol.New[*RequestPack](map[protocol.ResponseCode]string{
		SUCCESS:                     "Success",
		SERVER_ERROR:                "Server error",
		INVALID_STATUS:              "Invalid serve status",
		SESSION_NOT_FOUND:           "Session not found",
		ADMIN_SESSION_ALREADY_OPEN:  "Another administrative session is open",
		NOT_LOGGED:                  "Not logged",
		UNKNOWN_COMMAND:             "Unknown command",
		ERROR_KILLING_ADMIN_SESSION: "Error killing admin session",
		INVALID_CREDENTIAL:          "Invalid credential",
		PROTOCOL_ERROR:              "Protocol error",
		SOCKET:                      "Socket error",
		CONNECTION_LOST:             "Connection lost",
		NOT_ALLOWED_OPERATION:       "Operation not allowed",
		FILE_READING_ERROR:          "File reading error",
		FILE_WRITING_ERROR:          "File writing error",
		UNKNOWN_ERROR:               "Unknown error"})
)

func NewOperation(code protocol.OperationCode, description string) *protocol.Operation[*RequestPack] {
	return adminProtocol.Operation(code, description)
}

func NewError(respCode protocol.ResponseCode, message ...any) *protocol.ProtocolError {
	return adminProtocol.NewError(respCode, message...)
}
