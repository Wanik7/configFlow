# configFlow CLI (backup-cli)

configFlow CLI is a lightweight, efficient command-line utility for securely backing up and restoring configuration files. It supports standard gzip archiving and strong AES-256 CTR encryption for sensitive data, with support for batch job executions in parallel.

## Features

- **Archive Storage**: Automatically compress your configuration files into `gzip` format to save space.
- **Secure Storage**: Encrypt your configuration files using AES-256 CTR encryption with a custom generated master key.
- **Batch Backup Mode**: Perform multiple backup jobs in a single execution using a JSON configuration file.
- **Parallel Processing**: Automatically executes multiple batch jobs concurrently in a worker pool to speed up execution.
- **Interactive Configuration Builder**: Built-in interactive CLI wizard to easily generate backup configuration files.
- **Path Resolution**: Automatically resolves absolute paths and standard user home directory path expansions (`~/...`).
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

#### 1. Interactive Config Maker (`init`)
Launches a CLI configuration wizard to help you build a JSON file containing one or more backup jobs.

**Example:**
```bash
configflow init
```
This wizard will prompt you for:
- Source file path (supports `~` for your home directory)
- Backup destination path
- Storage type (`archive` or `secure`)
- Option to add more jobs, and finally the filename (defaults to `backup.json`)

#### 2. Generate an Encryption Key (`keygen`)
Generates a secure 256-bit (32-byte) key and prints it to the terminal in HEX format.

**Example:**
```bash
configflow keygen
# Output:
# Your Master Key (Save it securely!):
# a1b2c3d4e5f6...
```

#### 3. Backup File(s) (`backup`)
Backs up an existing configuration file or runs batch backup jobs.
*Note: You must specify either the `-in` and `-out` flags (for a single file) or the `-config` flag (for batch backup), but not both.*

*Flags:*
- `-in`: Path to the source file you want to backup.
- `-out`: Path to the destination backup file.
- `-key`: A 32-byte (64 hex characters) key for encryption. Enables secure mode.
- `-config`: Path to a JSON config file with batch backup jobs.

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

*Batch Config Mode (Parallel Execution):*
```bash
configflow backup -config ./backup.json
```

If your configuration contains `secure` jobs, specify the `-key` flag:
```bash
configflow backup -config ./backup.json -key <your_64_character_hex_key>
```

##### Config File Format
The configuration file is a JSON array of backup jobs. Each job requires `source_data`, `backup_dest`, and `storage_type`.
```json
[
  {
    "source_data": "~/app-config.json",
    "backup_dest": "./backups/app-config",
    "storage_type": "secure"
  },
  {
    "source_data": "./local-data.db",
    "backup_dest": "./backups/local-data",
    "storage_type": "archive"
  }
]
```

#### 4. Restore a File (`restore`)
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
- `engine.go`: Contains `SyncEngine`, which orchestrates backup and restore processes (including concurrent batch operations).
- `storage.go`: Implementations of the `Storage` interface (`ArchiveStorage`, `SecureStorage`).
- `config-maker.go`: Contains `RunConfigMaker`, the interactive wizard to generate config files.
- `parser.go`: Utility functions to parse JSON config files and resolve paths.
- `key-handler.go`: Key generation, encoding and decoding utilities.

## License

This project is open-source and available under the MIT License.
