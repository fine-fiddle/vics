package goimagehash

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"time"

	img_hash "github.com/corona10/goimagehash"
	"github.com/example/vics/core"
)

type Hasher struct{}

func NewHasher() *Hasher {
	return &Hasher{}
}

func (h *Hasher) HashAsset(ctx context.Context, asset core.Asset) ([]core.AssetPerceptualHash, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := os.Open(asset.Path)
	if err != nil {
		return nil, fmt.Errorf("open asset image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode asset image: %w", err)
	}

	now := time.Now().Unix()
	var hashes []core.AssetPerceptualHash
	var errs []error

	add := func(algorithm core.PerceptualHashAlgorithm, hash *img_hash.ImageHash, err error) {
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", algorithm, err))
			return
		}
		hashes = append(hashes, toCoreHash(asset.ID, algorithm, hash, now))
	}

	averageHash, err := img_hash.AverageHash(img)
	add(core.PerceptualHashAverage, averageHash, err)

	differenceHash, err := img_hash.DifferenceHash(img)
	add(core.PerceptualHashDifference, differenceHash, err)

	perceptionHash, err := img_hash.PerceptionHash(img)
	add(core.PerceptualHashPHash, perceptionHash, err)

	if len(hashes) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return hashes, nil
}

func toCoreHash(assetID core.AssetID, algorithm core.PerceptualHashAlgorithm, hash *img_hash.ImageHash, createdAt int64) core.AssetPerceptualHash {
	bitLength := hash.Bits()
	if bitLength <= 0 {
		bitLength = 64
	}
	hexWidth := (bitLength + 3) / 4
	if hexWidth < 16 {
		hexWidth = 16
	}

	return core.AssetPerceptualHash{
		AssetID:   assetID,
		Algorithm: algorithm,
		HashHex:   strings.ToLower(fmt.Sprintf("%0*x", hexWidth, hash.GetHash())),
		BitLength: bitLength,
		CreatedAt: createdAt,
	}
}
