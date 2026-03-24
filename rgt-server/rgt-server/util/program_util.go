package util

import (
	"fmt"
	"os"
	"path/filepath"
	"rgt-server/log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func ExePath() (string, error) {
	var err error
	var path string
	var fileInfo os.FileInfo

	prog := os.Args[0]
	path, err = filepath.Abs(prog)

	if err != nil {
		return "", err
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		if filepath.Ext(path) != "" {
			return "", err
		}
		path += ".exe"
		fileInfo, err = os.Stat(path)
		if err != nil {
			return "", err
		}
	}
	if fileInfo.Mode().IsDir() {
		return "", fmt.Errorf("%s is directory", path)
	}
	return path, nil
}

func ArgsToMap(list []string) map[string]string {
	args := make(map[string]string)

	for i := 0; i < len(list); i++ {
		option, value, found := strings.Cut(list[i], "=")
		if found {
			args[option] = value
		} else {
			args[list[i]] = "true"
		}
	}
	return args
}

func KillProcess(proc *process.Process, reason string) error {
	startTime, _ := proc.CreateTime()
	dateTime := time.Unix(0, startTime*int64(time.Millisecond))
	cmd, _ := proc.Cmdline()
	log.Infof("Process %d killed. Reason: %s Start: %v Cmd: %s", proc.Pid, reason, dateTime, cmd)
	return proc.Kill()
}

func KillProcessRecursive(proc *process.Process, reason string) error {
	startTime, _ := proc.CreateTime()
	dateTime := time.Unix(0, startTime*int64(time.Millisecond))
	cmd, _ := proc.Cmdline()
	children, err := proc.Children()
	if err != nil {
		for _, child := range children {
			KillProcessRecursive(child, "Child of "+strconv.Itoa(int(proc.Pid)))
		}
	}
	log.Infof("Process %d killed. Reason: '%s' Start: %v Cmd: '%s'", proc.Pid, reason, dateTime, cmd)
	return proc.Kill()
}

func ProcessEnvVar(proc *process.Process, varName string) (string, error) {
	if proc == nil {
		return "", fmt.Errorf("process not found")
	}
	vars, err := proc.Environ()
	if err != nil {
		return "", err
	}
	prefix := varName + "="
	varIndex := slices.IndexFunc(vars, func(value string) bool { return strings.HasPrefix(value, prefix) })
	if varIndex < 0 {
		return "", fmt.Errorf("variable '%s' not found in process %d", varName, proc.Pid)
	}
	return vars[varIndex][len(prefix):], nil
}
