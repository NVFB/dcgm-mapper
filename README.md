# DCGM Mapper

A lightweight tool for mapping GPU processes to their PIDs using NVIDIA's `nvidia-smi`. It can run as a one-shot command or as a daemon for continuous monitoring.

## Features

- **Single-shot mode**: Run once and exit
- **Daemon mode**: Continuous monitoring with configurable intervals
- **Graceful shutdown**: Handles SIGINT and SIGTERM signals
- **Automatic cleanup**: Old mapping files are removed before writing new ones
- **Configurable**: Command-line flags for directory, interval, and verbosity

## Building

```bash
# Build the binary
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## Usage

### Single Execution (Default)

Run once and update the GPU-to-PID mapping files:

```bash
dcgm-mapper
```

### Daemon Mode

Run continuously with periodic updates:

```bash
# Run with default 30-second interval
dcgm-mapper -daemon

# Run with custom interval
dcgm-mapper -daemon -interval 1m

# Run with custom directory and verbose logging
dcgm-mapper -daemon -interval 5m -dir /var/lib/dcgm-mapper -verbose
```

### Command-Line Flags

- `-daemon` - Run in daemon mode (continuous monitoring)
- `-interval duration` - Update interval in daemon mode (default: 30s)
  - Examples: `30s`, `1m`, `5m`, `1h`
- `-dir string` - Directory for mapping files (default: `/tmp/dcgm-job-mapping`)
- `-verbose` - Enable verbose logging with file locations

## Output

The tool creates one file per GPU in the specified directory. Each file:

- Is named with the GPU UUID (e.g., `GPU-12345678-1234-1234-1234-123456789012`)
- Contains one PID per line
- Represents processes currently running on that GPU

Example directory structure:

```
/tmp/dcgm-job-mapping/
├── GPU-12345678-1234-1234-1234-123456789012
├── GPU-87654321-4321-4321-4321-210987654321
└── GPU-ABCDEFAB-ABCD-ABCD-ABCD-ABCDEFABCDEF
```

## Running as a System Service

### Systemd (Linux)

Create a systemd service file at `/etc/systemd/system/dcgm-mapper.service`:

```ini
[Unit]
Description=DCGM GPU-to-PID Mapper
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/dcgm-mapper -daemon -interval 30s -dir /var/lib/dcgm-mapper
Restart=on-failure
RestartSec=10
User=root
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
# Copy binary to system location
sudo install -m 0755 bin/dcgm-mapper /usr/local/bin/

# Create directory for mapping files
sudo mkdir -p /var/lib/dcgm-mapper

# Reload systemd, enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable dcgm-mapper
sudo systemctl start dcgm-mapper

# Check status
sudo systemctl status dcgm-mapper

# View logs
sudo journalctl -u dcgm-mapper -f
```

## Requirements

- NVIDIA GPU with CUDA support
- `nvidia-smi` available in PATH
- Go 1.16+ (for building)

## Development

### Project Structure

```
.
├── main.go           # Main application code
├── main_test.go      # Unit tests
├── Makefile          # Build automation
├── go.mod            # Go module definition
├── .gitignore        # Git ignore patterns
└── README.md         # This file
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Format code
make fmt

# Run go vet
make vet
```

## License

Add your license here.
