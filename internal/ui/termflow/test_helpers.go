package termflow

import (
	"fmt"
	"os"
	"os/exec"
)

// buildTestBinary builds the test binary if it doesn't exist
func buildTestBinary() error {
	// Always build to ensure we have the latest changes
	cmd := exec.Command("go", "build", "-o", "/tmp/rigel-test", "../../../cmd/rigel")
	cmd.Env = append(os.Environ(), "PROVIDER=ollama")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
