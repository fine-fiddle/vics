package perceptual

import (
	"context"

	"github.com/example/vics/core"
)

type Hasher interface {
	HashAsset(ctx context.Context, asset core.Asset) ([]core.AssetPerceptualHash, error)
}
