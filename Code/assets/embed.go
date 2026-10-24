package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed embedded/*
var EmbeddedFS embed.FS

func ExtractEmbeddedADB() (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("embedded adb.exe is only for windows")
	}

	adbData, err := EmbeddedFS.ReadFile("embedded/adb.exe")
	if err != nil {
		return "", err
	}

	tempDir := filepath.Join(os.TempDir(), "duolingo_adb")
	_ = os.MkdirAll(tempDir, 0755)

	targetADB := filepath.Join(tempDir, "adb.exe")
	_ = os.WriteFile(targetADB, adbData, 0755)

	if dll1, err := EmbeddedFS.ReadFile("embedded/AdbWinApi.dll"); err == nil {
		_ = os.WriteFile(filepath.Join(tempDir, "AdbWinApi.dll"), dll1, 0644)
	}
	if dll2, err := EmbeddedFS.ReadFile("embedded/AdbWinUsbApi.dll"); err == nil {
		_ = os.WriteFile(filepath.Join(tempDir, "AdbWinUsbApi.dll"), dll2, 0644)
	}

	return targetADB, nil
}

func GetFiguresFS() (fs.FS, error) {
	sub, err := fs.Sub(EmbeddedFS, "embedded/Figures")
	if err != nil {
		return nil, fmt.Errorf("embedded/Figures not found: %w", err)
	}
	return sub, nil
}
