package core

type AssetID string

type Asset struct {
	ID        AssetID
	Path      string
	Hash      string // exact SHA-256 file hash, lowercase hex
	MimeType  string
	Width     int
	Height    int
	CreatedAt int64
	UpdatedAt int64
}
