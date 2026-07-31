package compile

import (
	"fmt"
	"os"
	"os/exec"
)

func Rs2Sp1Target(functionPath, workDir string) error {
	cmd := exec.Command("sh", "-c", "cargo proof build")
	cmd.Dir = functionPath
	// todo 环境变量需要测试，可能不起效，考虑 cp
	cmd.Env = append(os.Environ(), fmt.Sprintf("CARGO_TARGET_DIR=%s", workDir))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w, output: %s", err, out)
	}
	fmt.Println(string(out))
	return nil
}
