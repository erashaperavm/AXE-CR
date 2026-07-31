package bridge

func InputInt64(idx int64) int64 {
	// todo: 本地或网络套接字 Node Core 完成后接入，此处仅返回示例数据
	switch idx {
	case 1:
		return 0
	case 3:
		return 1
	default:
		return 2
	}
}

func InputBytes(idx int64) []byte {
	// todo: 本地或网络套接字 Node Core 完成后接入，此处仅返回示例数据
	switch idx {
	case 0:
		return []byte("hello")
	case 2:
		return []byte("world")
	default:
		return []byte("default")
	}
}

func OutputInt64(idx int64, data int64) {
	// todo: 本地或网络套接字 Node Core 完成后接入
	return
}

func OutputBytes(idx int64, data []byte) {
	// todo: 本地或网络套接字 Node Core 完成后接入
	return
}
