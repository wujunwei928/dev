package tools

import (
	"errors"
	"os/exec"
	"strings"
)

// TrimExplode split keys_str to slice
func TrimExplode(valueStr string, sep string) []string {
	valueList := make([]string, 0, strings.Count(valueStr, sep))
	for _, v := range strings.Split(valueStr, sep) {
		val := strings.TrimSpace(v)
		if len(val) != 0 {
			valueList = append(valueList, val)
		}
	}
	return valueList
}

func ExecCmd(cmd string) error {
	if len(cmd) <= 0 {
		return errors.New("cmd string is empty")
	}
	args := TrimExplode(cmd, " ")
	command := exec.Command(args[0], args[1:]...)
	return command.Start()
}
