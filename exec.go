package main

import (
	"os"
	"os/exec"
	"syscall"
)

// claudeArgv puts the prompt before --add-dir: that flag is variadic and would swallow a
// following positional, so `claude --add-dir <docs> "/quno:start x"` opens a session with no prompt.
func claudeArgv(cfg *config, prompt string, extra ...string) []string {
	argv := []string{"claude"}
	if prompt != "" {
		argv = append(argv, prompt)
	}
	argv = append(argv, "--add-dir", cfg.docs)
	return append(argv, extra...)
}

func execClaude(cfg *config, prompt string, extra ...string) error {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return err
	}
	return syscall.Exec(bin, claudeArgv(cfg, prompt, extra...), os.Environ())
}

func claudeCmd(cfg *config, dir, prompt string, extra ...string) (*exec.Cmd, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return nil, err
	}
	argv := claudeArgv(cfg, prompt, extra...)
	c := exec.Command(bin, argv[1:]...)
	c.Dir = dir
	return c, nil
}
