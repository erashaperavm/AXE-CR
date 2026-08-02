## 开发者文档：公开值提交标准格式

### 1. 目的
规范 zkVM 程序中公开输出（通过 `sp1_zkvm::io::commit` 提交）的二进制格式，使得链下服务（如 Go、Python）能够自动解析这些值，无需为每个程序单独编写解析逻辑。

### 2. 数据块结构
每个提交的数据块由三部分组成，按顺序紧密排列：

| 字段 | 大小 | 说明 |
|------|------|------|
| **类型标识** | 1 字节 | `0x00` 表示 `int64`，`0x01` 表示 `bytes` |
| **数据长度** | 4 字节 | 小端序的 `uint32`，表示后面 `Value` 的字节数 |
| **数据值** | N 字节 | 实际数据，对于 `int64` 固定为 8 字节的小端序有符号整数 |

多个数据块可以依次连接，最终一次性通过 `io::commit` 提交（或多次提交，效果相同，因为 SP1 会顺序拼接）。

### 3. 示例（Rust）

```rust
use sp1_zkvm::io;

const TYPE_INT64: u8 = 0x00;
const TYPE_BYTES: u8 = 0x01;

/// 提交一个 int64
pub fn commit_i64(v: i64) {
    let mut buf = Vec::with_capacity(1 + 4 + 8);
    buf.push(TYPE_INT64);
    buf.extend_from_slice(&(8u32).to_le_bytes());
    buf.extend_from_slice(&v.to_le_bytes());
    io::commit(&buf);
}

/// 提交一个字节切片
pub fn commit_bytes(data: &[u8]) {
    let mut buf = Vec::with_capacity(1 + 4 + data.len());
    buf.push(TYPE_BYTES);
    buf.extend_from_slice(&(data.len() as u32).to_le_bytes());
    buf.extend_from_slice(data);
    io::commit(&buf);
}

// 在 main 函数中使用
#[no_mangle]
pub fn main() {
    // 提交一个整数
    commit_i64(12345);
    // 提交一段文本
    commit_bytes(b"hello world");
    // 多次提交或一次提交多个块都可以
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