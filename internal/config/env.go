package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	instance *Env
	once     sync.Once
	mu       sync.RWMutex
)

type Env struct {
	GitToken         string `json:"git_token"`
	EnablePagination bool   `json:"enable_pagination"`
	Port             string `json:"port"`
}

// InitEnvFromFile initializes the singleton Env instance from the given file path.
// It only loads the config file once, subsequent calls will not reload the config.
// Returns an error if the file cannot be read or parsed.
func InitEnvFromFile(filePath string) error {
	var loadErr error

	once.Do(func() {
		// Read the file content
		data, err := os.ReadFile(filePath)
		if err != nil {
			loadErr = fmt.Errorf("failed to read config file: %w", err)
			return
		}

		// Unmarshal JSON data into temporary Env struct
		var env Env
		if err := json.Unmarshal(data, &env); err != nil {
			loadErr = fmt.Errorf("failed to parse config file: %w", err)
			return
		}

		// Set the singleton instance with write lock
		mu.Lock()
		instance = &env
		mu.Unlock()
	})

	return loadErr
}

func GetEnv() *Env {
	return instance
}
