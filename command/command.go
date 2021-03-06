package command

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"
)

const (
	VbsDir = "C:\\Glnt\\GlntSetup\\glnt\\cmd"
	OCR    = "ocr"
	GPMS   = "gpms"
	RELAY  = "relay"
)

func ServiceCheck(name string) error {
	isRunning := false
	switch name {
	case OCR:
		output, _, _ := Pipeline(TaskList(), FindStr("GlntProxySvr"))
		isRunning = runningCheck(output)

	case GPMS, RELAY:
		pid, err := ioutil.ReadFile(fmt.Sprintf("/tmp/%s.pid", name))
		if err != nil {
			return nil
		}

		cmd := exec.Command("powershell", "tasklist", fmt.Sprintf(` /FI "IMAGENAME eq javaw.exe"`))
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, _, _ := Pipeline(cmd, FindStr(string(pid)))

		isRunning = runningCheck(output)
	}

	if isRunning {
		return errors.New(fmt.Sprintf("[!] %s is running...!", name))
	}

	return nil
}

func runningCheck(output []byte) bool {
	serviceCount := strings.Count(string(output), "\n")
	if serviceCount > 0 {
		return true
	}

	return false
}

func ServiceStart(name string) error {
	cmd := exec.Command("wscript.exe", fmt.Sprintf(`./cmd/%s.vbs`, name))
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func ServiceStop(name string) error {
	var stop *exec.Cmd
	switch name {
	case OCR:
		stop = exec.Command("taskkill", "/f", "/im", "GlntProxySvr.exe")
	default:
		pid, err := ioutil.ReadFile(fmt.Sprintf("/tmp/%s.pid", name))
		if err != nil {
			return nil
		}
		stop = exec.Command("taskkill", "/f", "/pid", string(pid))

	}
	stop.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := stop.Run(); err != nil {
		return err
	}

	return nil
}

func taskKill(name string) *exec.Cmd {
	kill := exec.Command("taskkill", "/f", "im", name)
	kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return kill
}

func TaskList() *exec.Cmd {
	taskList := exec.Command("tasklist")
	taskList.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return taskList
}

func FindStr(val string) *exec.Cmd {
	findStr := exec.Command("findstr", val)
	findStr.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return findStr
}

// Pipeline cmd ????????? pipe ??????
func Pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	if len(cmds) < 1 {
		return nil, nil, nil
	}

	var output bytes.Buffer
	var stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// ??? ????????? stdin??? ?????? ????????? stdout??? ??????
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, err
		}
		// ??? ????????? stderrer??? ????????? ??????
		cmd.Stderr = &stderr
	}

	// ????????? ????????? ?????? ?????? ??? ?????? ??????
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// ??? ?????? ??????
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	// ??? ????????? ????????? ????????? ??????
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	return output.Bytes(), stderr.Bytes(), nil
}
