package perceptual

import (
	"encoding/hex"
	"fmt"
	"math/bits"
)

func HammingDistanceHex(a, b string) (int, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("hex strings have different lengths: %d and %d", len(a), len(b))
	}

	left, err := hex.DecodeString(a)
	if err != nil {
		return 0, fmt.Errorf("invalid first hex string: %w", err)
	}
	right, err := hex.DecodeString(b)
	if err != nil {
		return 0, fmt.Errorf("invalid second hex string: %w", err)
	}

	distance := 0
	for i := range left {
		distance += bits.OnesCount8(left[i] ^ right[i])
	}
	return distance, nil
}
