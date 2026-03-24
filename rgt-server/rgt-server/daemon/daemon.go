package daemon

type Daemon interface {
	GetName() string
	Start(args []string)
	Stop()
}

type Command uint32
type State uint32

const (
	Stopped         = State(1)
	StartPending    = State(2)
	StopPending     = State(3)
	Running         = State(4)
	ContinuePending = State(5)
	PausePending    = State(6)
	Paused          = State(7)
)

const (
	StopCmd        = Command(1)
	PauseCmd       = Command(2)
	ContinueCmd    = Command(3)
	InterrogateCmd = Command(4)
	ShutdownCmd    = Command(6)
)
