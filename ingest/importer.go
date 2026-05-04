package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/vics/core"
	"github.com/example/vics/perceptual"
)

type ImportOptions struct {
	RootPath          string
	Recursive         bool
	OriginalTagLang   core.Lang
	CreateMissingTags bool
}

type ImportResult struct {
	Imported int
	Skipped  int
	Warnings []ImportWarning
}

type ImportWarning struct {
	Path    string
	Code    string
	Message string
	Similar []core.AssetSimilarity
}

type Importer struct {
	Store  core.Store
	Hasher perceptual.Hasher
}

func (i *Importer) Import(ctx context.Context, opts ImportOptions) (ImportResult, error) {
	if i == nil || i.Store == nil {
		return ImportResult{}, errors.New("importer requires a store")
	}
	if opts.RootPath == "" {
		return ImportResult{}, errors.New("import root path is required")
	}

	var result ImportResult
	info, err := os.Stat(opts.RootPath)
	if err != nil {
		return result, fmt.Errorf("stat import root: %w", err)
	}
	if !info.IsDir() {
		return i.importFile(ctx, opts.RootPath, &result)
	}

	err = filepath.WalkDir(opts.RootPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			if path != opts.RootPath && !opts.Recursive {
				return filepath.SkipDir
			}
			return nil
		}
		_, err := i.importFile(ctx, path, &result)
		return err
	})
	if err != nil {
		return result, fmt.Errorf("walk import root: %w", err)
	}
	return result, nil
}

func (i *Importer) importFile(ctx context.Context, path string, result *ImportResult) (ImportResult, error) {
	if !isImagePath(path) {
		result.Skipped++
		return *result, nil
	}

	sum, err := sha256File(path)
	if err != nil {
		return *result, fmt.Errorf("hash file %q: %w", path, err)
	}

	width, height, mimeType, warning := imageMetadata(path)
	if warning != nil {
		result.Warnings = append(result.Warnings, *warning)
	}
	if mimeType == "" {
		mimeType = mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	}

	now := time.Now().Unix()
	asset := core.Asset{
		ID:        assetIDFromHash(sum),
		Path:      path,
		Hash:      sum,
		MimeType:  mimeType,
		Width:     width,
		Height:    height,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := i.Store.UpsertAsset(ctx, asset); err != nil {
		return *result, fmt.Errorf("upsert imported asset %q: %w", path, err)
	}

	if i.Hasher != nil {
		hashes, err := i.Hasher.HashAsset(ctx, asset)
		if err != nil {
			result.Warnings = append(result.Warnings, ImportWarning{
				Path:    path,
				Code:    "perceptual_hash_failed",
				Message: err.Error(),
			})
		}
		for _, hash := range hashes {
			if err := i.Store.UpsertAssetPerceptualHash(ctx, hash); err != nil {
				return *result, fmt.Errorf("upsert perceptual hash for %q: %w", path, err)
			}
		}
	}

	// TODO: Parse Japanese tag sidecars and map raw labels to semantic Tag concepts.
	result.Imported++
	return *result, nil
}

func isImagePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func imageMetadata(path string) (int, int, string, *ImportWarning) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, "", &ImportWarning{
			Path:    path,
			Code:    "image_open_failed",
			Message: err.Error(),
		}
	}
	defer file.Close()

	cfg, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", &ImportWarning{
			Path:    path,
			Code:    "image_decode_config_failed",
			Message: err.Error(),
		}
	}
	return cfg.Width, cfg.Height, "image/" + format, nil
}

func assetIDFromHash(hash string) core.AssetID {
	if len(hash) > 16 {
		return core.AssetID("asset_" + hash[:16])
	}
	return core.AssetID("asset_" + hash)
}
