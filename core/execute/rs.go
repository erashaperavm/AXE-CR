package execute

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type RsSp1Status struct {
	Status string `json:"status"`
}

type RsSp1Result struct {
	PvPath   string
	RpPath   string
	PfPath   string
	VkeyPath string
}

func RsSp1only(cwkdir, functionName, inputPath, mode, wkdir string) (RsSp1Result, error) {
	// 1. 创建命令
	cmd := exec.Command("exerssp1only", "--cwkdir", cwkdir, "--fnname", functionName, "--input", inputPath, "--mode", mode, "--wkdir", wkdir)

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
		return RsSp1Result{}, err
	}

	if result.Status != "success" {
		return RsSp1Result{}, err
	}

	basePath := filepath.Join(wkdir, "execution_out")
	pvPath := filepath.Join(basePath, "pv.txt")
	rpPath := filepath.Join(basePath, "report.json")
	pfPath := filepath.Join(basePath, "proof.bin")
	vkeyPath := filepath.Join(basePath, "vkey.txt")

	return RsSp1Result{
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
			return RsSp1Result{}, err
		}
		rpBytes, err := os.ReadFile(filepath.Join(basePath, "report.json"))
		if err != nil {
			return RsSp1Result{}, err
		}
		pfBytes, err := os.ReadFile(filepath.Join(basePath, "proof.bin"))
		if err != nil {
			return RsSp1Result{}, err
		}
		vkeyBytes, err := os.ReadFile(filepath.Join(basePath, "vkey.txt"))
		if err != nil {
			return RsSp1Result{}, err
		}

		// 解析内容
		origPv, err := hex.DecodeString(strings.TrimSpace(string(pvBytes)))
		if err != nil {
			return RsSp1Result{}, err
		}
		origRp := strings.TrimSpace(string(rpBytes))
		if len(vkeyBytes) != 32 {
			return RsSp1Result{}, axe-err.New("vkey's len should be 32")
		}
		var origVkey [32]byte
		copy(origVkey[:], vkeyBytes)

		return RsSp1Result{
			PublicValues: origPv,
			ReportJson:   origRp,
			Proof:        pfBytes,
			Vkey:         origVkey,
		}, nil
	*/
}
