---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description

<!-- A clear and concise description of what the bug is -->

## To Reproduce

Steps to reproduce the behavior:

1. Run command '...'
2. With flags '...'
3. See error

## Expected Behavior

<!-- A clear and concise description of what you expected to happen -->

## Actual Behavior

<!-- What actually happened -->

## Environment

- OS: [e.g., Ubuntu 22.04, macOS 14.0]
- Go Version: [e.g., 1.22.0]
- dcgm-mapper Version: [e.g., v1.0.0]
- GPU: [e.g., NVIDIA Tesla V100]
- nvidia-smi Version: [output of `nvidia-smi --version`]

## Logs

```
<!-- Paste relevant logs here. Run with -verbose flag for detailed logs -->
```

## Configuration

<!-- If running in daemon mode, include relevant flags and configuration -->

```bash
dcgm-mapper -daemon -interval 30s -dir /tmp/dcgm-job-mapping
```

## Additional Context

<!-- Add any other context about the problem here -->

## Possible Solution

<!-- Optional: Suggest a fix or reason for the bug -->

