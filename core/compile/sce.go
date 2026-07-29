package compile

import (
	axeerr "axe-cr/core/axe-err"
	"axe-cr/core/config"
	"encoding/json"
	"os"
	"path/filepath"
)

func SceFunctions(programPath, workdir string) error {
	// 列出 src/functions 目录下的 function name
	basePath := filepath.Join(programPath, "src", "functions")
	functions, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}
	var functionNames []string
	for _, entry := range functions {
		if entry.IsDir() {
			functionNames = append(functionNames, entry.Name())
		}
		continue
	}

	// 编译它们
	for _, functionName := range functionNames {
		// 读取配置文件
		cfgBytes, err := os.ReadFile(filepath.Join(filepath.Join(basePath, functionName, "axe_fn.json")))
		if err != nil {
			return err
		}

		// 解析配置
		var cfg config.AxeFunctionConfig
		err = json.Unmarshal(cfgBytes, &cfg)
		if err != nil {
			return err
		}

		// 根据语言和环境配置选用合适的编译函数
		switch cfg.SourceLanguage {
		case "rs":
			switch cfg.TargetVerify {
			case "sp1only":
				err := Rs2Sp1Target(filepath.Join(basePath, functionName), workdir)
				if err != nil {
					return err
				}
				continue
			case "joltonly":
				return axeerr.Unimplement
			case "sp1tdx":
				return axeerr.Unimplement
			case "jolttdx":
				return axeerr.Unimplement
			default:
				return axeerr.ExpectedTargetVrf
			}
		case "cpp":
			switch cfg.TargetVerify {
			case "sp1only":
				return axeerr.Unimplement
			case "joltonly":
				return axeerr.Unimplement
			case "sp1tdx":
				return axeerr.Unimplement
			case "jolttdx":
				return axeerr.Unimplement
			default:
				return axeerr.ExpectedTargetVrf
			}
		case "c":
			switch cfg.TargetVerify {
			case "sp1only":
				return axeerr.Unimplement
			case "joltonly":
				return axeerr.Unimplement
			case "sp1tdx":
				return axeerr.Unimplement
			case "jolttdx":
				return axeerr.Unimplement
			default:
				return axeerr.ExpectedTargetVrf
			}
		default:
			return axeerr.ExpectedLanguage
		}
	}

	return nil
}

func SmartContract(programPath, workdir string) error {
	scSetPath := filepath.Join(programPath, "src", "entries")
	scPathEntries, err := os.ReadDir(scSetPath)
	if err != nil {
		return err
	}
	var scPaths []string
	for _, scPathEntry := range scPathEntries {
		if !scPathEntry.IsDir() {
			scPaths = append(scPaths, scPathEntry.Name())
		}
		continue
	}

	// 编译为字节码

	// todo

	return nil
}
