package utils

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

// 类型常量
const (
	TypeInt64 = 0x00
	TypeBytes = 0x01
	TypeList  = 0x02
)

// PublicValue 公共输出值
type PublicValue struct {
	Type  byte
	Value interface{} // int64 或 []byte
}

// FirstCommit 第一次 commit 解析结果
type FirstCommit struct {
	Values []PublicValue // 列表形式存储所有公共值
}

// SecondCommit 第二次 commit 解析结果（加密数据列表）
type SecondCommit struct {
	EncryptedItems [][]byte // 加密数据列表
}

// ThirdCommit 第三次 commit 解析结果（哈希承诺列表）
type ThirdCommit struct {
	HashValues [][32]byte // SHA-256 哈希值列表
}

// RsSp1OnlyParsedOut 完整解析结果
type RsSp1OnlyParsedOut struct {
	FirstCommit  FirstCommit
	SecondCommit SecondCommit
	ThirdCommit  ThirdCommit
}
