package terminal

import (
	"errors"
	"fmt"
	"io"
	"net"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/service"
	"rgt-server/util"
	"sync"
	"sync/atomic"
	"time"
)

type TerminalHandler struct {
	id                   uint64
	session              *TerminalSession
	service              *TerminalEmulationService
	conn                 *net.TCPConn
	receivedPackets      chan *buffer.ByteBuffer
	packetsToSend        chan *buffer.ByteBuffer
	lastDataReadTime     time.Time
	lastAppOperationTime time.Time
	adminClients         map[uint64]*adminClient
	adminClientsMutex    sync.RWMutex
	remoteAddres         string
	endpoint             *TerminalHandler
	finished             atomic.Bool
	waitWorkers          sync.WaitGroup
	protocolVersion      int16
	connectionType       service.ConnectionType
	stats                *SessionStats
}

type requestPack struct {
	handler  *TerminalHandler
	opCode   protocol.OperationCode
	bodySize uint32
	packet   *buffer.ByteBuffer
}

type operationHandle func(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse)

var _ service.ConnectionHandler = &TerminalHandler{}

var final_packet *buffer.ByteBuffer = &buffer.ByteBuffer{}

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
		stats:           NewSessionStats(),
	}
}

func (h *TerminalHandler) Id() uint64 {
	return h.id
}

func (h *TerminalHandler) sessionId() int64 {
	if h.session != nil {
		return h.session.Id()
	}
	return 0
}

func (h *TerminalHandler) Connected() bool {
	return h.conn != nil && !h.finished.Load()
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
	h.stats.AddBytesSent(uint64(sent))
	return sent, err
}

func (h *TerminalHandler) readAll(readBuffer []byte) error {
	if h.Connected() {
		read, err := io.ReadFull(h.conn, readBuffer)
		log.Tracef("[%s;session=%d] TerminalHandler.readAll() read=%d data='%v' ", h.connectionType, h.sessionId(), read, buffer.Wrap(readBuffer))
		h.stats.AddBytesReceived(uint64(read))
		return err
	}
	return io.EOF
}

func (h *TerminalHandler) read(readBuffer []byte) (int, protocol.ErrorResponse) {
	if h.Connected() {
		// conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		// read, err := io.ReadFull(h.conn, readBuffer)
		read, err := h.conn.Read(readBuffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, EOFError
			} else {
				return read, NewError(SOCKET_ERROR, "error reading from ", h.GetRemoteAddr(), ": ", err)
			}
		} else if read == 0 {
			return 0, EOFError
		}
		log.Tracef("[%s;session=%d] TerminalHandler.read() read=%d data='%v' ", h.connectionType, h.sessionId(), read, buffer.Wrap(readBuffer))
		h.stats.AddBytesReceived(uint64(read))
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
	h.protocolVersion = packet.GetInt16()
	return h.runOperation(bodySize, opCode, packet)
}

func (h *TerminalHandler) runOperation(bodySize uint32, opCode protocol.OperationCode, packet *buffer.ByteBuffer) bool {
	var resp *buffer.ByteBuffer
	var err protocol.ErrorResponse

	execOperation, protocolError := terminalProtocol.FindOperation(opCode, h.protocolVersion)
	if protocolError != nil {
		h.sendError(NewError(PROTOCOL_ERROR, "[", h.connectionType, ";session=", h.sessionId(), "] Unknonwn operation: ", opCode))
		return false
	}

	pack := requestPack{
		handler:  h,
		opCode:   opCode,
		bodySize: bodySize,
		packet:   packet,
	}
	resp, err = execOperation.Execute(&pack)
	if resp != nil {
		if h.Send(resp) == nil {
			return true
		} else {
			err = EOFError
		}
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
		return NewError(SOCKET_ERROR, err)
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
		err = NewError(ADMIN_CLIENT_NOT_FOUND_ERROR)
	}
	return err
}

func (h *TerminalHandler) readFirstPacket() (*buffer.ByteBuffer, protocol.ErrorResponse) {
	var headerBuffer [protocol.FIRST_HEADER_SIZE]byte
	err := h.readAll(headerBuffer[:])
	if errors.Is(err, io.EOF) {
		return nil, EOFError
	} else if err != nil {
		return nil, NewError(SOCKET_ERROR, err)
	}
	header := buffer.Wrap(headerBuffer[:])
	magicNumber := header.GetInt32()
	if magicNumber != protocol.MAGIC_NUMBER {
		return nil, NewError(PROTOCOL_ERROR, "Invalid magic number in header: ", magicNumber)
	}
	bodySize := header.GetUInt32()
	if bodySize == 0 {
		return nil, NewError(PROTOCOL_ERROR, "Invalid body len in message: ", bodySize)
	}
	packet := buffer.NewLen(uint32(protocol.PACK_SIZE_FIELD_SIZE) + bodySize)
	packet.PutUInt32(bodySize)
	packet.PutByte(headerBuffer[protocol.FIRST_HEADER_SIZE-1])
	for packet.Remaining() > 0 {
		read, err := h.read(packet.RemainingSlice())
		if err != nil {
			return nil, err
		}
		packet.Skip(read)
	}
	packet.Flip()
	h.stats.AddPacketsReceived(1)
	return packet, nil
}

func (h *TerminalHandler) readPacket() (*buffer.ByteBuffer, protocol.ErrorResponse) {
	var headerBuffer [protocol.HEADER_SIZE]byte
	err := h.readAll(headerBuffer[:])
	// Ignore EOF if not started to read a packet
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, EOFError
		} else {
			return nil, NewError(SOCKET_ERROR, err)
		}
	}
	header := buffer.Wrap(headerBuffer[:])
	bodySize := header.GetUInt32()
	if bodySize == 0 {
		return nil, NewError(PROTOCOL_ERROR, "Invalid body len in message: ", bodySize)
	}
	packet := buffer.NewLen(uint32(protocol.PACK_SIZE_FIELD_SIZE) + bodySize)
	packet.Put(headerBuffer[:])
	for packet.Remaining() > 0 {
		read, err := h.read(packet.RemainingSlice())
		if err != nil {
			return nil, err
		}
		packet.Skip(read)
	}
	packet.Flip()
	h.stats.AddPacketsReceived(1)
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
	h.stats.AddPacketsSent(1)
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
					log.Debugf("[%s;session=%d] TerminalHandler.readPackets(). connection closed.", h.connectionType, h.sessionId())
				} else {
					log.Errorf("[%s;session=%d] TerminalHandler.readPackets(). error reading: %v", h.connectionType, h.sessionId(), err)
				}
			} else {
				log.Tracef("[%s;session=%d] TerminalHandler.readPackets(). Handler finished. socket error: %v", h.connectionType, h.sessionId(), err)
			}
			return
		}
		h.receivedPackets <- packet
		h.lastDataReadTime = time.Now()
	}
}

func (h *TerminalHandler) configConnection() {
	config := h.service.Config()
	if config.TerminalTCPWriteBufferSize().Get() > 0 {
		h.conn.SetWriteBuffer(int(config.TerminalTCPWriteBufferSize().Get()))
	}
	if config.TerminalTCPReadBufferSize().Get() > 0 {
		h.conn.SetReadBuffer(int(config.TerminalTCPReadBufferSize().Get()))
	}
	h.conn.SetKeepAlive(true)
	h.conn.SetKeepAlivePeriod(30 * time.Second)
	h.conn.SetLinger(3)
	h.conn.SetNoDelay(true)
}

func (h *TerminalHandler) Handle() {
	log.Debugf("TerminalHandler.Handle(). started. handle=%d addr=%s", h.id, h.GetRemoteAddr())
	defer h.handlePanic("unknown error in server (TerminalHandler.Handle)")

	h.configConnection()
	h.waitWorkers.Add(1)
	go h.sendPackets()
	if !h.handleNewConnection() {
		h.Close()
		log.Debugf("TerminalHandler.Handle(). error handling new connection. handle=%d", h.id)
		return
	}
	defer func() {
		h.service.sessionManager.CloseSession(h.sessionId())
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
	h.conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	h.conn.CloseRead()
	select {
	case h.packetsToSend <- final_packet:
	case <-time.After(3 * time.Second):
		log.Warnf("[%s;session=%d] TerminalHandler.Close(): timeout sending final_packet to packetsToSend", h.connectionType, h.sessionId())
	}
	select {
	case h.receivedPackets <- final_packet:
	case <-time.After(3 * time.Second):
		log.Warnf("[%s;session=%d] TerminalHandler.Close(): timeout sending final_packet to receivedPackets", h.connectionType, h.sessionId())
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.waitWorkers.Wait()
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Warnf("[%s;session=%d] TerminalHandler.Close(): timeout waiting for workers", h.connectionType, h.sessionId())
	}
	h.drainChannel(h.receivedPackets)
	close(h.receivedPackets)
	h.drainChannel(h.packetsToSend)
	close(h.packetsToSend)
	h.conn.Close()
	h.conn = nil
	log.Debugf("[%s;session=%d] TerminalHandler.Close(): closed", h.connectionType, h.sessionId())
	return nil
}

func (h *TerminalHandler) drainChannel(ch chan *buffer.ByteBuffer) {
	for len(ch) > 0 {
		<-ch
	}
}

func (h *TerminalHandler) sendError(err protocol.ErrorResponse) error {
	log.Debugf("[%s;session=%d] TerminalHandler.SendError(). error=%s", h.connectionType, h.sessionId(), err.Error())
	if err == EOFError {
		return nil
	}
	errBuff := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + buffer.STRING_HEADER_SIZE + len(err.Error())))
	resp := protocol.ResponseFromError(err)
	protocol.PutResponse(resp, errBuff)
	return h.Send(errBuff)
}

func (h *TerminalHandler) SendLogout(message string) {
	log.Debugf("[%s;session=%d] TerminalHandler.SendLogout() message='%s'", h.connectionType, h.sessionId(), message)
	buf := buffer.NewCapacity(uint32(buffer.UINT32_FIELD_SIZE + buffer.INT8_FIELD_SIZE + buffer.BOOLEAN_FIELD_SIZE + buffer.INT16_FIELD_SIZE + buffer.SLICE_HEADER_SIZE + len(message)))
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

func (h *TerminalHandler) SetEndpoint(endpoint *TerminalHandler) {
	h.endpoint = endpoint
}

func (h *TerminalHandler) GetEndpoint() *TerminalHandler {
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

func (h *TerminalHandler) GetStats() *SessionStats {
	if h.stats == nil {
		h.stats = NewSessionStats()
	}
	return h.stats
}
