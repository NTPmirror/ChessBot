package duolingo_test

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"

	"chessbot/Code/board"
	"chessbot/Code/duolingo"
	"chessbot/Code/recognizer"
)

func findTestDataDir() string {
	candidates := []string{
		filepath.Join("..", "..", "testdata", "screens"),
		filepath.Join("..", "testdata", "screens"),
		filepath.Join("testdata", "screens"),
		`C:\Users\ytika\Downloads\Telegram Desktop`,
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	return ""
}

func findFiguresDir() string {
	candidates := []string{
		filepath.Join("..", "..", "Figures"),
		filepath.Join("..", "Figures"),
		"Figures",
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	return "Figures"
}

func loadTestImage(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open sample image %s: %v", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("Failed to decode image %s: %v", path, err)
	}
	return img
}

func TestSampleScreensRecognition(t *testing.T) {
	testDataDir := findTestDataDir()
	if testDataDir == "" {
		t.Skip("Screens testdata directory not found, skipping test")
	}

	figuresDir := findFiguresDir()
	templates, err := recognizer.LoadTemplates(figuresDir)
	if err != nil {
		t.Logf("Warning: templates could not be loaded from %s: %v", figuresDir, err)
	}

	newuiDir := filepath.Join(testDataDir, "newui")

	t.Run("NewUILobbyCardDetection", func(t *testing.T) {
		cases := []struct {
			filename     string
			expectedCard int
		}{
			{"photo_2026-11-08_15-45-48.jpg", 0},
			{"photo_2026-11-08_15-45-44.jpg", 1},
			{"photo_2026-11-08_15-45-51.jpg", 2},
		}

		for _, tc := range cases {
			p := filepath.Join(newuiDir, tc.filename)
			if _, err := os.Stat(p); os.IsNotExist(err) {
				t.Logf("Sample file %s not found, skipping case", tc.filename)
				continue
			}

			img := loadTestImage(t, p)

			if !duolingo.IsLobbyScreen(img) {
				t.Errorf("[%s] Expected IsLobbyScreen to be true", tc.filename)
			}
			if !duolingo.IsNewLobbyScreen(img) {
				t.Errorf("[%s] Expected IsNewLobbyScreen to be true", tc.filename)
			}

			card := duolingo.GetNewLobbyCardSelection(img)
			if card != tc.expectedCard {
				t.Errorf("[%s] Card selection = %d, expected %d", tc.filename, card, tc.expectedCard)
			} else {
				t.Logf("[%s] OK: Correctly detected card %d", tc.filename, card)
			}

			btnType, _, _ := duolingo.DetectActionButton(img)
			if btnType != "" {
				t.Errorf("[%s] Expected no action button on lobby screen, got '%s'", tc.filename, btnType)
			}
		}
	})

	t.Run("NewUIGameplayPieceRecognition", func(t *testing.T) {
		p := filepath.Join(newuiDir, "photo_2026-11-08_15-45-55.jpg")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Skipf("File %s not found, skipping", p)
		}

		img := loadTestImage(t, p)

		if duolingo.IsLobbyScreen(img) {
			t.Errorf("photo_2026-11-08_15-45-55.jpg should not be detected as LobbyScreen")
		}

		if len(templates) > 0 {
			rawBoard := recognizer.Process(img, templates, 0.35625, 2500.0)
			count := board.CountBoardPieces(rawBoard)
			if count < 20 {
				t.Errorf("Expected at least 20 pieces recognized, got %d", count)
			} else {
				t.Logf("Successfully recognized %d pieces on board (Oscar y-ratio)", count)
			}
		}
	})

	t.Run("SampleActionButtons", func(t *testing.T) {
		cases := []struct {
			dir         string
			filename    string
			expectedBtn string
		}{
			{testDataDir, "ELO.jpg", "continue"},
			{testDataDir, "chest_mega.jpg", "continue"},
			{testDataDir, "chest_extra.jpg", "continue"},
			{newuiDir, "photo_2026-11-08_15-45-35.jpg", "continue"},
		}

		for _, tc := range cases {
			p := filepath.Join(tc.dir, tc.filename)
			if _, err := os.Stat(p); os.IsNotExist(err) {
				t.Logf("Sample file %s not found, skipping", tc.filename)
				continue
			}

			img := loadTestImage(t, p)
			btnType, x, y := duolingo.DetectActionButton(img)
			if btnType != tc.expectedBtn {
				t.Errorf("[%s] Expected action button '%s', got '%s'", tc.filename, tc.expectedBtn, btnType)
			} else {
				t.Logf("[%s] OK: detected '%s' at (%d, %d)", tc.filename, btnType, x, y)
			}
		}
	})

	t.Run("ChestScreenDetection", func(t *testing.T) {
		p := filepath.Join(testDataDir, "chest_mega.jpg")
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Skipf("File %s not found, skipping", p)
		}

		img := loadTestImage(t, p)
		if !duolingo.IsChestScreen(img) {
			t.Errorf("Expected IsChestScreen to be true for chest_mega.jpg")
		} else {
			t.Logf("chest_mega.jpg correctly detected as ChestScreen")
		}
	})
}

func Example() {
	fmt.Println("OCR/CV Verification Suite ready")
}
