// Package main provides a utility for mapping running GPU processes and managing mapping files.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type GPUProcess struct {
	GPU string
	PID string
}

// Parses `nvidia-smi` and returns a slice of GPUProcess structs
func getGPUProcesses() ([]GPUProcess, error) {
	cmd := exec.Command("nvidia-smi", "--query-compute-apps=gpu_uuid,pid", "--format=csv,noheader")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var processes []GPUProcess
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 2 {
			gpu := strings.TrimSpace(fields[0])
			pid := strings.TrimSpace(fields[1])
			processes = append(processes, GPUProcess{gpu, pid})
		}
	}
	return processes, nil
}

// Cleans up old mapping files in the directory
func cleanMappingFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist yet, nothing to clean
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: failed to remove old mapping file %s: %v", filePath, err)
			}
		}
	}
	return nil
}

// Writes mapping files: one file per GPU, each line a PID (job ID)
func writeMappingFiles(dir string, processes []GPUProcess) error {
	gpuMap := map[string][]string{}
	for _, proc := range processes {
		gpuMap[proc.GPU] = append(gpuMap[proc.GPU], proc.PID)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Clean old mapping files first
	if err := cleanMappingFiles(dir); err != nil {
		return fmt.Errorf("failed to clean old mapping files: %w", err)
	}

	// Write new mapping files
	for gpu, pids := range gpuMap {
		filePath := filepath.Join(dir, gpu)
		f, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		w := bufio.NewWriter(f)
		for _, pid := range pids {
			if _, err := fmt.Fprintln(w, pid); err != nil {
				if cerr := f.Close(); cerr != nil {
					log.Printf("Warning: failed to close file %s: %v", filePath, cerr)
				}
				return fmt.Errorf("failed to write pid to file %s: %w", filePath, err)
			}
		}
		if err := w.Flush(); err != nil {
			if cerr := f.Close(); cerr != nil {
				log.Printf("Warning: failed to close file %s: %v", filePath, cerr)
			}
			return fmt.Errorf("failed to flush writer for file %s: %w", filePath, err)
		}
		if cerr := f.Close(); cerr != nil {
			log.Printf("Warning: failed to close file %s: %v", filePath, cerr)
		}
	}
	return nil
}

// Performs a single update cycle
func updateMappings(mappingDir string) error {
	processes, err := getGPUProcesses()
	if err != nil {
		return fmt.Errorf("failed to get GPU processes: %w", err)
	}

	if err := writeMappingFiles(mappingDir, processes); err != nil {
		return fmt.Errorf("failed to write mapping files: %w", err)
	}

	log.Printf("Updated mapping files in %s (%d processes)", mappingDir, len(processes))
	return nil
}

// Runs in daemon mode with periodic updates
func runDaemon(ctx context.Context, mappingDir string, interval time.Duration) error {
	log.Printf("Starting daemon mode (interval: %v, directory: %s)", interval, mappingDir)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Perform initial update
	if err := updateMappings(mappingDir); err != nil {
		log.Printf("Error during initial update: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down gracefully...")
			return nil
		case <-ticker.C:
			if err := updateMappings(mappingDir); err != nil {
				log.Printf("Error updating mappings: %v", err)
			}
		}
	}
}

func main() {
	// Command-line flags
	daemon := flag.Bool("daemon", false, "Run in daemon mode (continuous monitoring)")
	interval := flag.Duration("interval", 30*time.Second, "Update interval in daemon mode (e.g., 30s, 1m, 5m)")
	mappingDir := flag.String("dir", "/tmp/dcgm-job-mapping", "Directory for mapping files")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	// Configure logging
	if !*verbose {
		log.SetFlags(log.LstdFlags)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	// Run in daemon mode or single-shot mode
	if *daemon {
		if err := runDaemon(ctx, *mappingDir, *interval); err != nil {
			log.Printf("Daemon error: %v", err)
			return
		}
	} else {
		// Single execution
		if err := updateMappings(*mappingDir); err != nil {
			log.Printf("Error: %v", err)
			return
		}
	}
}
