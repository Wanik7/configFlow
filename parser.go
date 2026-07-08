package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BackupJob struct {
	SourceData  string `json:"source_data"`
	BackupDest  string `json:"backup_dest"`
	StorageType string `json:"storage_type"` // archive or secure
}

func ParseConfigFile(filePath string) (jobs []BackupJob, err error) {
	targetFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer CloseSafe(targetFile)

	decoder := json.NewDecoder(targetFile)

	decoder.DisallowUnknownFields()

	err = decoder.Decode(&jobs)

	for i := range jobs {
		cleanPath, err := ResolvePath(jobs[i].SourceData)
		if err != nil {
			return nil, err
		}
		jobs[i].SourceData = cleanPath
	}

	return jobs, err
}

func ResolvePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get home directory: %w", err)
		}

		// Replace ~ via absolute path to home directory
		path = filepath.Join(home, path[1:])
	}

	// Make absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not convert to absolute path: %w", err)
	}

	return absPath, nil
}
