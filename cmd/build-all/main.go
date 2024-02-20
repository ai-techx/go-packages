package main

import (
	"github.com/google/logger"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// runCommandAll is a command line tool that builds all the packages in the current directory
func runCommandAll(directory string, command string, args ...string) {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				cmd := exec.Command(command, args...)
				cmd.Dir = path

				output, err := cmd.CombinedOutput()

				if err != nil {
					logger.Fatalf("Error building %s: %v\n%s", path, err, output)
				} else {
					logger.Infof("%s command in %s success: %s", command, path, output)
				}
			}
		}
		return nil
	})

	if err != nil {
		logger.Fatalf("Error walking the path %s: %v", directory, err)
	}
}

func main() {
	logger.Init("Logger", true, false, io.Discard)
	logger.Info("Building all packages")
	runCommandAll("pkg", "go", "build", ".")
	logger.Info("Testing all packages")
	runCommandAll("pkg", "go", "generate", "./...")
	runCommandAll("pkg", "go", "test", "./...")
}
