# Tracewrap

**NOTE!** This is a work-in-progress! I use it for my own development, but I haven't done much cross-platform standardization: I mostly work in linux.

This is public due to how it works--it relies on pulling down the latest public release for instrumenting my development applications. You're welcome to poke around and see what's under the hood, but as a solo engineer, **I'm currently supporting it for how I use it.**

_If this project interests you let me know: get in touch, star it, follow it, whatever. I'll gauge any interest, reevaluate._

---

**Tracewrap** is a Go tool that automatically instruments your applications for comprehensive tracing and performance analysis. With Tracewrap, you can gain deep insights into your application's function calls, memory usage, and execution times—all without changing a single line of your project’s code.

## Key Features

- **No-Code Instrumentation:**  
  Instrument your application by simply adding a `tracewrap.yaml` configuration file to your project directory. This means you don’t need to alter your source code to benefit from detailed tracing.  
  - **Why It Matters:**  
    - **Non-Invasive:** No need to sprinkle tracing code throughout your application.  
    - **Ease of Adoption:** Quickly integrate tracing into existing projects without refactoring or risking unintended side effects.  
    - **Universal Application:** Works with ~~any type of~~ some Go projects, regardless of structure, making it a simple tool that offers full instrumentation via configuration alone.
  
- **Function Call Instrumentation:**  
  Automatically wrap function calls to track entry, exit, parameters, return values, and performance metrics.
  
- **Call Graph Generation:**  
  Produces a `.dot` file that represents your application’s call graph. Convert it to a visual diagram with Graphviz to better understand your application's structure.
  
- **Flexible Configuration:**  
  [TO DO] Customize which files or functions are traced, adjust logging levels, and set output options using a simple YAML file.

- **Multiple Examples:**  
  Several self-contained example projects demonstrate different scenarios—simple math operations, concurrency, recursion, and more.

---

## Installation & Building

1. **Clone the Repository**  
   ```bash
   git clone https://github.com/mwiater/tracewrap.git
   cd tracewrap
   ```

2. **Build the Tracewrap Binary**  
   ```bash
   go build -o bin/tracewrap ./cmd/tracewrap
   ```
   This command compiles Tracewrap into an executable placed in the `bin/` directory.

---

## Basic Usage

Tracewrap’s primary strength is its ability to instrument your project without any code changes—simply add a `tracewrap.yaml` file. Here’s how to get started:

1. **Navigate to a Project Directory**  
   For example, if you want to instrument the `examples/simple` project:
   ```bash
   cd examples/simple
   ```

2. **Add a tracewrap.yaml**  
   Each example project includes a `tracewrap.yaml` file that configures the instrumentation. This file tells Tracewrap what to trace, which files to include, and where to output logs and call graphs.

3. **Run Tracewrap**  
   From within the example project directory, run:
   ```bash
   ../../bin/tracewrap buildTracedApplication --project . --config tracewrap.yaml
   ```
   - `--project .` instructs Tracewrap to instrument the current directory.
   - `--config tracewrap.yaml` specifies the configuration file.

4. **Inspect Output**  
   Tracewrap creates a `tracewrap` directory containing:
   ```
   tracewrap/
   ├── callgraph.dot
   └── tracewrap.log
   ```

5. **Generate a Visual Call Graph (Optional)**  
   Install Graphviz (e.g., `sudo apt-get install graphviz` on Ubuntu), then convert the `.dot` file to a PNG:
   ```bash
   dot -Tpng tracewrap/callgraph.dot -o tracewrap/callgraph.png
   ```
   Open `tracewrap/callgraph.png` to visualize your application's function call structure.

---

## Example Projects

Our repository includes several self-contained example projects under the `examples/` directory. Each example demonstrates a different use case of Tracewrap’s instrumentation—all without requiring any modifications to the original source code. Here’s a brief overview:

1. **Simple**  
   - **Path:** `examples/simple`  
   - **Description:** Demonstrates basic math functions (addition, subtraction, multiplication, division).  
   - **Focus:** Illustrates basic function calls and error handling with minimal setup.

2. **Concurrency**  
   - **Path:** `examples/concurrency`  
   - **Description:** Showcases concurrent job processing with goroutines and channels.  
   - **Focus:** Tests Tracewrap’s ability to capture parallel function calls and concurrent execution.

3. **Recursive**  
   - **Path:** `examples/recursive`  
   - **Description:** Calculates Fibonacci numbers using a naive recursive approach.  
   - **Focus:** Highlights deep call stacks and recursive function invocations.

### How to Run an Example

1. **Initialize the Module (if not already done):**
   ```bash
   go mod init github.com/mwiater/tracewrap-<example-name>
   ```
2. **Build Tracewrap from the Repository Root:**
   ```bash
   go build -o bin/tracewrap ./cmd/main.go
   ```
3. **Move into the Example Directory:**
   ```bash
   cd examples/<example-name>
   ```
4. **Run Tracewrap:**
   ```bash
   ../../bin/tracewrap buildTracedApplication --name=<example-name> --project . --config tracewrap.yaml
   ```
5. **Generate the Call Graph:**
   ```bash
   dot -Tpng tracewrap/callgraph.dot -o tracewrap/callgraph.png
   ```

Refer to each example’s README.md for additional details and sample outputs.

## Tracewrap Commands

`go run cmd/main.go list commands`

```
  tracewrap                              tracewrap is a tool for building instrumented Go applications.
    tracewrap buildTracedApplication     Build and run an instrumented version of the application
    tracewrap completion                 Generate the autocompletion script for the specified shell
      tracewrap completion bash          Generate the autocompletion script for bash
      tracewrap completion fish          Generate the autocompletion script for fish
      tracewrap completion powershell    Generate the autocompletion script for powershell
      tracewrap completion zsh           Generate the autocompletion script for zsh
    tracewrap generate                   Generate various artifacts for tracewrap.
      tracewrap generate callgraph       Generate a call graph from a tracewrap log file.
      tracewrap generate callgraphImage  Generate a PNG image from a callgraph.dot file.
    tracewrap help                       Help about any command
    tracewrap list                       Group commands for listing resources
      tracewrap list commands            List all available commands and subcommands in two columns

```

## Testing

Only some prelimiary tests at the moment.

First, build the binary: `go build -o bin/tracewrap ./cmd/main.go`

Then, run the tests: `go test -count=1 -v ./...`

## Dev Notes

### Tags

#### Proxy issue

List local tags: `git fetch origin --prune --tags && git tag` 

Or latest tag across all branches: `git describe --tags $(git rev-list --tags --max-count=1)`

Remove local tags: `git tag | xargs git tag -d`

List remote tags: `git ls-remote --tags origin`

Remove remote tags: `git ls-remote --tags origin | awk '{print $2}' | sed 's/refs\/tags\///' | xargs -I {} git push --delete origin {}`

`go list -m -versions github.com/mwiater/tracewrap`

Proxy is not in sync with github: `github.com/mwiater/tracewrap v0.1.0 v0.2.0 v0.3.0 v0.4.0 v0.5.0 v0.6.0 v0.7.0 v0.8.0 v0.9.0 v0.10.0 v0.11.0 v0.12.0 v0.13.0 v0.14.0`

Add to shell (these are necessary for now until I can figuroute how to build the workspace without them):

```
export GONOPROXY=github.com/mwiater/tracewrap
export GONOSUMDB=github.com/mwiater/tracewrap
export GOROOT=$(go env GOROOT)
export GOPATH=$(go env GOPATH)
```


List Versions: `go list -m -versions github.com/mwiater/tracewrap`

Correct: `github.com/mwiater/tracewrap v0.1.0 v0.2.0`

If needed, clean out modcache: `go clean -cache -modcache -i`

---

## License

This project is licensed under the [MIT License](LICENSE).

---
