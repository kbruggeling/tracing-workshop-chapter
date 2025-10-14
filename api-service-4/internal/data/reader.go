package data

import (
	"io"
	"time"
)

// ReadAll reads all data from the given reader and validates it
func ReadAll(r io.Reader) ([]byte, error) {
	// Read all data from the reader
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Validate the data with a small processing delay
	validateData(data)

	return data, nil
}

// validateData validates the data and simulates processing time
func validateData(data []byte) {
	// Simulate data validation (3ms processing time)
	time.Sleep(3 * time.Millisecond)
}