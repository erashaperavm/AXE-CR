package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AxeFunctionConfig struct {
	SourceLanguage string `json:"source_language"`
	TargetVerify   string `json:"target_vrf"`
	FunctionName   string `json:"function_name"`
	ProveType      string `json:"prove_type"`
}

// AxeProjectConfig represents the project-level configuration (axe.json at the program root).
type AxeProjectConfig struct {
	ProgramName string `json:"program_name"`
}

type FunctionMeta struct {
	ReqInput     []string `json:"req_input"`
	ReqOutput    []string `json:"req_output"`
	TargetVerify string   `json:"target_verify"`
}

// LoadFunctionConfig reads the axe.json configuration for a specific function.
func LoadFunctionConfig(functionPath string) (*AxeFunctionConfig, error) {
	cfgPath := filepath.Join(functionPath, "axe_fn.json")
	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var cfg AxeFunctionConfig
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadProjectConfig reads the project-level axe.json.
func LoadProjectConfig(programPath string) (*AxeProjectConfig, error) {
	cfgPath := filepath.Join(programPath, "axe_pkg.json")
	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var cfg AxeProjectConfig
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
