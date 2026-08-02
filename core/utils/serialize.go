package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

const (
	TypeInt64 = 0x00
	TypeBytes = 0x01
)

func SerializeSP1input(ints []int64, bytes [][]byte, swkdir string) (string, error) {
	// 构建参数列表
	args := []string{"--swkdir", swkdir}

	// 每个 int 单独作为参数
	for _, n := range ints {
		args = append(args, "--int", strconv.FormatInt(n, 10))
	}

	// 每个 bytes 单独作为参数
	for _, b := range bytes {
		args = append(args, "--byte", hex.EncodeToString(b))
	}

	cmd := exec.Command("serializesp1in", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w, output: %s", err, out)
	}

	// 解析 JSON 响应
	var resp Response
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w, output: %s", err, out)
	}

	if resp.Status == "error" {
		return "", fmt.Errorf("serializesp1in error: %s", resp.Message)
	}

	return resp.Path, nil
}

func SerializeSP1output(hexData string) ([]RsSp1OnlyParsedOut, error) {
	// 去除空白并解码
	hexStr := strings.TrimSpace(hexData)
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex decode failed: %w", err)
	}

	var result []RsSp1OnlyParsedOut
	r := bytes.NewReader(raw)

	for r.Len() > 0 {
		// 1. 读取类型（1 字节）
		var typ byte
		if err := binary.Read(r, binary.LittleEndian, &typ); err != nil {
			return nil, fmt.Errorf("read type failed: %w", err)
		}
		if typ != TypeInt64 && typ != TypeBytes {
			return nil, fmt.Errorf("unknown type 0x%02x", typ)
		}

		// 2. 读取长度（4 字节，小端序）
		var length uint32
		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return nil, fmt.Errorf("read length failed: %w", err)
		}

		// 3. 读取值数据
		valueBytes := make([]byte, length)
		if _, err := io.ReadFull(r, valueBytes); err != nil {
			return nil, fmt.Errorf("read value failed: %w", err)
		}

		// 4. 根据类型解析
		var parsed RsSp1OnlyParsedOut
		parsed.Type = typ
		if typ == TypeInt64 {
			if length != 8 {
				return nil, fmt.Errorf("int64 must have length 8, got %d", length)
			}
			parsed.Val = int64(binary.LittleEndian.Uint64(valueBytes))
		} else { // TypeBytes
			parsed.Val = valueBytes
		}
		result = append(result, parsed)
	}
	return result, nil
}
