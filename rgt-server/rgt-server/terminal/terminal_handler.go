package terminal

import (
	"errors"
	"fmt"
	"io"
	"net"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"rgt-server/service"
	"rgt-server/util"
	"sync"
	"sync/atomic"
	"time"
)

type TerminalHandler struct {
	id                   uint64
	session              *server.Session
	service              *TerminalEmulationService
	conn                 *net.TCPConn
	receivedPackets      chan *buffer.ByteBuffer
	packetsToSend        chan *buffer.ByteBuffer
	lastDataReadTime     time.Time
	lastAppOperationTime time.Time
	adminClients         map[uint64]*adminClient
	adminClientsMutex    sync.RWMutex
	remoteAddres         string
	endpoint             service.TerminalConnectionHandler
	finished             atomic.Bool
	waitWorkers          sync.WaitGroup
	protocolVersion      int16
	connectionType       service.ConnectionType
	headerBuffer         []byte
}

type requestPack struct {
	handler  *TerminalHandler
	opCode   protocol.OperationCode
	bodySize uint32
	packet   *buffer.ByteBuffer
}

type operationHandle func(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse)

var _ service.ConnectionHandler = &TerminalHandler{}

var final_packet *buffer.ByteBuffer = buffer.Wrap([]byte{})

var operations map[protocol.OperationCode]operationHandle = make(map[protocol.OperationCode]operationHandle)

func newHandler(handlerId uint64, conn *net.TCPConn, terminalService *TerminalEmulationService) *TerminalHandler {
	return &TerminalHandler{id: handlerId,
		conn:            conn,
		remoteAddres:    conn.RemoteAddr().String(),
		protocolVersion: SERVER_PROTOCOL_VERSION,
		session:         nil,
		connectionType:  service.UNKNOWN,
		endpoint:        nil,
		service:         terminalService,
		receivedPackets: make(chan *buffer.ByteBuffer, 1024),
		packetsToSend:   make(chan *buffer.ByteBuffer, 1024),
		adminClients:    make(map[uint64]*adminClient),
		headerBuffer:    make([]byte, protocol.HEADER_SIZE),
	}
}

func (h *TerminalHandler) Id() uint64 {
	return h.id
}

func (h *TerminalHandler) sessionId() int64 {
	if h.session != nil {
		return h.session.Id
	}
	return 0
}

func (h *TerminalHandler) Connected() bool {
	return h.conn != nil
}

func (h *TerminalHandler) Send(buf *buffer.ByteBuffer) error {
	if !h.Connected() {
		log.Debugf("[%s;session=%d] TerminalHandler.Send(). error sending data: connection closed", h.connectionType, h.sessionId())
		return io.EOF
	}
	if h.finished.Load() {
		log.Debugf("[%s;session=%d] TerminalHandler.Send(). error sending data: handler finished", h.connectionType, h.sessionId())
		return io.EOF
	}
	h.packetsToSend <- buf
	log.Tracef("[%s;session=%d] TerminalHandler.Send() sent data='%v' ", h.connectionType, h.sessionId(), buf)
	return nil
}

func (h *TerminalHandler) write(buf []byte) (int, error) {
	if !h.Connected() {
		return 0, fmt.Errorf("connection lost to %v", h.GetRemoteAddr())
	}
	sent, err := h.conn.Write(buf)
	log.Tracef("[%s;session=%d] TerminalHandler.write(). sent=%d data='%v'", h.connectionType, h.sessionId(), sent, buffer.Wrap(buf))
	return sent, err
}

func (h *TerminalHandler) readAll(readBuffer []byte) error {
	if h.Connected() {
		read, err := io.ReadFull(h.conn, readBuffer)
		log.Tracef("[%s;session=%d] TerminalHandler.readAll() read=%d data='%v' ", h.connectionType, h.sessionId(), read, buffer.Wrap(readBuffer))
		return err
	}
	return nil
}

func (h *TerminalHandler) read(readBuffer []byte) (int, protocol.ErrorResponse) {
	if h.Connected() {
		read, err := h.conn.Read(readBuffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, EOFError
			} else {
				return read, NewError(SOCKET, "error reading from ", h.GetRemoteAddr(), ": ", err)
			}
		} else if read == 0 {
			return 0, EOFError
		}
		log.Tracef("[%s;session=%d] TerminalHandler.read() read=%d data='%v' ", h.connectionType, h.sessionId(), read, buffer.Wrap(readBuffer))
		return read, nil
	} else {
		log.Debugf("[%s;session=%d] TerminalHandler.read(). connection closed.", h.connectionType, h.sessionId())
	}
	return 0, EOFError
}

func (h *TerminalHandler) handleNewConnection() bool {
	packet, err := h.readFirstPacket()
	if err != nil {
		h.sendError(err)
		return false
	}
	bodySize := packet.GetUInt32()
	opCode := protocol.OperationCode(packet.GetUInt8())

	return h.runOperation(bodySize, opCode, packet)
}

func (h *TerminalHandler) processTeLogin(packet *buffer.ByteBuffer) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("TerminalHandler.processTeLogin(). handler=", h.id)
	h.connectionType = service.TERMINAL
	h.protocolVersion = packet.GetInt16()
	proto, err := findProtocol[*TeLoginRequest, *TeLoginResponse](TRM_TE_LOGIN, h.protocolVersion)
	if err != nil {
		return nil, err
	}
	session, err := teLogin(h.service, proto.GetRequest(packet), h)
	if err != nil {
		return nil, err
	}
	h.session = session
	config := h.service.server.Config()
	response := NewTeLoginResponse(session.Id, config.TeLogLevel().Get(), config.TeLogPathName().Get())
	proto.PutResponse(response, packet)
	return packet, nil
}

func (h *TerminalHandler) processAppLogin(packet *buffer.ByteBuffer) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("TerminalHandler.processAppLogin(). handler=", h.id)
	h.connectionType = service.APPLICATION
	h.protocolVersion = packet.GetInt16()
	proto, err := findProtocol[*AppLoginRequest, *AppLoginResponse](TRM_APP_LOGIN, h.protocolVersion)
	if err != nil {
		return nil, err
	}
	session, err := appLogin(h.service.server, proto.GetRequest(packet), h)
	if err != nil {
		return nil, err
	}
	session.SetStatus(server.SESS_READY)
	h.session = session
	response := &AppLoginResponse{LogLevel: h.service.server.Config().AppLogLevel().Get(),
		LogPathName: util.RelativePathToAbsolute(h.service.server.Config().AppLogPathName().Get())}
	proto.PutResponse(response, packet)
	return packet, nil
}

func (h *TerminalHandler) processExecStandaloneApp(packet *buffer.ByteBuffer) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("TerminalHandler.processExecStandaloneApp(). handler=", h.id)
	h.connectionType = service.LAUNCHER
	h.protocolVersion = packet.GetInt16()
	proto, err := findProtocol[*AppExecRequest, *AppExecResponse](TRM_STANDALONE_APP_EXEC, h.protocolVersion)
	if err != nil {
		return nil, err
	}
	session, err := executeStandaloneApp(h.service, proto.GetRequest(packet), h, h.protocolVersion)
	if err != nil {
		return nil, err
	}
	h.session = session
	response := &AppExecResponse{SessionId: session.Id,
		Pid: session.AppPid}
	proto.PutResponse(response, packet)
	return packet, nil
}

func (h *TerminalHandler) runOperation(bodySize uint32, opCode protocol.OperationCode, packet *buffer.ByteBuffer) bool {
	var resp *buffer.ByteBuffer
	var err protocol.ErrorResponse
	if execOperation, found := operations[opCode]; found {
		pack := &requestPack{handler: h,
			opCode:   opCode,
			bodySize: bodySize,
			packet:   packet}
		resp, err = execOperation(pack)
		if resp != nil {
			if h.Send(resp) == nil {
				return true
			} else {
				err = EOFError
			}
		}
	} else {
		err = NewError(PROTOCOL, "[", h.connectionType, ";session=", h.sessionId(), "] Unknonwn operation: ", opCode)
	}
	if err != nil {
		h.sendError(err)
	}
	return err != nil
}

func (h *TerminalHandler) sendToEndpoint(packet *buffer.ByteBuffer) protocol.ErrorResponse {
	if h.endpoint == nil {
		log.Debugf("[%s;session=%d] Undefined endpoint", h.connectionType, h.sessionId())
		return nil
	}
	h.lastAppOperationTime = time.Now()
	err := h.endpoint.Send(packet)
	if errors.Is(err, io.EOF) {
		return EOFError
	} else if err != nil {
		return NewError(SOCKET, err)
	}
	return nil
}

func (h *TerminalHandler) sendToAdminClient(adminId uint64, packet *buffer.ByteBuffer) protocol.ErrorResponse {
	var err protocol.ErrorResponse
	h.adminClientsMutex.RLock()
	admin, found := h.adminClients[adminId]
	h.adminClientsMutex.RUnlock()
	if found {
		admin.ProcessPacket(packet)
	} else {
		err = NewError(ADMIN_CLIENT_NOT_FOUND)
	}
	return err
}

func (h *TerminalHandler) readFirstPacket() (*buffer.ByteBuffer, protocol.ErrorResponse) {
	err := h.readAll(h.headerBuffer)
	if errors.Is(err, io.EOF) {
		return nil, EOFError
	} else if err != nil {
		return nil, NewError(SOCKET, err)
	}
	header := buffer.Wrap(h.headerBuffer)
	magicNumber := header.GetInt32()
	if magicNumber != protocol.MAGIC_NUMBER {
		return nil, NewError(PROTOCOL, "Invalid magic number in header: ", magicNumber)
	}
	bodySize := header.GetUInt32()
	if bodySize == 0 {
		return nil, NewError(PROTOCOL, "Invalid body len in message: ", bodySize)
	}
	packet := buffer.NewLen(uint32(protocol.PACK_SIZE_FIELD_SIZE) + bodySize)
	packet.PutUInt32(bodySize)
	packet.PutByte(h.headerBuffer[protocol.FIRST_HEADER_SIZE-1])
	for packet.Remaining() > 0 {
		read, err := h.read(packet.RemainingSlice())
		if err != nil {
			return nil, err
		}
		packet.Skip(read)
	}
	packet.Flip()
	return packet, nil
}

func (h *TerminalHandler) readPacket() (*buffer.ByteBuffer, protocol.ErrorResponse) {
	err := h.readAll(h.headerBuffer)
	// Ignore EOF if not started to read a packet
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, EOFError
		} else {
			return nil, NewError(SOCKET, err)
		}
	}
	header := buffer.Wrap(h.headerBuffer)
	bodySize := header.GetUInt32()
	if bodySize == 0 {
		return nil, NewError(PROTOCOL, "Invalid body len in message: ", bodySize)
	}
	packet := buffer.NewLen(uint32(protocol.PACK_SIZE_FIELD_SIZE) + bodySize)
	packet.Put(h.headerBuffer)
	for packet.Remaining() > 0 {
		read, err := h.read(packet.RemainingSlice())
		if err != nil {
			return nil, err
		}
		packet.Skip(read)
	}
	packet.Flip()
	return packet, nil
}

func (h *TerminalHandler) sendPacket(packet *buffer.ByteBuffer) bool {
	if !h.Connected() {
		return false
	}
	for packet.Remaining() > 0 {
		writed, err := h.write(packet.RemainingSlice())
		packet.Skip(writed)
		if err != nil {
			log.Debugf("[%s;session=%d] TerminalHandler.sendPacket(). error writing: %v", h.connectionType, h.sessionId(), err)
			return false
		}
	}
	return true
}

func (h *TerminalHandler) handlePanic(message string) {
	if err := recover(); err != nil {
		log.Errorf("[%s;session=%d] %s: %v\n%s", h.connectionType, h.sessionId(), message, err, util.FullStack())
	}
}

func (h *TerminalHandler) finishWorker(workerName string) {
	h.waitWorkers.Done()
	log.Debugf("[%s;session=%d] %s: finished", h.connectionType, h.sessionId(), workerName)
}

func (h *TerminalHandler) sendPackets() {
	log.Debugf("TerminalHandler.sendPackets(): started")
	defer h.handlePanic("unknown error in server (TerminalHandler.sendPackets)")
	defer h.finishWorker("TerminalHandler.sendPackets()")
	for packet := range h.packetsToSend {
		if packet == final_packet {
			return
		} else if !h.sendPacket(packet) {
			log.Debugf("[%s;session=%d] TerminalHandler.sendPackets(). finishing by failure to send packet", h.connectionType, h.sessionId())
			return
		}
	}
}

func (h *TerminalHandler) processAppPackets() {
	log.Debugf("TerminalHandler.processaAppPackets(): started")
	defer h.handlePanic("unknown error in server (TerminalHandler.processAppPackets)")
	defer h.finishWorker("TerminalHandler.processAppPackets()")
	for packet := range h.receivedPackets {
		if packet == final_packet {
			return
		}
		h.handleAppPacket(packet)
	}
}

func (h *TerminalHandler) handleAppPacket(packet *buffer.ByteBuffer) {
	bodySize := packet.GetUInt32()
	opCode := protocol.OperationCode(packet.GetUInt8())
	h.runOperation(bodySize, opCode, packet)
}

func (h *TerminalHandler) processTrmPackets() {
	log.Debugf("TerminalHandler.processTrmPackets(): started")
	defer h.handlePanic("unknown error in server (TerminalHandler.processTrmPackets)")
	defer h.finishWorker("TerminalHandler.processTrmPackets()")
	for packet := range h.receivedPackets {
		if packet == final_packet {
			return
		}
		h.handleTrmPacket(packet)
	}
}

func (h *TerminalHandler) handleTrmPacket(packet *buffer.ByteBuffer) {
	bodySize := packet.GetUInt32()
	respCode := protocol.ResponseCode(packet.GetUInt16())
	if respCode == ADMIN_REQUEST_RESPONSE {
		h.runOperation(bodySize, ADMIN_REQUEST_RESP_OP_CODE, packet)
	} else {
		h.runOperation(bodySize, TRM_TE_APP_RESPONSE, packet)
	}
}

func (h *TerminalHandler) readPackets() {
	log.Debugf("TerminalHandler.readPackets(): started")
	defer log.Debugf("[%s;session=%d] TerminalHandler.readPackets(): finished", h.connectionType, h.sessionId())
	for {
		packet, err := h.readPacket()
		if err != nil {
			if !h.finished.Load() {
				if errors.Is(err, EOFError) {
					log.Infof("[%s;session=%d] TerminalHandler.readPackets(). connection closed.", h.connectionType, h.sessionId())
				} else {
					log.Errorf("[%s;session=%d] TerminalHandler.readPackets(). error reading: %v", h.connectionType, h.sessionId(), err)
				}
			} else {
				log.Debugf("[%s;session=%d] TerminalHandler.readPackets(). error reading: %v", h.connectionType, h.sessionId(), err)
			}
			return
		}
		h.receivedPackets <- packet
		h.lastDataReadTime = time.Now()
	}
}

func (h *TerminalHandler) Handle() {
	log.Debugf("TerminalHandler.Handle(). started. handle=%d addr=%s", h.id, h.GetRemoteAddr())
	defer h.handlePanic("unknown error in server (TerminalHandler.Handle)")

	h.waitWorkers.Add(1)
	go h.sendPackets()
	if !h.handleNewConnection() {
		h.Close()
		log.Debugf("TerminalHandler.Handle(). error handling new connection. handle=%d", h.id)
		return
	}
	defer func() {
		h.service.server.CloseSession(h.sessionId())
		log.Debugf("[%s;session=%d] TerminalHandler.Handle(). finished. handle=%d", h.connectionType, h.sessionId(), h.id)
	}()
	if h.connectionType == service.TERMINAL {
		h.waitWorkers.Add(1)
		go h.processTrmPackets()
	} else {
		h.waitWorkers.Add(1)
		go h.processAppPackets()
	}
	h.readPackets()
}

func (h *TerminalHandler) GetLastDataReadTime() time.Time {
	return h.lastDataReadTime
}

func (h *TerminalHandler) GetLastAppOperationTime() time.Time {
	return h.lastAppOperationTime
}

func (h *TerminalHandler) Close() error {
	log.Debugf("[%s;session=%d] TerminalHandler.Close()", h.connectionType, h.sessionId())
	if !h.finished.CompareAndSwap(false, true) {
		log.Debugf("[%s;session=%d] TerminalHandler.Close(): close already called", h.connectionType, h.sessionId())
		return nil
	}
	h.conn.CloseRead()
	// Use select with timeout to prevent blocking if channels are full
	select {
	case h.packetsToSend <- final_packet:
	case <-time.After(100 * time.Millisecond):
		log.Debugf("[%s;session=%d] TerminalHandler.Close(): timeout sending final_packet to packetsToSend", h.connectionType, h.sessionId())
	}
	select {
	case h.receivedPackets <- final_packet:
	case <-time.After(100 * time.Millisecond):
		log.Debugf("[%s;session=%d] TerminalHandler.Close(): timeout sending final_packet to receivedPackets", h.connectionType, h.sessionId())
	}
	h.waitWorkers.Wait()
	close(h.receivedPackets)
	close(h.packetsToSend)
	h.conn.Close()
	log.Debugf("[%s;session=%d] TerminalHandler.Close(): closed", h.connectionType, h.sessionId())
	return nil
}

func (h *TerminalHandler) sendError(err protocol.ErrorResponse) error {
	log.Debugf("[%s;session=%d] TerminalHandler.SendError(). error=%s", h.connectionType, h.sessionId(), err.Error())
	if err == EOFError {
		return nil
	}
	proto := protocol.NewDefault()
	errMsg := err.Error()
	errBuff := buffer.NewCapacity(uint32(8 + len(errMsg)))
	resp := protocol.ResponseFromError(err)
	proto.PutResponse(resp, errBuff)
	errSend := h.Send(errBuff)
	return errSend
}

func (h *TerminalHandler) SendLogout(message string) {
	log.Debugf("[%s;session=%d] TerminalHandler.SendLogout() message='%s'", h.connectionType, h.sessionId(), message)
	buf := buffer.NewCapacity(uint32(len(message) + 12))
	buf.PutUInt32(0)
	buf.PutInt8(int8(TRM_APP_LOGOUT))
	buf.PutBool(false) // screen update?
	buf.PutInt16(0)    // Tones?
	buf.PutString(message)
	pos := buf.Position()
	buf.Flip()
	buf.PutUInt32(uint32(pos - 4))
	buf.Rewind()
	h.Send(buf)
}

func (h *TerminalHandler) SetEndpoint(endpoint service.TerminalConnectionHandler) {
	h.endpoint = endpoint
}

func (h *TerminalHandler) GetEndpoint() service.TerminalConnectionHandler {
	return h.endpoint
}

func (h *TerminalHandler) GetRemoteAddr() string {
	return h.remoteAddres
}

func (h *TerminalHandler) RegisterAdminClient(conn service.ConnectionHandler) service.AdminClient {
	h.adminClientsMutex.Lock()
	defer h.adminClientsMutex.Unlock()
	admin, found := h.adminClients[conn.Id()]
	if !found {
		admin = newAdminClient(h, conn)
		h.adminClients[conn.Id()] = admin
	}
	return admin
}

func (h *TerminalHandler) UnregisterAdminClient(conn service.ConnectionHandler) {
	h.adminClientsMutex.Lock()
	defer h.adminClientsMutex.Unlock()
	delete(h.adminClients, conn.Id())
}

func registerOperation(op protocol.OperationCode, opFun operationHandle) {
	operations[op] = opFun
	log.Debugf("TerminalHandler.registerOperation() code=%v description=%s", op, GetOperationCodeDescription(op))
}

func TrmAppSuccessResponse() *buffer.ByteBuffer {
	proto := protocol.NewDefault()
	respBuf := buffer.NewCapacity(8)
	proto.PutTrmAppResponse(protocol.NewResponse(SUCCESS), respBuf)
	return respBuf
}

func TrmAppErrorResponse(responseCode protocol.ResponseCode, messages ...any) *buffer.ByteBuffer {
	errorMessage := fmt.Sprint(messages...)
	proto := protocol.NewDefault()
	respBuf := buffer.NewCapacity(uint32(6 + len(errorMessage)))
	respBuf.PutString(errorMessage)
	proto.PutTrmAppResponse(protocol.NewResponse(responseCode), respBuf)
	return respBuf
}
