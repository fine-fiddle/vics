package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/example/vics/core"
	"github.com/example/vics/perceptual"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

var _ core.Store = (*Store)(nil)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := New(db)
	if err := store.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate sqlite schema: %w", err)
	}
	return nil
}

func (s *Store) UpsertAsset(ctx context.Context, asset core.Asset) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO assets (id, path, hash, mime_type, width, height, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			path = excluded.path,
			hash = excluded.hash,
			mime_type = excluded.mime_type,
			width = excluded.width,
			height = excluded.height,
			updated_at = excluded.updated_at
	`, string(asset.ID), asset.Path, strings.ToLower(asset.Hash), asset.MimeType, asset.Width, asset.Height, asset.CreatedAt, asset.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert asset %q: %w", asset.ID, err)
	}
	return nil
}

func (s *Store) GetAsset(ctx context.Context, id core.AssetID) (core.Asset, error) {
	if err := s.ensureDB(); err != nil {
		return core.Asset{}, err
	}
	asset, err := scanAsset(s.db.QueryRowContext(ctx, `
		SELECT id, path, hash, mime_type, width, height, created_at, updated_at
		FROM assets
		WHERE id = ?
	`, string(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return core.Asset{}, fmt.Errorf("asset %q: %w", id, core.ErrNotFound)
	}
	if err != nil {
		return core.Asset{}, fmt.Errorf("get asset %q: %w", id, err)
	}
	return asset, nil
}

func (s *Store) SearchAssets(ctx context.Context, query core.AssetQuery) (core.AssetSearchResult, error) {
	if err := s.ensureDB(); err != nil {
		return core.AssetSearchResult{}, err
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	where, args := buildAssetFilter("a", query.IncludeTags, query.ExcludeTags, query.Text)

	var total int
	countSQL := "SELECT COUNT(*) FROM assets a WHERE " + where
	if err := s.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return core.AssetSearchResult{}, fmt.Errorf("count matching assets: %w", err)
	}

	assetSQL := `
		SELECT a.id, a.path, a.hash, a.mime_type, a.width, a.height, a.created_at, a.updated_at
		FROM assets a
		WHERE ` + where + `
		ORDER BY ` + assetOrderBy(query.Sort) + `
		LIMIT ? OFFSET ?
	`
	assetArgs := append(append([]any{}, args...), limit, offset)
	rows, err := s.db.QueryContext(ctx, assetSQL, assetArgs...)
	if err != nil {
		return core.AssetSearchResult{}, fmt.Errorf("search assets: %w", err)
	}
	defer rows.Close()

	var assets []core.Asset
	for rows.Next() {
		asset, err := scanAsset(rows)
		if err != nil {
			return core.AssetSearchResult{}, fmt.Errorf("scan asset search result: %w", err)
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return core.AssetSearchResult{}, fmt.Errorf("iterate asset search results: %w", err)
	}

	facets, err := s.searchFacets(ctx, where, args)
	if err != nil {
		return core.AssetSearchResult{}, err
	}

	return core.AssetSearchResult{
		Assets: assets,
		Total:  total,
		Facets: facets,
	}, nil
}

func (s *Store) UpsertTag(ctx context.Context, tag core.Tag) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tags (id, kind, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			kind = excluded.kind,
			updated_at = excluded.updated_at
	`, string(tag.ID), string(tag.Kind), tag.CreatedAt, tag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert tag %q: %w", tag.ID, err)
	}
	return nil
}

func (s *Store) GetTag(ctx context.Context, id core.TagID) (core.Tag, error) {
	if err := s.ensureDB(); err != nil {
		return core.Tag{}, err
	}
	var tag core.Tag
	err := s.db.QueryRowContext(ctx, `
		SELECT id, kind, created_at, updated_at
		FROM tags
		WHERE id = ?
	`, string(id)).Scan(&tag.ID, &tag.Kind, &tag.CreatedAt, &tag.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Tag{}, fmt.Errorf("tag %q: %w", id, core.ErrNotFound)
	}
	if err != nil {
		return core.Tag{}, fmt.Errorf("get tag %q: %w", id, err)
	}
	return tag, nil
}

func (s *Store) DeleteTag(ctx context.Context, id core.TagID) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete tag transaction: %w", err)
	}
	defer tx.Rollback()

	for _, query := range []string{
		"DELETE FROM asset_tags WHERE tag_id = ?",
		"DELETE FROM tag_labels WHERE tag_id = ?",
		"DELETE FROM tags WHERE id = ?",
	} {
		if _, err := tx.ExecContext(ctx, query, string(id)); err != nil {
			return fmt.Errorf("delete tag %q: %w", id, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete tag %q: %w", id, err)
	}
	return nil
}

func (s *Store) UpsertTagLabel(ctx context.Context, label core.TagLabel) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	normalized := normalizeLabel(label.Normalized)
	if normalized == "" {
		normalized = normalizeLabel(label.Text)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tag_labels (tag_id, language, text, normalized, is_primary, source)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(tag_id, language, normalized) DO UPDATE SET
			text = excluded.text,
			is_primary = excluded.is_primary,
			source = excluded.source
	`, string(label.TagID), string(label.Language), label.Text, normalized, boolToInt(label.IsPrimary), string(label.Source))
	if err != nil {
		return fmt.Errorf("upsert label %q/%q/%q: %w", label.TagID, label.Language, normalized, err)
	}
	return nil
}

func (s *Store) ListTagLabels(ctx context.Context, tagID core.TagID) ([]core.TagLabel, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag_id, language, text, normalized, is_primary, source
		FROM tag_labels
		WHERE tag_id = ?
		ORDER BY is_primary DESC, language ASC, text ASC
	`, string(tagID))
	if err != nil {
		return nil, fmt.Errorf("list labels for tag %q: %w", tagID, err)
	}
	defer rows.Close()

	var labels []core.TagLabel
	for rows.Next() {
		label, err := scanTagLabel(rows)
		if err != nil {
			return nil, fmt.Errorf("scan tag label: %w", err)
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tag labels: %w", err)
	}
	return labels, nil
}

func (s *Store) SuggestTags(ctx context.Context, query core.TagSuggestQuery) ([]core.TagSuggestion, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	conditions := []string{"tl.normalized LIKE ?"}
	args := []any{normalizeLabel(query.Prefix) + "%"}
	if query.Language != "" {
		conditions = append(conditions, "tl.language = ?")
		args = append(args, string(query.Language))
	}
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			t.id,
			tl.text,
			tl.language,
			tl.text,
			t.kind,
			COUNT(at.asset_id) AS asset_count
		FROM tag_labels tl
		JOIN tags t ON t.id = tl.tag_id
		LEFT JOIN asset_tags at ON at.tag_id = t.id
		WHERE `+strings.Join(conditions, " AND ")+`
		GROUP BY t.id, tl.text, tl.language, t.kind, tl.is_primary
		ORDER BY tl.is_primary DESC, asset_count DESC, tl.text ASC
		LIMIT ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("suggest tags: %w", err)
	}
	defer rows.Close()

	var suggestions []core.TagSuggestion
	for rows.Next() {
		var suggestion core.TagSuggestion
		if err := rows.Scan(
			&suggestion.TagID,
			&suggestion.DisplayText,
			&suggestion.Language,
			&suggestion.MatchedText,
			&suggestion.Kind,
			&suggestion.AssetCount,
		); err != nil {
			return nil, fmt.Errorf("scan tag suggestion: %w", err)
		}
		suggestions = append(suggestions, suggestion)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tag suggestions: %w", err)
	}
	return suggestions, nil
}

func (s *Store) AddTagToAsset(ctx context.Context, assetID core.AssetID, tagID core.TagID, source core.TagAssignmentSource) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO asset_tags (asset_id, tag_id, source, confidence, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(asset_id, tag_id) DO UPDATE SET
			source = excluded.source,
			confidence = excluded.confidence
	`, string(assetID), string(tagID), string(source), 1.0, now)
	if err != nil {
		return fmt.Errorf("add tag %q to asset %q: %w", tagID, assetID, err)
	}
	return nil
}

func (s *Store) RemoveTagFromAsset(ctx context.Context, assetID core.AssetID, tagID core.TagID) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM asset_tags
		WHERE asset_id = ? AND tag_id = ?
	`, string(assetID), string(tagID))
	if err != nil {
		return fmt.Errorf("remove tag %q from asset %q: %w", tagID, assetID, err)
	}
	return nil
}

func (s *Store) ListAssetTags(ctx context.Context, assetID core.AssetID) ([]core.Tag, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.kind, t.created_at, t.updated_at
		FROM tags t
		JOIN asset_tags at ON at.tag_id = t.id
		WHERE at.asset_id = ?
		ORDER BY t.id ASC
	`, string(assetID))
	if err != nil {
		return nil, fmt.Errorf("list tags for asset %q: %w", assetID, err)
	}
	defer rows.Close()

	var tags []core.Tag
	for rows.Next() {
		var tag core.Tag
		if err := rows.Scan(&tag.ID, &tag.Kind, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan asset tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate asset tags: %w", err)
	}
	return tags, nil
}

func (s *Store) UpsertAssetPerceptualHash(ctx context.Context, hash core.AssetPerceptualHash) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO asset_perceptual_hashes (asset_id, algorithm, hash_hex, bit_length, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(asset_id, algorithm) DO UPDATE SET
			hash_hex = excluded.hash_hex,
			bit_length = excluded.bit_length,
			created_at = excluded.created_at
	`, string(hash.AssetID), string(hash.Algorithm), strings.ToLower(hash.HashHex), hash.BitLength, hash.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert perceptual hash %q/%q: %w", hash.AssetID, hash.Algorithm, err)
	}
	return nil
}

func (s *Store) ListAssetPerceptualHashes(ctx context.Context, assetID core.AssetID) ([]core.AssetPerceptualHash, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT asset_id, algorithm, hash_hex, bit_length, created_at
		FROM asset_perceptual_hashes
		WHERE asset_id = ?
		ORDER BY algorithm ASC
	`, string(assetID))
	if err != nil {
		return nil, fmt.Errorf("list perceptual hashes for asset %q: %w", assetID, err)
	}
	defer rows.Close()

	var hashes []core.AssetPerceptualHash
	for rows.Next() {
		var hash core.AssetPerceptualHash
		if err := rows.Scan(&hash.AssetID, &hash.Algorithm, &hash.HashHex, &hash.BitLength, &hash.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan perceptual hash: %w", err)
		}
		hashes = append(hashes, hash)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate perceptual hashes: %w", err)
	}
	return hashes, nil
}

func (s *Store) FindSimilarAssets(ctx context.Context, query core.SimilarityQuery) ([]core.AssetSimilarity, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	algorithm := query.Algorithm
	if algorithm == "" {
		algorithm = core.PerceptualHashPHash
	}
	maxDistance := query.MaxDistance
	if maxDistance <= 0 {
		maxDistance = 10
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	sourceHash := strings.ToLower(query.HashHex)
	if query.AssetID != nil {
		err := s.db.QueryRowContext(ctx, `
			SELECT hash_hex
			FROM asset_perceptual_hashes
			WHERE asset_id = ? AND algorithm = ?
		`, string(*query.AssetID), string(algorithm)).Scan(&sourceHash)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("perceptual hash for asset %q algorithm %q: %w", *query.AssetID, algorithm, core.ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("get source perceptual hash: %w", err)
		}
	}
	if sourceHash == "" {
		return nil, errors.New("similarity query requires AssetID or HashHex")
	}

	where, filterArgs := buildAssetFilter("a", query.IncludeTags, query.ExcludeTags, "")
	args := append([]any{string(algorithm)}, filterArgs...)
	similarSQL := `
		SELECT a.id, a.path, a.hash, a.mime_type, a.width, a.height, a.created_at, a.updated_at, aph.hash_hex
		FROM asset_perceptual_hashes aph
		JOIN assets a ON a.id = aph.asset_id
		WHERE aph.algorithm = ? AND ` + where
	if query.AssetID != nil {
		similarSQL += " AND aph.asset_id <> ?"
		args = append(args, string(*query.AssetID))
	}

	rows, err := s.db.QueryContext(ctx, similarSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("load candidate perceptual hashes: %w", err)
	}
	defer rows.Close()

	var similarities []core.AssetSimilarity
	for rows.Next() {
		asset, hashHex, err := scanAssetWithHash(rows)
		if err != nil {
			return nil, fmt.Errorf("scan similarity candidate: %w", err)
		}
		distance, err := perceptual.HammingDistanceHex(sourceHash, hashHex)
		if err != nil {
			return nil, fmt.Errorf("compare perceptual hash for asset %q: %w", asset.ID, err)
		}
		if distance <= maxDistance {
			similarities = append(similarities, core.AssetSimilarity{
				Asset:    asset,
				Distance: distance,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate similarity candidates: %w", err)
	}

	sort.SliceStable(similarities, func(i, j int) bool {
		if similarities[i].Distance == similarities[j].Distance {
			return similarities[i].Asset.ID < similarities[j].Asset.ID
		}
		return similarities[i].Distance < similarities[j].Distance
	})
	if len(similarities) > limit {
		similarities = similarities[:limit]
	}
	return similarities, nil
}

func (s *Store) searchFacets(ctx context.Context, where string, args []any) ([]core.TagFacet, error) {
	facetSQL := `
		SELECT at.tag_id, COUNT(*) AS count
		FROM asset_tags at
		JOIN assets a ON a.id = at.asset_id
		WHERE ` + where + `
		GROUP BY at.tag_id
		ORDER BY count DESC, at.tag_id ASC
		LIMIT 50
	`
	rows, err := s.db.QueryContext(ctx, facetSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("load search facets: %w", err)
	}
	defer rows.Close()

	var facets []core.TagFacet
	for rows.Next() {
		var facet core.TagFacet
		if err := rows.Scan(&facet.TagID, &facet.Count); err != nil {
			return nil, fmt.Errorf("scan search facet: %w", err)
		}
		facets = append(facets, facet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search facets: %w", err)
	}
	return facets, nil
}

func buildAssetFilter(assetAlias string, includeTags []core.TagID, excludeTags []core.TagID, text string) (string, []any) {
	conditions := []string{"1 = 1"}
	var args []any

	for i, tagID := range includeTags {
		tagAlias := fmt.Sprintf("at_include_%d", i)
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM asset_tags %s WHERE %s.asset_id = %s.id AND %s.tag_id = ?)",
			tagAlias,
			tagAlias,
			assetAlias,
			tagAlias,
		))
		args = append(args, string(tagID))
	}

	if len(excludeTags) > 0 {
		conditions = append(conditions, fmt.Sprintf(
			"NOT EXISTS (SELECT 1 FROM asset_tags at_exclude WHERE at_exclude.asset_id = %s.id AND at_exclude.tag_id IN (%s))",
			assetAlias,
			placeholders(len(excludeTags)),
		))
		for _, tagID := range excludeTags {
			args = append(args, string(tagID))
		}
	}

	normalizedText := normalizeLabel(text)
	if normalizedText != "" {
		pattern := "%" + normalizedText + "%"
		conditions = append(conditions, fmt.Sprintf(
			`(LOWER(%s.path) LIKE ? OR EXISTS (
				SELECT 1
				FROM asset_tags at_text
				JOIN tag_labels tl_text ON tl_text.tag_id = at_text.tag_id
				WHERE at_text.asset_id = %s.id AND tl_text.normalized LIKE ?
			))`,
			assetAlias,
			assetAlias,
		))
		args = append(args, pattern, pattern)
	}

	return strings.Join(conditions, " AND "), args
}

func assetOrderBy(mode core.SortMode) string {
	switch mode {
	case core.SortFilename:
		return "a.path ASC, a.id ASC"
	case core.SortRecentlyAdded, core.SortRelevance, core.SortSimilarity, "":
		return "a.created_at DESC, a.id ASC"
	default:
		return "a.created_at DESC, a.id ASC"
	}
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

func normalizeLabel(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (s *Store) ensureDB() error {
	if s == nil || s.db == nil {
		return errors.New("sqlite store has no database")
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAsset(row scanner) (core.Asset, error) {
	var asset core.Asset
	err := row.Scan(
		&asset.ID,
		&asset.Path,
		&asset.Hash,
		&asset.MimeType,
		&asset.Width,
		&asset.Height,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	return asset, err
}

func scanAssetWithHash(row scanner) (core.Asset, string, error) {
	var asset core.Asset
	var hashHex string
	err := row.Scan(
		&asset.ID,
		&asset.Path,
		&asset.Hash,
		&asset.MimeType,
		&asset.Width,
		&asset.Height,
		&asset.CreatedAt,
		&asset.UpdatedAt,
		&hashHex,
	)
	return asset, hashHex, err
}

func scanTagLabel(row scanner) (core.TagLabel, error) {
	var label core.TagLabel
	var isPrimary int
	err := row.Scan(
		&label.TagID,
		&label.Language,
		&label.Text,
		&label.Normalized,
		&isPrimary,
		&label.Source,
	)
	label.IsPrimary = isPrimary != 0
	return label, err
}
