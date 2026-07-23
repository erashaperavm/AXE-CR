package main

import (
	"axe-cr/core/compile"
	"axe-cr/core/config"
	"axe-cr/core/execute"
	"axe-cr/core/verify"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Output represents a unified JSON output structure for all commands.
type Output struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// printOutput serializes and prints the output in JSON format, then exits.
func printOutput(status, message string, data interface{}) {
	out := Output{
		Status:  status,
		Message: message,
		Data:    data,
	}
	jsonBytes, _ := json.Marshal(out)
	fmt.Println(string(jsonBytes))
	if status == "error" {
		os.Exit(1)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "axe",
		Short: "A toolchain for AXE program.",
		Long:  "A powerful toolchain for AXE Smart Contract Programming Language.",
	}

	// for functions
	rootCmd.AddCommand(compileCmd())
	rootCmd.AddCommand(executeCmd())
	rootCmd.AddCommand(verifyCmd())

	// for whole sce program
	rootCmd.AddCommand(sceCompileCmd()) // todo
	rootCmd.AddCommand(sceExecuteCmd()) // todo
	rootCmd.AddCommand(sceVerifyCmd())  // todo

	if err := rootCmd.Execute(); err != nil {
		out, _ := json.Marshal(map[string]string{"status": "error", "message": err.Error()})
		fmt.Println(string(out))
		os.Exit(1)
	}
}

func compileCmd() *cobra.Command {
	var (
		functionPath string
		workDir      string
	)

	cmd := &cobra.Command{
		Use:     "compile",
		Short:   "compile your function to executable files.",
		Long:    "compile your function (include rs, c, cpp, sc) to executable files.",
		Example: "axe compile -fp ./my_func -w ./workdir",
		Run: func(cmd *cobra.Command, args []string) {
			fcfg, err := config.LoadFunctionConfig(functionPath)
			if err != nil {
				printOutput("error", fmt.Sprintf("failed to load function config: %v", err), nil)
				return
			}
			switch fcfg.SourceLanguage {
			case "rs":
				switch fcfg.TargetVerify {
				case "sp1only":
					err := compile.Rs2Sp1Target(functionPath, workDir)
					if err != nil {
						printOutput("error", fmt.Sprintf("compile rs to sp1 target failed: %v", err), nil)
						return
					}
					printOutput("success", "compile successfully!", nil)
					return
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			case "cpp":
				switch fcfg.TargetVerify {
				case "sp1only":
					// todo
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			case "c":
				switch fcfg.TargetVerify {
				case "sp1only":
					// todo
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			default:
				printOutput("error", fmt.Sprintf("unsupported source language: %s", fcfg.SourceLanguage), nil)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&functionPath, "fnpath", "fp", "./my_func", "path to the function")
	cmd.Flags().StringVarP(&workDir, "wkdir", "w", "./compile_workdir", "path to the workdir to save results")

	return cmd
}

func executeCmd() *cobra.Command {
	var (
		cwkdir       string
		functionPath string
		inputPath    string
		wkdir        string
	)

	cmd := &cobra.Command{
		Use:     "execute",
		Short:   "execute function in local.",
		Long:    "execute function in local, this should be used by worker nodes, or test builder.",
		Example: "./axe execute -cw ./cpmpile_workdir -fp ./my_func -i ./input.bin -w ./workdir",
		Run: func(cmd *cobra.Command, args []string) {
			// Load function configuration
			fcfg, err := config.LoadFunctionConfig(functionPath)
			if err != nil {
				printOutput("error", fmt.Sprintf("failed to load function config: %v", err), nil)
				return
			}

			switch fcfg.SourceLanguage {
			case "rs":
				switch fcfg.TargetVerify {
				case "sp1only":
					res, err := execute.RsSp1only(cwkdir, fcfg.FunctionName, inputPath, fcfg.ProveType, wkdir)
					if err != nil {
						printOutput("error", fmt.Sprintf("execute rs sp1 failed: %v", err), nil)
						return
					}
					data := map[string]interface{}{
						"public_values_path": res.PvPath,
						"report_path":        res.RpPath,
						"proof_path":         res.PfPath,
						"vkey_path":          res.VkeyPath,
					}
					printOutput("success", "execution completed", data)
					return
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			case "cpp":
				switch fcfg.TargetVerify {
				case "sp1only":
					// todo
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			case "c":
				switch fcfg.TargetVerify {
				case "sp1only":
					// todo
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			default:
				printOutput("error", fmt.Sprintf("unsupported source language: %s", fcfg.SourceLanguage), nil)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&cwkdir, "cwkdir", "cw", "./compile_workdir", "path to the workdir to compile results")
	cmd.Flags().StringVarP(&functionPath, "fnpath", "fp", "./my_func", "path to the function")
	cmd.Flags().StringVarP(&inputPath, "in", "i", ",/input.bin", "path to input file")
	cmd.Flags().StringVarP(&wkdir, "wkdir", "w", "./execute_workdir", "path to the workdir to save results")

	return cmd
}

func verifyCmd() *cobra.Command {
	var (
		cwkdir       string
		ewkdir       string
		functionPath string
	)

	cmd := &cobra.Command{
		Use:     "verify",
		Short:   "verify function execute result.",
		Long:    "verify function execute result, this should be used by verify nodes, or caller.",
		Example: "./axe verify -cw ./compile_workdir -ew ./execute_workdir -fp ./my_func",
		Run: func(cmd *cobra.Command, args []string) {
			// Load function configuration
			fcfg, err := config.LoadFunctionConfig(functionPath)
			if err != nil {
				printOutput("error", fmt.Sprintf("failed to load function config: %v", err), nil)
				return
			}

			switch fcfg.SourceLanguage {
			case "rs":
				switch fcfg.TargetVerify {
				case "sp1only":
					err := verify.RsSp1only(cwkdir, ewkdir, fcfg.FunctionName)
					if err != nil {
						printOutput("error", fmt.Sprintf("verify rs sp1 only failed: %v", err), nil)
						return
					}
					printOutput("success", "verify successful", nil)
					return
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			case "cpp":
				switch fcfg.TargetVerify {
				case "sp1only":
					// todo
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			case "c":
				switch fcfg.TargetVerify {
				case "sp1only":
					// todo
				case "joltonly":
					// todo
				case "sp1tdx":
					// todo
				case "jolttdx":
					// todo
				default:
					printOutput("error", fmt.Sprintf("unsupported target verify: %s", fcfg.TargetVerify), nil)
					return
				}
			default:
				printOutput("error", fmt.Sprintf("unsupported source language: %s", fcfg.SourceLanguage), nil)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&cwkdir, "cwkdir", "cw", "./compile_workdir", "path to workdir to compile result")
	cmd.Flags().StringVarP(&ewkdir, "ewkdir", "ew", "./execute_workdir", "path to workdir to execute result")
	cmd.Flags().StringVarP(&functionPath, "fnpath", "fp", "./my_func", "path to the function")

	return cmd
}

func sceCompileCmd() *cobra.Command {
	var (
		programPath string
		wkdir       string
	)

	cmd := cobra.Command{
		Use:     "compile-sce",
		Short:   "compile the whole sce program. ",
		Long:    "compile the whole sce program automatically through only one instruction. ",
		Example: "./axe compile-sce -p ./my_program -w ./sce_compile_workdir",
		Run: func(cmd *cobra.Command, args []string) {
			err := compile.SceFunctions(programPath, wkdir)
			if err != nil {
				printOutput("error", fmt.Sprintf("failed to compile inside functions: %v", err), nil)
				return
			}
			printOutput("success", "compile successfully!", nil)
			return
		},
	}

	cmd.Flags().StringVarP(&programPath, "path", "p", "./my_program", "path to your sce program")
	cmd.Flags().StringVarP(&wkdir, "wkdir", "w", "./sce_compile_workdir", "path to workdir to save compile result")

	return &cmd
}

func sceExecuteCmd() *cobra.Command {
	var ()

	cmd := cobra.Command{
		Use:     "execute-sce",
		Short:   "execute the whole sce program. ",
		Long:    "execute the whole sce program automatically through only one instruction. ",
		Example: "axe execute-sce ",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	cmd.Flags()

	return &cmd
}

func sceVerifyCmd() *cobra.Command {
	var (
		programPath  string
		programName  string
		targetVerify string
	)

	cmd := cobra.Command{
		Use:     "verify-sce",
		Short:   "verify the whole sce program. ",
		Long:    "verify the whole sce program automatically through only one struct. ",
		Example: "axe verify-sce ",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	cmd.Flags().StringVarP(&programPath, "path", "p", "", "path to your sce program. ")
	cmd.Flags().StringVarP(&programName, "name", "n", "", "your program name. ")
	cmd.Flags().StringVarP(&targetVerify, "targetvrf", "tv", "", "target platform for verification. ")

	return &cmd
}
