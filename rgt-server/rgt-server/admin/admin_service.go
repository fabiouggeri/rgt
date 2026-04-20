package admin

import (
	"fmt"
	"net"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"rgt-server/service"
	"rgt-server/util"
	"strconv"
	"sync"
	"sync/atomic"
)

type AdminOperation int8

type AdminService struct {
	name           string
	server         *server.Server
	listener       *net.TCPListener
	currHandlerId  atomic.Uint64
	handlers       map[uint64]*AdminHandler
	handlerLock    sync.Mutex
	handlerEditing *AdminHandler
	status         atomic.Value // stores service.ServiceStatus
}

var protocols map[protocol.OperationCode]map[int]any = make(map[protocol.OperationCode]map[int]any)

func NewService(serviceName string, srv *server.Server) *AdminService {
	s := &AdminService{name: serviceName,
		server:         srv,
		listener:       nil,
		handlers:       make(map[uint64]*AdminHandler),
		handlerEditing: nil}
	s.status.Store(service.STOPPED)
	return s
}

func (s *AdminService) GetName() string {
	return s.name
}

func (s *AdminService) Start(wait *sync.WaitGroup) error {
	if s.listener == nil {
		log.Infof("Starting service %s...", s.name)
		s.setStatus(service.STARTING)
		err := s.createListener()
		if err != nil {
			return err
		}
		wait.Add(1)
		go s.adminService(wait)
		log.Infof("Service %s started.", s.name)
	} else {
		log.Warn("Listener already running")
	}
	return nil
}

func (s *AdminService) createListener() error {
	var err error
	var proto string
	if s.server.Config().Address().Get() == "" {
		proto = "tcp4"
	} else {
		proto = "tcp"
	}
	addr, err := net.ResolveTCPAddr(proto, s.server.Config().Address().Get()+":"+strconv.Itoa(int(s.server.Config().AdminPort().Get())))
	if err != nil {
		if proto == "tcp4" {
			proto = "tcp"
			addr, err = net.ResolveTCPAddr("tcp", s.server.Config().Address().Get()+":"+strconv.Itoa(int(s.server.Config().AdminPort().Get())))
			if err != nil {
				return fmt.Errorf("error starting %s listener: %v", s.name, err)
			}
		} else {
			return fmt.Errorf("error starting %s listener: %v", s.name, err)
		}
	}
	s.listener, err = net.ListenTCP(proto, addr)
	if err != nil {
		return fmt.Errorf("error starting %s listener: %v", s.name, err)
	}
	return nil
}

func (s *AdminService) closeListener() {
	if s.listener != nil {
		s.setStatus(service.STOPPING)
		s.listener.Close()
		s.listener = nil
		s.setStatus(service.STOPPED)
	}
}

func (s *AdminService) registerHandler(id uint64, handler *AdminHandler) {
	s.handlerLock.Lock()
	defer s.handlerLock.Unlock()
	if s.handlerEditing == nil {
		s.handlerEditing = handler
		handler.readOnly = false
	} else {
		handler.readOnly = true
	}
	s.handlers[id] = handler
}

func (s *AdminService) unregisterHandler(id uint64) {
	s.handlerLock.Lock()
	defer s.handlerLock.Unlock()
	if s.handlerEditing != nil && s.handlerEditing.id == id {
		s.handlerEditing = nil
	}
	delete(s.handlers, id)
}

func (s *AdminService) Stop() error {
	if s.listener != nil {
		log.Infof("Stopping service %s...", s.name)
		s.setStatus(service.STOPPING)
		s.closeListener()
		log.Infof("Service %s stopped.", s.name)
	} else {
		log.Warnf("Service %s is not running", s.name)
	}
	return nil
}

func (s *AdminService) GetStatus() service.ServiceStatus {
	return s.status.Load().(service.ServiceStatus)
}

func (s *AdminService) setStatus(status service.ServiceStatus) {
	s.status.Store(status)
}

func (s *AdminService) adminService(wait *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("[ADMIN] unknown error in server(AdminService.adminService): %v\n%s", err, util.FullStack())
		}
	}()
	defer s.closeListener()
	defer wait.Done()

	s.setStatus(service.STARTED)
	for s.GetStatus() == service.STARTED {
		c, err := s.listener.AcceptTCP()
		if err != nil {
			if s.GetStatus() == service.STARTED {
				log.Error("[ADMIN] error listening client connections: ", err)
			}
			break
		}
		handlerId := s.currHandlerId.Add(1)
		s.configConnection(c)
		h := newHandler(handlerId, c, s)
		s.registerHandler(handlerId, h)
		go h.Handle()
	}
}

func (s *AdminService) configConnection(c *net.TCPConn) {
	if s.server.Config().AdminTCPWriteBufferSize().Get() > 0 {
		c.SetWriteBuffer(int(s.server.Config().AdminTCPWriteBufferSize().Get()))
	}
	if s.server.Config().AdminTCPReadBufferSize().Get() > 0 {
		c.SetReadBuffer(int(s.server.Config().AdminTCPReadBufferSize().Get()))
	}
}

func (s *AdminService) GetType() service.ServiceType {
	return service.SERVICE_ADMIN
}

func (s *AdminService) PauseAccepting() {
	// Admin service always accepts connections
}

func (s *AdminService) ResumeAccepting() {
	// Admin service always accepts connections
}

func (s *AdminService) IsAccepting() bool {
	return s.listener != nil
}

func (s *AdminService) GetHandlerEditing() *AdminHandler {
	return s.handlerEditing
}

func findProtocol[T protocol.Request, S protocol.Response](op protocol.OperationCode, version int16) (*protocol.Protocol[T, S], protocol.ErrorResponse) {
	versions, found := protocols[op]
	if !found {
		return nil, NewError(PROTOCOL_ERROR, "[ADMIN] protocol not found or operation ", op)
	}
	for i := version; i >= 0; i-- {
		proto := versions[int(i)]
		if proto != nil {
			return proto.(*protocol.Protocol[T, S]), nil
		}
	}
	return nil, NewError(PROTOCOL_ERROR, "[ADMIN] protocol version (", version, ") not found")
}

func registerProtocol(op protocol.OperationCode, version int, proto any) {
	versions, found := protocols[op]
	if !found {
		versions = make(map[int]any)
		protocols[op] = versions
	}
	versions[version] = proto
}
