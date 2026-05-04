package core

import (
	"context"
	"errors"
)

var ErrNilStore = errors.New("vics: nil store")

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Search(ctx context.Context, q AssetQuery) (AssetSearchResult, error) {
	store, err := s.storeOrErr()
	if err != nil {
		return AssetSearchResult{}, err
	}
	if q.Limit <= 0 {
		q.Limit = 100
	}
	return store.SearchAssets(ctx, q)
}

func (s *Service) SuggestTags(ctx context.Context, prefix string, lang Lang) ([]TagSuggestion, error) {
	store, err := s.storeOrErr()
	if err != nil {
		return nil, err
	}
	return store.SuggestTags(ctx, TagSuggestQuery{
		Prefix:   prefix,
		Language: lang,
		Limit:    20,
	})
}

func (s *Service) AddTag(ctx context.Context, assetID AssetID, tagID TagID) error {
	store, err := s.storeOrErr()
	if err != nil {
		return err
	}
	return store.AddTagToAsset(ctx, assetID, tagID, TagAssignmentUser)
}

func (s *Service) RemoveTag(ctx context.Context, assetID AssetID, tagID TagID) error {
	store, err := s.storeOrErr()
	if err != nil {
		return err
	}
	return store.RemoveTagFromAsset(ctx, assetID, tagID)
}

func (s *Service) FindSimilar(ctx context.Context, q SimilarityQuery) ([]AssetSimilarity, error) {
	store, err := s.storeOrErr()
	if err != nil {
		return nil, err
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Algorithm == "" {
		q.Algorithm = PerceptualHashPHash
	}
	if q.MaxDistance <= 0 {
		q.MaxDistance = 10
	}
	return store.FindSimilarAssets(ctx, q)
}

func (s *Service) AddOrUpdatePerceptualHash(ctx context.Context, h AssetPerceptualHash) error {
	store, err := s.storeOrErr()
	if err != nil {
		return err
	}
	return store.UpsertAssetPerceptualHash(ctx, h)
}

func (s *Service) storeOrErr() (Store, error) {
	if s == nil || s.store == nil {
		return nil, ErrNilStore
	}
	return s.store, nil
}
