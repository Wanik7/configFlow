package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

func ComandHelpMessage(f *flag.FlagSet) func() {
	return func() {
		fmt.Printf("Usage: configflow %s [flags]\n", f.Name())
		fmt.Println("\nFlags:")
		f.PrintDefaults()
	}
}

func main() {
	// keygen command && its flags
	keygenCmd := flag.NewFlagSet("keygen", flag.ExitOnError)
	toTerminal := keygenCmd.Bool("t", false, "Print key into terminal")
	keyDest := keygenCmd.String("out", "", "Point to key file")

	// custom help massage for keygen
	keygenCmd.Usage = ComandHelpMessage(keygenCmd)

	// backup command && its flags
	backupCmd := flag.NewFlagSet("backup", flag.ExitOnError)
	inFile := backupCmd.String("in", "", "Point to target file (REQUIRED)")
	outName := backupCmd.String("out", "", "Point to destination file (REQUIRED)")
	keyStr := backupCmd.String("key", "", "Use provided 32-byte hex key for encryption (enables secure mode)")

	// custom help massage for backup
	backupCmd.Usage = ComandHelpMessage(backupCmd)

	// standart help massage
	flag.Usage = func() {
		fmt.Println("configFlow CLI — Utility that stores your config backups")
		fmt.Println("\nUsage:\n  configflow [command] [flags]")
		fmt.Println("\nCommands:\n  keygen       Generate a new 256-bit encryption key\n  backup       Backup an existing config file")
		fmt.Println("\nUse \"configflow [command] -h\" for more information about a command.")
	}

	// If user didn't enter any command
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	// Check which command the user entered first
	switch os.Args[1] {
	case "keygen":
		// Parse flags that come after the keygen word
		keygenCmd.Parse(os.Args[2:])

		// KEYGEN LOGIC
		if *toTerminal {
			// Generate 32 random bytes
			key, err := GenerateKey()
			if err != nil {
				fmt.Println("Error generating key:", err)
				return
			}
			// Print key in terminal in beautiful HEX format (%x)
			fmt.Printf("Your Master Key (Save it securely!):\n%x\n", key)
		} else {
			if *keyDest == "" {
				fmt.Println("Error: -out flag is required.")
				keygenCmd.Usage()
				return
			}
			// Create new 32 byte key
			key, err := GenerateKey()
			if err != nil {
				fmt.Println("Error generating key:", err)
				return
			}
			// Write key to file
			err = os.WriteFile(*keyDest, key, 0600)
			if err != nil {
				fmt.Println("Error writing key file:", err)
				return
			}
			fmt.Printf("Key written to: %s\n", *keyDest)
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

		var key [32]byte
		if *keyStr != "" {
			// Decoding HEX-string into bytes
			decodedKey, err := hex.DecodeString(*keyStr)
			if err != nil {
				fmt.Println("Error: invalid hex key provided")
				return
			}
			// Check for correct length
			if len(decodedKey) != 32 {
				fmt.Println("Error: key must be exactly 32 bytes (64 hex characters)")
				return
			}
			// Copy correct bytes into our array
			copy(key[:], decodedKey)
		}

		engine.RegisterStorage(&SecureStorage{backupDir: backupDir, secretKey: key})

		if err := engine.Backup(*inFile, *outName, storageType); err != nil {
			fmt.Println("Error during backup:", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		flag.Usage()
	}
}
