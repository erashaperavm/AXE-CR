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
)

// FirstCommit 第一次 commit 解析结果
type FirstCommit struct {
	Type  byte        // 0x00 或 0x01
	Value interface{} // int64 或 []byte
}

// SecondCommit 第二次 commit 解析结果
type SecondCommit struct {
	PrivateMemSeq uint64   // 隐私内存序号
	HashValue     [32]byte // 32字节的SHA-256哈希值
}

// RsSp1OnlyParsedOut 完整解析结果
type RsSp1OnlyParsedOut struct {
	FirstCommit  FirstCommit  // 第一次 commit 的公开数据
	SecondCommit SecondCommit // 第二次 commit 的隐私内存数据
}
