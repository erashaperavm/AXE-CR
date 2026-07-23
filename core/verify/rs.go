package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type RsSp1Status struct {
	Status string `json:"status"`
}

func RsSp1only(cwkdir, ewkdir, functionName string) error {
	// 1. 创建命令
	cmd := exec.Command("vrfrssp1only", "--cwkdir", cwkdir, "--ewkdir", ewkdir, "--fnname", functionName)

	// 2. 阻塞执行并直接获取 stdout 的所有字节
	raw, err := cmd.Output()
	if err != nil {
		// 如果命令执行失败、退出码非 0，或者根本找不到可执行文件，都会直接在这里报错
		fmt.Fprintf(os.Stderr, "execute failed: %v\n", err)

		// 拿到更详细的错误信息（比如子进程打印在 stderr 的错误）
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			fmt.Fprintf(os.Stderr, "stderr output: %s\n", string(exitErr.Stderr))
		}

		os.Exit(1)
	}

	// 3. 反序列化结果
	var result RsSp1Status
	if err = json.Unmarshal(raw, &result); err != nil {
		return err
	}

	if result.Status != "success" {
		return err
	}

	return nil
}
