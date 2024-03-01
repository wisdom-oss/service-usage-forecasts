package helpers

import "os/exec"

// CallAlgorithm calls the algorithm with the needed arguments.
func CallAlgorithm(algorithmPath, dataFile, outputFile, parameterFile string) error {
	cmd := exec.Command(algorithmPath, dataFile, outputFile, parameterFile)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
