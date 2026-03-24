package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"regexp"
	"rgt-server/buffer"
	"rgt-server/protocol"
	"rgt-server/security"
	"rgt-server/terminal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type execution struct {
	args               []string
	envVars            []string
	exePathname        string
	username           string
	password           string
	workingDir         string
	server             string
	connection         net.Conn
	serverPort         uint16
	keepAlive          uint16
	captureOutput      bool
	killLostConnection bool
	running            bool
}

type serverRequest struct {
	body *buffer.ByteBuffer
	code protocol.OperationCode
}

const (
	INVALID_COMMAND_LINE = 1
	SOCKET_ERROR         = 2
)

var (
	Version = "dev"
)

func showHelp() {
	fmt.Printf("RGT Launcher %v", Version)
	fmt.Println()
	fmt.Printf("Usage:\n")
	fmt.Printf("   %s [options] \\path\\to\\app.exe arg1 arg2 ... argN\n", os.Args[0])
	fmt.Println()
	fmt.Printf("Options:\n")
	fmt.Printf("   --server=ip|hostname             (default=value of TSNODEADDR env var)\n")
	fmt.Printf("      * Address of RGT Server where the app must execute. If not informed the\n")
	fmt.Printf("        address will be get from TSNODEADDR env var.\n")
	fmt.Println()
	fmt.Printf("   --serverPort=hexPort             (default=value of TSSOCKET env var or 1DE6)\n")
	fmt.Printf("      * Port number in hexadecimal format.\n")
	fmt.Println()
	fmt.Printf("   --workDir=\\path\\to\\workdir    (default=home of user running RGT Server)\n")
	fmt.Printf("      * Directory to set as current working directory. Values of remote\n")
	fmt.Printf("        environment variables can be set in the format %c%s%c.\n", '%', "ENVVAR", '%')
	fmt.Println()
	fmt.Printf("   --username=user\n")
	fmt.Printf("      * User with authorization to run application. Not necessary if the server\n")
	fmt.Printf("        is configured for anyone runs standalone apps.\n")
	fmt.Println()
	fmt.Printf("   --password=pswd\n")
	fmt.Printf("      * Password of the user. If not informed and user war informed, it will\n")
	fmt.Printf("        be requested. Not necessary if the server is configured for anyone \n")
	fmt.Printf("        runs standalone apps.\n")
	//	fmt.Println()
	//	fmt.Printf("   --attach=session-id\n")
	//	fmt.Printf("      * Session to be attached. This option is useful when the connection is\n")
	//	fmt.Printf("        lost and app keeps running\n")
	fmt.Println()
	fmt.Printf("   --output=true|false              (default=true)\n")
	fmt.Printf("      * Capture application output?\n")
	fmt.Println()
	fmt.Printf("   --killLost=true|false  (default=false)\n")
	fmt.Printf("      * Kill application if connection is lost?\n")
	fmt.Println()
	fmt.Printf("   --keepAlive=999                  (default=15)\n")
	fmt.Printf("      * Time interval to send keepalive. Value 0 disable keep alive.\n")
	fmt.Println()
	fmt.Printf("   --env=var[:value]\n")
	fmt.Printf("      * Environments variables to set remotely. The variables can be set as\n")
	fmt.Printf("        follows:\n")
	fmt.Printf("         --env=Var       --> set Var with local environment variable\n")
	fmt.Printf("         --env=Var*      --> set variables with all local environment variables\n")
	fmt.Printf("                             that match mask\n")
	fmt.Printf("         --env=Var:Value --> set variable with defined value\n")
}

func setEnvVars(exec *execution, envVar string) error {
	i := strings.IndexRune(envVar, '*')
	if i >= 0 {
		i = strings.IndexRune(envVar, ':')
		if i >= 0 {
			return fmt.Errorf("masked env var can't receive value")
		}
		reg, err := regexp.Compile("(?i)^" + strings.ReplaceAll(envVar, "*", ".*") + "$")
		if err != nil {
			return err
		}
		for _, env := range os.Environ() {
			varName, _, found := strings.Cut(env, "=")
			if found && reg.MatchString(varName) {
				exec.envVars = append(exec.envVars, env)
			}
		}
	} else {
		varName, varValue, found := strings.Cut(envVar, ":")
		if found {
			exec.envVars = append(exec.envVars, varName+"="+varValue)
		} else {
			varValue = os.Getenv(envVar)
			if varValue != "" {
				exec.envVars = append(exec.envVars, envVar+"="+varValue)
			}
		}
	}
	return nil
}

func getOption(exec *execution, arg string) error {
	option, value, hasValue := strings.Cut(arg, "=")
	option = strings.ToLower(option)
	switch option {

	case "output":
		exec.captureOutput = !hasValue || !strings.EqualFold(value, "false")

	case "killlost":
		exec.killLostConnection = !hasValue || strings.EqualFold(value, "true")

	case "server":
		if hasValue {
			exec.server = value
		} else {
			return fmt.Errorf("server option must receive a host address")
		}

	case "workdir":
		if hasValue {
			exec.workingDir = value
		} else {
			return fmt.Errorf("server option must receive a host address")
		}

	case "serverport":
		if hasValue {
			val, valErr := strconv.ParseUint(value, 16, 16)
			if valErr == nil {
				exec.serverPort = uint16(val)
			} else {
				return fmt.Errorf("invalid value to serverPort option: %v", value)
			}
		} else {
			return fmt.Errorf("server port option must receive a value between 1 and 65535")
		}

	case "keepalive":
		if hasValue {
			val, valErr := strconv.ParseUint(value, 10, 16)
			if valErr == nil {
				exec.keepAlive = uint16(val)
			} else {
				return fmt.Errorf("invalid value to keepAlive option: %v", value)
			}
		} else {
			return fmt.Errorf("keep alive option must receive a value between 0 and 65535")
		}

	case "env":
		if hasValue {
			return setEnvVars(exec, value)
		} else {
			return fmt.Errorf("env option must receive environment variable name (with or without value) or mask")
		}

	case "username":
		if hasValue {
			exec.username = value
		} else {
			return fmt.Errorf("username must receive the login name of the user")
		}

	case "password":
		if hasValue {
			exec.password = "{DESede}" + hex.EncodeToString(security.GetCipher("DESede").Encrypt([]byte(value)))
		} else {
			return fmt.Errorf("password must receive the key pass of the user")
		}

	default:
		return fmt.Errorf("unknown option: %s", option)
	}
	return nil
}

func defaultServerPort() (uint16, error) {
	hexPort := os.Getenv("TSSOCKET")
	if hexPort == "" {
		hexPort = "1DE6"
	}
	port, err := strconv.ParseUint(hexPort, 16, 16)
	if err != nil {
		return 7654, fmt.Errorf("invalid port number in TSSOCKET env var: %v", err)
	}
	return uint16(port), nil
}

func parseCommandLine(args []string) (*execution, error) {
	var err error
	if len(args) == 0 {
		showHelp()
		return nil, nil
	}
	exec := &execution{
		captureOutput:      true,
		killLostConnection: false,
		keepAlive:          15,
		server:             os.Getenv("TSNODEADDR"),
		envVars:            make([]string, 0, 8),
		args:               make([]string, 0, 8)}

	exec.serverPort, err = defaultServerPort()
	if err != nil {
		return nil, err
	}

	alreadySetPathName := false
	for _, arg := range args {
		var err error
		if strings.HasPrefix(arg, "--") {
			if !alreadySetPathName {
				err = getOption(exec, arg[2:])
			} else {
				return nil, fmt.Errorf("options must precede the executable name")
			}
		} else if !alreadySetPathName {
			exec.exePathname = arg
			alreadySetPathName = true
		} else {
			exec.args = append(exec.args, arg)
		}
		if err != nil {
			return nil, err
		}
	}
	err = exec.validateOptions()
	if err != nil {
		return nil, err
	}
	err = exec.readPassword()
	if err != nil {
		return nil, err
	}
	return exec, err
}

func (exec *execution) validateOptions() error {
	if exec.exePathname == "" {
		return fmt.Errorf("enter executable path and name after options")
	}
	if exec.server == "" {
		return fmt.Errorf("set TSNODEADDR environement variable or enter --server option")
	}
	return nil
}

func (exec *execution) readPassword() error {
	if strings.TrimSpace(exec.username) != "" && strings.TrimSpace(exec.password) == "" {
		print("Password: ")
		pswd, err := term.ReadPassword(int(syscall.Stdin))
		println()
		if err != nil {
			return fmt.Errorf("error reading password: %v", err)
		} else if len(pswd) == 0 {
			return fmt.Errorf("password not informed and it is mandatory")
		}
		exec.password = "{DESede}" + hex.EncodeToString(security.GetCipher("DESede").Encrypt(pswd))
	}
	return nil
}

func (exec *execution) connectToServer() bool {
	conn, err := net.Dial("tcp", net.JoinHostPort(exec.server, fmt.Sprintf("%d", exec.serverPort)))
	if err != nil {
		fmt.Printf("Error connecting to server %s: %v", exec.server, err)
		return false
	}
	exec.connection = conn
	return true
}

func write(conn net.Conn, buffer []byte) error {
	if conn != nil {
		_, err := conn.Write(buffer)
		if err != nil {
			return fmt.Errorf("error writing to %v: %v", conn.RemoteAddr(), err)
		}
	}
	return nil
}

func readAll(conn net.Conn, buffer []byte) error {
	dataLen := len(buffer)
	read, err := io.ReadFull(conn, buffer)
	if err != nil {
		return err
	}
	if read < int(dataLen) {
		return fmt.Errorf("insuficient data. read: %v waiting: %v", read, dataLen)
	}
	return nil
}

func read(conn net.Conn, buffer []byte) (int, error) {
	read, err := conn.Read(buffer)
	if err != nil {
		return 0, fmt.Errorf("error reading from %v: %v", conn.RemoteAddr(), err)
	}
	return read, nil
}

func (exec *execution) createRequest() *terminal.AppExecRequest {
	appExecReq := &terminal.AppExecRequest{
		BaseRequest:           protocol.BaseRequest{OperationCode: terminal.TRM_STANDALONE_APP_EXEC},
		ProtocolVersion:       terminal.TRM_STANDALONE_APP_PROTOCOL_VERSION,
		Username:              exec.username,
		Password:              exec.password,
		TerminalAddress:       exec.connection.LocalAddr().String(),
		CaptureOutput:         exec.captureOutput,
		KillAppLostConnection: exec.killLostConnection,
		WorkingDir:            exec.workingDir,
		ExePathName:           exec.exePathname,
		EnvVars:               make([]string, 0, len(exec.envVars)),
		Arguments:             make([]string, 0, len(exec.args))}
	appExecReq.EnvVars = append(appExecReq.EnvVars, exec.envVars...)
	appExecReq.Arguments = append(appExecReq.Arguments, exec.args...)
	osUser, err := user.Current()
	if err != nil {
		fmt.Println("Error trying to get OS username: ", err)
		return nil
	}
	appExecReq.OsUser = osUser.Username
	return appExecReq
}

func readResponse(conn net.Conn, proto *protocol.Protocol[*terminal.AppExecRequest, *terminal.AppExecResponse]) error {
	header := make([]byte, protocol.RESPONSE_HEADER_SIZE)
	read, err := io.ReadFull(conn, header)
	if err != nil {
		return err
	}
	if read < int(protocol.RESPONSE_HEADER_SIZE) {
		return fmt.Errorf("insuficient data. read: %d waiting: %d", read, protocol.RESPONSE_HEADER_SIZE)
	}
	headerBuf := buffer.Wrap(header)
	bodySize := headerBuf.GetInt32() - 2
	if bodySize < 0 {
		return fmt.Errorf("invalid body size in response from server: %d", bodySize+2)
	}
	bodyBuf := buffer.NewCapacity(uint32(bodySize) + 2)
	bodyBuf.PutUInt16(headerBuf.GetUInt16())
	if bodySize > 0 {
		body := make([]byte, bodySize)
		err := readAll(conn, body)
		if err != nil {
			return fmt.Errorf("error reading packet body: %v", err)
		}
		bodyBuf.Put(body)
	}
	bodyBuf.Flip()
	resp := proto.GetResponse(bodyBuf)
	if resp.GetCode() != terminal.SUCCESS {
		if resp.GetMessage() == "" {
			return fmt.Errorf("response error: %s", terminal.GetResponseCodeDescription(resp.GetCode()))
		} else {
			return fmt.Errorf("response error: %s", resp.GetMessage())
		}
	}
	fmt.Printf("------------------------\n")
	fmt.Printf("     Execution Info\n")
	fmt.Printf("------------------------\n")
	fmt.Printf("Session: %d\n", resp.SessionId)
	fmt.Printf("PID ...: %d\n", resp.Pid)
	fmt.Printf("------------------------\n")
	return nil
}

func (exec *execution) readRequest() (*serverRequest, error) {
	var toRead int32
	var bytesRead int
	headerBuffer := make([]byte, protocol.HEADER_SIZE)
	err := readAll(exec.connection, headerBuffer)
	if err != nil {
		return nil, err
	}
	header := buffer.Wrap(headerBuffer)
	bodySize := header.GetInt32() - 1
	opCode := protocol.OperationCode(header.GetUInt8())
	body := buffer.NewCapacity(uint32(bodySize))

	ioBuffer := make([]byte, protocol.DEFAULT_IO_BUFFER_SIZE)
	if int(bodySize) < len(ioBuffer) {
		toRead = bodySize
	} else {
		toRead = int32(len(ioBuffer))
	}
	for err == nil && bodySize > 0 && toRead > 0 {
		bytesRead, err = read(exec.connection, ioBuffer[:toRead])
		body.Put(ioBuffer[:bytesRead])
		bodySize -= int32(bytesRead)
		if int(bodySize) < len(ioBuffer) {
			toRead = bodySize
		} else {
			toRead = int32(len(ioBuffer))
		}
	}
	if err != nil {
		return nil, err
	}
	body.Flip()
	return &serverRequest{code: opCode, body: body}, nil
}

func (exec *execution) outputOperation(proto *protocol.Protocol[*terminal.AppOutputRequest, *protocol.BaseResponse], req *serverRequest) bool {
	var err error
	outReq := proto.GetRequest(req.body)
	if outReq.Error {
		_, err = os.Stderr.Write(outReq.Output)
	} else {
		_, err = os.Stdout.Write(outReq.Output)
	}
	if err != nil {
		fmt.Printf("Error printing remote process output: %v", err)
		return false
	}
	return true
}

func (exec *execution) statusOperation(proto *protocol.Protocol[*terminal.AppStatusRequest, *protocol.BaseResponse], req *serverRequest) int {
	statusReq := proto.GetRequest(req.body)
	if statusReq.Message != "" {
		fmt.Printf("\n%s", statusReq.Message)
	}
	return int(statusReq.ExitCode)
}

func (exec *execution) waitAppFinish() int {
	exitCode := 0
	outputProto := terminal.CreateSendOutputProtocol()
	statusProto := terminal.CreateSendStatusProtocol()
	for exec.running {
		req, err := exec.readRequest()
		if err != nil {
			if err != io.EOF {
				fmt.Println("error waiting server data: ", err)
				exitCode = SOCKET_ERROR
			}
			return exitCode
		}
		switch req.code {
		case terminal.TRM_STANDALONE_APP_SEND_OUTPUT:
			exec.running = exec.outputOperation(outputProto, req)

		case terminal.TRM_STANDALONE_APP_SEND_STATUS:
			exitCode = exec.statusOperation(statusProto, req)
			exec.running = false

		case terminal.TRM_APP_KEEP_ALIVE:
			exec.running = true

		default:
			fmt.Printf("Unknown command received: %d", req.code)
			exec.running = false
		}
	}
	return exitCode
}

func (exec *execution) launchApp() bool {
	appExecReq := exec.createRequest()
	if appExecReq == nil {
		return false
	}
	execAppProto := terminal.CreateAppExecProtocol()
	buf := buffer.New()
	execAppProto.PutRequestFirstOp(appExecReq, buf)
	err := write(exec.connection, buf.GetBytes())
	if err != nil {
		fmt.Print(err)
		return false
	}
	err = readResponse(exec.connection, execAppProto)
	if err != nil {
		fmt.Print(err)
		return false
	}
	exec.running = true
	return true
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Print("stacktrace from panic: \n", string(debug.Stack()))
		}
	}()
	// args := []string{
	// 	"--server=127.0.0.1",
	// 	"--env=MEDNODEADDR:172.23.3.104",
	// 	"--env=MEDCS:db1core2072clone",
	// 	"--env=MEDSOCKET:19C8",
	// 	"w:\\sistemas\\legado\\siret-exe-3.11.0.2\\target\\classes\\win64-msc16.x-hb3.2.x\\siret-exe-3.11.0.2-win64-msc16.x-hb3.2.x.exe",
	// 	"AG0101", "AG0101", "R", "1", "1", "123455"}
	// exec, err := parseCommandLine(args)
	exec, err := parseCommandLine(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(INVALID_COMMAND_LINE)
	} else if exec == nil {
		return
	}

	if exec.connectToServer() && exec.launchApp() {
		exitCode := exec.waitAppFinish()
		exec.connection.Close()
		os.Exit(exitCode)
	}
}
