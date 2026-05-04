package perceptual

import "testing"

func TestHammingDistanceHex(t *testing.T) {
	tests := []struct {
		name    string
		a       string
		b       string
		want    int
		wantErr bool
	}{
		{name: "identical", a: "deadbeef", b: "deadbeef", want: 0},
		{name: "all bits differ", a: "00", b: "ff", want: 8},
		{name: "nibbles differ", a: "0f", b: "f0", want: 8},
		{name: "uppercase accepted", a: "0F", b: "f0", want: 8},
		{name: "length mismatch", a: "00", b: "0000", wantErr: true},
		{name: "invalid hex", a: "zz", b: "00", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HammingDistanceHex(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("distance = %d, want %d", got, tt.want)
			}
		})
	}
}
