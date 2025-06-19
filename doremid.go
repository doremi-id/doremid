// Package doremid provides functions for generation and manipulation of musical note-based IDs.
// It supports both random ID generation and sequential ID generation with position-based conversion.
package doremid

import (
	"math/rand"
	"strings"
	"time"
)

// Generator holds the configuration and lookup tables for efficient ID generation
type Generator struct {
	// ID generation parameters
	JustIntonationDigits   int    // Number of musical note pairs in the first part
	EqualTemperamentDigits int    // Number of characters in the second part
	Separator              string // String used to separate the two parts of the ID

	// Musical note names as byte slices for better performance
	justIntonationBytes [][]byte
	// Character set as byte array for direct indexing
	equalTemperamentBytes []byte
	// Cached lengths
	justIntonationLen   int
	equalTemperamentLen int
	// Lookup maps for O(1) reverse conversion
	justIntonationMap   map[string]int
	equalTemperamentMap map[byte]int
	// Random number generator with proper seeding
	rand *rand.Rand
}

// Config defines the configuration for ID generation
type Config struct {
	// JustIntonationDigits specifies the number of musical note pairs in the first part
	JustIntonationDigits int

	// EqualTemperamentDigits specifies the number of characters in the second part
	EqualTemperamentDigits int

	// Separator is the string used to separate the two parts of the ID
	Separator string
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		JustIntonationDigits:   4,
		EqualTemperamentDigits: 5,
		Separator:              "-",
	}
}

// New creates a new ID generator with optimized lookup tables
func New(config Config) *Generator {
	g := &Generator{
		JustIntonationDigits:   config.JustIntonationDigits,
		EqualTemperamentDigits: config.EqualTemperamentDigits,
		Separator:              config.Separator,
		justIntonationBytes: [][]byte{
			[]byte("do"), []byte("re"), []byte("mi"), []byte("fa"),
			[]byte("so"), []byte("la"), []byte("ti"),
		},
		equalTemperamentBytes: []byte("0123456789ab"),
		rand:                  rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Cache lengths
	g.justIntonationLen = len(g.justIntonationBytes)
	g.equalTemperamentLen = len(g.equalTemperamentBytes)

	// Build lookup maps for O(1) reverse conversion
	g.justIntonationMap = make(map[string]int, g.justIntonationLen)
	for i, note := range g.justIntonationBytes {
		g.justIntonationMap[string(note)] = i
	}

	g.equalTemperamentMap = make(map[byte]int, g.equalTemperamentLen)
	for i, char := range g.equalTemperamentBytes {
		g.equalTemperamentMap[char] = i
	}

	return g
}

// NewWithDefaults creates a new generator with default configuration
func NewWithDefaults() *Generator {
	return New(DefaultConfig())
}

// NewID generates a random ID based on the generator's configuration.
// It creates an ID with two parts: a musical note part and an alphanumeric part,
// separated by the configured separator.
func (g *Generator) NewID() string {
	// Pre-estimate capacity: just part longest element is 2 bytes, equal part is 1 byte
	capacity := g.JustIntonationDigits*2 + len(g.Separator) + g.EqualTemperamentDigits
	result := make([]byte, 0, capacity)

	// Generate musical note part using optimized byte arrays
	for i := 0; i < g.JustIntonationDigits; i++ {
		result = append(result, g.justIntonationBytes[g.rand.Intn(g.justIntonationLen)]...)
	}

	// Add separator
	result = append(result, g.Separator...)

	// Generate alphanumeric part using direct byte indexing
	for i := 0; i < g.EqualTemperamentDigits; i++ {
		result = append(result, g.equalTemperamentBytes[g.rand.Intn(g.equalTemperamentLen)])
	}

	return string(result)
}

// BatchGenerateRandomIDs generates a batch of unique random IDs.
//
// Parameters:
//   - count: number of unique random IDs to generate
//
// Returns a slice of unique random IDs. Returns empty slice if count <= 0
// or count exceeds maximum possible combinations.
// Uses random sampling from all possible positions to ensure uniqueness without collision checking.
func (g *Generator) BatchGenerateRandomIDs(count int64) []string {
	if count <= 0 {
		return []string{}
	}

	maxCombinations := g.MaxCombinations()

	// Check if count exceeds maximum possible combinations
	if count > maxCombinations {
		return []string{}
	}

	// Generate random sample of positions without replacement
	positions := g.randomSample(int(maxCombinations), int(count))

	// Convert positions to IDs
	ids := make([]string, count)
	for i, pos := range positions {
		ids[i] = g.PositionToID(int64(pos))
	}

	return ids
}

// randomSample generates count unique random numbers from range [0, max).
// Uses reservoir sampling algorithm for efficient sampling without replacement.
func (g *Generator) randomSample(max, count int) []int {
	if count >= max {
		// Return all positions shuffled if count equals or exceeds max
		positions := make([]int, max)
		for i := 0; i < max; i++ {
			positions[i] = i
		}
		// Shuffle the entire array using Fisher-Yates
		for i := max - 1; i > 0; i-- {
			j := g.rand.Intn(i + 1)
			positions[i], positions[j] = positions[j], positions[i]
		}
		return positions[:count]
	}

	// For smaller samples, use a more straightforward approach
	// Create a set to track used positions
	used := make(map[int]bool)
	positions := make([]int, 0, count)

	// Generate unique random positions
	for len(positions) < count {
		pos := g.rand.Intn(max)
		if !used[pos] {
			used[pos] = true
			positions = append(positions, pos)
		}
	}

	return positions
}

// MaxCombinations returns the maximum number of unique IDs that can be generated
// with the current configuration.
func (g *Generator) MaxCombinations() int64 {
	justMax := int64(g.intPow(g.justIntonationLen, g.JustIntonationDigits))
	equalMax := int64(g.intPow(g.equalTemperamentLen, g.EqualTemperamentDigits))
	return justMax * equalMax
}

// BatchGenerateIDs generates a batch of sequential IDs starting from a specific position.
//
// Parameters:
//   - count: number of IDs to generate (will be limited by maximum combinations)
//   - startPosition: starting position in the sequence (0-based)
//
// Returns a slice of sequential IDs. The actual count may be less than requested
// if it would exceed the maximum possible combinations or go beyond valid positions.
func (g *Generator) BatchGenerateIDs(count int64, startPosition int64) []string {
	if count <= 0 || startPosition < 0 {
		return []string{}
	}

	maxCombinations := g.MaxCombinations()

	// Check if startPosition is beyond maximum
	if startPosition >= maxCombinations {
		return []string{}
	}

	// Limit count to not exceed maximum combinations
	if startPosition+count > maxCombinations {
		count = maxCombinations - startPosition
	}

	// Additional safety check
	if count <= 0 {
		return []string{}
	}

	ids := make([]string, count)
	for i := int64(0); i < count; i++ {
		ids[i] = g.PositionToID(startPosition + i)
	}
	return ids
}

// IDToPosition converts an ID back to its position in the sequential order.
//
// Parameters:
//   - id: the ID string to parse
//
// Returns:
//   - position in the sequence (0-based)
//   - -1 if the ID format is invalid
func (g *Generator) IDToPosition(id string) int64 {
	// Split ID by separator
	parts := strings.Split(id, g.Separator)
	if len(parts) != 2 {
		return -1
	}

	justPart := parts[0]
	equalPart := parts[1]

	// Validate part lengths
	if len(justPart) != g.JustIntonationDigits*2 || len(equalPart) != g.EqualTemperamentDigits {
		return -1
	}

	// Parse musical note part using O(1) map lookup
	justValue := int64(0)
	for i := 0; i < len(justPart); i += 2 {
		if i+1 >= len(justPart) {
			return -1 // Length is not a multiple of 2
		}
		twoChar := justPart[i : i+2]
		if index, found := g.justIntonationMap[twoChar]; found {
			justValue = justValue*int64(g.justIntonationLen) + int64(index)
		} else {
			return -1
		}
	}

	// Parse alphanumeric part using O(1) map lookup
	equalValue := int64(0)
	for _, char := range []byte(equalPart) {
		if index, found := g.equalTemperamentMap[char]; found {
			equalValue = equalValue*int64(g.equalTemperamentLen) + int64(index)
		} else {
			return -1
		}
	}

	// Calculate total position
	return justValue*int64(g.intPow(g.equalTemperamentLen, g.EqualTemperamentDigits)) + equalValue
}

// PositionToID generates an ID based on its position in the sequential order.
//
// Parameters:
//   - position: position in the sequence (0-based)
//
// Returns:
//   - the corresponding ID string
//   - empty string if position is negative
func (g *Generator) PositionToID(position int64) string {
	if position < 0 {
		return ""
	}

	// Calculate maximum value for alphanumeric part
	equalMax := int64(g.intPow(g.equalTemperamentLen, g.EqualTemperamentDigits))

	// Separate values for musical note and alphanumeric parts
	justValue := position / equalMax
	equalValue := position % equalMax

	// Pre-estimate capacity for efficiency
	capacity := g.JustIntonationDigits*2 + len(g.Separator) + g.EqualTemperamentDigits
	result := make([]byte, 0, capacity)

	// Generate musical note part
	justDigits := make([]int, g.JustIntonationDigits)
	temp := justValue
	for i := g.JustIntonationDigits - 1; i >= 0; i-- {
		justDigits[i] = int(temp % int64(g.justIntonationLen))
		temp /= int64(g.justIntonationLen)
	}

	for _, digit := range justDigits {
		result = append(result, g.justIntonationBytes[digit]...)
	}

	// Add separator
	result = append(result, g.Separator...)

	// Generate alphanumeric part using direct byte indexing
	equalDigits := make([]int, g.EqualTemperamentDigits)
	temp = equalValue
	for i := g.EqualTemperamentDigits - 1; i >= 0; i-- {
		equalDigits[i] = int(temp % int64(g.equalTemperamentLen))
		temp /= int64(g.equalTemperamentLen)
	}

	for _, digit := range equalDigits {
		result = append(result, g.equalTemperamentBytes[digit])
	}

	return string(result)
}

// intPow calculates integer power using binary exponentiation.
// This is a helper function for efficient power calculation.
func (g *Generator) intPow(base, exp int) int {
	result := 1
	for exp > 0 {
		if exp%2 == 1 {
			result *= base
		}
		base *= base
		exp /= 2
	}
	return result
}
