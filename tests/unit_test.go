package tests

import (
	"testing"
)

// TestBatchSplitting tests splitting codes into batches
func TestBatchSplitting(t *testing.T) {
	tests := []struct {
		name      string
		codes     []string
		batchSize int
		expected  int
	}{
		{
			name:      "empty codes",
			codes:     []string{},
			batchSize: 10,
			expected:  0,
		},
		{
			name:      "single batch",
			codes:     []string{"A", "B", "C"},
			batchSize: 10,
			expected:  1,
		},
		{
			name:      "multiple batches",
			codes:     []string{"A", "B", "C", "D", "E"},
			batchSize: 2,
			expected:  3,
		},
		{
			name:      "exact batch size",
			codes:     []string{"A", "B", "C", "D"},
			batchSize: 2,
			expected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := splitIntoBatches(tt.codes, tt.batchSize)
			if len(batches) != tt.expected {
				t.Errorf("expected %d batches, got %d", tt.expected, len(batches))
			}
		})
	}
}

// TestCodeDeduplication tests removing duplicate codes
func TestCodeDeduplication(t *testing.T) {
	tests := []struct {
		name     string
		codes    []string
		expected int
	}{
		{
			name:     "no duplicates",
			codes:    []string{"A", "B", "C"},
			expected: 3,
		},
		{
			name:     "with duplicates",
			codes:    []string{"A", "B", "A", "C", "B"},
			expected: 3,
		},
		{
			name:     "all duplicates",
			codes:    []string{"A", "A", "A"},
			expected: 1,
		},
		{
			name:     "empty",
			codes:    []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateCodes(tt.codes)
			if len(result) != tt.expected {
				t.Errorf("expected %d unique codes, got %d", tt.expected, len(result))
			}
		})
	}
}

// TestFilterEmptyCodes tests removing empty strings from codes
func TestFilterEmptyCodes(t *testing.T) {
	tests := []struct {
		name     string
		codes    []string
		expected int
	}{
		{
			name:     "no empty codes",
			codes:    []string{"A", "B", "C"},
			expected: 3,
		},
		{
			name:     "with empty codes",
			codes:    []string{"A", "", "B", "", "C"},
			expected: 3,
		},
		{
			name:     "all empty",
			codes:    []string{"", "", ""},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEmptyCodes(tt.codes)
			if len(result) != tt.expected {
				t.Errorf("expected %d non-empty codes, got %d", tt.expected, len(result))
			}
		})
	}
}

// Helper functions for testing

func splitIntoBatches(codes []string, batchSize int) [][]string {
	var batches [][]string

	for i := 0; i < len(codes); i += batchSize {
		end := i + batchSize
		if end > len(codes) {
			end = len(codes)
		}
		batches = append(batches, codes[i:end])
	}

	return batches
}

func deduplicateCodes(codes []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, code := range codes {
		if !seen[code] {
			seen[code] = true
			result = append(result, code)
		}
	}

	return result
}

func filterEmptyCodes(codes []string) []string {
	result := []string{}
	for _, code := range codes {
		if code != "" {
			result = append(result, code)
		}
	}
	return result
}
