package core

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("vics: not found")

type Store interface {
	// Assets
	UpsertAsset(ctx context.Context, asset Asset) error
	GetAsset(ctx context.Context, id AssetID) (Asset, error)
	SearchAssets(ctx context.Context, query AssetQuery) (AssetSearchResult, error)

	// Tags
	UpsertTag(ctx context.Context, tag Tag) error
	GetTag(ctx context.Context, id TagID) (Tag, error)
	DeleteTag(ctx context.Context, id TagID) error

	// Labels
	UpsertTagLabel(ctx context.Context, label TagLabel) error
	ListTagLabels(ctx context.Context, tagID TagID) ([]TagLabel, error)
	SuggestTags(ctx context.Context, query TagSuggestQuery) ([]TagSuggestion, error)

	// Assignments
	AddTagToAsset(ctx context.Context, assetID AssetID, tagID TagID, source TagAssignmentSource) error
	RemoveTagFromAsset(ctx context.Context, assetID AssetID, tagID TagID) error
	ListAssetTags(ctx context.Context, assetID AssetID) ([]Tag, error)

	// Perceptual hashes
	UpsertAssetPerceptualHash(ctx context.Context, hash AssetPerceptualHash) error
	ListAssetPerceptualHashes(ctx context.Context, assetID AssetID) ([]AssetPerceptualHash, error)
	FindSimilarAssets(ctx context.Context, query SimilarityQuery) ([]AssetSimilarity, error)
}
