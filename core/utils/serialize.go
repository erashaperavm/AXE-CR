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

// SerializeSP1output 解析完整的 SP1 输出
func SerializeSP1output(hexData string) (*RsSp1OnlyParsedOut, error) {
	// 去除空白并解码
	hexStr := strings.TrimSpace(hexData)
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex decode failed: %w", err)
	}

	r := bytes.NewReader(raw)
	result := &RsSp1OnlyParsedOut{}

	// ========== 解析第一次 commit ==========
	// 1. 读取类型（1 字节）
	var typ byte
	if err := binary.Read(r, binary.LittleEndian, &typ); err != nil {
		return nil, fmt.Errorf("read first commit type failed: %w", err)
	}
	if typ != TypeInt64 && typ != TypeBytes {
		return nil, fmt.Errorf("unknown type 0x%02x in first commit", typ)
	}

	// 2. 读取长度（4 字节，小端序）
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, fmt.Errorf("read first commit length failed: %w", err)
	}

	// 3. 读取值数据
	valueBytes := make([]byte, length)
	if _, err := io.ReadFull(r, valueBytes); err != nil {
		return nil, fmt.Errorf("read first commit value failed: %w", err)
	}

	// 4. 根据类型解析
	result.FirstCommit.Type = typ
	if typ == TypeInt64 {
		if length != 8 {
			return nil, fmt.Errorf("int64 must have length 8, got %d", length)
		}
		result.FirstCommit.Value = int64(binary.LittleEndian.Uint64(valueBytes))
	} else { // TypeBytes
		result.FirstCommit.Value = valueBytes
	}

	// ========== 解析第二次 commit ==========
	// 检查是否还有剩余数据（应该正好是 40 字节：8字节序号 + 32字节哈希）
	remaining := r.Len()
	if remaining != 40 {
		return nil, fmt.Errorf("expected 40 bytes for second commit, got %d", remaining)
	}

	// 5. 读取隐私内存序号（8 字节，小端序）
	if err := binary.Read(r, binary.LittleEndian, &result.SecondCommit.PrivateMemSeq); err != nil {
		return nil, fmt.Errorf("read private memory sequence failed: %w", err)
	}

	// 6. 读取哈希值（32 字节）
	if _, err := io.ReadFull(r, result.SecondCommit.HashValue[:]); err != nil {
		return nil, fmt.Errorf("read hash value failed: %w", err)
	}

	// 验证读取完毕
	if r.Len() != 0 {
		return nil, fmt.Errorf("unexpected extra data after second commit: %d bytes", r.Len())
	}

	return result, nil
}

// 辅助函数：格式化输出结果
func (p *RsSp1OnlyParsedOut) String() string {
	var firstStr string
	switch p.FirstCommit.Type {
	case TypeInt64:
		firstStr = fmt.Sprintf("int64 = %d", p.FirstCommit.Value.(int64))
	case TypeBytes:
		data := p.FirstCommit.Value.([]byte)
		if len(data) > 0 {
			firstStr = fmt.Sprintf("bytes = %x (string: %s)", data, string(data))
		} else {
			firstStr = "bytes = (empty)"
		}
	}

	secondStr := fmt.Sprintf("private_mem_seq = %d, hash = %x",
		p.SecondCommit.PrivateMemSeq,
		p.SecondCommit.HashValue[:])

	return fmt.Sprintf("FirstCommit: {%s}\nSecondCommit: {%s}", firstStr, secondStr)
}
