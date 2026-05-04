package ingest

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/vics/core"
)

func TestImporterImportsImagesAndSkipsOtherFiles(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	writeTinyPNG(t, filepath.Join(root, "emoji.png"))
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("not an image"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	store := &importRecordingStore{}
	importer := &Importer{Store: store}

	result, err := importer.Import(ctx, ImportOptions{RootPath: root})
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if result.Imported != 1 {
		t.Fatalf("imported = %d, want 1", result.Imported)
	}
	if result.Skipped != 1 {
		t.Fatalf("skipped = %d, want 1", result.Skipped)
	}
	if len(store.assets) != 1 {
		t.Fatalf("stored assets = %d, want 1", len(store.assets))
	}
	asset := store.assets[0]
	if asset.Width != 1 || asset.Height != 1 {
		t.Fatalf("dimensions = %dx%d, want 1x1", asset.Width, asset.Height)
	}
	if asset.MimeType != "image/png" {
		t.Fatalf("mime type = %q, want image/png", asset.MimeType)
	}
}

func writeTinyPNG(t *testing.T, path string) {
	t.Helper()

	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatalf("DecodeString returned error: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}

type importRecordingStore struct {
	assets []core.Asset
	hashes []core.AssetPerceptualHash
}

func (s *importRecordingStore) UpsertAsset(ctx context.Context, asset core.Asset) error {
	s.assets = append(s.assets, asset)
	return nil
}

func (s *importRecordingStore) GetAsset(ctx context.Context, id core.AssetID) (core.Asset, error) {
	return core.Asset{}, nil
}

func (s *importRecordingStore) SearchAssets(ctx context.Context, query core.AssetQuery) (core.AssetSearchResult, error) {
	return core.AssetSearchResult{}, nil
}

func (s *importRecordingStore) UpsertTag(ctx context.Context, tag core.Tag) error {
	return nil
}

func (s *importRecordingStore) GetTag(ctx context.Context, id core.TagID) (core.Tag, error) {
	return core.Tag{}, nil
}

func (s *importRecordingStore) DeleteTag(ctx context.Context, id core.TagID) error {
	return nil
}

func (s *importRecordingStore) UpsertTagLabel(ctx context.Context, label core.TagLabel) error {
	return nil
}

func (s *importRecordingStore) ListTagLabels(ctx context.Context, tagID core.TagID) ([]core.TagLabel, error) {
	return nil, nil
}

func (s *importRecordingStore) SuggestTags(ctx context.Context, query core.TagSuggestQuery) ([]core.TagSuggestion, error) {
	return nil, nil
}

func (s *importRecordingStore) AddTagToAsset(ctx context.Context, assetID core.AssetID, tagID core.TagID, source core.TagAssignmentSource) error {
	return nil
}

func (s *importRecordingStore) RemoveTagFromAsset(ctx context.Context, assetID core.AssetID, tagID core.TagID) error {
	return nil
}

func (s *importRecordingStore) ListAssetTags(ctx context.Context, assetID core.AssetID) ([]core.Tag, error) {
	return nil, nil
}

func (s *importRecordingStore) UpsertAssetPerceptualHash(ctx context.Context, hash core.AssetPerceptualHash) error {
	s.hashes = append(s.hashes, hash)
	return nil
}

func (s *importRecordingStore) ListAssetPerceptualHashes(ctx context.Context, assetID core.AssetID) ([]core.AssetPerceptualHash, error) {
	return nil, nil
}

func (s *importRecordingStore) FindSimilarAssets(ctx context.Context, query core.SimilarityQuery) ([]core.AssetSimilarity, error) {
	return nil, nil
}
