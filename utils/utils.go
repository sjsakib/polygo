package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func CopyFile(src, dest string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, content, 0644)
}

func CopyDirectoryWithFiles(src, dest string, ignore []string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		for _, ignore := range ignore {
			if strings.Contains(path, ignore) {
				return nil
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, relPath)

		err = os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, content, 0644)
	})
}
