//go:build unit

package druckauftrag_repo

import (
	"testing"
	"time"
)

func TestBackoffDauer(t *testing.T) {
	tests := []struct {
		versuch int
		want    time.Duration
	}{
		{versuch: 0, want: 0},
		{versuch: 1, want: 5 * time.Second},
		{versuch: 2, want: 15 * time.Second},
		{versuch: 3, want: 30 * time.Second},
		{versuch: 4, want: 60 * time.Second},
		{versuch: 5, want: 180 * time.Second},
		{versuch: 6, want: 0},
		{versuch: 7, want: 0},
	}

	for _, tt := range tests {
		got := backoffDauer(tt.versuch)
		if got != tt.want {
			t.Errorf("backoffDauer(%d) = %v, want %v", tt.versuch, got, tt.want)
		}
	}
}
