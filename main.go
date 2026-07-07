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
		fmt.Println("\nFlags:")
		f.PrintDefaults()
	}
}

func CreateDestDir(backupDir string, permissions os.FileMode) error {
	if err := os.MkdirAll(backupDir, permissions); err != nil {
		return fmt.Errorf("could not create backup directory: %w", err)
	}
	return nil
}

func main() {
	// keygen command && its flags
	keygenCmd := flag.NewFlagSet("keygen", flag.ExitOnError)
	toTerminal := keygenCmd.Bool("t", false, "Print key into terminal")
	keyDest := keygenCmd.String("out", "", "Point to key file")

	// custom help message for keygen
	keygenCmd.Usage = CommandHelpMessage(keygenCmd)

	// backup command && its flags
	backupCmd := flag.NewFlagSet("backup", flag.ExitOnError)
	inFile := backupCmd.String("in", "", "Point to target file (REQUIRED)")
	outName := backupCmd.String("out", "", "Point to destination file (REQUIRED)")
	keyStr := backupCmd.String("key", "", "Use provided 32-byte hex key for encryption (enables secure mode)")

	// custom help message for backup
	backupCmd.Usage = CommandHelpMessage(backupCmd)

	// standard help message
	flag.Usage = func() {
		binName := filepath.Base(os.Args[0])
		fmt.Printf("%s CLI — Utility that stores your config backups\n", binName)
		fmt.Printf("\nUsage:\n  %s [command] [flags]\n", binName)
		fmt.Println("\nCommands:\n  keygen       Generate a new 256-bit encryption key\n  backup       Backup an existing config file")
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
		// Parse flags that come after the keygen word
		keygenCmd.Parse(os.Args[2:])

		// Generate 32 random bytes
		key, err := keyHandler.GenerateKey()
		if err != nil {
			fmt.Println("Error generating key:", err)
			return
		}

		// KEYGEN LOGIC
		if *toTerminal {
			// Encoding key to HEX-string
			encodedKey := keyHandler.EncodeKey(key)
			fmt.Printf("Your Master Key (Save it securely!):\n%s\n", encodedKey)
		} else {
			if *keyDest == "" {
				fmt.Println("Error: -out flag is required.")
				keygenCmd.Usage()
				return
			}
			const keyPermissions = 0600
			const keyDir = "./backupKeys/"

			if err := CreateDestDir(keyDir, keyPermissions); err != nil {
				fmt.Println("Error creating key file:", err)
				return
			}
			// Writing key to file
			if err := os.WriteFile(keyDir+*keyDest+".key", key, keyPermissions); err != nil {
				fmt.Println("Error writing key file:", err)
				return
			}
			fmt.Printf("Key written to: %s\n", keyDir+*keyDest+".key")
		}

	case "backup":
		const backupDir = "./backups"
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			fmt.Println("Error creating backup directory:", err)
			return
		}

		// Parse flags that come after the backup word
		backupCmd.Parse(os.Args[2:])

		// Checking required flags
		if *inFile == "" || *outName == "" {
			fmt.Println("Error: -in and -out flags are required.")
			backupCmd.Usage()
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

		if err := engine.Backup(*inFile, *outName, storageType); err != nil {
			fmt.Println("Error during backup:", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		flag.Usage()
	}
}
