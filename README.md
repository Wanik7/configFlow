# configFlow CLI (backup-cli)

configFlow CLI is a lightweight, efficient command-line utility for securely backing up and restoring configuration files. It supports standard gzip archiving and strong AES-256 CTR encryption for sensitive data.

## Features

- **Archive Storage**: Automatically compress your configuration files into `gzip` format to save space.
- **Secure Storage**: Encrypt your configuration files using AES-256 CTR encryption with a custom generated master key.
- **Key Generation**: Built-in 256-bit encryption key generator.
- **Restore**: Restore your backed up files — both archived and encrypted.
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
Generates a secure 256-bit (32-byte) key and prints it to the terminal in HEX format.

**Example:**
```bash
configflow keygen
# Output:
# Your Master Key (Save it securely!):
# a1b2c3d4e5f6...
```

#### 2. Backup a File (`backup`)
Backs up an existing configuration file.
If an encryption key is provided, it uses **Secure Storage** (AES-256 CTR encryption). If no key is provided, it defaults to **Archive Storage** (Gzip compression).

*Flags:*
- `-in`: Path to the source file you want to backup **(Required)**.
- `-out`: Path to the destination backup file. The directory will be created automatically **(Required)**.
- `-key`: A 32-byte (64 hex characters) key for encryption. Enables secure mode.

**Examples:**

*Archive Mode (Gzip):*
```bash
configflow backup -in ./config.json -out ./backups/my-config
# Result: Creates ./backups/my-config.gz
```

*Secure Mode (AES-256 Encryption):*
```bash
configflow backup -in ./config.json -out ./backups/my-secure-config -key <your_64_character_hex_key>
# Result: Creates ./backups/my-secure-config.enc
```

#### 3. Restore a File (`restore`)
Restores a previously backed up file. The storage type is detected automatically by file extension (`.enc` → secure, `.gz` → archive).

*Flags:*
- `-in`: Path to the backup file **(Required)**.
- `-out`: Path to restore the file to **(Required)**.
- `-key`: A 32-byte (64 hex characters) key for decryption (required for `.enc` files).

**Examples:**

*Restore from Archive:*
```bash
configflow restore -in ./backups/my-config.gz -out ./restored-config.json
```

*Restore from Encrypted Backup:*
```bash
configflow restore -in ./backups/my-secure-config.enc -out ./restored-config.json -key <your_64_character_hex_key>
```

## Project Structure

- `main.go`: Entry point, parses CLI arguments and flags.
- `engine.go`: Contains `SyncEngine`, which orchestrates the backup and restore processes using registered storages.
- `storage.go`: Implementations of the `Storage` interface (`ArchiveStorage`, `SecureStorage`).
- `key-handler.go`: Key generation, encoding and decoding utilities.

## License

This project is open-source and available under the MIT License.
