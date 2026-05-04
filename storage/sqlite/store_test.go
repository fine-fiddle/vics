package sqlite

import (
	"context"
	"strings"
	"testing"

	"github.com/example/vics/core"
)

func TestStoreUpsertAndGetAsset(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	asset := testAsset("asset_mountain", "mountain.png", "hash_mountain")
	if err := store.UpsertAsset(ctx, asset); err != nil {
		t.Fatalf("UpsertAsset returned error: %v", err)
	}

	got, err := store.GetAsset(ctx, asset.ID)
	if err != nil {
		t.Fatalf("GetAsset returned error: %v", err)
	}
	if got != asset {
		t.Fatalf("asset = %+v, want %+v", got, asset)
	}
}

func TestStoreTagLabelsAndSuggestions(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	upsertTagWithLabels(t, ctx, store, core.Tag{
		ID:        "tag_mountain",
		Kind:      core.TagKindObject,
		CreatedAt: 1,
		UpdatedAt: 1,
	}, []core.TagLabel{
		{TagID: "tag_mountain", Language: core.LangJapanese, Text: "山", IsPrimary: true, Source: core.LabelSourceOriginal},
		{TagID: "tag_mountain", Language: core.LangEnglish, Text: "mountain", IsPrimary: true, Source: core.LabelSourceTranslated},
		{TagID: "tag_mountain", Language: core.LangRomaji, Text: "yama", IsPrimary: true, Source: core.LabelSourceTranslated},
	})

	labels, err := store.ListTagLabels(ctx, "tag_mountain")
	if err != nil {
		t.Fatalf("ListTagLabels returned error: %v", err)
	}
	if len(labels) != 3 {
		t.Fatalf("labels length = %d, want 3", len(labels))
	}

	english, err := store.SuggestTags(ctx, core.TagSuggestQuery{
		Prefix:   "mo",
		Language: core.LangEnglish,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("SuggestTags English returned error: %v", err)
	}
	if len(english) != 1 || english[0].DisplayText != "mountain" {
		t.Fatalf("english suggestions = %+v, want mountain", english)
	}

	romaji, err := store.SuggestTags(ctx, core.TagSuggestQuery{
		Prefix:   "ya",
		Language: core.LangRomaji,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("SuggestTags Romaji returned error: %v", err)
	}
	if len(romaji) != 1 || romaji[0].DisplayText != "yama" {
		t.Fatalf("romaji suggestions = %+v, want yama", romaji)
	}
}

func TestStoreAddTagToAssetAndListAssetTags(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	asset := testAsset("asset_mountain", "mountain.png", "hash_mountain")
	tag := testTag("tag_mountain", core.TagKindObject)
	mustUpsertAsset(t, ctx, store, asset)
	mustUpsertTag(t, ctx, store, tag)

	if err := store.AddTagToAsset(ctx, asset.ID, tag.ID, core.TagAssignmentImported); err != nil {
		t.Fatalf("AddTagToAsset returned error: %v", err)
	}
	tags, err := store.ListAssetTags(ctx, asset.ID)
	if err != nil {
		t.Fatalf("ListAssetTags returned error: %v", err)
	}
	if len(tags) != 1 || tags[0].ID != tag.ID {
		t.Fatalf("asset tags = %+v, want tag_mountain", tags)
	}
}

func TestStoreSearchAssetsTags(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	setupSearchFixture(t, ctx, store)

	includeMountain, err := store.SearchAssets(ctx, core.AssetQuery{
		IncludeTags: []core.TagID{"tag_mountain"},
		Sort:        core.SortFilename,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("SearchAssets include returned error: %v", err)
	}
	if assetIDs(includeMountain.Assets) != "asset_blue_mountain,asset_mountain" {
		t.Fatalf("include mountain ids = %s", assetIDs(includeMountain.Assets))
	}

	excludeFace, err := store.SearchAssets(ctx, core.AssetQuery{
		ExcludeTags: []core.TagID{"tag_face"},
		Sort:        core.SortFilename,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("SearchAssets exclude returned error: %v", err)
	}
	if assetIDs(excludeFace.Assets) != "asset_blue_mountain,asset_mountain" {
		t.Fatalf("exclude face ids = %s", assetIDs(excludeFace.Assets))
	}

	includeAll, err := store.SearchAssets(ctx, core.AssetQuery{
		IncludeTags: []core.TagID{"tag_mountain", "tag_blue"},
		Sort:        core.SortFilename,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("SearchAssets include all returned error: %v", err)
	}
	if assetIDs(includeAll.Assets) != "asset_blue_mountain" {
		t.Fatalf("include all ids = %s, want asset_blue_mountain", assetIDs(includeAll.Assets))
	}
}

func TestStorePerceptualHashesAndFindSimilar(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	source := testAsset("asset_source", "source.png", "hash_source")
	closeMatch := testAsset("asset_close", "close.png", "hash_close")
	farMatch := testAsset("asset_far", "far.png", "hash_far")
	for _, asset := range []core.Asset{source, closeMatch, farMatch} {
		mustUpsertAsset(t, ctx, store, asset)
	}

	for _, hash := range []core.AssetPerceptualHash{
		{AssetID: source.ID, Algorithm: core.PerceptualHashPHash, HashHex: "0000000000000000", BitLength: 64, CreatedAt: 1},
		{AssetID: closeMatch.ID, Algorithm: core.PerceptualHashPHash, HashHex: "0000000000000001", BitLength: 64, CreatedAt: 1},
		{AssetID: farMatch.ID, Algorithm: core.PerceptualHashPHash, HashHex: "00000000000000ff", BitLength: 64, CreatedAt: 1},
	} {
		if err := store.UpsertAssetPerceptualHash(ctx, hash); err != nil {
			t.Fatalf("UpsertAssetPerceptualHash returned error: %v", err)
		}
	}

	hashes, err := store.ListAssetPerceptualHashes(ctx, source.ID)
	if err != nil {
		t.Fatalf("ListAssetPerceptualHashes returned error: %v", err)
	}
	if len(hashes) != 1 || hashes[0].HashHex != "0000000000000000" {
		t.Fatalf("hashes = %+v, want source phash", hashes)
	}

	similar, err := store.FindSimilarAssets(ctx, core.SimilarityQuery{
		HashHex:     "0000000000000000",
		Algorithm:   core.PerceptualHashPHash,
		MaxDistance: 5,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("FindSimilarAssets returned error: %v", err)
	}
	if len(similar) != 2 {
		t.Fatalf("similar length = %d, want 2 including source when searching by hash", len(similar))
	}

	similarByAsset, err := store.FindSimilarAssets(ctx, core.SimilarityQuery{
		AssetID:     &source.ID,
		Algorithm:   core.PerceptualHashPHash,
		MaxDistance: 5,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("FindSimilarAssets by asset returned error: %v", err)
	}
	if len(similarByAsset) != 1 {
		t.Fatalf("similar by asset length = %d, want 1", len(similarByAsset))
	}
	if similarByAsset[0].Asset.ID != closeMatch.ID || similarByAsset[0].Distance != 1 {
		t.Fatalf("similar by asset = %+v, want close match at distance 1", similarByAsset)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})
	return store
}

func setupSearchFixture(t *testing.T, ctx context.Context, store *Store) {
	t.Helper()

	for _, tag := range []core.Tag{
		testTag("tag_mountain", core.TagKindObject),
		testTag("tag_blue", core.TagKindColor),
		testTag("tag_face", core.TagKindObject),
	} {
		mustUpsertTag(t, ctx, store, tag)
	}

	for _, asset := range []core.Asset{
		testAsset("asset_mountain", "mountain.png", "hash_mountain"),
		testAsset("asset_blue_mountain", "blue_mountain.png", "hash_blue_mountain"),
		testAsset("asset_face", "face.png", "hash_face"),
	} {
		mustUpsertAsset(t, ctx, store, asset)
	}

	mustAddTag(t, ctx, store, "asset_mountain", "tag_mountain")
	mustAddTag(t, ctx, store, "asset_blue_mountain", "tag_mountain")
	mustAddTag(t, ctx, store, "asset_blue_mountain", "tag_blue")
	mustAddTag(t, ctx, store, "asset_face", "tag_face")
}

func upsertTagWithLabels(t *testing.T, ctx context.Context, store *Store, tag core.Tag, labels []core.TagLabel) {
	t.Helper()
	mustUpsertTag(t, ctx, store, tag)
	for _, label := range labels {
		if err := store.UpsertTagLabel(ctx, label); err != nil {
			t.Fatalf("UpsertTagLabel returned error: %v", err)
		}
	}
}

func testAsset(id core.AssetID, path string, hash string) core.Asset {
	return core.Asset{
		ID:        id,
		Path:      path,
		Hash:      hash,
		MimeType:  "image/png",
		Width:     16,
		Height:    16,
		CreatedAt: 1,
		UpdatedAt: 2,
	}
}

func testTag(id core.TagID, kind core.TagKind) core.Tag {
	return core.Tag{
		ID:        id,
		Kind:      kind,
		CreatedAt: 1,
		UpdatedAt: 2,
	}
}

func mustUpsertAsset(t *testing.T, ctx context.Context, store *Store, asset core.Asset) {
	t.Helper()
	if err := store.UpsertAsset(ctx, asset); err != nil {
		t.Fatalf("UpsertAsset returned error: %v", err)
	}
}

func mustUpsertTag(t *testing.T, ctx context.Context, store *Store, tag core.Tag) {
	t.Helper()
	if err := store.UpsertTag(ctx, tag); err != nil {
		t.Fatalf("UpsertTag returned error: %v", err)
	}
}

func mustAddTag(t *testing.T, ctx context.Context, store *Store, assetID core.AssetID, tagID core.TagID) {
	t.Helper()
	if err := store.AddTagToAsset(ctx, assetID, tagID, core.TagAssignmentImported); err != nil {
		t.Fatalf("AddTagToAsset returned error: %v", err)
	}
}

func assetIDs(assets []core.Asset) string {
	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		ids = append(ids, string(asset.ID))
	}
	return strings.Join(ids, ",")
}
