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
	adminProtocol.Operation(ADM_GET_USERS, "Get users").Version(0).Executor(serverGetUsers)
	adminProtocol.Operation(ADM_SET_USERS, "Set users").Version(0).Executor(serverSetUsers)
	adminProtocol.Operation(ADM_SAVE_USERS, "Save users").Version(0).Executor(serverSaveUsers)
	adminProtocol.Operation(ADM_LOAD_USERS, "Load users").Version(0).Executor(serverLoadUsers)
	adminProtocol.Operation(ADM_ADD_USER, "Add user").Version(0).Executor(serverAddUser)
	adminProtocol.Operation(ADM_REMOVE_USER, "Remove user").Version(0).Executor(serverRemoveUser)
}

func (resp *GetUsersResponse) FromBuffer(buf *buffer.ByteBuffer) {
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
}

func (resp *GetUsersResponse) ToBuffer(buf *buffer.ByteBuffer) {
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

func serverGetUsers(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverGetUsers()")
	srv := pack.handler.service.server
	resp := &GetUsersResponse{}
	if srv.GetUserRepository() != nil {
		resp.userAuthentication = true
		resp.users = append(resp.users, srv.GetUserRepository().GetUsers()...)
	} else {
		resp.userAuthentication = false
	}
	respBuf := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE +
		(len(resp.users) * (buffer.STRING_HEADER_SIZE + 20 + buffer.STRING_HEADER_SIZE + 20 + buffer.DATE_SIZE))))
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}

func (r *SetUsersRequest) FromBuffer(buf *buffer.ByteBuffer) {
	usersCount := int(buf.GetInt32())
	for i := 0; i < usersCount; i++ {
		user := &server.TerminalUser{Username: buf.GetString(), Password: buf.GetString()}
		expiration := buf.GetBool()
		if expiration {
			t := buf.GetDate()
			user.Expiration = &t
		}
		r.users = append(r.users, user)

	}
}

func (req *SetUsersRequest) ToBuffer(buf *buffer.ByteBuffer) {
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

func serverSetUsers(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverSetUsers()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	req := &SetUsersRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
	srv.GetUserRepository().ClearUsers()
	for _, u := range req.users {
		srv.GetUserRepository().AddUser(u)
	}
	return protocol.SuccessResponse(), nil
}

func serverSaveUsers(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverSaveUsers()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	srv.GetUserRepository().Save()
	return protocol.SuccessResponse(), nil
}

func serverLoadUsers(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverLoadUsers()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	srv.GetUserRepository().Load()
	return protocol.SuccessResponse(), nil
}

func (r *AddUserRequest) bufferToAddUserRequest(buf *buffer.ByteBuffer) {
	user := &server.TerminalUser{Username: buf.GetString(), Password: buf.GetString()}
	expiration := buf.GetBool()
	if expiration {
		t := buf.GetDate()
		user.Expiration = &t
	}
	r.user = user
}

func (r *AddUserRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.user.Username)
	buf.PutString(r.user.Password)
	buf.PutBool(r.user.Expiration != nil)
	if r.user.Expiration != nil {
		buf.PutDate(*r.user.Expiration)
	}
}

func serverAddUser(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverAddUser()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	req := AddUserRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
	ok, addErr := srv.GetUserRepository().AddUser(req.user)
	if addErr != nil {
		return nil, NewError(SERVER_ERROR, "Error adding terminal user: ", addErr)
	}
	if !ok {
		return nil, NewError(SERVER_ERROR, "User ", req.user.Username, " already exists")
	}
	return protocol.SuccessResponse(), nil
}

func (r *RemoveUserRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.username = buf.GetString()
}

func (r *RemoveUserRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.username)
}

func serverRemoveUser(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_users_operations.serverRemoveUser()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetUserRepository() == nil {
		return nil, NewError(SERVER_ERROR, "Server without user repository configured")
	}
	req := RemoveUserRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
	user := srv.GetUserRepository().RemoveUser(req.username)
	if user == nil {
		return nil, NewError(SERVER_ERROR, "User ", req.username, " not found")
	}
	return protocol.SuccessResponse(), nil
}
