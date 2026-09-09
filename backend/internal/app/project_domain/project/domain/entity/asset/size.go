package asset

type Size struct {
	bytes int64
}

func NewSize(bytes int64) Size {
	return Size{bytes: bytes}
}

func (s Size) Bytes() int64 {
	return s.bytes
}

func (s Size) Equals(other Size) bool {
	return s.bytes == other.bytes
}

func (s Size) IsEmpty() bool {
	return s.bytes == 0
}
