package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// RunConfigMaker launches the interactive config builder.
// It asks the user for backup jobs and saves the result as a JSON file.
func RunConfigMaker() {
	scanner := bufio.NewScanner(os.Stdin)

	var jobs []BackupJob
	fmt.Println("⚙️  ConfigFlow — Interactive Configuration Maker")
	fmt.Println("================================================")
	fmt.Println("This wizard will help you create a JSON config file for batch backups.")
	fmt.Println()

	jobNum := 1
	for {
		fmt.Printf("--- Job #%d ---\n", jobNum)

		// Source path
		fmt.Print("Source file path: ")
		scanner.Scan()
		sourcePath := strings.TrimSpace(scanner.Text())
		if sourcePath == "" {
			fmt.Println("Error: source path cannot be empty.")
			continue
		}

		// Backup destination
		fmt.Print("Backup destination name: ")
		scanner.Scan()
		backupDest := strings.TrimSpace(scanner.Text())
		if backupDest == "" {
			fmt.Println("Error: backup destination cannot be empty.")
			continue
		}

		// Storage type selection
		fmt.Print("Storage type (archive/secure) [archive]: ")
		scanner.Scan()
		storageType := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if storageType == "" {
			storageType = "archive"
		}
		if storageType != "archive" && storageType != "secure" {
			fmt.Printf("Error: unknown storage type %q. Must be \"archive\" or \"secure\".\n", storageType)
			continue
		}

		jobs = append(jobs, BackupJob{
			SourceData:  sourcePath,
			BackupDest:  backupDest,
			StorageType: storageType,
		})

		jobNum++

		// Ask to add more
		fmt.Print("\nAdd another job? (y/n) [n]: ")
		scanner.Scan()
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			break
		}
		fmt.Println()
	}

	fmt.Println("\n================================================")

	// Output file name
	fmt.Print("Save config as [backup.json]: ")
	scanner.Scan()
	fileName := strings.TrimSpace(scanner.Text())
	if fileName == "" {
		fileName = "backup.json"
	}

	// Ensure .json extension
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	// Serialize to JSON
	jsonBytes, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		fmt.Println("Error generating JSON:", err)
		return
	}

	// Write to file
	if err := os.WriteFile(fileName, jsonBytes, 0644); err != nil {
		fmt.Println("Error saving config file:", err)
		return
	}

	fmt.Printf("✅ Config saved to %s (%d job(s))\n", fileName, len(jobs))
}
