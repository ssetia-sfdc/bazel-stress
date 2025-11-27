package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

// LinearNoise generates a linear noise profile using a recursive midpoint displacement algorithm
type LinearNoise struct {
	keypoints []float64
	shift     int
}

// NewLinearNoise creates a new LinearNoise generator
func NewLinearNoise(shift int) *LinearNoise {
	size := 1 << shift
	ln := &LinearNoise{
		keypoints: make([]float64, size+1),
		shift:     shift,
	}
	ln.generate()
	return ln
}

// Size returns the number of keypoints (excluding the last one)
func (ln *LinearNoise) Size() int {
	return 1 << ln.shift
}

// Sum returns the sum of all keypoints
func (ln *LinearNoise) Sum() float64 {
	sum := 0.0
	for i := 0; i < len(ln.keypoints)-1; i++ {
		sum += ln.keypoints[i]
	}
	return sum
}

// Values returns all keypoint values (excluding the last one)
func (ln *LinearNoise) Values() []float64 {
	return ln.keypoints[:len(ln.keypoints)-1]
}

func (ln *LinearNoise) generate() {
	// Initialize endpoints with random values
	ln.keypoints[0] = randomFloat()
	ln.keypoints[1<<ln.shift] = randomFloat()

	// Generate recursively
	ln.generateAt(1<<(ln.shift-1), 1<<(ln.shift-1), 1.0)
}

func (ln *LinearNoise) generateAt(at, length int, scale float64) {
	for length > 0 {
		x := (randomFloat() - 0.5) * 2.0 * scale
		v := (math.Abs(ln.keypoints[at-length]+ln.keypoints[at+length]) / 2.0) + x
		if v < 0 {
			ln.keypoints[at] = 0
		} else if v > 1.0 {
			ln.keypoints[at] = 1.0
		} else {
			ln.keypoints[at] = v
		}

		if length >= 2 {
			ln.generateAt(at-length/2, length/2, scale/2)
		}
		length /= 2
		at += length
		scale /= 2
	}
}

func randomFloat() float64 {
	b := make([]byte, 8)
	rand.Read(b)
	// Convert to float64 in range [0, 1)
	val := float64(uint64(b[0])|uint64(b[1])<<8|uint64(b[2])<<16|uint64(b[3])<<24|uint64(b[4])<<32|uint64(b[5])<<40|uint64(b[6])<<48|uint64(b[7])<<56) / float64(1<<64)
	return val
}

// Target represents a Bazel target
type Target struct {
	Rule       string
	Name       string
	Deps       []string
	InputFiles []string // List of input file paths relative to the directory
}

// Package represents a collection of targets
type Package struct {
	Targets      []Target
	FilegroupSrc []string // Source files for filegroup (if input directory provided)
}

const packageTemplate = `# Generated package for remote ex stress test
load(":stress.bzl", "remote_ex_rule")
{{- if .FilegroupSrc}}
filegroup(
    name = "input_files",
    srcs = [
{{- range .FilegroupSrc}}
        "{{.}}",
{{- end}}
    ],
    visibility = ["//visibility:public"],
)
{{- end}}
{{range .Targets}}
{{.Render}}{{end}}
`

const targetTemplate = `{{.Rule}}(
    name = "{{.Name}}",
    usecs = {{.RandomUsecs}},
{{- if .InputFiles}}
{{- if eq (len .InputFiles) 1}}
    input_files = ["{{index .InputFiles 0}}"],
{{- else}}
    input_files = [
{{- range .InputFiles}}
        "{{.}}",
{{- end}}
    ],
{{- end}}
{{- end}}
{{- if .Deps}}
{{- if eq (len .Deps) 1}}
    deps = ["{{index .Deps 0}}"],
{{- else}}
    deps = [
{{- range .Deps}}
        "{{.}}",
{{- end}}
    ],
{{- end}}
{{- end}}
)
`

// Render renders a target to BUILD file format
func (t Target) Render() string {
	tmpl, err := template.New("target").Parse(targetTemplate)
	if err != nil {
		panic(err)
	}

	type TargetData struct {
		Rule        string
		Name        string
		Deps        []string
		InputFiles  []string
		RandomUsecs int
	}

	data := TargetData{
		Rule:        t.Rule,
		Name:        t.Name,
		Deps:        t.Deps,
		InputFiles:  t.InputFiles,
		RandomUsecs: randomInt(1000000),
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// Render renders a package to BUILD file format
func (p Package) Render() string {
	tmpl, err := template.New("package").Parse(packageTemplate)
	if err != nil {
		panic(err)
	}

	type TargetRenderData struct {
		Render string
	}

	type PackageData struct {
		Targets      []TargetRenderData
		FilegroupSrc []string
	}

	renderData := make([]TargetRenderData, len(p.Targets))
	for i, t := range p.Targets {
		renderData[i] = TargetRenderData{
			Render: t.Render(),
		}
	}

	data := PackageData{
		Targets:      renderData,
		FilegroupSrc: p.FilegroupSrc,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func randomInt(max int) int {
	b := make([]byte, 4)
	rand.Read(b)
	val := int(uint32(b[0])|uint32(b[1])<<8|uint32(b[2])<<16|uint32(b[3])<<24) % max
	if val < 0 {
		val = -val
	}
	return val
}

func randomHex() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <shift> <desired_targets> [input_dir]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  shift: power-of-2 degree for linear noise\n")
		fmt.Fprintf(os.Stderr, "  desired_targets: approximate number of targets to generate\n")
		fmt.Fprintf(os.Stderr, "  input_dir: (optional) directory containing pre-generated input files\n")
		os.Exit(1)
	}

	shift, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid shift: %v\n", err)
		os.Exit(1)
	}

	desiredActions, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid desired_targets: %v\n", err)
		os.Exit(1)
	}

	var inputDir string
	var availableFiles []string
	var inputDirLabel string // Path to use in BUILD file
	if len(os.Args) >= 4 {
		inputDir = os.Args[3]

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

		// Resolve the input directory path
		var resolvedDir string
		if filepath.IsAbs(inputDir) {
			resolvedDir = inputDir
		} else {
			// Relative path - resolve relative to workspace root
			if workspaceRoot != "" {
				resolvedDir = filepath.Join(workspaceRoot, inputDir)
			} else {
				// Fallback: use current working directory
				resolvedDir = inputDir
			}
			// Clean the path
			resolvedDir = filepath.Clean(resolvedDir)
		}

		// Read all files from the directory
		entries, err := os.ReadDir(resolvedDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input directory %s: %v\n", resolvedDir, err)
			if !filepath.IsAbs(inputDir) && workspaceRoot != "" {
				fmt.Fprintf(os.Stderr, "  (resolved from workspace root: %s + %s)\n", workspaceRoot, inputDir)
			}
			os.Exit(1)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				availableFiles = append(availableFiles, entry.Name())
			}
		}

		if len(availableFiles) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: No files found in input directory %s\n", resolvedDir)
		} else {
			fmt.Fprintf(os.Stderr, "Found %d input files in %s\n", len(availableFiles), resolvedDir)
		}

		// Store the path to use in BUILD file
		// For BUILD files, we want to use the original path (relative to workspace root)
		// not the resolved absolute path
		if filepath.IsAbs(inputDir) {
			// Absolute path - use as-is
			inputDirLabel = inputDir
		} else {
			// Relative path - use the cleaned original path (not the resolved absolute one)
			inputDirLabel = filepath.Clean(inputDir)
			if inputDirLabel == "." {
				inputDirLabel = ""
			}
		}
	}

	noise := NewLinearNoise(shift)
	noiseSum := noise.Sum()
	noiseSize := noise.Size()

	noiseFactor := float64(desiredActions-noiseSize) / noiseSum

	var levels [][]Target
	values := noise.Values()

	for _, point := range values {
		width := int(noiseFactor*point) + 1
		level := make([]Target, width)
		for i := 0; i < width; i++ {
			level[i] = Target{
				Rule:       "remote_ex_rule",
				Name:       randomHex(),
				Deps:       []string{},
				InputFiles: []string{}, // Will be set below
			}
		}
		levels = append(levels, level)
	}

	// Flatten all targets first
	var allTargets []Target
	for _, level := range levels {
		allTargets = append(allTargets, level...)
	}

	// Assign input files to targets
	if len(availableFiles) > 0 && inputDirLabel != "" {
		for i := range allTargets {
			// Randomly decide how many input files this target should have (1 to 3, or up to available)
			maxFiles := 3
			if len(availableFiles) < maxFiles {
				maxFiles = len(availableFiles)
			}
			numFiles := randomInt(maxFiles) + 1

			// Randomly select files (without replacement within the same target to avoid duplicates)
			// Use a map to track selected files for this target
			selectedIndices := make(map[int]bool)
			for len(selectedIndices) < numFiles {
				fileIdx := randomInt(len(availableFiles))
				selectedIndices[fileIdx] = true
			}

			// Build the list of selected file paths
			selectedFiles := make([]string, 0, numFiles)
			for fileIdx := range selectedIndices {
				fileName := availableFiles[fileIdx]
				// Create file path for Bazel
				// If inputDirLabel is empty, file is in current directory
				// Otherwise, use the directory path + filename
				var filePath string
				if inputDirLabel == "" || inputDirLabel == "." {
					filePath = fileName
				} else if filepath.IsAbs(inputDirLabel) {
					// For absolute paths, use the full path
					// Note: Bazel may require these files to be accessible at build time
					// User may need to set up filegroups or ensure files are in the sandbox
					filePath = filepath.Join(inputDirLabel, fileName)
					// Use forward slashes for consistency
					filePath = filepath.ToSlash(filePath)
				} else {
					// Relative path - join directory and filename
					filePath = filepath.Join(inputDirLabel, fileName)
					// Use forward slashes for Bazel (works on all platforms)
					filePath = filepath.ToSlash(filePath)
				}
				selectedFiles = append(selectedFiles, filePath)
			}
			allTargets[i].InputFiles = selectedFiles
		}
	}

	// Store InputFiles assignments by target name before adding dependencies
	inputFilesMap := make(map[string][]string)
	for i := range allTargets {
		if len(allTargets[i].InputFiles) > 0 {
			inputFilesMap[allTargets[i].Name] = allTargets[i].InputFiles
		}
	}

	// Add dependencies: each target in level i depends on one random target from level i-1
	for i := 1; i < len(levels); i++ {
		prevLevel := levels[i-1]
		for j := range levels[i] {
			// Pick a random target from previous level
			randomIdx := randomInt(len(prevLevel))
			depName := ":" + prevLevel[randomIdx].Name
			levels[i][j].Deps = []string{depName}
		}
	}

	// Regenerate allTargets with updated dependencies, preserving InputFiles assignments
	allTargets = []Target{}
	for _, level := range levels {
		for _, target := range level {
			// Restore InputFiles if they were assigned
			if inputFiles, ok := inputFilesMap[target.Name]; ok {
				target.InputFiles = inputFiles
			}
			allTargets = append(allTargets, target)
		}
	}

	// Create filegroup if input files were provided and using relative paths
	// For absolute paths, we skip filegroup (user needs to handle file access)
	var filegroupSrc []string
	if len(availableFiles) > 0 && inputDirLabel != "" && !filepath.IsAbs(inputDirLabel) {
		// Add all files to filegroup sources
		for _, fileName := range availableFiles {
			if inputDirLabel == "." || inputDirLabel == "" {
				filegroupSrc = append(filegroupSrc, fileName)
			} else {
				filePath := filepath.ToSlash(filepath.Join(inputDirLabel, fileName))
				filegroupSrc = append(filegroupSrc, filePath)
			}
		}
	}

	pkg := Package{
		Targets:      allTargets,
		FilegroupSrc: filegroupSrc,
	}
	fmt.Print(pkg.Render())
}
