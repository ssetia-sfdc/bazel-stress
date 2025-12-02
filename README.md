Generates a desired number of actions according to a linear noise
shape profile. The generated actions consume a deterministic amount
of cpu time and depend on zero or one targets in order to establish
their position in the overall organization of the build graph.

## Usage

### Generate BUILD file

```bash
bazel run //tools/gen:gen <shift> <desired_targets> [input_dir] > BUILD
```

**Parameters:**
- `shift`: A power-of-2 degree of the linear noise generated. The higher
  the `shift`, the more levels to the graph. For example:
  - `shift=2` creates 2² = 4 levels
  - `shift=5` creates 2⁵ = 32 levels
- `desired_targets`: A total count of targets that should be generated.
  This is desired, not exact, and may shift slightly to ensure no empty
  levels.
- `input_dir`: (Optional) Directory containing input files.
  If provided, files from this directory (single level only, subdirectories are not scanned) will be randomly selected and assigned to actions
  (some actions get 1 file, some get multiple files). Files may be reused across different actions.
  If not provided, actions will have no input files.
  
  **Path handling:**
  - Relative paths should be relative to the workspace root (where the BUILD file is generated)
  - Absolute paths are supported, but Bazel may require additional setup to access files outside the workspace
  - Only files directly in the specified directory are considered (subdirectories are ignored)
  - Recommended: Use relative paths within the workspace (e.g., `input_files` instead of `/tmp/input_files`)

**Example:**
```bash
# Generate ~100 targets with 8 levels, using pre-generated input files
bazel run //tools/gen:gen 3 100 input_files > BUILD

# Generate ~1000 targets with 32 levels, using pre-generated input files
bazel run //tools/gen:gen 5 1000 input_files > BUILD

# Generate targets without input files (faster generation, but no file I/O in stress test)
bazel run //tools/gen:gen 5 1000 > BUILD
```

### Run the stress test

```bash
bazel build :all
```

This builds all generated targets. Each target:
- Runs a CPU-bound task (busy-wait) for a random duration
- Reads random lines from one or more input files (file I/O operations)
- Writes selected lines to an output file
- Has dependencies forming a build graph
- Tests Bazel's parallel build capabilities with realistic file operations

**Note:** If input files were specified during BUILD file generation, each action will be randomly assigned 1-3 input files from the provided directory. Only files directly in the directory are considered (subdirectories are not scanned). Files may be reused across different actions.

**Goal:** Shortest time for a given parameter set wins.

Feel free to use any Bazel options you want to improve your score:
```bash
bazel build :all --jobs=8  # Control parallelism
bazel build :all --remote_cache=...  # Use remote caching
```

## Implementation

This project is implemented in Go and uses Bazel's Bzlmod for dependency management.

**Main components:**
- `tools/gen/gen.go`: Generates BUILD files using a linear noise algorithm, randomly selecting files from input directories (single level only)
- `tools/ex/ex.go`: Executable that busy-waits, reads random lines from one or more input files, and writes to output files
- `stress.bzl`: Bazel rule definition for the stress test targets with support for multiple input files per action
- `MODULE.bazel`: Bzlmod configuration (Bazel 8+)

## Requirements

- Bazel 8.0 or later (uses Bzlmod by default)
- Go toolchain (managed by Bazel via rules_go)
