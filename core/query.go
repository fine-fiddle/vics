package core

type SortMode string

const (
	SortRecentlyAdded SortMode = "recently_added"
	SortFilename      SortMode = "filename"
	SortRelevance     SortMode = "relevance"
	SortSimilarity    SortMode = "similarity"
)

type AssetQuery struct {
	IncludeTags []TagID
	ExcludeTags []TagID

	Text string

	SimilarToAssetID *AssetID
	SimilarHashHex   string
	HashAlgorithm    PerceptualHashAlgorithm
	MaxHashDistance  int

	Limit  int
	Offset int
	Sort   SortMode
}

type AssetSearchResult struct {
	Assets []Asset
	Total  int
	Facets []TagFacet
}

type TagFacet struct {
	TagID TagID
	Count int
}

type TagSuggestion struct {
	TagID       TagID
	DisplayText string
	Language    Lang
	MatchedText string
	Kind        TagKind
	AssetCount  int
}

type TagSuggestQuery struct {
	Prefix   string
	Language Lang
	Limit    int
}

type SimilarityQuery struct {
	AssetID     *AssetID
	HashHex     string
	Algorithm   PerceptualHashAlgorithm
	MaxDistance int
	Limit       int

	IncludeTags []TagID
	ExcludeTags []TagID
}

type AssetSimilarity struct {
	Asset    Asset
	Distance int
}
