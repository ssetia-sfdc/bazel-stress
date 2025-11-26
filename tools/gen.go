package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os"
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
	Rule      string
	Name      string
	Deps      []string
	InputFile string
}

// Package represents a collection of targets
type Package struct {
	Targets    []Target
	InputFiles []string
}

const packageTemplate = `# Generated package for remote ex stress test
load(":stress.bzl", "remote_ex_rule")
{{range .InputFiles}}
{{.}}{{end}}
{{range .Targets}}
{{.Render}}{{end}}
`

const targetTemplate = `{{.Rule}}(
    name = "{{.Name}}",
    usecs = {{.RandomUsecs}},
{{- if .InputFile}}
    input_file = ":{{.InputFile}}",
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
		InputFile   string
		RandomUsecs int
	}

	data := TargetData{
		Rule:        t.Rule,
		Name:        t.Name,
		Deps:        t.Deps,
		InputFile:   t.InputFile,
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
		Targets    []TargetRenderData
		InputFiles []string
	}

	renderData := make([]TargetRenderData, len(p.Targets))
	for i, t := range p.Targets {
		renderData[i] = TargetRenderData{
			Render: t.Render(),
		}
	}

	data := PackageData{
		Targets:    renderData,
		InputFiles: p.InputFiles,
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

// generateInputFileContent generates random lines of text for an input file
func generateInputFileContent(numLines int) string {
	var builder bytes.Buffer
	for i := 0; i < numLines; i++ {
		// Generate a random line (20-80 characters)
		lineLen := randomInt(60) + 20
		for j := 0; j < lineLen; j++ {
			// Random printable ASCII character (avoid problematic chars for shell)
			char := randomInt(62)
			if char < 26 {
				builder.WriteByte(byte('a' + char))
			} else if char < 52 {
				builder.WriteByte(byte('A' + char - 26))
			} else {
				builder.WriteByte(byte('0' + char - 52))
			}
		}
		builder.WriteString("\\n")
	}
	return builder.String()
}

// generateGenruleForInputFile generates a Bazel genrule target for an input file
func generateGenruleForInputFile(name string, content string) string {
	// Escape the content for Bazel (replace newlines and quotes)
	escapedContent := ""
	for _, char := range content {
		switch char {
		case '\n':
			escapedContent += "\\n"
		case '"':
			escapedContent += "\\\""
		case '\\':
			escapedContent += "\\\\"
		default:
			escapedContent += string(char)
		}
	}

	return fmt.Sprintf(`genrule(
    name = "%s",
    outs = ["%s.txt"],
    cmd = "echo -e \"%s\" > $@",
)

`, name, name, escapedContent)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <shift> <desired_targets>\n", os.Args[0])
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
				Rule:      "remote_ex_rule",
				Name:      randomHex(),
				Deps:      []string{},
				InputFile: "", // Will be set below
			}
		}
		levels = append(levels, level)
	}

	// Generate input files for all targets
	var inputFileNames []string
	inputFileMap := make(map[string]string) // Maps target name to input file name

	// Flatten all targets first to generate input files
	var allTargets []Target
	for _, level := range levels {
		allTargets = append(allTargets, level...)
	}

	// Generate a unique input file for each target
	for i := range allTargets {
		inputFileName := "input_" + allTargets[i].Name
		inputFileNames = append(inputFileNames, inputFileName)
		inputFileMap[allTargets[i].Name] = inputFileName
		allTargets[i].InputFile = inputFileName
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

	// Regenerate allTargets with updated dependencies, preserving InputFile assignments
	allTargets = []Target{}
	for _, level := range levels {
		for _, target := range level {
			// Restore InputFile from the map
			if inputFile, ok := inputFileMap[target.Name]; ok {
				target.InputFile = inputFile
			}
			allTargets = append(allTargets, target)
		}
	}

	// Generate genrule targets for input files
	var inputFileRules []string
	for _, inputFileName := range inputFileNames {
		// Generate random content (10-50 lines per file)
		numLines := randomInt(40) + 10
		content := generateInputFileContent(numLines)
		rule := generateGenruleForInputFile(inputFileName, content)
		inputFileRules = append(inputFileRules, rule)
	}

	pkg := Package{
		Targets:    allTargets,
		InputFiles: inputFileRules,
	}
	fmt.Print(pkg.Render())
}

