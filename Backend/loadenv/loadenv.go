package loadenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadEnv(filePath string) error {
	candidates := []string{}
	if filePath != "" {
		candidates = append(candidates, filePath)
	} else {
		candidates = append(candidates, ".env", "Backend/.env")
	}

	var lastErr error
	var file *os.File
	for _, path := range candidates {
		f, err := os.Open(path)
		if err != nil {
			lastErr = err
			continue
		}
		file = f
		break
	}
	if file == nil {
		return fmt.Errorf("could not open .env file: %v", lastErr)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Skipping invalid line: %s\n", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		// Don't overwrite vars already set in the process environment
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			fmt.Printf("Error setting environment variable: %s=%s\n", key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	return nil
}
