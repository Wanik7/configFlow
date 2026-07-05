# configFlow CLI (backup-cli)

configFlow CLI is a lightweight, efficient command-line utility for securely backing up configuration files. It supports standard gzip archiving and strong AES-256 CTR encryption for sensitive data.

## Features

- **Archive Storage**: Automatically compress your configuration files into `gzip` format to save space.
- **Secure Storage**: Encrypt your configuration files using AES-256 CTR encryption with a custom generated master key.
- **Key Generation**: Built-in 256-bit encryption key generator.
- **Modular Storage Engine**: Easy to extend with new storage mechanisms in the future thanks to the `Storage` interface pattern.

## Installation

Ensure you have [Go](https://go.dev/) installed, then navigate to the project directory and build the executable:

```bash
cd backup-cli
# On Linux / macOS:
go build -o configflow

# On Windows:
go build -o configflow.exe
```

## Usage

```bash
# On Linux / macOS
./configflow [command] [flags]

# On Windows
.\configflow.exe [command] [flags]
```

### Commands

#### 1. Generate an Encryption Key (`keygen`)
Generates a secure 256-bit (32-byte) key used for encrypting your backups.

*Flags:*
- `-t`: Print the generated key to the terminal in HEX format.
- `-out`: Save the generated key directly to a specified file.

**Examples:**
```bash
# Print key to terminal
configflow keygen -t

# Save key to a file
configflow keygen -out ./secret.key
```

#### 2. Backup a File (`backup`)
Backs up an existing configuration file into a `./backups` directory. 
If an encryption key is provided, it uses **Secure Storage** (AES-256 CTR encryption). If no key is provided, it defaults to **Archive Storage** (Gzip compression).

*Flags:*
- `-in`: Path to the source file you want to backup **(Required)**.
- `-out`: Name of the destination file inside the `./backups` directory **(Required)**.
- `-key`: A 32-byte (64 hex characters) key for encryption. Enables secure mode.

**Examples:**

*Archive Mode (Gzip):*
```bash
configflow backup -in ./config.json -out my-config
# Result: Creates ./backups/my-config.gz
```

*Secure Mode (AES-256 Encryption):*
```bash
# Assuming you generated a key using 'keygen -t'
configflow backup -in ./config.json -out my-secure-config -key <your_64_character_hex_key>
# Result: Creates ./backups/my-secure-config.enc
```

## Project Structure

- `main.go`: Entry point, parses CLI arguments and flags.
- `engine.go`: Contains `SyncEngine`, which orchestrates the backup process using registered storages.
- `storage.go`: Implementations of the `Storage` interface (`ArchiveStorage`, `SecureStorage`).
- `utils.go`: Helper functions for key generation and resource closing.

## License

This project is open-source and available under the MIT License.
