package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <usecs> <output_file>\n", os.Args[0])
		os.Exit(1)
	}

	busyWaitUsecs, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid usecs: %v\n", err)
		os.Exit(1)
	}

	output := os.Args[2]

	start := time.Now()
	targetDuration := time.Duration(busyWaitUsecs) * time.Microsecond

	for {
		now := time.Now()
		elapsed := now.Sub(start)
		if elapsed >= targetDuration {
			break
		}
	}

	// Create the output file
	file, err := os.OpenFile(output, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	file.Close()
}

