package execute

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type RsSp1CpuStatus struct {
	Status string `json:"status"`
}

type RsSp1CpuResult struct {
	PvPath   string
	RpPath   string
	PfPath   string
	VkeyPath string
}

func RsSp1Cpu(programPath, programName, inputPath, mode string) (RsSp1CpuResult, error) {
	// 1. 创建命令
	cmd := exec.Command("exerssp1onlycpu", "--path", programPath, "--name", programName, "--input", inputPath, "--mode", mode)

	// 2. 阻塞执行并直接获取 stdout 的所有字节
	raw, err := cmd.Output()
	if err != nil {
		// 如果命令执行失败、退出码非 0，或者根本找不到可执行文件，都会直接在这里报错
		fmt.Fprintf(os.Stderr, "execute failed: %v\n", err)

		// 如果你想拿到更详细的错误信息（比如子进程打印在 stderr 的错误）
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			fmt.Fprintf(os.Stderr, "stderr output: %s\n", string(exitErr.Stderr))
		}

		os.Exit(1)
	}

	// 3. 反序列化结果
	var result RsSp1CpuStatus
	if err = json.Unmarshal(raw, &result); err != nil {
		return RsSp1CpuResult{}, err
	}

	if result.Status != "success" {
		return RsSp1CpuResult{}, err
	}

	basePath := filepath.Join(programPath, "execution_out", programName)
	pvPath := filepath.Join(basePath, "pv.txt")
	rpPath := filepath.Join(basePath, "report.json")
	pfPath := filepath.Join(basePath, "proof.bin")
	vkeyPath := filepath.Join(basePath, "vkey.txt")

	return RsSp1CpuResult{
		PvPath:   pvPath,
		RpPath:   rpPath,
		PfPath:   pfPath,
		VkeyPath: vkeyPath,
	}, nil

	/*
		// 读取文件内容
		basePath := filepath.Join(programPath, "execution_out", programName)
		pvBytes, err := os.ReadFile(filepath.Join(basePath, "pv.txt"))
		if err != nil {
			return RsSp1CpuResult{}, err
		}
		rpBytes, err := os.ReadFile(filepath.Join(basePath, "report.json"))
		if err != nil {
			return RsSp1CpuResult{}, err
		}
		pfBytes, err := os.ReadFile(filepath.Join(basePath, "proof.bin"))
		if err != nil {
			return RsSp1CpuResult{}, err
		}
		vkeyBytes, err := os.ReadFile(filepath.Join(basePath, "vkey.txt"))
		if err != nil {
			return RsSp1CpuResult{}, err
		}

		// 解析内容
		origPv, err := hex.DecodeString(strings.TrimSpace(string(pvBytes)))
		if err != nil {
			return RsSp1CpuResult{}, err
		}
		origRp := strings.TrimSpace(string(rpBytes))
		if len(vkeyBytes) != 32 {
			return RsSp1CpuResult{}, errors.New("vkey's len should be 32")
		}
		var origVkey [32]byte
		copy(origVkey[:], vkeyBytes)

		return RsSp1CpuResult{
			PublicValues: origPv,
			ReportJson:   origRp,
			Proof:        pfBytes,
			Vkey:         origVkey,
		}, nil
	*/
}
