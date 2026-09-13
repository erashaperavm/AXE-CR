## 开发者文档：公开值提交标准格式

### 1. 目的
规范 zkVM 程序中公开输出（通过 `sp1_zkvm::io::commit` 提交）的二进制格式，使得链下服务（如 Go、Python）能够自动解析这些值，无需为每个程序单独编写解析逻辑。

### 2. 数据块结构
提交的数据块按顺序紧密排列：

#### 第一次 commit：公共输出

| 字段       | 大小   | 说明                                  |
|----------|------|-------------------------------------|
| **类型标识** | 1 字节 | `0x00` 表示 `int64`，`0x01` 表示 `bytes` |
| **数据长度** | 4 字节 | 小端序的 `uint32`，表示后面 `Value` 的字节数     |
| **数据值**  | N 字节 | 实际数据，对于 `int64` 固定为 8 字节的小端序有符号整数   |

#### 第二次 commit： 隐私输出(就是加密公共输出)

| 字段       | 大小   | 说明                              |
|----------|------|---------------------------------|
| **数据长度** | 4 字节 | 小端序的 `uint32`，表示后面 `Value` 的字节数 |
| **数据值**  | N 字节 | 加密数据 []byte                     |

#### 第三次 commit： 加密前的明文哈希承诺

对公共输出和隐私输出原始值哈希承诺（隐私内存表示，公共内存），个数 = 公共输出个数 + 隐私输出个数

| 字段       | 大小   | 说明             |
|----------|------|----------------|
| **数据值**  | N 字节 | 32 位 SHA256 哈希 |

多个数据块可以依次连接，最终分为三次 `io::commit` 提交。

### 3. 示例（Rust）

```rust
use sp1_zkvm::io;
use sha2::{Sha256, Digest};

// 类型常量
const TYPE_INT64: u8 = 0x00;
const TYPE_BYTES: u8 = 0x01;
const TYPE_LIST: u8 = 0x02;

/// 公共输出值的枚举类型
pub enum PublicValue {
    Int64(i64),
    Bytes(Vec<u8>),
}

/// 加密数据项
pub struct EncryptedItem {
    pub data: Vec<u8>,  // 加密后的数据
}

/// 哈希承诺项
pub struct HashCommitmentItem {
    pub plaintext: Vec<u8>,  // 原始明文，用于计算哈希
}

/// 编码单个元素
fn encode_int64(value: i64) -> Vec<u8> {
    let mut buf = Vec::with_capacity(1 + 4 + 8);
    buf.push(TYPE_INT64);
    buf.extend_from_slice(&(8u32).to_le_bytes());
    buf.extend_from_slice(&value.to_le_bytes());
    buf
}

fn encode_bytes(data: &[u8]) -> Vec<u8> {
    let mut buf = Vec::with_capacity(1 + 4 + data.len());
    buf.push(TYPE_BYTES);
    buf.extend_from_slice(&(data.len() as u32).to_le_bytes());
    buf.extend_from_slice(data);
    buf
}

/// 编码加密数据项（没有类型标识，只有长度前缀）
fn encode_encrypted_item(data: &[u8]) -> Vec<u8> {
    let mut buf = Vec::with_capacity(4 + data.len());
    buf.extend_from_slice(&(data.len() as u32).to_le_bytes());
    buf.extend_from_slice(data);
    buf
}

/// 编码哈希承诺项（只有32字节哈希值）
fn encode_hash_commitment_item(plaintext: &[u8]) -> Vec<u8> {
    let hash = Sha256::digest(plaintext);
    hash.to_vec()
}

/// 第一次 commit：公共输出（支持混合类型列表）
/// 
/// 格式：[Type: 0x02] [Length: 4字节] [元素1][元素2]...
/// 每个元素格式：[Type: 1字节] [Length: 4字节] [Value: N字节]
pub fn commit_public_mixed(values: &[PublicValue]) {
    let mut elements = Vec::new();
    for value in values {
        match value {
            PublicValue::Int64(v) => elements.push(encode_int64(*v)),
            PublicValue::Bytes(v) => elements.push(encode_bytes(v)),
        }
    }
    
    // 计算总长度
    let total_len: u32 = elements.iter()
        .map(|elem| elem.len() as u32)
        .sum();
    
    let mut buf = Vec::with_capacity(1 + 4 + total_len as usize);
    
    // 列表头：Type + Length
    buf.push(TYPE_LIST);
    buf.extend_from_slice(&total_len.to_le_bytes());
    
    // 添加所有元素
    for elem in elements {
        buf.extend_from_slice(&elem);
    }
    
    io::commit(&buf);
}

/// 第二次 commit：隐私输出（加密数据列表）
/// 
/// 格式：[Type: 0x02] [Length: 4字节] [元素1][元素2]...
/// 每个元素格式：[Length: 4字节] [Value: N字节]
pub fn commit_encrypted_list(items: &[EncryptedItem]) {
    let mut elements = Vec::new();
    for item in items {
        elements.push(encode_encrypted_item(&item.data));
    }
    
    // 计算总长度
    let total_len: u32 = elements.iter()
        .map(|elem| elem.len() as u32)
        .sum();
    
    let mut buf = Vec::with_capacity(1 + 4 + total_len as usize);
    
    // 列表头：Type + Length
    buf.push(TYPE_LIST);
    buf.extend_from_slice(&total_len.to_le_bytes());
    
    // 添加所有元素
    for elem in elements {
        buf.extend_from_slice(&elem);
    }
    
    io::commit(&buf);
}

/// 第三次 commit：明文哈希承诺列表
/// 
/// 格式：[Type: 0x02] [Length: 4字节] [元素1][元素2]...
/// 每个元素格式：[Value: 32字节]
pub fn commit_hash_commitment_list(items: &[HashCommitmentItem]) {
    let mut elements = Vec::new();
    for item in items {
        elements.push(encode_hash_commitment_item(&item.plaintext));
    }
    
    // 计算总长度
    let total_len: u32 = elements.iter()
        .map(|elem| elem.len() as u32)
        .sum();
    
    let mut buf = Vec::with_capacity(1 + 4 + total_len as usize);
    
    // 列表头：Type + Length
    buf.push(TYPE_LIST);
    buf.extend_from_slice(&total_len.to_le_bytes());
    
    // 添加所有元素
    for elem in elements {
        buf.extend_from_slice(&elem);
    }
    
    io::commit(&buf);
}

// ============ 使用示例 ============

#[no_mangle]
pub fn main() {
    // 第一次 commit：公共输出（混合类型列表）
    commit_public_mixed(&[
        PublicValue::Int64(42),
        PublicValue::Bytes(b"hello".to_vec()),
        PublicValue::Int64(100),
    ]);
    
    // 第二次 commit：加密数据列表
    let plaintext1 = b"sensitive data 1";
    let plaintext2 = b"sensitive data 2";
    let encrypted1 = encrypt_data(plaintext1, &[0u8; 32]);
    let encrypted2 = encrypt_data(plaintext2, &[0u8; 32]);
    
    commit_encrypted_list(&[
        EncryptedItem { data: encrypted1 },
        EncryptedItem { data: encrypted2 },
    ]);
    
    // 第三次 commit：哈希承诺列表
    commit_hash_commitment_list(&[
        HashCommitmentItem { plaintext: plaintext1.to_vec() },
        HashCommitmentItem { plaintext: plaintext2.to_vec() },
    ]);
}

// 示例加密函数（实际使用时替换为真实加密逻辑）
fn encrypt_data(data: &[u8], key: &[u8]) -> Vec<u8> {
    let mut result = data.to_vec();
    for (i, byte) in result.iter_mut().enumerate() {
        *byte ^= key[i % key.len()];
    }
    result
}
```

### 4. 解析方式
链下服务（如 Go）按照上述格式逐块读取即可。已经提供了 Go 解析示例，其他语言类似。

### 5. 注意事项
- **整数必须使用小端序**，与 SP1 内部保持一致。
- `int64` 的长度固定为 8，请勿自行变更。
- 如果未来需要扩展更多类型（如 `uint64`、`bool` 等），可以约定新的类型标识，并保持向前兼容。

### 6.为什么不用“每个值独立提交”？
如果每个值单独调用 `io::commit`，虽然 Rust 端简单，但 Go 端无法区分各段数据的类型和边界（除非事先约定固定顺序和类型，这牺牲了灵活性）。采用长度前缀加类型标记的方案，做到了**自描述**，真正实现自动化解析。