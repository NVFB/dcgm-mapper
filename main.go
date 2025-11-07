package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// Writes mapping files: one file per GPU, each line a PID (job ID)
func writeMappingFiles(dir string, processes []GPUProcess) error {
	gpuMap := map[string][]string{}
	for _, proc := range processes {
		gpuMap[proc.GPU] = append(gpuMap[proc.GPU], proc.PID)
	}
	os.MkdirAll(dir, 0755)
	for gpu, pids := range gpuMap {
		filePath := filepath.Join(dir, gpu)
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		w := bufio.NewWriter(f)
		for _, pid := range pids {
			fmt.Fprintln(w, pid)
		}
		w.Flush()
		f.Close()
	}
	return nil
}

func main() {
	mappingDir := "/tmp/dcgm-job-mapping" // Change to your mapping directory
	processes, err := getGPUProcesses()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get GPU processes:", err)
		os.Exit(1)
	}
	err = writeMappingFiles(mappingDir, processes)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write mapping files:", err)
		os.Exit(1)
	}
	fmt.Println("Mapping files written to", mappingDir)
}
