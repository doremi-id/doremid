package doremid

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewID(t *testing.T) {
	tests := []struct {
		name                   string
		justIntonationDigits   int
		equalTemperamentDigits int
		separator              string
	}{
		{
			name:                   "basic test",
			justIntonationDigits:   2,
			equalTemperamentDigits: 3,
			separator:              "-",
		},
		{
			name:                   "no separator",
			justIntonationDigits:   1,
			equalTemperamentDigits: 2,
			separator:              "",
		},
		{
			name:                   "long ID",
			justIntonationDigits:   5,
			equalTemperamentDigits: 8,
			separator:              "_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := New(Config{
				JustIntonationDigits:   tt.justIntonationDigits,
				EqualTemperamentDigits: tt.equalTemperamentDigits,
				Separator:              tt.separator,
			})
			id := generator.NewID()

			// Check ID is not empty
			if id == "" {
				t.Error("generated ID should not be empty")
			}

			// Check ID format
			if tt.separator == "" {
				// No separator case, check total length directly
				expectedLen := tt.justIntonationDigits*2 + tt.equalTemperamentDigits
				if len(id) != expectedLen {
					t.Errorf("ID length without separator is incorrect, expected %d, got %d", expectedLen, len(id))
				}
			} else {
				// With separator case
				parts := strings.Split(id, tt.separator)
				if len(parts) != 2 {
					t.Errorf("ID format is incorrect, expected 2 parts, got %d parts", len(parts))
				}

				// Check just part length
				justPart := parts[0]
				expectedJustLen := tt.justIntonationDigits * 2
				if len(justPart) != expectedJustLen {
					t.Errorf("just part length is incorrect, expected %d, got %d", expectedJustLen, len(justPart))
				}

				// Check equal part length
				equalPart := parts[1]
				if len(equalPart) != tt.equalTemperamentDigits {
					t.Errorf("equal part length is incorrect, expected %d, got %d", tt.equalTemperamentDigits, len(equalPart))
				}
			}
		})
	}
}

func TestBatchGenerateIDs(t *testing.T) {
	generator := New(Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 2,
		Separator:              "-",
	})

	tests := []struct {
		name          string
		count         int64
		startPosition int64
	}{
		{
			name:          "generate 5 IDs",
			count:         5,
			startPosition: 0,
		},
		{
			name:          "generate 3 IDs starting from position 10",
			count:         3,
			startPosition: 10,
		},
		{
			name:          "generate 0 IDs",
			count:         0,
			startPosition: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := generator.BatchGenerateIDs(tt.count, tt.startPosition)

			// Check count
			if int64(len(ids)) != tt.count {
				t.Errorf("generated ID count is incorrect, expected %d, got %d", tt.count, len(ids))
			}

			// Check format of each ID
			for i, id := range ids {
				parts := strings.Split(id, generator.Separator)
				if len(parts) != 2 {
					t.Errorf("ID[%d] format is incorrect", i)
				}
			}

			// Check sequential ordering of IDs (adjacent IDs should be consecutive)
			for i := 0; i < len(ids)-1; i++ {
				pos1 := generator.IDToPosition(ids[i])
				pos2 := generator.IDToPosition(ids[i+1])
				if pos2 != pos1+1 {
					t.Errorf("ID sequence is not consecutive, positions %d and %d", pos1, pos2)
				}
			}
		})
	}
}

func TestIDToPosition(t *testing.T) {
	generator := New(Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 2,
		Separator:              "-",
	})

	tests := []struct {
		name     string
		id       string
		expected int64
	}{
		{
			name:     "first ID",
			id:       "do-00",
			expected: 0,
		},
		{
			name:     "known position ID",
			id:       "do-01",
			expected: 1,
		},
		{
			name:     "invalid ID - format error",
			id:       "do00",
			expected: -1,
		},
		{
			name:     "invalid ID - just part length error",
			id:       "d-00",
			expected: -1,
		},
		{
			name:     "invalid ID - equal part length error",
			id:       "do-0",
			expected: -1,
		},
		{
			name:     "invalid ID - contains illegal characters",
			id:       "dx-00",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.IDToPosition(tt.id)
			if result != tt.expected {
				t.Errorf("expected position %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestPositionToID(t *testing.T) {
	generator := New(Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 2,
		Separator:              "-",
	})

	tests := []struct {
		name     string
		position int64
		expected string
	}{
		{
			name:     "position 0",
			position: 0,
			expected: "do-00",
		},
		{
			name:     "position 1",
			position: 1,
			expected: "do-01",
		},
		{
			name:     "negative position",
			position: -1,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.PositionToID(tt.position)
			if result != tt.expected {
				t.Errorf("expected ID '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPositionToIDAndIDToPosition(t *testing.T) {
	generator := New(Config{
		JustIntonationDigits:   2,
		EqualTemperamentDigits: 3,
		Separator:              "-",
	})

	// Test round-trip conversion consistency
	positions := []int64{0, 1, 10, 100, 500, 1000}

	for _, pos := range positions {
		t.Run(fmt.Sprintf("round-trip test position %d", pos), func(t *testing.T) {
			// Position -> ID -> Position
			id := generator.PositionToID(pos)
			if id == "" {
				t.Errorf("ID generated from position %d is empty", pos)
				return
			}

			backPos := generator.IDToPosition(id)
			if backPos != pos {
				t.Errorf("round-trip conversion failed: original position %d, ID '%s', converted back position %d", pos, id, backPos)
			}
		})
	}
}

func TestIntPow(t *testing.T) {
	// Since intPow is a package-private function, we test it indirectly through public functions
	// Here we test PositionToID results to verify intPow correctness
	generator := New(Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 2,
		Separator:              "-",
	})

	// Test some known positions and expected IDs
	testCases := []struct {
		position int64
		expected string
	}{
		{0, "do-00"},
		{12, "do-10"},  // 12 = 0*144 + 12, equal part: 12 = 1*12 + 0
		{144, "re-00"}, // 144 = 12*12
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("indirect test intPow position %d", tc.position), func(t *testing.T) {
			result := generator.PositionToID(tc.position)
			if result != tc.expected {
				t.Errorf("position %d expected ID '%s', got '%s'", tc.position, tc.expected, result)
			}
		})
	}
}

func TestLargeNumbers(t *testing.T) {
	generator := New(Config{
		JustIntonationDigits:   3,
		EqualTemperamentDigits: 4,
		Separator:              "_",
	})

	// Test handling of large numbers
	largePositions := []int64{10000, 50000, 100000}

	for _, pos := range largePositions {
		t.Run(fmt.Sprintf("large number test %d", pos), func(t *testing.T) {
			id := generator.PositionToID(pos)
			if id == "" {
				t.Errorf("ID generated from large position %d is empty", pos)
				return
			}

			backPos := generator.IDToPosition(id)
			if backPos != pos {
				t.Errorf("large number round-trip conversion failed: original position %d, converted back position %d", pos, backPos)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("minimum parameters", func(t *testing.T) {
		generator := New(Config{
			JustIntonationDigits:   1,
			EqualTemperamentDigits: 1,
			Separator:              "",
		})

		id := generator.NewID()
		if len(id) != 3 { // 2 chars just + 1 char equal
			t.Errorf("minimum parameters ID length is incorrect, expected 3, got %d", len(id))
		}
	})

	t.Run("verify all notes can be generated", func(t *testing.T) {
		generator := New(Config{
			JustIntonationDigits:   1,
			EqualTemperamentDigits: 1,
			Separator:              "-",
		})

		// Generate enough IDs to cover all possible combinations
		seen := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			id := generator.NewID()
			seen[id] = true
		}

		if len(seen) < 10 { // Should see at least multiple different IDs
			t.Errorf("too few ID variations generated, only %d types", len(seen))
		}
	})

	t.Run("test default config", func(t *testing.T) {
		generator := NewWithDefaults()

		id := generator.NewID()
		if id == "" {
			t.Error("generated ID with default config should not be empty")
		}

		// Check default separator
		if !strings.Contains(id, "-") {
			t.Error("default config should use '-' as separator")
		}
	})
}

func TestMaxCombinations(t *testing.T) {
	tests := []struct {
		name                   string
		justIntonationDigits   int
		equalTemperamentDigits int
		expectedMax            int64
	}{
		{
			name:                   "1x1 configuration",
			justIntonationDigits:   1,
			equalTemperamentDigits: 1,
			expectedMax:            7 * 12, // 7^1 * 12^1 = 84
		},
		{
			name:                   "2x2 configuration",
			justIntonationDigits:   2,
			equalTemperamentDigits: 2,
			expectedMax:            49 * 144, // 7^2 * 12^2 = 7056
		},
		{
			name:                   "default configuration",
			justIntonationDigits:   4,
			equalTemperamentDigits: 5,
			expectedMax:            2401 * 248832, // 7^4 * 12^5 = 597,394,032
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := New(Config{
				JustIntonationDigits:   tt.justIntonationDigits,
				EqualTemperamentDigits: tt.equalTemperamentDigits,
				Separator:              "-",
			})

			maxCombinations := generator.MaxCombinations()
			if maxCombinations != tt.expectedMax {
				t.Errorf("expected max combinations %d, got %d", tt.expectedMax, maxCombinations)
			}
		})
	}
}

func TestBatchGenerateIDsWithLimits(t *testing.T) {
	// Use small configuration for easier testing
	generator := New(Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 1,
		Separator:              "-",
	})

	maxCombinations := generator.MaxCombinations() // Should be 84

	tests := []struct {
		name          string
		count         int64
		startPosition int64
		expectedCount int64
		shouldBeEmpty bool
	}{
		{
			name:          "normal case within limits",
			count:         5,
			startPosition: 0,
			expectedCount: 5,
			shouldBeEmpty: false,
		},
		{
			name:          "count exceeds remaining combinations",
			count:         10,
			startPosition: 80, // Only 4 combinations left (84-80=4)
			expectedCount: 4,
			shouldBeEmpty: false,
		},
		{
			name:          "start position at maximum",
			count:         5,
			startPosition: 84, // At maximum, no combinations left
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "start position beyond maximum",
			count:         5,
			startPosition: 100,
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "negative count",
			count:         -5,
			startPosition: 0,
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "negative start position",
			count:         5,
			startPosition: -1,
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "zero count",
			count:         0,
			startPosition: 0,
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "request all combinations",
			count:         maxCombinations,
			startPosition: 0,
			expectedCount: maxCombinations,
			shouldBeEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := generator.BatchGenerateIDs(tt.count, tt.startPosition)

			if tt.shouldBeEmpty {
				if len(ids) != 0 {
					t.Errorf("expected empty result, got %d IDs", len(ids))
				}
				return
			}

			if int64(len(ids)) != tt.expectedCount {
				t.Errorf("expected %d IDs, got %d", tt.expectedCount, len(ids))
			}

			// Verify all IDs are valid and sequential
			for i, id := range ids {
				expectedPos := tt.startPosition + int64(i)
				actualPos := generator.IDToPosition(id)
				if actualPos != expectedPos {
					t.Errorf("ID[%d] position mismatch: expected %d, got %d", i, expectedPos, actualPos)
				}
			}
		})
	}
}

func TestBatchGenerateRandomIDs(t *testing.T) {
	generator := New(Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 1,
		Separator:              "-",
	})

	maxCombinations := generator.MaxCombinations() // Should be 84

	tests := []struct {
		name          string
		count         int64
		expectedCount int64
		shouldBeEmpty bool
	}{
		{
			name:          "generate 5 unique random IDs",
			count:         5,
			expectedCount: 5,
			shouldBeEmpty: false,
		},
		{
			name:          "generate 1 random ID",
			count:         1,
			expectedCount: 1,
			shouldBeEmpty: false,
		},
		{
			name:          "generate 0 IDs",
			count:         0,
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "generate negative count",
			count:         -1,
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "generate all possible combinations",
			count:         maxCombinations,
			expectedCount: maxCombinations,
			shouldBeEmpty: false,
		},
		{
			name:          "generate more than maximum combinations",
			count:         maxCombinations + 1,
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "generate half of maximum combinations",
			count:         maxCombinations / 2,
			expectedCount: maxCombinations / 2,
			shouldBeEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := generator.BatchGenerateRandomIDs(tt.count)

			// Check count
			if int64(len(ids)) != tt.expectedCount {
				t.Errorf("expected %d IDs, got %d", tt.expectedCount, len(ids))
			}

			if tt.shouldBeEmpty {
				if len(ids) != 0 {
					t.Errorf("expected empty result, got %d IDs", len(ids))
				}
				return
			}

			// Check format of each ID
			for i, id := range ids {
				if id == "" {
					t.Errorf("ID[%d] should not be empty", i)
				}

				// Verify ID format
				parts := strings.Split(id, generator.Separator)
				if len(parts) != 2 {
					t.Errorf("ID[%d] '%s' has incorrect format", i, id)
				}

				// Check part lengths
				justPart := parts[0]
				equalPart := parts[1]

				expectedJustLen := generator.JustIntonationDigits * 2
				if len(justPart) != expectedJustLen {
					t.Errorf("ID[%d] just part length incorrect: expected %d, got %d", i, expectedJustLen, len(justPart))
				}

				if len(equalPart) != generator.EqualTemperamentDigits {
					t.Errorf("ID[%d] equal part length incorrect: expected %d, got %d", i, generator.EqualTemperamentDigits, len(equalPart))
				}

				// Verify ID can be converted back to valid position
				pos := generator.IDToPosition(id)
				if pos < 0 || pos >= maxCombinations {
					t.Errorf("ID[%d] '%s' converts to invalid position %d", i, id, pos)
				}
			}

			// Check uniqueness - this should be 100% for the new implementation
			uniqueIDs := make(map[string]bool)
			uniquePositions := make(map[int64]bool)

			for i, id := range ids {
				if uniqueIDs[id] {
					t.Errorf("Duplicate ID found: '%s' at index %d", id, i)
				}
				uniqueIDs[id] = true

				pos := generator.IDToPosition(id)
				if uniquePositions[pos] {
					t.Errorf("Duplicate position found: %d for ID '%s' at index %d", pos, id, i)
				}
				uniquePositions[pos] = true
			}

			// Verify we have exactly the expected number of unique IDs
			if int64(len(uniqueIDs)) != tt.expectedCount {
				t.Errorf("Expected %d unique IDs, got %d", tt.expectedCount, len(uniqueIDs))
			}

			if int64(len(uniquePositions)) != tt.expectedCount {
				t.Errorf("Expected %d unique positions, got %d", tt.expectedCount, len(uniquePositions))
			}
		})
	}
}

func TestRandomSample(t *testing.T) {
	generator := New(Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 1,
		Separator:              "-",
	})

	tests := []struct {
		name  string
		max   int
		count int
	}{
		{"sample 5 from 10", 10, 5},
		{"sample 1 from 10", 10, 1},
		{"sample 10 from 10", 10, 10},
		{"sample 3 from 100", 100, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := generator.randomSample(tt.max, tt.count)

			// Check count
			if len(sample) != tt.count {
				t.Errorf("expected %d samples, got %d", tt.count, len(sample))
			}

			// Check uniqueness
			unique := make(map[int]bool)
			for _, pos := range sample {
				if unique[pos] {
					t.Errorf("duplicate position in sample: %d", pos)
				}
				unique[pos] = true

				// Check range
				if pos < 0 || pos >= tt.max {
					t.Errorf("position %d out of range [0, %d)", pos, tt.max)
				}
			}

			// Verify uniqueness count
			if len(unique) != tt.count {
				t.Errorf("expected %d unique positions, got %d", tt.count, len(unique))
			}
		})
	}
}
