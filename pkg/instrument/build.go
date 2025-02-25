package instrument

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// BuildInstrumentedBinary runs the necessary Go commands ("go mod tidy", "go get", and "go build")
// in the workspace directory to build the instrumented binary. It preserves environment variables.
// It returns the path to the built binary and an error if any command fails.
//
// Parameters:
//   - workspace (string): the path to the workspace directory.
//
// Returns:
//   - string: the path to the built instrumented binary.
//   - error: an error object if any step in the build process fails.
func BuildInstrumentedBinary(workspace string) (string, error) {
	fmt.Println("Running 'go mod tidy' in workspace:", workspace)
	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Dir = workspace
	cmdTidy.Env = os.Environ()
	out, err := cmdTidy.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("go mod tidy failed: %v, output: %s", err, string(out))
	}
	fmt.Println("go mod tidy completed successfully.")

	fmt.Println("Running 'go get github.com/mwiater/tracewrap@latest' in workspace:", workspace)
	cmdGet := exec.Command("go", "get", "github.com/mwiater/tracewrap@latest")
	cmdGet.Dir = workspace
	cmdGet.Env = os.Environ()
	out, err = cmdGet.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get tracewrap repository: %v, output: %s", err, string(out))
	}
	fmt.Println("Tracewrap repository acquired successfully.")

	binaryName := "tracedApp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(workspace, binaryName)
	fmt.Println("Building instrumented binary:", binaryPath)
	cmdBuild := exec.Command("go", "build", "-o", binaryPath)
	cmdBuild.Dir = workspace
	cmdBuild.Env = os.Environ()
	out, err = cmdBuild.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %v, output: %s", err, string(out))
	}
	fmt.Println("Binary built successfully at:", binaryPath)
	return binaryPath, nil
}

// RunInstrumentedBinary executes the built binary located at binaryPath with any additional command-line arguments.
// It sets the standard output and error to the current process's output streams and preserves environment variables.
//
// Parameters:
//   - binaryPath (string): the path to the instrumented binary.
//   - args ([]string): a slice of strings representing additional arguments to pass to the binary.
//
// Returns:
//   - error: an error object if the binary execution fails.
func RunInstrumentedBinary(binaryPath string, args []string) error {
	fmt.Println("Running instrumented binary:", binaryPath, "with args:", args)
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}
