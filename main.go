package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func CommandHelpMessage(f *flag.FlagSet) func() {
	return func() {
		fmt.Printf("Usage: %s %s [flags]\n", filepath.Base(os.Args[0]), f.Name())
		fmt.Println("Flags:")
		f.PrintDefaults()
	}
}

func main() {
	// backup command && its flags
	backupCmd := flag.NewFlagSet("backup", flag.ExitOnError)
	inFile := backupCmd.String("in", "", "Point to target file (REQUIRED)")
	outName := backupCmd.String("out", "", "Point to destination file (REQUIRED)")
	keyStr := backupCmd.String("key", "", "Use provided 32-byte hex key for encryption (enables secure mode)")

	configFile := backupCmd.String("config", "", "Path to JSON config file with batch backup jobs")

	// custom help message for backup
	backupCmd.Usage = CommandHelpMessage(backupCmd)

	// restore command && its flags
	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	restoreIn := restoreCmd.String("in", "", "Path to the backup file (REQUIRED)")
	restoreOut := restoreCmd.String("out", "", "Path to restore the file to (REQUIRED)")
	restoreKey := restoreCmd.String("key", "", "32-byte hex key for decryption (required for .enc files)")

	// custom help message for restore
	restoreCmd.Usage = CommandHelpMessage(restoreCmd)

	// standard help message
	flag.Usage = func() {
		binName := filepath.Base(os.Args[0])
		fmt.Printf("%s CLI — Utility that stores your config backups\n", binName)
		fmt.Printf("\nUsage:\n  %s [command] [flags]\n", binName)
		fmt.Println("\nCommands:\n  keygen       Generate a new 256-bit encryption key\n  backup       Backup an existing config file\n  restore      Restore a file from backup")
		fmt.Printf("\nUse \"%s [command] -h\" for more information about a command.\n", binName)
	}

	// If user didn't enter any command
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	// Check which command the user entered first
	switch os.Args[1] {
	case "keygen":
		keyHandler := NewKeyHandler()

		// Generate 32 random bytes
		key, err := keyHandler.GenerateKey()
		if err != nil {
			fmt.Println("Error generating key:", err)
			return
		}

		// Encoding key to HEX-string
		encodedKey := keyHandler.EncodeKey(key)
		fmt.Printf("Your Master Key (Save it securely!):\n%s\n", encodedKey)

	case "backup":
		// Parse flags that come after the backup word
		backupCmd.Parse(os.Args[2:])

		// Validate: either (-in and -out) or (-config) must be provided, not both
		hasInOut := *inFile != "" || *outName != ""
		hasConfig := *configFile != ""

		if !hasInOut && !hasConfig {
			fmt.Println("Error: (-in and -out) or (-config) flags are required.")
			backupCmd.Usage()
			return
		}

		if hasInOut && hasConfig {
			fmt.Println("Error: (-in and -out) and (-config) flags cannot be used together.")
			backupCmd.Usage()
			return
		}

		// Build the list of backup jobs
		var jobs []BackupJob

		if hasConfig {
			var err error
			jobs, err = ParseConfigFile(*configFile)
			if err != nil {
				fmt.Println("Error while parsing backup jobs config file:", err)
				return
			}
		} else {
			if *inFile == "" || *outName == "" {
				fmt.Println("Error: both -in and -out flags are required.")
				backupCmd.Usage()
				return
			}

			storageType := "archive"
			if *keyStr != "" {
				storageType = "secure"
			}

			jobs = []BackupJob{
				{
					SourceData:  *inFile,
					BackupDest:  *outName,
					StorageType: storageType,
				},
			}
		}

		// Create engine and register storages
		engine := NewSyncEngine()

		// Always register ArchiveStorage as the default
		engine.RegisterStorage(&ArchiveStorage{backupDir: "."})

		// Additionally register SecureStorage if a key is provided
		if *keyStr != "" {
			keyHandler := NewKeyHandler()
			key, err := keyHandler.DecodeKey(*keyStr)
			if err != nil {
				fmt.Println("Error decoding key:", err)
				return
			}
			engine.RegisterStorage(&SecureStorage{backupDir: ".", secretKey: key})
		}

		// Calculate optimal worker count: no more than CPU cores or job count
		numWorkers := min(len(jobs), runtime.NumCPU())

		engine.ParallelBackup(jobs, numWorkers)

	case "restore":
		restoreCmd.Parse(os.Args[2:])

		if *restoreIn == "" || *restoreOut == "" {
			fmt.Println("Error: -in and -out flags are required.")
			restoreCmd.Usage()
			return
		}

		// Determine storage type by file extension
		storageType := "archive"
		if filepath.Ext(*restoreIn) == ".enc" {
			storageType = "secure"
		}

		// Extract directory and filename from -in path
		backupDir := filepath.Dir(*restoreIn)
		backupName := filepath.Base(*restoreIn)

		fmt.Printf("Restoring from: %s -> %s using [%s] storage\n", *restoreIn, *restoreOut, storageType)

		engine := NewSyncEngine()

		engine.RegisterStorage(&ArchiveStorage{backupDir: backupDir})

		if storageType == "secure" {
			if *restoreKey == "" {
				fmt.Println("Error: -key flag is required for encrypted backups.")
				restoreCmd.Usage()
				return
			}
			keyHandler := NewKeyHandler()
			key, err := keyHandler.DecodeKey(*restoreKey)
			if err != nil {
				fmt.Println("Error decoding key:", err)
				return
			}
			engine.RegisterStorage(&SecureStorage{backupDir: backupDir, secretKey: key})
		}

		if err := engine.Restore(backupName, *restoreOut, storageType); err != nil {
			fmt.Println("Error during restore:", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		flag.Usage()
	}
}
