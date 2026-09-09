package dto

type AssetInfo struct {
	ID   string
	Name string
}

type AssetFile struct {
	Name    string
	Content []byte
	Size    int64
}
