package bridge

import (
	"axe-cr/core/execute"
	"axe-cr/core/utils"
	"fmt"
	"path/filepath"
)

func CallRS(exeDir, wkdir, funName, targetVerify string, ints []int64, bytes [][]byte, mode string) (*execute.RsSp1Result, error) {
	funTargetPath := filepath.Join(exeDir, "function", funName)

	// build input file
	inputFilePath, err := utils.SerializeSP1input(ints, bytes, wkdir)
	if err != nil {
		return nil, err
	}

	switch targetVerify {
	case "sp1only":
		res, err := execute.RsSp1only(funTargetPath, funName, inputFilePath, mode, wkdir)
		if err != nil {
			return nil, err
		}
		return &res, nil
	case "joltonly":
		// todo
		return nil, nil
	case "sp1tdx":
		// todo
		return nil, nil
	case "jolttdx":
		// todo
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown verify target: %s", targetVerify)
	}
}
