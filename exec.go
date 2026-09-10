package main

import (
	"os"
	"os/exec"
	"syscall"
)

func execClaude(cfg *config, prompt string, extra ...string) error {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return err
	}
	argv := append([]string{"claude", "--add-dir", cfg.docs}, extra...)
	if prompt != "" {
		argv = append(argv, prompt)
	}
	return syscall.Exec(bin, argv, os.Environ())
}
