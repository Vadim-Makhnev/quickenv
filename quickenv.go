package quickenv

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Version of the quickenv package.
const Version = "1.0.0"

// LoadOptions configures how environment variables are loaded.
type LoadOptions struct {
	// Pathname is the path of the env file to load (default: ".env")
	Pathname string

	// Overwrite existing environment variables (default: false)
	Overwrite bool

	// Debug enables debug logging (default: false)
	Debug bool

	// MaxLevels limits how many directories up to search for the env file (default: 3)
	MaxLevels int
}

// DefaultLoadOptions returns the default loading options
func DefaultLoadOptions() *LoadOptions {
	return &LoadOptions{
		Pathname:  ".env",
		Overwrite: false,
		Debug:     false,
		MaxLevels: 3,
	}
}

// Load loads environment variables from the specified file.
// If no pathname is provided, it defaults to ".env" in the current directory.
// Returns the number of variables loaded and any error encountered.
func Load(opts ...*LoadOptions) (int, error) {
	options := parseOptions(opts...)

	filePath, err := findEnvFile(options.Pathname, options.MaxLevels)
	if err != nil {
		return 0, fmt.Errorf("quickenv: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("quickenv: failed to open %s:%w", filePath, err)
	}
	defer file.Close()

	return loadFromReader(file, options)
}

// MustLoad is like Load but panics if an error occurs.
// Useful for initialization in main() functions.
func MustLoad(opts ...*LoadOptions) int {
	count, err := Load(opts...)
	if err != nil {
		panic(fmt.Sprintf("quickenv: %s", err))
	}

	return count
}

// Helper functions

// parseOptions processes the provided LoadOptions and applies default values
// for missing or invalid fields. Always returns a valid *LoadOptions.
//
// If no options are provided, uses DefaultLoadOptions().
// Otherwise, creates a copy of the provided options and ensures:
//   - Pathname defaults to ".env" if empty,
//   - MaxLevels defaults to 3 if <= 0.
func parseOptions(opts ...*LoadOptions) *LoadOptions {
	if len(opts) > 0 && opts[0] != nil {
		result := *opts[0] // Make a copy to avoid modifying the original

		// Set default filename if not specified
		if result.Pathname == "" {
			result.Pathname = ".env"
		}

		// Set default max levels if invalid
		if result.MaxLevels <= 0 {
			result.MaxLevels = 3
		}

		return &result
	}

	// No valid options provided → use defaults
	return DefaultLoadOptions()
}

// findEnvFile looks for a file named pathname starting in the current directory.
// If not found and maxLevels > 0, it searches up to maxLevels levels in parent directories.
// Returns the path on success, or an error if not found.
func findEnvFile(pathname string, maxLevels int) (string, error) {
	// Step 1: Check in the current directory (e.g. /home/user/project/cmd/api/.env)
	if _, err := os.Stat(pathname); err == nil {
		return pathname, nil
	}

	// Step 2: Start from current working directory (e.g. /home/user/project/cmd/api)
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current dir: %w", err)
	}

	// Step 3: Climb up to parent directories, maxLevels times
	// Will perform at most 'maxLevels' steps, and stops early if root is reached.
	// Example: /home/user/project/cmd/api → /home/user/project/cmd → ...
	for range maxLevels {
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root (/ or C:\)
		}
		dir = parent                         // e.g. /home/user/project/cmd
		path := filepath.Join(dir, pathname) // e.g. /home/vadim/project/cmd/.env
		if _, err := os.Stat(path); err == nil {
			return path, nil // found
		}
	}

	return "", fmt.Errorf("env file not found: %s", pathname)
}

// loadFromReader reads environment variables from an io.Reader (e.g. file, buffer).
// Parses each non-empty, non-comment line as KEY=VALUE, optionally with quotes and 'export' prefix.
// Skips invalid lines and logs them if Debug is enabled.
// Only sets a variable if:
//   - Overwrite is true, OR
//   - The variable is not already set in the environment.
//
// Returns the number of successfully loaded variables and any critical read error.
// Parsing errors do not stop execution but are logged when Debug = true.
func loadFromReader(reader io.Reader, options *LoadOptions) (int, error) {
	scanner := bufio.NewScanner(reader)
	loaded := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		key, value, err := parseLine(line)
		if err != nil {
			if options.Debug {
				fmt.Fprintf(os.Stderr, "quickenv: [DEBUG] skip invalid line %q: %v\n", line, err)
			}
			continue
		}

		// Set environment variable
		if options.Overwrite || os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return loaded, fmt.Errorf("failed to set %s: %w", key, err)
			}
			loaded++

			if options.Debug {
				mask := "***"
				if len(value) < 5 {
					mask = strings.Repeat("*", len(value))
				}
				fmt.Fprintf(os.Stderr, "quickenv: [DEBUG] set %s=%s\n", key, mask)
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return loaded, fmt.Errorf("read error: %w", err)
	}
	return loaded, nil
}

// Supports quoted values and the optional "export" prefix.
// Only the first unquoted '=' is treated as delimiter.
// Returns the key, value, and nil error on success.
// Returns empty strings and an error if the line is invalid.
func parseLine(line string) (string, string, error) {
	// Handle export keyword
	line = strings.TrimPrefix(line, "export")

	// Find the first equals sign that's not in quotes
	equalsIndex := -1
	inQuotes := false
	var quoteChar rune

loop:
	for i, char := range line {
		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			}
		case char == '=' && !inQuotes:
			equalsIndex = i
			break loop
		}
	}

	if equalsIndex == -1 {
		return "", "", fmt.Errorf("invalid line format, missing equals sign")
	}

	key := strings.TrimSpace(line[:equalsIndex])
	value := strings.TrimSpace(line[equalsIndex+1:])

	// Validate key
	if key == "" {
		return "", "", fmt.Errorf("empty key")
	}

	if !isValidEnvKey(key) {
		return "", "", fmt.Errorf("invalid key format: %s", key)
	}

	// Remove surrounding quotes from value
	value = unquoteValue(value)

	return key, value, nil
}

// isValidEnvKey checks if a string is a valid environment variable name.
// Rules:
//   - Must not be empty
//   - First character must be a letter or underscore
//   - Subsequent characters may be letters, digits, or underscores
func isValidEnvKey(key string) bool {
	if key == "" {
		return false
	}

	// First character must be letter or underscore
	firstChar := rune(key[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return false
	}

	// Rest can be letters, numbers, or underscores
	for _, char := range key[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}

	return true
}

// unquoteValue strips surrounding single or double quotes if both are present and matching.
// Returns the original string otherwise.
func unquoteValue(value string) string {
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

// GetEnv returns the value of the environmnet variable named by the key.
// It returns the defaultValue if the variable is not present.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvOrPanic returns the value of the environment variable or panics if not set.
func GetEnvOrPanic(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("quickenv: required environment variable %s is not set", key))
}
