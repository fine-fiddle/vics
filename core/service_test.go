package core

import (
	"context"
	"testing"
)

func TestServiceDefaultsSearchLimit(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store)

	if _, err := service.Search(context.Background(), AssetQuery{}); err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if store.searchQuery.Limit != 100 {
		t.Fatalf("search limit = %d, want 100", store.searchQuery.Limit)
	}
}

func TestServiceDefaultsSuggestLimit(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store)

	if _, err := service.SuggestTags(context.Background(), "mo", LangEnglish); err != nil {
		t.Fatalf("SuggestTags returned error: %v", err)
	}
	if store.suggestQuery.Limit != 20 {
		t.Fatalf("suggest limit = %d, want 20", store.suggestQuery.Limit)
	}
	if store.suggestQuery.Prefix != "mo" || store.suggestQuery.Language != LangEnglish {
		t.Fatalf("suggest query = %+v, want prefix mo and language en", store.suggestQuery)
	}
}

func TestServiceAddTagUsesUserSource(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store)

	if err := service.AddTag(context.Background(), "asset_1", "tag_mountain"); err != nil {
		t.Fatalf("AddTag returned error: %v", err)
	}
	if store.addTagSource != TagAssignmentUser {
		t.Fatalf("source = %q, want %q", store.addTagSource, TagAssignmentUser)
	}
}

func TestServiceDefaultsFindSimilar(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store)

	if _, err := service.FindSimilar(context.Background(), SimilarityQuery{}); err != nil {
		t.Fatalf("FindSimilar returned error: %v", err)
	}
	if store.similarityQuery.Limit != 50 {
		t.Fatalf("similarity limit = %d, want 50", store.similarityQuery.Limit)
	}
	if store.similarityQuery.Algorithm != PerceptualHashPHash {
		t.Fatalf("algorithm = %q, want %q", store.similarityQuery.Algorithm, PerceptualHashPHash)
	}
	if store.similarityQuery.MaxDistance != 10 {
		t.Fatalf("max distance = %d, want 10", store.similarityQuery.MaxDistance)
	}
}

type recordingStore struct {
	searchQuery     AssetQuery
	suggestQuery    TagSuggestQuery
	similarityQuery SimilarityQuery
	addTagSource    TagAssignmentSource
}

func (s *recordingStore) UpsertAsset(ctx context.Context, asset Asset) error {
	return nil
}

func (s *recordingStore) GetAsset(ctx context.Context, id AssetID) (Asset, error) {
	return Asset{}, nil
}

func (s *recordingStore) SearchAssets(ctx context.Context, query AssetQuery) (AssetSearchResult, error) {
	s.searchQuery = query
	return AssetSearchResult{}, nil
}

func (s *recordingStore) UpsertTag(ctx context.Context, tag Tag) error {
	return nil
}

func (s *recordingStore) GetTag(ctx context.Context, id TagID) (Tag, error) {
	return Tag{}, nil
}

func (s *recordingStore) DeleteTag(ctx context.Context, id TagID) error {
	return nil
}

func (s *recordingStore) UpsertTagLabel(ctx context.Context, label TagLabel) error {
	return nil
}

func (s *recordingStore) ListTagLabels(ctx context.Context, tagID TagID) ([]TagLabel, error) {
	return nil, nil
}

func (s *recordingStore) SuggestTags(ctx context.Context, query TagSuggestQuery) ([]TagSuggestion, error) {
	s.suggestQuery = query
	return nil, nil
}

func (s *recordingStore) AddTagToAsset(ctx context.Context, assetID AssetID, tagID TagID, source TagAssignmentSource) error {
	s.addTagSource = source
	return nil
}

func (s *recordingStore) RemoveTagFromAsset(ctx context.Context, assetID AssetID, tagID TagID) error {
	return nil
}

func (s *recordingStore) ListAssetTags(ctx context.Context, assetID AssetID) ([]Tag, error) {
	return nil, nil
}

func (s *recordingStore) UpsertAssetPerceptualHash(ctx context.Context, hash AssetPerceptualHash) error {
	return nil
}

func (s *recordingStore) ListAssetPerceptualHashes(ctx context.Context, assetID AssetID) ([]AssetPerceptualHash, error) {
	return nil, nil
}

func (s *recordingStore) FindSimilarAssets(ctx context.Context, query SimilarityQuery) ([]AssetSimilarity, error) {
	s.similarityQuery = query
	return nil, nil
}
