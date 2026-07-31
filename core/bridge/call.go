package bridge

import (
	"axe-cr/core/execute"
	"fmt"
	"path/filepath"
)

func CallRS(exeDir, funName, targetVerify string, input [][]byte) ([][]byte, error) {
	funTargetPath := filepath.Join(exeDir, "function", funName)

	switch targetVerify {
	case "sp1only":
		res, err := execute.RsSp1only(funTargetPath, funName)
		if err != nil {
			return nil, err
		}
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
