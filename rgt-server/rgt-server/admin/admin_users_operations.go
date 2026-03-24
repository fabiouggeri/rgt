package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
)

type GetUsersResponse struct {
	protocol.BaseResponse
	users              []*server.TerminalUser
	userAuthentication bool
}

type SetUsersRequest struct {
	protocol.BaseRequest
	users []*server.TerminalUser
}

type AddUserRequest struct {
	protocol.BaseRequest
	user *server.TerminalUser
}

type RemoveUserRequest struct {
	protocol.BaseRequest
	username string
}

func init() {
	registerOperation(ADM_GET_USERS, serverGetUsers)
	registerOperation(ADM_SET_USERS, serverSetUsers)
	registerOperation(ADM_SAVE_USERS, serverSaveUsers)
	registerOperation(ADM_LOAD_USERS, serverLoadUsers)
	registerOperation(ADM_ADD_USER, serverAddUser)
	registerOperation(ADM_REMOVE_USER, serverRemoveUser)
	registerProtocol(ADM_GET_USERS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToGetUsersResponse, getUsersResponseToBuffer))
	registerProtocol(ADM_SET_USERS, 0, protocol.New(bufferToSetUsersRequest, setUsersRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_SAVE_USERS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_LOAD_USERS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_ADD_USER, 0, protocol.New(bufferToAddUserRequest, addUserRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_REMOVE_USER, 0, protocol.New(bufferToRemoveUserRequest, removeUserRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
}

func bufferToGetUsersResponse(buf *buffer.ByteBuffer) *GetUsersResponse {
	resp := &GetUsersResponse{}
	usersCount := int(buf.GetInt32())
	resp.userAuthentication = usersCount >= 0
	for i := 0; i < usersCount; i++ {
		user := &server.TerminalUser{Username: buf.GetString(), Password: buf.GetString()}
		expiration := buf.GetBool()
		if expiration {
			t := buf.GetDate()
			user.Expiration = &t
		}
		resp.users = append(resp.users, user)
	}
	return resp
}

func getUsersResponseToBuffer(resp *GetUsersResponse, buf *buffer.ByteBuffer) {
	if resp.userAuthentication {
		buf.PutInt32(int32(len(resp.users)))
		for _, user := range resp.users {
			buf.PutString(user.Username)
			buf.PutString(user.Password)
			if user.Expiration != nil {
				buf.PutBool(true)
				buf.PutDate(*user.Expiration)
			} else {
				buf.PutBool(false)
			}
		}
	} else {
		buf.PutInt32(server.AUTH_DISABLE)
	}
}

func serverGetUsers(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverGetUsers()")
	proto, err := findProtocol[*protocol.BaseRequest, *GetUsersResponse](ADM_GET_USERS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	resp := &GetUsersResponse{}
	if srv.GetUserRepository() != nil {
		resp.userAuthentication = true
		resp.users = append(resp.users, srv.GetUserRepository().GetUsers()...)
	} else {
		resp.userAuthentication = false
	}
	respBuf := buffer.New()
	proto.PutResponse(resp, respBuf)
	return respBuf, nil
}

func bufferToSetUsersRequest(buf *buffer.ByteBuffer) *SetUsersRequest {
	req := &SetUsersRequest{}
	usersCount := int(buf.GetInt32())
	for i := 0; i < usersCount; i++ {
		user := &server.TerminalUser{Username: buf.GetString(), Password: buf.GetString()}
		expiration := buf.GetBool()
		if expiration {
			t := buf.GetDate()
			user.Expiration = &t
		}
		req.users = append(req.users, user)

	}
	return req
}

func setUsersRequestToBuffer(req *SetUsersRequest, buf *buffer.ByteBuffer) {
	buf.PutInt32(int32(len(req.users)))
	for _, u := range req.users {
		buf.PutString(u.Username)
		buf.PutString(u.Password)
		buf.PutBool(u.Expiration != nil)
		if u.Expiration != nil {
			buf.PutDate(*u.Expiration)
		}
	}
}

func serverSetUsers(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverSetUsers()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	proto, err := findProtocol[*SetUsersRequest, *protocol.BaseResponse](ADM_SET_USERS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	srv.GetUserRepository().ClearUsers()
	for _, u := range req.users {
		srv.GetUserRepository().AddUser(u)
	}
	return SuccessAdminResponse(), nil
}

func serverSaveUsers(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverSaveUsers()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	_, err := findProtocol[*protocol.BaseRequest, *protocol.BaseResponse](ADM_SAVE_USERS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv.GetUserRepository().Save()
	return SuccessAdminResponse(), nil
}

func serverLoadUsers(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverLoadUsers()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	_, err := findProtocol[*protocol.BaseRequest, *protocol.BaseResponse](ADM_LOAD_USERS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv.GetUserRepository().Load()
	return SuccessAdminResponse(), nil
}

func bufferToAddUserRequest(buf *buffer.ByteBuffer) *AddUserRequest {
	user := &server.TerminalUser{Username: buf.GetString(), Password: buf.GetString()}
	expiration := buf.GetBool()
	if expiration {
		t := buf.GetDate()
		user.Expiration = &t
	}
	return &AddUserRequest{user: user}
}

func addUserRequestToBuffer(req *AddUserRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.user.Username)
	buf.PutString(req.user.Password)
	buf.PutBool(req.user.Expiration != nil)
	if req.user.Expiration != nil {
		buf.PutDate(*req.user.Expiration)
	}
}

func serverAddUser(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverAddUser()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	proto, err := findProtocol[*AddUserRequest, *protocol.BaseResponse](ADM_ADD_USER, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	ok, addErr := srv.GetUserRepository().AddUser(req.user)
	if addErr != nil {
		return nil, NewError(SERVER_ERROR, "Error adding terminal user: ", addErr)
	}
	if !ok {
		return nil, NewError(SERVER_ERROR, "User ", req.user.Username, " already exists")
	}
	return SuccessAdminResponse(), nil
}

func bufferToRemoveUserRequest(buf *buffer.ByteBuffer) *RemoveUserRequest {
	return &RemoveUserRequest{username: buf.GetString()}
}

func removeUserRequestToBuffer(req *RemoveUserRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.username)
}

func serverRemoveUser(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverRemoveUser()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	proto, err := findProtocol[*RemoveUserRequest, *protocol.BaseResponse](ADM_REMOVE_USER, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	user := srv.GetUserRepository().RemoveUser(req.username)
	if user == nil {
		return nil, NewError(SERVER_ERROR, "User ", req.username, " not found")
	}
	return SuccessAdminResponse(), nil
}
