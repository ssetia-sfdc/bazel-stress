Generates a desired number of actions according to a linear noise
shape profile. The generated actions consume a deterministic amount
of cpu time and depend on zero or one targets in order to establish
their position in the overall organization of the build graph.

## Usage

### Generate BUILD file

```bash
bazel run tools:gen [shift] [desired_targets] > BUILD
```

**Parameters:**
- `shift`: A power-of-2 degree of the linear noise generated. The higher
  the `shift`, the more levels to the graph. For example:
  - `shift=2` creates 2² = 4 levels
  - `shift=5` creates 2⁵ = 32 levels
- `desired_targets`: A total count of targets that should be generated.
  This is desired, not exact, and may shift slightly to ensure no empty
  levels.

**Example:**
```bash
# Generate ~100 targets with 8 levels
bazel run tools:gen 3 100 > BUILD

# Generate ~1000 targets with 32 levels
bazel run tools:gen 5 1000 > BUILD
```

### Run the stress test

```bash
bazel build :all
```

This builds all generated targets. Each target:
- Runs a CPU-bound task (busy-wait) for a random duration
- Has dependencies forming a build graph
- Tests Bazel's parallel build capabilities

**Goal:** Shortest time for a given parameter set wins.

Feel free to use any Bazel options you want to improve your score:
```bash
bazel build :all --jobs=8  # Control parallelism
bazel build :all --remote_cache=...  # Use remote caching
```

## Implementation

This project is implemented in Go and uses Bazel's Bzlmod for dependency management.

**Main components:**
- `tools/gen.go`: Generates BUILD files using a linear noise algorithm
- `tools/ex.go`: Executable that busy-waits for a specified time and creates output files
- `stress.bzl`: Bazel rule definition for the stress test targets
- `MODULE.bazel`: Bzlmod configuration (Bazel 8+)

## Requirements

- Bazel 8.0 or later (uses Bzlmod by default)
- Go toolchain (managed by Bazel via rules_go)
