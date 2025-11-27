package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output_dir> <num_files> <min_size_bytes> <max_size_bytes>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s /tmp/input_files 1000 1024 1048576\n", os.Args[0])
		os.Exit(1)
	}

	outputDirArg := os.Args[1]
	numFiles, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid num_files: %v\n", err)
		os.Exit(1)
	}

	minSize, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid min_size_bytes: %v\n", err)
		os.Exit(1)
	}

	maxSize, err := strconv.Atoi(os.Args[4])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid max_size_bytes: %v\n", err)
		os.Exit(1)
	}

	if minSize < 0 || maxSize < 0 {
		fmt.Fprintf(os.Stderr, "Size values must be non-negative\n")
		os.Exit(1)
	}

	if minSize > maxSize {
		fmt.Fprintf(os.Stderr, "min_size_bytes must be <= max_size_bytes\n")
		os.Exit(1)
	}

	if numFiles <= 0 {
		fmt.Fprintf(os.Stderr, "num_files must be positive\n")
		os.Exit(1)
	}

	// Get workspace root - Bazel sets BUILD_WORKSPACE_DIRECTORY when running via 'bazel run'
	workspaceRoot := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	if workspaceRoot == "" {
		// Fallback: try to find workspace root by looking for MODULE.bazel or WORKSPACE
		cwd, err := os.Getwd()
		if err == nil {
			// Walk up the directory tree to find MODULE.bazel or WORKSPACE
			dir := cwd
			for {
				if _, err := os.Stat(filepath.Join(dir, "MODULE.bazel")); err == nil {
					workspaceRoot = dir
					break
				}
				if _, err := os.Stat(filepath.Join(dir, "WORKSPACE")); err == nil {
					workspaceRoot = dir
					break
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					break // Reached root
				}
				dir = parent
			}
		}
	}

	// Resolve the output directory path
	var outputDir string
	if filepath.IsAbs(outputDirArg) {
		outputDir = outputDirArg
	} else {
		// Relative path - resolve relative to workspace root
		if workspaceRoot != "" {
			outputDir = filepath.Join(workspaceRoot, outputDirArg)
		} else {
			// Fallback: use current working directory
			outputDir = outputDirArg
		}
		// Clean the path
		outputDir = filepath.Clean(outputDir)
	}

	// Create output directory if it doesn't exist
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate files
	for i := 0; i < numFiles; i++ {
		fileName := fmt.Sprintf("file_%06d.txt", i)
		filePath := filepath.Join(outputDir, fileName)

		// Generate random size between minSize and maxSize
		sizeRange := maxSize - minSize + 1
		fileSize := minSize
		if sizeRange > 1 {
			fileSize += randomInt(sizeRange)
		}

		// Generate file with lines (for line-based reading)
		// Each line will be 20-100 characters, so we can estimate number of lines
		avgLineSize := 60 // Average line size including newline
		numLines := fileSize / avgLineSize
		if numLines < 1 {
			numLines = 1
		}

		// Build content with lines
		var content []byte
		remainingSize := fileSize
		for lineNum := 0; lineNum < numLines && remainingSize > 0; lineNum++ {
			// Last line: use remaining size
			if lineNum == numLines-1 {
				lineSize := remainingSize
				if lineSize > 0 {
					lineContent := make([]byte, lineSize)
					rand.Read(lineContent)
					// Make content printable ASCII
					for j := range lineContent {
						lineContent[j] = byte(32 + (int(lineContent[j]) % 95))
					}
					content = append(content, lineContent...)
				}
			} else {
				// Regular line: random size between 20-100 chars
				lineSize := randomInt(80) + 20
				if lineSize > remainingSize-1 {
					lineSize = remainingSize - 1
				}
				if lineSize > 0 {
					lineContent := make([]byte, lineSize)
					rand.Read(lineContent)
					// Make content printable ASCII
					for j := range lineContent {
						lineContent[j] = byte(32 + (int(lineContent[j]) % 95))
					}
					content = append(content, lineContent...)
					content = append(content, '\n')
					remainingSize -= lineSize + 1
				}
			}
		}

		// Trim to exact size if we went over
		if len(content) > fileSize {
			content = content[:fileSize]
		}

		// Write file
		err := os.WriteFile(filePath, content, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		if (i+1)%100 == 0 {
			fmt.Fprintf(os.Stderr, "Generated %d/%d files...\n", i+1, numFiles)
		}
	}

	fmt.Fprintf(os.Stderr, "Successfully generated %d files in %s\n", numFiles, outputDir)
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	b := make([]byte, 4)
	rand.Read(b)
	val := int(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)
	if val < 0 {
		val = -val
	}
	return val % max
}

