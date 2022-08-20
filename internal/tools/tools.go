package tools

import (
	"fmt"
	"os/exec"
)

func IsCommandExists(command string) {
	path, err := exec.LookPath(command)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(path)
}
