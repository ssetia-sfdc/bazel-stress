package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <usecs> <input_file> <output_file>\n", os.Args[0])
		os.Exit(1)
	}

	busyWaitUsecs, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid usecs: %v\n", err)
		os.Exit(1)
	}

	inputFile := os.Args[2]
	outputFile := os.Args[3]

	// Busy-wait for the specified duration
	start := time.Now()
	targetDuration := time.Duration(busyWaitUsecs) * time.Microsecond

	for {
		now := time.Now()
		elapsed := now.Sub(start)
		if elapsed >= targetDuration {
			break
		}
	}

	// If no input file provided, create empty output
	if inputFile == "" {
		outFile, err := os.OpenFile(outputFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		outFile.Close()
		return
	}

	// Read all lines from input file
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	if len(lines) == 0 {
		fmt.Fprintf(os.Stderr, "Input file is empty\n")
		os.Exit(1)
	}

	// Select random lines (between 1 and all lines, but at least 1)
	numLinesToSelect := randomInt(len(lines)) + 1
	if numLinesToSelect > len(lines) {
		numLinesToSelect = len(lines)
	}

	// Randomly select lines
	selectedIndices := make(map[int]bool)
	for len(selectedIndices) < numLinesToSelect {
		idx := randomInt(len(lines))
		selectedIndices[idx] = true
	}

	// Write selected lines to output file
	outFile, err := os.OpenFile(outputFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for idx := range selectedIndices {
		writer.WriteString(lines[idx])
		writer.WriteString("\n")
	}
	writer.Flush()
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	b := make([]byte, 4)
	rand.Read(b)
	val := int(uint32(b[0])|uint32(b[1])<<8|uint32(b[2])<<16|uint32(b[3])<<24)
	if val < 0 {
		val = -val
	}
	return val % max
}

