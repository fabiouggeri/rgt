package admin

import (
	"errors"
	"fmt"
	"io"
	"net"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/util"
)

type AdminHandler struct {
	id              uint64
	username        string
	conn            *net.TCPConn
	service         *AdminService
	headerBuffer    []byte
	protocolVersion int16
	readOnly        bool
}

type requestPack struct {
	operation protocol.OperationCode
	handler   *AdminHandler
	body      []byte
}

type operationHandle func(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse)

var operations map[protocol.OperationCode]operationHandle = make(map[protocol.OperationCode]operationHandle)

func newHandler(handlerId uint64, conn *net.TCPConn, adminService *AdminService) *AdminHandler {
	return &AdminHandler{id: handlerId,
		readOnly:        true,
		protocolVersion: ADMIN_PROTOCOL_VERSION,
		conn:            conn,
		service:         adminService,
		headerBuffer:    make([]byte, protocol.HEADER_SIZE)}
}

func (h *AdminHandler) Id() uint64 {
	return h.id
}

func (h *AdminHandler) GetRemoteAddr() string {
	return h.conn.RemoteAddr().String()
}

func (h *AdminHandler) GetUsername() string {
	return h.username
}

func (h *AdminHandler) SetUsername(username string) {
	h.username = username
}

func (h *AdminHandler) Send(buf *buffer.ByteBuffer) error {
	if h.conn != nil {
		//_, err := h.conn.Write(buf.RemainingSlice())
		_, err := h.conn.Write(buf.GetBytes())
		if err != nil {
			return fmt.Errorf("error writing to %v: %v", h.GetRemoteAddr(), err)
		}
	}
	return nil
}

func (h *AdminHandler) Close() error {
	h.service.unregisterHandler(h.id)
	err := h.conn.Close()
	if err != nil {
		return NewError(SOCKET, err)
	}
	return nil
}

func (h *AdminHandler) Connected() bool {
	return h.conn != nil
}

func (h *AdminHandler) finishHandle() {
	h.Close()
}

func (h *AdminHandler) readPacket() (*requestPack, error) {
	_, err := io.ReadFull(h.conn, h.headerBuffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Error("[ADMIN] Socket channel closed when reading header")
			return nil, nil
		} else {
			return nil, NewError(SOCKET, "Socket channel closed when reading header: ", err.Error())
		}
	}
	buf := buffer.Wrap(h.headerBuffer)
	bodyLen := buf.GetUInt32()
	if bodyLen == 0 {
		log.Error("[ADMIN] Invalid body len in admin message: ", bodyLen)
		return nil, NewError(PROTOCOL_ERROR, "Invalid body len in admin message: ", bodyLen)
	}
	bodyLen--
	opCode := protocol.OperationCode(buf.GetUInt8())
	pack := &requestPack{operation: opCode, handler: h}
	if bodyLen > 0 {
		pack.body = make([]byte, bodyLen)
		_, err = io.ReadFull(h.conn, pack.body)
		if err != nil {
			err = NewError(SOCKET, "[ADMIN] error reading data: ", err.Error())
			pack = nil
		}
	}
	return pack, err
}

func (h *AdminHandler) processRequest(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	execOperation, found := operations[pack.operation]
	if found {
		return execOperation(pack)
	} else {
		return nil, NewError(PROTOCOL_ERROR, "protocol not found ", pack.operation)
	}
}

func (h *AdminHandler) Handle() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("[ADMIN] unknown error in server(AdminHandler.Handle): %v\n%s", err, util.FullStack())
		}
	}()
	defer h.finishHandle()
	log.Debug("[ADMIN] Serving client ", h.GetRemoteAddr())
	for {
		pack, err := h.readPacket()
		if err != nil {
			log.Error("[ADMIN] ", err)
			break
		} else if pack == nil {
			break
		} else {
			respBuf, requestErr := h.processRequest(pack)
			if requestErr != nil {
				err = h.sendError(requestErr)
			} else if respBuf != nil {
				err = h.sendResponse(respBuf)
			}
			if err != nil {
				log.Error("[ADMIN] ", requestErr)
				break
			}
		}
	}
}

func (h *AdminHandler) sendError(err protocol.ErrorResponse) error {
	proto := protocol.NewDefault()
	errMsg := err.Error()
	errBuff := buffer.NewCapacity(uint32(8 + len(errMsg)))
	resp := protocol.ResponseFromError(err)
	proto.PutResponse(resp, errBuff)
	return h.sendResponse(errBuff)
}

func (h *AdminHandler) sendResponse(buf *buffer.ByteBuffer) error {
	err := h.Send(buf)
	if err != nil {
		return NewError(SOCKET, err.Error())
	}
	return nil
}

func registerOperation(op protocol.OperationCode, opFun operationHandle) {
	operations[op] = opFun
}

func SuccessAdminResponse() *buffer.ByteBuffer {
	proto := protocol.NewDefault()
	respBuf := buffer.NewCapacity(8)
	proto.PutResponse(protocol.NewResponse(SUCCESS), respBuf)
	return respBuf
}
