package compile

import (
	"fmt"
	"os/exec"
)

func Rs2Sp1CPUTarget(rustProgramPath string) error {
	cmd := exec.Command("cd", rustProgramPath)

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	cmd = exec.Command("cargo", "proof", "build")

	out, err = cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}
