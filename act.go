package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const versionFilePath = "Code/config/version.go"

const (
	windowsPlatformToolsURL = "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
	linuxPlatformToolsURL   = "https://dl.google.com/android/repository/platform-tools-latest-linux.zip"
)

func getVersion() string {
	data, err := os.ReadFile(versionFilePath)
	if err != nil {
		return "2.3"
	}
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "Version") || strings.HasPrefix(l, "var Version") {
			parts := strings.Split(l, "\"")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "2.3"
}

func setVersion(newVer string) {
	newVer = strings.TrimSpace(newVer)
	if newVer == "" {
		fmt.Println("Usage: go run act.go set-version <version_name>")
		return
	}
	content := fmt.Sprintf("package config\n\nvar (\n\tVersion = \"%s\"\n\tAuthor  = \"ytika\"\n)\n", newVer)
	if err := os.WriteFile(versionFilePath, []byte(content), 0644); err != nil {
		fmt.Printf("Failed to write %s: %v\n", versionFilePath, err)
		return
	}
	fmt.Printf("✓ Version updated to: %s\n", newVer)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func downloadFile(url, dstPath string) error {
	_ = os.MkdirAll(filepath.Dir(dstPath), 0755)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status: %s", resp.Status)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractFileFromZip(zipPath, entrySuffix, dstPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, entrySuffix) {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			_ = os.MkdirAll(filepath.Dir(dstPath), 0755)
			outFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			return err
		}
	}

	return fmt.Errorf("entry '%s' not found in zip archive", entrySuffix)
}

func ensureADB(targetOS, outDir string) error {
	if targetOS == "android" {
		// On Android (Termux + root), adb is not needed (screencap and input run natively via su)
		return nil
	}

	cacheDir := filepath.Join(".adb_cache", targetOS)
	_ = os.MkdirAll(cacheDir, 0755)

	if targetOS == "windows" {
		targetADB := filepath.Join(cacheDir, "adb.exe")
		targetDLL1 := filepath.Join(cacheDir, "AdbWinApi.dll")
		targetDLL2 := filepath.Join(cacheDir, "AdbWinUsbApi.dll")

		if _, err := os.Stat("adb.exe"); err == nil {
			_ = copyFile("adb.exe", targetADB)
			_ = copyFile("AdbWinApi.dll", targetDLL1)
			_ = copyFile("AdbWinUsbApi.dll", targetDLL2)
		}

		if _, err := os.Stat(targetADB); err != nil {
			fmt.Println("  [ADB] Downloading official Android Platform-Tools for Windows...")
			zipPath := filepath.Join(cacheDir, "platform-tools.zip")
			if err := downloadFile(windowsPlatformToolsURL, zipPath); err != nil {
				return fmt.Errorf("failed to download Windows ADB: %w", err)
			}
			_ = extractFileFromZip(zipPath, "platform-tools/adb.exe", targetADB)
			_ = extractFileFromZip(zipPath, "platform-tools/AdbWinApi.dll", targetDLL1)
			_ = extractFileFromZip(zipPath, "platform-tools/AdbWinUsbApi.dll", targetDLL2)
			_ = os.Remove(zipPath)
		}

		_ = copyFile(targetADB, filepath.Join(outDir, "adb.exe"))
		_ = copyFile(targetDLL1, filepath.Join(outDir, "AdbWinApi.dll"))
		_ = copyFile(targetDLL2, filepath.Join(outDir, "AdbWinUsbApi.dll"))

	} else if targetOS == "linux" {
		targetADB := filepath.Join(cacheDir, "adb")

		if _, err := os.Stat(targetADB); err != nil {
			fmt.Println("  [ADB] Downloading official Android Platform-Tools for Linux...")
			zipPath := filepath.Join(cacheDir, "platform-tools.zip")
			if err := downloadFile(linuxPlatformToolsURL, zipPath); err != nil {
				return fmt.Errorf("failed to download Linux ADB: %w", err)
			}
			_ = extractFileFromZip(zipPath, "platform-tools/adb", targetADB)
			_ = os.Chmod(targetADB, 0755)
			_ = os.Remove(zipPath)
		}

		dstADB := filepath.Join(outDir, "adb")
		_ = copyFile(targetADB, dstADB)
		_ = os.Chmod(dstADB, 0755)
	}

	return nil
}

func prepareEmbedded(targetOS string) func() {
	embeddedDir := filepath.Join("Code", "assets", "embedded")
	_ = os.MkdirAll(embeddedDir, 0755)

	_ = copyDir("Figures", filepath.Join(embeddedDir, "Figures"))

	if targetOS == "windows" {
		_ = copyFile("adb.exe", filepath.Join(embeddedDir, "adb.exe"))
		_ = copyFile("AdbWinApi.dll", filepath.Join(embeddedDir, "AdbWinApi.dll"))
		_ = copyFile("AdbWinUsbApi.dll", filepath.Join(embeddedDir, "AdbWinUsbApi.dll"))
	}

	return func() {
		_ = os.Remove(filepath.Join(embeddedDir, "adb.exe"))
		_ = os.Remove(filepath.Join(embeddedDir, "AdbWinApi.dll"))
		_ = os.Remove(filepath.Join(embeddedDir, "AdbWinUsbApi.dll"))
		_ = os.RemoveAll(filepath.Join(embeddedDir, "Figures"))
	}
}

func buildTarget(targetOS string) {
	ver := getVersion()
	folderName := "Windows"
	binName := "ChessBot.exe"
	goos := targetOS
	goarch := "amd64"

	if targetOS == "linux" {
		folderName = "Linux"
		binName = "chessbot"
	} else if targetOS == "android" {
		folderName = "Android"
		binName = "chessbot-android"
		goos = "android"
		goarch = "arm64"
	}

	outDir := filepath.Join("Builds", ver, folderName)
	fmt.Printf("==> Building for %s (v%s, arch=%s) into %s ...\n", strings.Title(targetOS), ver, goarch, outDir)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Cannot create directory %s: %v\n", outDir, err)
		return
	}

	// 1. Prepare ADB
	if targetOS != "android" {
		fmt.Println("  [1/4] Preparing platform ADB binary...")
		if err := ensureADB(targetOS, outDir); err != nil {
			fmt.Printf("  Warning: could not prepare ADB: %v\n", err)
		}
	} else {
		fmt.Println("  [1/4] Android root mode: direct screencap & input via su (no external ADB needed)")
	}

	// 2. Prepare embedded figures & assets
	cleanup := prepareEmbedded(targetOS)
	defer cleanup()

	targetBin := filepath.Join(outDir, binName)
	fmt.Printf("  [2/4] Compiling %s (GOOS=%s, GOARCH=%s)...\n", targetBin, goos, goarch)

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", targetBin, "main.go")
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		return
	}

	// 3. Write sample config
	fmt.Println("  [3/4] Writing sample config.jsonc...")
	cfgDst := filepath.Join(outDir, "config.jsonc")
	if data, err := os.ReadFile("config.jsonc"); err == nil {
		_ = os.WriteFile(cfgDst, data, 0644)
	}

	// 4. Create convenient start script for Termux
	if targetOS == "android" {
		startSh := filepath.Join(outDir, "start.sh")
		scriptContent := "#!/data/data/com.termux/files/usr/bin/sh\nchmod +x chessbot-android\nsu -c ./chessbot-android \"$@\"\n"
		_ = os.WriteFile(startSh, []byte(scriptContent), 0755)
	}

	fmt.Println("  [4/4] Finalizing build...")
	fmt.Printf("✓ %s build complete in %s\n\n", strings.Title(targetOS), outDir)
}

func build(target string) {
	target = strings.ToLower(strings.TrimSpace(target))
	switch target {
	case "linux":
		buildTarget("linux")
	case "windows":
		buildTarget("windows")
	case "android", "termux":
		buildTarget("android")
	case "all":
		buildTarget("windows")
		buildTarget("linux")
		buildTarget("android")
	default:
		buildTarget("windows")
		buildTarget("linux")
		buildTarget("android")
	}
}

func cleanProject() {
	fmt.Println("==> Cleaning project temporary files...")
	patterns := []string{
		"*.exe",
		"config.json",
		"*scratch*",
		"*test*",
	}
	count := 0
	for _, pat := range patterns {
		matches, _ := filepath.Glob(pat)
		for _, m := range matches {
			if m != "adb.exe" {
				_ = os.RemoveAll(m)
				count++
				fmt.Printf("  Removed: %s\n", m)
			}
		}
	}
	fmt.Printf("✓ Project cleaned (removed %d items)\n", count)
}

func printHelp() {
	fmt.Printf("ChessBot Build & Management Tool (act.go) - Created by %s\n", "ytika")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run act.go <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  set-version <version>             Set application version in Code/config/version.go")
	fmt.Println("  build [windows|linux|android|all] Compile standalone release (defaults to all)")
	fmt.Println("                                    - windows: AMD64 exe + ADB & DLLs")
	fmt.Println("                                    - linux: AMD64 ELF + Linux ADB")
	fmt.Println("                                    - android: ARM64 ELF for Termux with root (KernelSU/Magisk)")
	fmt.Println("  clean                             Remove temporary binaries, test files, and logs")
	fmt.Println("  help, h, -h, --help               Show this help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run act.go build android")
	fmt.Println("  go run act.go build windows")
	fmt.Println("  go run act.go build all")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "help", "h", "-h", "--help", "-help":
		printHelp()
	case "set-version":
		if len(os.Args) < 3 {
			fmt.Println("Error: please specify a version name. Example: go run act.go set-version 1.2.0")
			return
		}
		setVersion(os.Args[2])
	case "build":
		target := "all"
		if len(os.Args) >= 3 {
			target = os.Args[2]
		}
		build(target)
	case "clean":
		cleanProject()
	default:
		fmt.Printf("Unknown command '%s'. Run 'go run act.go help' for usage.\n", cmd)
	}
}
