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

// SerializeSP1output 解析完整的 SP1 输出（三次 commit）
func SerializeSP1output(hexData string) (*RsSp1OnlyParsedOut, error) {
	// 去除空白并解码
	hexStr := strings.TrimSpace(hexData)
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex decode failed: %w", err)
	}

	r := bytes.NewReader(raw)
	result := &RsSp1OnlyParsedOut{}

	// ========== 解析第一次 commit（公共输出列表） ==========
	firstValues, err := parseList(r)
	if err != nil {
		return nil, fmt.Errorf("parse first commit failed: %w", err)
	}
	result.FirstCommit.Values = firstValues

	// ========== 解析第二次 commit（加密数据列表） ==========
	encryptedItems, err := parseEncryptedList(r)
	if err != nil {
		return nil, fmt.Errorf("parse second commit failed: %w", err)
	}
	result.SecondCommit.EncryptedItems = encryptedItems

	// ========== 解析第三次 commit（哈希承诺列表） ==========
	hashValues, err := parseHashList(r)
	if err != nil {
		return nil, fmt.Errorf("parse third commit failed: %w", err)
	}
	result.ThirdCommit.HashValues = hashValues

	// 验证读取完毕
	if r.Len() != 0 {
		return nil, fmt.Errorf("unexpected extra data after third commit: %d bytes", r.Len())
	}

	return result, nil
}

// parseList 解析公共值列表（第一次 commit）
// 格式：[Type: 0x02] [Length: 4字节] [元素1][元素2]...
// 每个元素格式：[Type: 1字节] [Length: 4字节] [Value: N字节]
func parseList(r *bytes.Reader) ([]PublicValue, error) {
	// 1. 读取类型（1 字节）
	var typ byte
	if err := binary.Read(r, binary.LittleEndian, &typ); err != nil {
		return nil, err
	}

	if typ != TypeList {
		return nil, fmt.Errorf("expected list type 0x%02x, got 0x%02x", TypeList, typ)
	}

	// 2. 读取总长度（4 字节）
	var totalLen uint32
	if err := binary.Read(r, binary.LittleEndian, &totalLen); err != nil {
		return nil, err
	}

	// 3. 读取所有元素数据
	listData := make([]byte, totalLen)
	if _, err := io.ReadFull(r, listData); err != nil {
		return nil, err
	}

	// 4. 解析列表中的每个元素
	listReader := bytes.NewReader(listData)
	var values []PublicValue

	for listReader.Len() > 0 {
		value, err := parseSingleValue(listReader)
		if err != nil {
			return nil, err
		}
		values = append(values, *value)
	}

	return values, nil
}

// parseEncryptedList 解析加密数据列表（第二次 commit）
// 格式：[Type: 0x02] [Length: 4字节] [元素1][元素2]...
// 每个元素格式：[Length: 4字节] [Value: N字节]
func parseEncryptedList(r *bytes.Reader) ([][]byte, error) {
	// 1. 读取类型（1 字节）
	var typ byte
	if err := binary.Read(r, binary.LittleEndian, &typ); err != nil {
		return nil, err
	}

	if typ != TypeList {
		return nil, fmt.Errorf("expected list type 0x%02x, got 0x%02x", TypeList, typ)
	}

	// 2. 读取总长度（4 字节）
	var totalLen uint32
	if err := binary.Read(r, binary.LittleEndian, &totalLen); err != nil {
		return nil, err
	}

	// 3. 读取所有元素数据
	listData := make([]byte, totalLen)
	if _, err := io.ReadFull(r, listData); err != nil {
		return nil, err
	}

	// 4. 解析列表中的每个元素
	listReader := bytes.NewReader(listData)
	var items [][]byte

	for listReader.Len() > 0 {
		// 读取长度（4 字节）
		var length uint32
		if err := binary.Read(listReader, binary.LittleEndian, &length); err != nil {
			return nil, err
		}

		// 读取数据
		data := make([]byte, length)
		if _, err := io.ReadFull(listReader, data); err != nil {
			return nil, err
		}

		items = append(items, data)
	}

	return items, nil
}

// parseHashList 解析哈希承诺列表（第三次 commit）
// 格式：[Type: 0x02] [Length: 4字节] [元素1][元素2]...
// 每个元素格式：[Value: 32字节]
func parseHashList(r *bytes.Reader) ([][32]byte, error) {
	// 1. 读取类型（1 字节）
	var typ byte
	if err := binary.Read(r, binary.LittleEndian, &typ); err != nil {
		return nil, err
	}

	if typ != TypeList {
		return nil, fmt.Errorf("expected list type 0x%02x, got 0x%02x", TypeList, typ)
	}

	// 2. 读取总长度（4 字节）
	var totalLen uint32
	if err := binary.Read(r, binary.LittleEndian, &totalLen); err != nil {
		return nil, err
	}

	// 3. 读取所有元素数据
	listData := make([]byte, totalLen)
	if _, err := io.ReadFull(r, listData); err != nil {
		return nil, err
	}

	// 4. 解析列表中的每个元素（每个元素固定32字节）
	listReader := bytes.NewReader(listData)
	var hashValues [][32]byte

	for listReader.Len() > 0 {
		if listReader.Len() < 32 {
			return nil, fmt.Errorf("incomplete hash data: %d bytes remaining, need 32", listReader.Len())
		}

		var hash [32]byte
		if _, err := io.ReadFull(listReader, hash[:]); err != nil {
			return nil, err
		}

		hashValues = append(hashValues, hash)
	}

	return hashValues, nil
}

// parseSingleValue 解析单个 TLV 格式的值
func parseSingleValue(r *bytes.Reader) (*PublicValue, error) {
	// 1. 读取类型（1 字节）
	var typ byte
	if err := binary.Read(r, binary.LittleEndian, &typ); err != nil {
		return nil, err
	}

	if typ != TypeInt64 && typ != TypeBytes {
		return nil, fmt.Errorf("unknown type 0x%02x", typ)
	}

	// 2. 读取长度（4 字节，小端序）
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	// 3. 读取值数据
	valueBytes := make([]byte, length)
	if _, err := io.ReadFull(r, valueBytes); err != nil {
		return nil, err
	}

	// 4. 根据类型解析
	result := &PublicValue{Type: typ}
	if typ == TypeInt64 {
		if length != 8 {
			return nil, fmt.Errorf("int64 must have length 8, got %d", length)
		}
		result.Value = int64(binary.LittleEndian.Uint64(valueBytes))
	} else { // TypeBytes
		result.Value = valueBytes
	}

	return result, nil
}

// 辅助函数：打印解析结果
func (p *RsSp1OnlyParsedOut) String() string {
	var sb strings.Builder

	sb.WriteString("=== First Commit (Public Values) ===\n")
	for i, v := range p.FirstCommit.Values {
		switch v.Type {
		case TypeInt64:
			sb.WriteString(fmt.Sprintf("  [%d] int64 = %d\n", i, v.Value.(int64)))
		case TypeBytes:
			data := v.Value.([]byte)
			if len(data) > 0 {
				sb.WriteString(fmt.Sprintf("  [%d] bytes = %x (string: %s)\n", i, data, string(data)))
			} else {
				sb.WriteString(fmt.Sprintf("  [%d] bytes = (empty)\n", i))
			}
		}
	}

	sb.WriteString("\n=== Second Commit (Encrypted Items) ===\n")
	for i, item := range p.SecondCommit.EncryptedItems {
		sb.WriteString(fmt.Sprintf("  [%d] encrypted data (%d bytes): %x\n", i, len(item), item))
	}

	sb.WriteString("\n=== Third Commit (Hash Commitments) ===\n")
	for i, hash := range p.ThirdCommit.HashValues {
		sb.WriteString(fmt.Sprintf("  [%d] hash: %x\n", i, hash[:]))
	}

	return sb.String()
}
