package admin

import (
	"fmt"
	"rgt-server/protocol"
)

type AdminError struct {
	message   string
	errorCode protocol.ResponseCode
}

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
	ADM_CANCEL                protocol.OperationCode = 126
	ADM_UNKNOWN               protocol.OperationCode = 127
	ADM_MIN_OP_CODE           protocol.OperationCode = ADM_LOGIN
	ADM_MAX_OP_CODE           protocol.OperationCode = ADM_GET_SESSION_STATS

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
	UNKNOWN_ERROR               protocol.ResponseCode = 127

	ADMIN_PROTOCOL_VERSION int16 = 7
)

var reponseCodes = map[protocol.ResponseCode]string{
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
	UNKNOWN_ERROR:               "Unknown error"}

func NewError(respCode protocol.ResponseCode, message ...any) *AdminError {
	return &AdminError{errorCode: respCode,
		message: fmt.Sprint(message...)}
}

func (e *AdminError) GetResponseCode() protocol.ResponseCode {
	return e.errorCode
}

func (e *AdminError) Error() string {
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
