package capture

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteAtomic(path string, raw []byte) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), ".safe-svc-capture-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer func() { _ = os.Remove(name) }()
	if err = temp.Chmod(0o600); err != nil {
		return err
	}
	if _, err = temp.Write(raw); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = os.Link(name, path); err != nil {
		return fmt.Errorf("claim destination: %w", err)
	}
	return nil
}
