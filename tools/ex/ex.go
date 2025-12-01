package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <usecs> <input_file1> [input_file2 ...] <output_file>\n", os.Args[0])
		os.Exit(1)
	}

	busyWaitUsecs, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid usecs: %v\n", err)
		os.Exit(1)
	}

	// Last argument is output file, all others (after usecs) are input files
	outputFile := os.Args[len(os.Args)-1]
	inputFiles := os.Args[2 : len(os.Args)-1]

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

	// If no input files provided, create empty output
	if len(inputFiles) == 0 || (len(inputFiles) == 1 && inputFiles[0] == "") {
		outFile, err := os.OpenFile(outputFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		outFile.Close()
		return
	}

	// Read all lines from all input files
	var allLines []string
	for _, inputFile := range inputFiles {
		if inputFile == "" {
			continue
		}
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file %s: %v\n", inputFile, err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			allLines = append(allLines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			file.Close()
			fmt.Fprintf(os.Stderr, "Error reading input file %s: %v\n", inputFile, err)
			os.Exit(1)
		}
		file.Close()
	}

	if len(allLines) == 0 {
		fmt.Fprintf(os.Stderr, "All input files are empty\n")
		os.Exit(1)
	}

	// Select random lines (between 1 and all lines, but at least 1)
	numLinesToSelect := randomInt(len(allLines)) + 1
	if numLinesToSelect > len(allLines) {
		numLinesToSelect = len(allLines)
	}

	// Randomly select lines
	selectedIndices := make(map[int]bool)
	for len(selectedIndices) < numLinesToSelect {
		idx := randomInt(len(allLines))
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
		writer.WriteString(allLines[idx])
		writer.WriteString("\n")
	}
	writer.Flush()
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	return rand.Intn(max)
}
