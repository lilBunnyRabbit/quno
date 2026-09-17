package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
)

const (
	defaultKeep    = 14
	snapshotSuffix = ".bundle.age"
	historyExclude = ".DS_Store\n.trash/\n.obsidian/workspace*.json\n"
)

func stateDir() string {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "quno")
	}
	return filepath.Join(home(), ".local", "share", "quno")
}

// vaultKey keeps a scratch QUNO_DOCS from sharing history or snapshots with the real vault.
func (c *config) vaultKey() string {
	sum := sha1.Sum([]byte(c.docs))
	return filepath.Base(c.docs) + "-" + hex.EncodeToString(sum[:])[:6]
}

func (c *config) historyDir() string {
	if c.history != "" {
		return c.history
	}
	return filepath.Join(stateDir(), "history", c.vaultKey()+".git")
}

func (c *config) snapshotDir() string {
	if c.snapshots != "" {
		return c.snapshots
	}
	return filepath.Join(stateDir(), "snapshots", c.vaultKey())
}

// vaultGit never runs git init inside the vault: the git dir lives outside it and is always
// passed explicitly, so the vault holds no .git and repo detection from inside it still fails.
func vaultGit(cfg *config, args ...string) *exec.Cmd {
	argv := append([]string{"--git-dir=" + cfg.historyDir(), "--work-tree=" + cfg.docs}, args...)
	c := exec.Command("git", argv...)
	c.Dir = cfg.docs
	return c
}

func ensureHistory(cfg *config) error {
	dir := cfg.historyDir()
	if _, err := os.Stat(filepath.Join(dir, "HEAD")); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return err
	}
	if msg, err := exec.Command("git", "init", "--quiet", "--bare", dir).CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %s", strings.TrimSpace(string(msg)))
	}
	return os.WriteFile(filepath.Join(dir, "info", "exclude"), []byte(historyExclude), 0o644)
}

func cmdSync(cfg *config, args []string) error {
	title := ""
	if len(args) == 2 && args[0] == "-m" {
		title = args[1]
	} else if len(args) != 0 {
		return fmt.Errorf("usage: quno sync [-m <title>]")
	}
	sha, changed, err := syncVault(cfg, title)
	if err != nil {
		return err
	}
	if changed == 0 {
		fmt.Println(out.dim("nothing to sync"))
		return nil
	}
	fmt.Printf("%s %s  %d changed  %s\n", out.green("✓"), out.cyan(sha), changed, out.dim(contract(cfg.historyDir())))
	return nil
}

func syncVault(cfg *config, title string) (string, int, error) {
	if err := ensureHistory(cfg); err != nil {
		return "", 0, err
	}
	if msg, err := vaultGit(cfg, "add", "-A").CombinedOutput(); err != nil {
		return "", 0, fmt.Errorf("git add: %s", strings.TrimSpace(string(msg)))
	}
	status, err := vaultGit(cfg, "status", "--porcelain").Output()
	if err != nil {
		return "", 0, err
	}
	lines := strings.Split(strings.TrimRight(string(status), "\n"), "\n")
	if lines[0] == "" {
		return "", 0, nil
	}
	if title == "" {
		title = "chore: sync " + time.Now().Format("2006-01-02 15:04")
	}
	body := lines
	if len(body) > 20 {
		body = append(body[:20:20], fmt.Sprintf("and %d more", len(lines)-20))
	}
	message := title + "\n\n- " + strings.Join(body, "\n- ") + "\n"
	if msg, err := vaultGit(cfg, "commit", "--quiet", "-m", message).CombinedOutput(); err != nil {
		return "", 0, fmt.Errorf("git commit: %s", strings.TrimSpace(string(msg)))
	}
	sha, err := vaultGit(cfg, "rev-parse", "--short", "HEAD").Output()
	return strings.TrimSpace(string(sha)), len(lines), err
}

func cmdGit(cfg *config, args []string) error {
	if err := ensureHistory(cfg); err != nil {
		return err
	}
	c := vaultGit(cfg, args...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

func cmdSnapshot(cfg *config, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "keygen":
			return snapshotKeygen(cfg)
		case "restore":
			return cmdRestore(args[1:])
		case "ls":
			for _, name := range snapshotFiles(cfg) {
				fmt.Println(filepath.Join(cfg.snapshotDir(), name))
			}
			return nil
		}
		return fmt.Errorf("usage: quno snapshot [keygen | ls | restore <file> <dir> [-i <identity file>]]")
	}
	path, err := snapshot(cfg, time.Now())
	if err != nil {
		return err
	}
	size := int64(0)
	if fi, err := os.Stat(path); err == nil {
		size = fi.Size()
	}
	fmt.Printf("%s %s  %s\n", out.green("✓"), contract(path), out.dim(fmt.Sprintf("%d KB, keeping %d", (size+1023)/1024, cfg.keepCount())))
	return nil
}

func (c *config) keepCount() int {
	if c.keep > 0 {
		return c.keep
	}
	return defaultKeep
}

func snapshot(cfg *config, now time.Time) (string, error) {
	if cfg.recipient == "" {
		return "", fmt.Errorf("no recipient under [vault] in %s; run: quno snapshot keygen", contract(cfg.tomlPath()))
	}
	recipient, err := age.ParseX25519Recipient(cfg.recipient)
	if err != nil {
		return "", fmt.Errorf("recipient: %w", err)
	}
	if _, _, err := syncVault(cfg, ""); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp("", "quno-snapshot-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	bundle := filepath.Join(tmp, "vault.bundle")
	if msg, err := vaultGit(cfg, "bundle", "create", "--quiet", bundle, "--all").CombinedOutput(); err != nil {
		return "", fmt.Errorf("git bundle: %s", strings.TrimSpace(string(msg)))
	}
	if err := os.MkdirAll(cfg.snapshotDir(), 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(cfg.snapshotDir(), filepath.Base(cfg.docs)+"-"+now.Format("20060102-150405")+snapshotSuffix)
	if err := encryptFile(bundle, path, recipient); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, pruneSnapshots(cfg)
}

func encryptFile(src, dst string, recipient age.Recipient) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	outFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer outFile.Close()
	w, err := age.Encrypt(outFile, recipient)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, in); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return outFile.Sync()
}

func snapshotFiles(cfg *config) []string {
	entries, _ := os.ReadDir(cfg.snapshotDir())
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), snapshotSuffix) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

func pruneSnapshots(cfg *config) error {
	names := snapshotFiles(cfg)
	for i := 0; i < len(names)-cfg.keepCount(); i++ {
		if err := os.Remove(filepath.Join(cfg.snapshotDir(), names[i])); err != nil {
			return err
		}
	}
	return nil
}

func snapshotKeygen(cfg *config) error {
	if cfg.recipient != "" {
		return fmt.Errorf("recipient already set in %s; delete that line first, older snapshots stay readable only with the old key", contract(cfg.tomlPath()))
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return err
	}
	if err := writeRecipient(cfg.tomlPath(), identity.Recipient().String()); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "public key written to %s\nsecret key below, shown once and stored nowhere: put it in your password manager.\nwithout it no snapshot can be restored.\n\n", contract(cfg.tomlPath()))
	fmt.Println(identity.String())
	return nil
}

func writeRecipient(path, recipient string) error {
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	line := fmt.Sprintf("recipient = %q", recipient)
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) == "[vault]" {
			lines = append(lines[:i+1], append([]string{line}, lines[i+1:]...)...)
			return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	text := strings.TrimRight(string(raw), "\n")
	if text != "" {
		text += "\n\n"
	}
	return os.WriteFile(path, []byte(text+"[vault]\n"+line+"\n"), 0o644)
}

func cmdRestore(args []string) error {
	identityFile := ""
	var rest []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-i" && i+1 < len(args) {
			identityFile = args[i+1]
			i++
			continue
		}
		rest = append(rest, args[i])
	}
	if len(rest) != 2 {
		return fmt.Errorf("usage: quno snapshot restore <file> <dir> [-i <identity file>]")
	}
	secret, err := readIdentity(identityFile)
	if err != nil {
		return err
	}
	if err := restoreSnapshot(rest[0], expand(rest[1]), secret); err != nil {
		return err
	}
	fmt.Printf("%s restored into %s\n", out.green("✓"), contract(expand(rest[1])))
	fmt.Println(out.dim("it is a normal clone with the full history; before using it as the vault move its .git out:"))
	fmt.Println(out.dim("  mv <dir>/.git <history path> && git --git-dir=<history path> config core.bare true"))
	return nil
}

func readIdentity(file string) (string, error) {
	if file != "" {
		raw, err := os.ReadFile(expand(file))
		if err != nil {
			return "", err
		}
		for _, l := range strings.Split(string(raw), "\n") {
			if l = strings.TrimSpace(l); strings.HasPrefix(l, "AGE-SECRET-KEY-") {
				return l, nil
			}
		}
		return "", fmt.Errorf("no AGE-SECRET-KEY line in %s", file)
	}
	if v := strings.TrimSpace(os.Getenv("QUNO_AGE_KEY")); v != "" {
		return v, nil
	}
	fmt.Fprint(os.Stderr, "secret key (AGE-SECRET-KEY-...): ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func restoreSnapshot(file, target, secret string) error {
	identity, err := age.ParseX25519Identity(secret)
	if err != nil {
		return fmt.Errorf("secret key: %w", err)
	}
	if entries, err := os.ReadDir(target); err == nil && len(entries) > 0 {
		return fmt.Errorf("%s exists and is not empty", target)
	}
	in, err := os.Open(expand(file))
	if err != nil {
		return err
	}
	defer in.Close()
	r, err := age.Decrypt(in, identity)
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "quno-restore-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	bundle := filepath.Join(tmp, "vault.bundle")
	f, err := os.Create(bundle)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if msg, err := exec.Command("git", "clone", "--quiet", bundle, target).CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %s", strings.TrimSpace(string(msg)))
	}
	return nil
}
