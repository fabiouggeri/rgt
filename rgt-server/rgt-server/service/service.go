package service

import "sync"

type ServiceStatus uint8
type ServiceType uint8

const (
	STOPPED  ServiceStatus = 0
	STARTING ServiceStatus = 1
	STARTED  ServiceStatus = 2
	PAUSED   ServiceStatus = 3
	STOPPING ServiceStatus = 4

	SERVICE_EMULATION ServiceType = 0x01
	SERVICE_ADMIN     ServiceType = 0x02
	SERVICE_ALL       ServiceType = 0xFF
)

type Service interface {
	GetName() string
	Start(w *sync.WaitGroup) error
	Stop() error
	GetStatus() ServiceStatus
	GetType() ServiceType
	Pause()
	Resume()
	IsAccepting() bool
}
