package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

		// KEYGEN LOGIC

		// Encoding key to HEX-string
		encodedKey := keyHandler.EncodeKey(key)
		fmt.Printf("Your Master Key (Save it securely!):\n%s\n", encodedKey)

	case "backup":
		// Parse flags that come after the backup word
		backupCmd.Parse(os.Args[2:])

		// Checking required flags
		if *inFile == "" || *outName == "" {
			fmt.Println("Error: -in and -out flags are required.")
			backupCmd.Usage()
			return
		}

		// Extract directory and filename from -out path
		backupDir := filepath.Dir(*outName)
		backupName := filepath.Base(*outName)

		if err := os.MkdirAll(backupDir, 0755); err != nil {
			fmt.Println("Error creating backup directory:", err)
			return
		}

		// Determine backup strategy
		storageType := "archive"
		if *keyStr != "" {
			storageType = "secure"
		}

		fmt.Printf("Starting backup for: %s -> saved as: %s using [%s] storage\n", *inFile, *outName, storageType)

		engine := NewSyncEngine()

		engine.RegisterStorage(&ArchiveStorage{backupDir: backupDir})

		if *keyStr != "" {
			keyHandler := NewKeyHandler()
			// Decoding HEX-string into bytes
			key, err := keyHandler.DecodeKey(*keyStr)
			if err != nil {
				fmt.Println("Error decoding key:", err)
				return
			}

			engine.RegisterStorage(&SecureStorage{backupDir: backupDir, secretKey: key})
		}

		if err := engine.Backup(*inFile, backupName, storageType); err != nil {
			fmt.Println("Error during backup:", err)
		}

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
