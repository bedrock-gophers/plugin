package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

func writeGeneratedGoFile(path, src string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, src)
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return err
	}
	return nil
}
