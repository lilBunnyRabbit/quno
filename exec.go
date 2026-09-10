package main

import (
	"os"
	"os/exec"
	"syscall"
)

// execClaude puts the prompt before --add-dir: that flag is variadic and would swallow a
// following positional, so `claude --add-dir <docs> "/quno:start x"` opens a session with no prompt.
func execClaude(cfg *config, prompt string, extra ...string) error {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return err
	}
	argv := []string{"claude"}
	if prompt != "" {
		argv = append(argv, prompt)
	}
	argv = append(argv, "--add-dir", cfg.docs)
	argv = append(argv, extra...)
	return syscall.Exec(bin, argv, os.Environ())
}
