package GeeCache

// A ByteView holds an immutable view of bytes
// 主要用于表示缓存值
type ByteView struct {
	b []byte // 存储真实的缓存值，byte可支持任意的数据类型的存储
}

// Len returns the view's length
func (v ByteView) Len() int {
	return len(v.b) // 实现Lenders() int 实现接口，可返回内存大小
}

// ByteSlice returns a copy of the data as a byte slice
// 返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// String returns the data as a string, making a copy if necessary
func (v ByteView) String() string {
	return string(v.b)
}
