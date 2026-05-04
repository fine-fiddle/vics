package core

type PerceptualHashAlgorithm string

const (
	PerceptualHashAverage    PerceptualHashAlgorithm = "ahash"
	PerceptualHashDifference PerceptualHashAlgorithm = "dhash"
	PerceptualHashPHash      PerceptualHashAlgorithm = "phash"
)

type AssetPerceptualHash struct {
	AssetID   AssetID
	Algorithm PerceptualHashAlgorithm
	HashHex   string // canonical lowercase hex string
	BitLength int
	CreatedAt int64
}

type HashDistance struct {
	AssetID   AssetID
	Algorithm PerceptualHashAlgorithm
	Distance  int
}
