package main

import (
	"fmt"

	"github.com/doremi-id/doremid"
)

func main() {
	// Basic usage
	generator := doremid.New(doremid.Config{
		JustIntonationDigits:   4,
		EqualTemperamentDigits: 5,
		Separator:              "-",
	})

	// Generate random ID
	id := generator.NewID()
	fmt.Printf("Random ID: %s\n", id)

	// Batch generate unique random IDs
	randomIDs := generator.BatchGenerateRandomIDs(3)
	fmt.Printf("Unique random batch IDs: %v\n", randomIDs)

	// Demonstrate uniqueness with larger batch
	largerBatch := generator.BatchGenerateRandomIDs(10)
	fmt.Printf("Larger unique batch (%d IDs): %v\n", len(largerBatch), largerBatch)

	// Batch generate sequential IDs
	batchIDs := generator.BatchGenerateIDs(3, 0)
	fmt.Printf("Sequential batch IDs: %v\n", batchIDs)

	// ID and position conversion
	position := generator.IDToPosition(batchIDs[0])
	fmt.Printf("ID '%s' position: %d\n", batchIDs[0], position)

	convertedID := generator.PositionToID(position)
	fmt.Printf("Position %d to ID: %s\n", position, convertedID)

	// Check maximum combinations
	maxCombs := generator.MaxCombinations()
	fmt.Printf("Max combinations: %s\n", formatNumber(maxCombs))

	// Use default configuration
	defaultGen := doremid.NewWithDefaults()
	defaultID := defaultGen.NewID()
	fmt.Printf("Default config ID: %s\n", defaultID)

	// Different configuration example
	smallGen := doremid.New(doremid.Config{
		JustIntonationDigits:   1,
		EqualTemperamentDigits: 2,
		Separator:              "-",
	})
	smallID := smallGen.NewID()
	fmt.Printf("Small config ID: %s\n", smallID)
	fmt.Printf("Small config max combinations: %d\n", smallGen.MaxCombinations())

	// Batch generation with limits example
	limitedIDs := smallGen.BatchGenerateIDs(10, 80) // Automatically limited
	fmt.Printf("Limited batch generation: %d IDs\n", len(limitedIDs))
}

// formatNumber formats large numbers with thousand separators
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	str := fmt.Sprintf("%d", n)
	result := ""

	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(char)
	}

	return result
}
