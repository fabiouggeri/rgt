package terminal

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

type TerminalEmulationService struct {
	name          string
	listener      *net.TCPListener
	currHandlerId atomic.Uint64
	server        *server.Server
	status        atomic.Value // stores service.ServiceStatus
	paused        atomic.Bool
	waitGroup     *sync.WaitGroup
	listenerMu    sync.Mutex
}

var protocols map[protocol.OperationCode]map[int]interface{} = make(map[protocol.OperationCode]map[int]interface{})

func NewService(serviceName string, srv *server.Server) *TerminalEmulationService {
	s := &TerminalEmulationService{name: serviceName,
		listener: nil,
		server:   srv}
	s.status.Store(service.STOPPED)
	return s
}

func (s *TerminalEmulationService) GetName() string {
	return s.name
}

func (s *TerminalEmulationService) Start(wait *sync.WaitGroup) error {
	if s.listener == nil {
		log.Infof("Starting service %s...", s.name)
		s.setStatus(service.STARTING)
		err := s.createListener()
		if err != nil {
			return err
		}
		s.waitGroup = wait
		wait.Add(1)
		go s.terminalEmulationService(wait)
		log.Infof("Service %s started.", s.name)
	} else {
		log.Warn("Listener already running")
	}
	return nil
}

func (s *TerminalEmulationService) Stop() error {
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

func (s *TerminalEmulationService) GetStatus() service.ServiceStatus {
	return s.status.Load().(service.ServiceStatus)
}

func (s *TerminalEmulationService) setStatus(status service.ServiceStatus) {
	s.status.Store(status)
}

func (s *TerminalEmulationService) createListener() error {
	var err error
	var proto string
	if s.server.Config().Address().Get() == "" {
		proto = "tcp4"
	} else {
		proto = "tcp"
	}
	addr, err := net.ResolveTCPAddr(proto, s.server.Config().Address().Get()+":"+strconv.Itoa(int(s.server.Config().EmulationPort().Get())))
	if err != nil {
		if proto == "tcp4" {
			proto = "tcp"
			addr, err = net.ResolveTCPAddr("tcp", s.server.Config().Address().Get()+":"+strconv.Itoa(int(s.server.Config().EmulationPort().Get())))
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
	log.Debugf("TerminalEmulationService.createListener(). created for service %s.", s.name)
	return nil
}

func (s *TerminalEmulationService) closeListener() {
	if s.listener != nil {
		l := s.listener
		s.listener = nil
		s.setStatus(service.STOPPING)
		l.Close()
		s.setStatus(service.STOPPED)
		log.Debugf("TerminalEmulationService.closeListener(). Listener closed for service %s.", s.name)
	}
}

func (s *TerminalEmulationService) terminalEmulationService(wait *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(TerminalEmulationService.terminalEmulationService): %v\n%s", err, util.FullStack())
		}
	}()
	defer wait.Done()
	defer func() {
		if !s.paused.Load() {
			s.closeListener()
		}
	}()

	s.setStatus(service.STARTED)
	for s.GetStatus() == service.STARTED {
		s.listenerMu.Lock()
		l := s.listener
		s.listenerMu.Unlock()
		if l == nil {
			break
		}
		c, err := l.AcceptTCP()
		if err != nil {
			if s.paused.Load() {
				log.Infof("TerminalEmulationService. Listener paused for service %s.", s.name)
			} else if s.GetStatus() == service.STARTED {
				log.Error("error listening client connections: ", err)
			}
			break
		}
		handlerId := s.currHandlerId.Add(1)
		h := newHandler(handlerId, c, s)
		go h.Handle()
	}
}

func (s *TerminalEmulationService) GetType() service.ServiceType {
	return service.SERVICE_EMULATION
}

func (s *TerminalEmulationService) PauseAccepting() {
	if s.paused.CompareAndSwap(false, true) {
		log.Infof("TerminalEmulationService.PauseAccepting(). Pausing service %s.", s.name)
		s.listenerMu.Lock()
		if s.listener != nil {
			s.listener.Close()
			s.listener = nil
		}
		s.listenerMu.Unlock()
	}
}

func (s *TerminalEmulationService) ResumeAccepting() {
	if s.paused.CompareAndSwap(true, false) {
		log.Infof("TerminalEmulationService.ResumeAccepting(). Resuming service %s.", s.name)
		s.listenerMu.Lock()
		err := s.createListener()
		s.listenerMu.Unlock()
		if err != nil {
			log.Errorf("TerminalEmulationService.ResumeAccepting(). Error recreating listener: %v", err)
			s.paused.Store(true)
			return
		}
		if s.waitGroup != nil {
			s.waitGroup.Add(1)
			go s.terminalEmulationService(s.waitGroup)
		}
	}
}

func (s *TerminalEmulationService) IsAccepting() bool {
	return !s.paused.Load() && s.listener != nil
}

func findProtocol[T protocol.Request, S protocol.Response](op protocol.OperationCode, version int16) (*protocol.Protocol[T, S], protocol.ErrorResponse) {
	versions, found := protocols[op]
	if !found {
		return nil, NewError(PROTOCOL, "protocol not found or operation ", op)
	}
	for i := version; i >= 0; i-- {
		proto := versions[int(i)]
		if proto != nil {
			return proto.(*protocol.Protocol[T, S]), nil
		}
	}
	return nil, NewError(PROTOCOL, "protocol version (", version, ") not found")
}

func registerProtocol(op protocol.OperationCode, version int, proto interface{}) {
	versions, found := protocols[op]
	if !found {
		versions = make(map[int]interface{})
		protocols[op] = versions
		log.Debugf("terminal.registerProtocol(). Protocol registered for operation %d.", op)
	}
	versions[version] = proto
}
