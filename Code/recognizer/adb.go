package recognizer

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"chessbot/Code/assets"
	"chessbot/Code/engine"
)

var (
	isDirectRootMode bool
	detectedSuPath   string
)

func InitDeviceMode() string {
	if runtime.GOOS == "android" || runtime.GOARCH == "arm64" || runtime.GOARCH == "arm" {
		if _, err := exec.LookPath("screencap"); err == nil {
			isDirectRootMode = true
			return "direct (Android root / shell)"
		}
		for _, su := range []string{"su", "/system/bin/su", "/system/xbin/su", "/data/adb/ksu/bin/su", "/data/adb/magisk/su"} {
			if _, err := exec.LookPath(su); err == nil {
				isDirectRootMode = true
				detectedSuPath = su
				return fmt.Sprintf("direct (%s / KernelSU / Magisk)", su)
			}
		}
	}

	return "adb"
}

func FindADBPath() string {
	mode := InitDeviceMode()
	if isDirectRootMode {
		return mode
	}

	if runtime.GOOS == "windows" {
		if embeddedADB, err := assets.ExtractEmbeddedADB(); err == nil && embeddedADB != "" {
			return embeddedADB
		}
	}

	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{"adb.exe", ".\\adb.exe", "adb"}
	} else {
		candidates = []string{"adb", "./adb", "/usr/bin/adb", "/usr/local/bin/adb"}
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		if runtime.GOOS == "windows" {
			candidates = append([]string{filepath.Join(exeDir, "adb.exe")}, candidates...)
		} else {
			candidates = append([]string{filepath.Join(exeDir, "adb")}, candidates...)
		}
	}

	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			abs, err := filepath.Abs(p)
			if err != nil {
				abs = p
			}
			if runtime.GOOS != "windows" {
				_ = os.Chmod(abs, 0755)
			}
			return abs
		}
	}

	if path, err := exec.LookPath("adb"); err == nil {
		if runtime.GOOS != "windows" {
			_ = os.Chmod(path, 0755)
		}
		return path
	}

	return "adb"
}

func CaptureScreen(adbPath string) (image.Image, error) {
	var cmd *exec.Cmd
	if isDirectRootMode {
		if detectedSuPath != "" {
			cmd = exec.Command(detectedSuPath, "-c", "screencap -p")
		} else {
			cmd = exec.Command("screencap", "-p")
		}
	} else {
		cmd = exec.Command(adbPath, "exec-out", "screencap", "-p")
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	img, _, err := image.Decode(&out)
	return img, err
}

func Tap(adbPath string, x, y int) error {
	var cmd *exec.Cmd
	if isDirectRootMode {
		if detectedSuPath != "" {
			cmd = exec.Command(detectedSuPath, "-c", fmt.Sprintf("input tap %d %d", x, y))
		} else {
			cmd = exec.Command("input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
		}
	} else {
		cmd = exec.Command(adbPath, "shell", "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	}
	return cmd.Run()
}

func ExecuteMove(adbPath string, screenW, screenH int, yRatio float64, m engine.Move) error {
	cellW := float64(screenW) / 8.0
	cellH := cellW
	startY := float64(screenH) * yRatio

	fromX := int((float64(m.FromC) + 0.5) * cellW)
	fromY := int(startY + (float64(m.FromR)+0.5)*cellH)

	toX := int((float64(m.ToC) + 0.5) * cellW)
	toY := int(startY + (float64(m.ToR)+0.5)*cellH)

	isPromotion := len(m.Piece) >= 2 && m.Piece[1] == 'P' && (m.ToR == 0 || m.ToR == 7)

	if err := Tap(adbPath, fromX, fromY); err != nil {
		return err
	}
	time.Sleep(180 * time.Millisecond)

	if err := Tap(adbPath, toX, toY); err != nil {
		return err
	}

	if isPromotion {
		time.Sleep(550 * time.Millisecond)
		qX, qY := GetPromotionQueenTap(screenW, screenH, yRatio, m.ToR, m.ToC)
		_ = Tap(adbPath, qX, qY)
		time.Sleep(450 * time.Millisecond)
	}

	return nil
}
