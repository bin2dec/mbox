package mbox

import "bytes"

// ScanMessages is a split function for bufio.Scanner that returns mbox messages.
func ScanMessages(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Try to find the token head in the given data.
	low := bytes.Index(data, []byte("From "))
	if low < 0 {
		// If we can't find the head, then we should try to request more data to process.
		return 0, nil, nil
	}

	// Check if we're at EOF, and if so, return the token with the rest of the data.
	if atEOF {
		return len(data), data[low:], nil
	}

	// Try to find the token tail. Initially, the tail index is relative to the token head.
	high := bytes.Index(data[low+5:], []byte("\nFrom "))
	if high < 0 {
		return 0, nil, nil
	}
	high += low + 5

	// Return the token.
	return high, data[low:high], nil
}
