package main

import (
	"axe-cr/core/compile"
	"axe-cr/core/execute"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	programPath    string
	sourceLanguage string
	targetVerify   string
	targetExecute  string
)

var (
	inputPath   string
	proveMode   string
	programName string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "axe",
		Short: "A toolchain for AXE project.",
		Long:  "A powerful toolchain for AXE Smart Contract Programming Language.",
	}

	rootCmd.PersistentFlags() // todo

	rootCmd.AddCommand(compileCmd())
	rootCmd.AddCommand(executeCmd())
	rootCmd.AddCommand(rwChainCmd())
	rootCmd.AddCommand(devCmd()) // execute without real ntwk and prover
	rootCmd.AddCommand(verifyCmd())

	if err := rootCmd.Execute(); err != nil {
		out, _ := json.Marshal(map[string]string{"status": "error", "message": err.Error()})
		fmt.Println(string(out))
		os.Exit(1)
	}
}

func compileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "compile",
		Short:   "compile your project to sce-pkg.",
		Long:    "compile your axe project(include rs, c, cpp, sc) to sce-pkg.",
		Example: "axe compile -p ./my_project -sl rs -tv sp1onlycpu -te cpu",
		Run: func(cmd *cobra.Command, args []string) {
			switch sourceLanguage {
			case "rs":
				switch targetVerify {
				case "sp1onlycpu":
					switch targetExecute {
					case "cpu":
						err := compile.Rs2Sp1CPUTarget(programPath)
						if err != nil {
							out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("compile rs to sp1 target failed: %v", err)})
							fmt.Println(string(out))
							return
						}
						out, _ := json.Marshal(map[string]string{"status": "success", "message": "compile successfully!"})
						fmt.Println(string(out))
						return
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "joltonlycpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "sp1tdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolttdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				default:
					fmt.Println("unsupported target verify", targetVerify)
					return
				}
			case "cpp":
				switch targetVerify {
				case "sp1":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolt":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "sp1tdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolttdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				default:
					fmt.Println("unsupported target verify", targetVerify)
					return
				}
			case "c":
				switch targetVerify {
				case "sp1":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolt":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "sp1tdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolttdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				default:
					fmt.Println("unsupported target verify", targetVerify)
					return
				}
			default:
				fmt.Println("unsupported source language", sourceLanguage)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&programPath, "path", "p", "", "path to the program")
	cmd.Flags().StringVarP(&sourceLanguage, "language", "sl", "", "source language")
	cmd.Flags().StringVarP(&targetVerify, "targetvrf", "tv", "", "target platform for verification")
	cmd.Flags().StringVarP(&targetExecute, "targetexe", "te", "", "target platform for execution")

	return cmd
}

func executeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute",
		Short:   "execute sce-pkg in local.",
		Long:    "execute sce-pkg in local, this should be used by worker nodes, or test builder.",
		Example: "./axe execute -p ./my_program.scepkg -in ./input.bin -pm groth16 -tv sp1onlycpu -te cpu",
		Run: func(cmd *cobra.Command, args []string) {
			switch sourceLanguage {
			case "rs":
				switch targetVerify {
				case "sp1onlycpu":
					switch targetExecute {
					case "cpu":
						res, err := execute.RsSp1Cpu(programPath, programName, inputPath, proveMode)
						if err != nil {
							out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("execute rs sp1 cpu failed: %v", err)})
							fmt.Println(string(out))
							return
						}
						output := map[string]interface{}{
							"status":             "success",
							"public_values_path": res.PvPath,
							"report_path":        res.RpPath,
							"proof_path":         res.PfPath,
							"vkey_path":          res.VkeyPath,
						}
						jsonBytes, _ := json.Marshal(output)
						fmt.Println(string(jsonBytes))
						return
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "joltonlycpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "sp1tdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolttdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				default:
					fmt.Println("unsupported target verify", targetVerify)
					return
				}
			case "cpp":
				switch targetVerify {
				case "sp1":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolt":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "sp1tdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolttdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				default:
					fmt.Println("unsupported target verify", targetVerify)
					return
				}
			case "c":
				switch targetVerify {
				case "sp1":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolt":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "sp1tdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				case "jolttdxcpu":
					switch targetExecute {
					case "cpu":
						// todo
					case "gpu":
						// todo
					default:
						out, _ := json.Marshal(map[string]string{"status": "error", "message": fmt.Sprintf("unsupported target execute: %s", targetExecute)})
						fmt.Println(string(out))
						return
					}
				default:
					fmt.Println("unsupported target verify", targetVerify)
					return
				}
			default:
				fmt.Println("unsupported source language", sourceLanguage)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&programPath, "path", "p", "", "path to the program")
	cmd.Flags().StringVarP(&programName, "name", "n", "", "your project name")
	cmd.Flags().StringVarP(&targetVerify, "targetvrf", "tv", "", "target platform for verification")
	cmd.Flags().StringVarP(&targetExecute, "targetexe", "te", "", "target platform for execution")
	cmd.Flags().StringVarP(&inputPath, "input", "in", "", "input file path")
	cmd.Flags().StringVarP(&proveMode, "mode", "pm", "core", "prove mode")

	return cmd
}

func rwChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rw",
		Short:   "rw the blockchain.",
		Long:    "read or write blockchain.",
		Example: "todo",                                     // todo
		Run:     func(cmd *cobra.Command, args []string) {}, // todo
	}

	cmd.Flags() // todo

	return cmd
}

func devCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dev",
		Short:   "use a developer environment to execute sce-pkg.",
		Long:    "start a developer environment to execute sce-pkg without real ntwk and prover to reduce develop cycles and time.",
		Example: "todo",                                     // todo
		Run:     func(cmd *cobra.Command, args []string) {}, // todo
	}

	cmd.Flags() // todo

	return cmd
}

func verifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify",
		Short:   "verify sce-pkg execute result.",
		Long:    "verify sce-pkg execute result, this should be used by verify nodes, or caller.",
		Example: "todo",                                     // todo
		Run:     func(cmd *cobra.Command, args []string) {}, // todo
	}

	cmd.Flags() // todo

	return cmd
}
