package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// CallAlgorithm calls the algorithm with the needed arguments.
func CallAlgorithm(algorithmPath, dataFile, outputFile, parameterFile string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("python", algorithmPath, dataFile, outputFile, parameterFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error while executing algorithm: %s", stderr.String())
	}
	return nil
}
