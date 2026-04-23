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
	teListener    atomic.Pointer[net.TCPListener]
	appListener   atomic.Pointer[net.TCPListener]
	currHandlerId atomic.Uint64
	server        *server.Server
	status        atomic.Value // stores service.ServiceStatus
	paused        atomic.Bool
	waitGroup     *sync.WaitGroup
}

var protocols map[protocol.OperationCode]map[int]any = make(map[protocol.OperationCode]map[int]any)

func NewService(serviceName string, srv *server.Server) *TerminalEmulationService {
	s := &TerminalEmulationService{
		name:   serviceName,
		server: srv,
	}
	s.status.Store(service.STOPPED)
	configureLaunchAppSemaphore(srv.Config().MaxConcurrentLaunchingApps().Get())
	srv.Config().MaxConcurrentLaunchingApps().SetHook(configureLaunchAppSemaphore)

	return s
}

func (s *TerminalEmulationService) GetName() string {
	return s.name
}

func (s *TerminalEmulationService) Start(wait *sync.WaitGroup) error {
	if s.GetStatus() == service.STOPPED {
		s.setStatus(service.STARTING)
		log.Infof("Starting service %s...", s.name)
		appPort := s.server.Config().AppEmulationPort().Get()
		tePort := s.server.Config().EmulationPort().Get()
		address := s.server.Config().Address().Get()
		if appPort == tePort {
			teListener, err := s.createListener("TE/APP", address, tePort)
			if err != nil {
				return err
			}
			s.teListener.Store(teListener)
			s.setStatus(service.STARTED)
			wait.Add(1)
			go s.listenConnections("TE/APP", teListener)
		} else {
			appListener, err := s.createListener("APP", address, appPort)
			if err != nil {
				return err
			}
			s.appListener.Store(appListener)
			teListener, err := s.createListener("TE", address, tePort)
			if err != nil {
				return err
			}
			s.teListener.Store(teListener)
			s.setStatus(service.STARTED)
			wait.Add(1)
			go s.listenConnections("APP", appListener)
			wait.Add(1)
			go s.listenConnections("TE", teListener)
		}
		s.waitGroup = wait
		log.Infof("Service %s started.", s.name)
	} else {
		log.Warnf("Service %s already running", s.name)
	}
	return nil
}

func (s *TerminalEmulationService) Stop() error {
	if s.GetStatus() == service.STARTED {
		s.setStatus(service.STOPPING)
		log.Infof("Stopping service %s...", s.name)
		if s.appListener.Load() == nil {
			s.closeListener("TE/APP", s.teListener.Swap(nil))
		} else {
			s.closeListener("TE", s.teListener.Swap(nil))
			s.closeListener("APP", s.appListener.Swap(nil))
		}
		s.setStatus(service.STOPPED)
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

func (s *TerminalEmulationService) createListener(name string, address string, port uint16) (*net.TCPListener, error) {
	var err error
	var proto string
	if address == "" {
		proto = "tcp4"
	} else {
		proto = "tcp"
	}
	addr, err := net.ResolveTCPAddr(proto, address+":"+strconv.Itoa(int(port)))
	if err != nil {
		if proto == "tcp4" {
			proto = "tcp"
			addr, err = net.ResolveTCPAddr("tcp", address+":"+strconv.Itoa(int(port)))
			if err != nil {
				return nil, fmt.Errorf("error starting %s listener of %s service: %v", name, s.name, err)
			}
		} else {
			return nil, fmt.Errorf("error starting %s listener of %s service: %v", name, s.name, err)
		}
	}

	listener, err := net.ListenTCP(proto, addr)
	if err != nil {
		return nil, fmt.Errorf("error starting %s listener of %s service: %v", name, s.name, err)
	}
	log.Debugf("TerminalEmulationService.createListener(). listener %s created for service %s.", name, s.name)
	return listener, nil
}

func (s *TerminalEmulationService) closeListener(name string, listener *net.TCPListener) {
	if listener != nil {
		listener.Close()
		log.Debugf("TerminalEmulationService.closeListener(). Listener %s closed for service %s.", name, s.name)
	}
}

func (s *TerminalEmulationService) listenConnections(name string, listener *net.TCPListener) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(TerminalEmulationService.listenConnections): %v\n%s", err, util.FullStack())
		}
	}()
	defer s.waitGroup.Done()

	for s.GetStatus() == service.STARTED {
		if s.paused.Load() {
			log.Infof("TerminalEmulationService. Listener %s paused for service %s.", name, s.name)
			break
		}
		c, err := listener.AcceptTCP()
		if err != nil {
			if s.paused.Load() {
				log.Infof("TerminalEmulationService. Listener %s paused for service %s.", name, s.name)
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
	if s.GetStatus() == service.STARTED && s.paused.CompareAndSwap(false, true) {
		var listenerLabel string
		log.Infof("TerminalEmulationService.PauseAccepting(). Pausing service %s.", s.name)
		if s.appListener.Load() == nil {
			listenerLabel = "TE/APP"
		} else {
			listenerLabel = "TE"
		}
		s.closeListener(listenerLabel, s.teListener.Swap(nil))
	}
}

func (s *TerminalEmulationService) ResumeAccepting() {
	if s.GetStatus() == service.STARTED && s.paused.CompareAndSwap(true, false) {
		log.Infof("TerminalEmulationService.ResumeAccepting(). Resuming service %s.", s.name)
		var listenerLabel string
		if s.appListener.Load() == nil {
			listenerLabel = "TE/APP"
		} else {
			listenerLabel = "TE"
		}
		teListener, err := s.createListener(listenerLabel, s.server.Config().Address().Get(), s.server.Config().EmulationPort().Get())
		if err != nil {
			log.Errorf("TerminalEmulationService.ResumeAccepting(). Error recreating listener: %v", err)
			s.paused.Store(true)
			return
		}
		s.teListener.Store(teListener)
		s.waitGroup.Add(1)
		go s.listenConnections(listenerLabel, teListener)
	}
}

func (s *TerminalEmulationService) IsAccepting() bool {
	return !s.paused.Load() && s.teListener.Load() != nil
}

func findProtocol[T protocol.Request, S protocol.Response](op protocol.OperationCode, version int16) (*protocol.Protocol[T, S], protocol.ErrorResponse) {
	versions, found := protocols[op]
	if !found {
		return nil, NewError(PROTOCOL_ERROR, "protocol not found or operation ", op)
	}
	for i := version; i >= 0; i-- {
		proto := versions[int(i)]
		if proto != nil {
			return proto.(*protocol.Protocol[T, S]), nil
		}
	}
	return nil, NewError(PROTOCOL_ERROR, "protocol version (", version, ") not found")
}

func registerProtocol(op protocol.OperationCode, version int, proto any) {
	versions, found := protocols[op]
	if !found {
		versions = make(map[int]any)
		protocols[op] = versions
		log.Debugf("terminal.registerProtocol(). Protocol registered for operation %d.", op)
	}
	versions[version] = proto
}
