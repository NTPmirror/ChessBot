package recognizer

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"chessbot/Code/assets"
)

type PieceTemplate struct {
	Name      string
	BasePiece string
	Grays     [32 * 32]float64
	BgLum     float64
}

func ToBasePiece(tmplName string) string {
	if len(tmplName) < 2 {
		return tmplName
	}
	color := tmplName[0]
	piece := tmplName[1]
	if (color == 'w' || color == 'b') && strings.ContainsRune("PNBRQK", rune(piece)) {
		return string(color) + string(piece)
	}
	return tmplName
}

func CalcCornerBg(grays [32 * 32]float64) float64 {
	return (grays[0] + grays[31] + grays[31*32] + grays[31*32+31]) / 4.0
}

func LoadTemplates(figuresDir string) ([]PieceTemplate, error) {
	if figFS, err := assets.GetFiguresFS(); err == nil {
		if tmpls, err := loadFromFS(figFS); err == nil && len(tmpls) > 0 {
			log.Printf("Loaded %d embedded piece templates", len(tmpls))
			return tmpls, nil
		}
	}

	entries, err := os.ReadDir(figuresDir)
	if err != nil {
		return nil, fmt.Errorf("directory %s not found: %w", figuresDir, err)
	}

	var tmpls []PieceTemplate
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".png") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(e.Name()), "screen") {
			continue
		}

		fPath := filepath.Join(figuresDir, e.Name())
		f, err := os.Open(fPath)
		if err != nil {
			continue
		}

		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}

		rawName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		basePiece := ToBasePiece(rawName)
		grays := To32x32Grayscale(img)

		tmpls = append(tmpls, PieceTemplate{
			Name:      rawName,
			BasePiece: basePiece,
			Grays:     grays,
			BgLum:     CalcCornerBg(grays),
		})
	}

	if len(tmpls) == 0 {
		return nil, fmt.Errorf("no valid piece templates found in %s", figuresDir)
	}

	log.Printf("Loaded %d piece templates from disk", len(tmpls))
	return tmpls, nil
}

func loadFromFS(fileSystem fs.FS) ([]PieceTemplate, error) {
	entries, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return nil, err
	}

	var tmpls []PieceTemplate
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".png") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(e.Name()), "screen") {
			continue
		}

		f, err := fileSystem.Open(e.Name())
		if err != nil {
			continue
		}

		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}

		rawName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		basePiece := ToBasePiece(rawName)
		grays := To32x32Grayscale(img)

		tmpls = append(tmpls, PieceTemplate{
			Name:      rawName,
			BasePiece: basePiece,
			Grays:     grays,
			BgLum:     CalcCornerBg(grays),
		})
	}

	return tmpls, nil
}
