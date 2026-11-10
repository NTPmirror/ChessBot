package config

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ChessComConfig struct {
	FiguresDir    string  `json:"figures_dir"`
	YRatio        float64 `json:"y_ratio"`
	AutoStartGame bool    `json:"auto_start_game"`
	AutoDismiss   bool    `json:"auto_dismiss"`
}

type DuolingoConfig struct {
	FiguresDir    string  `json:"figures_dir"`
	YRatio        float64 `json:"y_ratio"`
	YRatioOscar   float64 `json:"y_ratio_oscar"`
	YRatioPeople  float64 `json:"y_ratio_people"`
	GameType      string  `json:"game_type"`
	AutoClaimXP   bool    `json:"auto_claim_xp"`
	AutoStartGame bool    `json:"auto_start_game"`
}

type Config struct {
	Service string `json:"service"`

	IntervalMs         int     `json:"interval_ms"`
	WebPort            int     `json:"web_port"`
	EnableWeb          bool    `json:"enable_web"`
	MaxErrorThreshold  float64 `json:"max_error_threshold"`
	StabilityThreshold int     `json:"stability_threshold"`
	AutoPlay           bool    `json:"auto_play"`

	Humanize       bool `json:"humanize"`
	MoveDelayMinMs int  `json:"move_delay_min_ms"`
	MoveDelayMaxMs int  `json:"move_delay_max_ms"`
	TapDelayMinMs  int  `json:"tap_delay_min_ms"`
	TapDelayMaxMs  int  `json:"tap_delay_max_ms"`

	Duolingo DuolingoConfig `json:"duolingo"`
	ChessCom ChessComConfig `json:"chesscom"`

	// Deprecated / backwards compatibility fields (for root config fallback)
	FiguresDir         string  `json:"figures_dir,omitempty"`
	YRatio             float64 `json:"y_ratio,omitempty"`
	YRatioOscar        float64 `json:"y_ratio_oscar,omitempty"`
	YRatioPeople       float64 `json:"y_ratio_people,omitempty"`
	GameType           string  `json:"game_type,omitempty"`
	AutoClaimXP        *bool   `json:"auto_claim_xp,omitempty"`
	AutoStartGame      *bool   `json:"auto_start_game,omitempty"`
}

func stripComments(data []byte) []byte {
	var out bytes.Buffer
	inString := false
	escape := false
	inLineComment := false
	inBlockComment := false

	n := len(data)
	for i := 0; i < n; i++ {
		ch := data[i]

		if inLineComment {
			if ch == '\n' || ch == '\r' {
				inLineComment = false
				out.WriteByte(ch)
			}
			continue
		}

		if inBlockComment {
			if ch == '*' && i+1 < n && data[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if inString {
			out.WriteByte(ch)
			if escape {
				escape = false
			} else if ch == '\\' {
				escape = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			out.WriteByte(ch)
			continue
		}

		if ch == '/' && i+1 < n {
			if data[i+1] == '/' {
				inLineComment = true
				i++
				continue
			} else if data[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}
		}

		out.WriteByte(ch)
	}

	return out.Bytes()
}

func LoadConfig(path string) Config {
	cfg := Config{
		Service:            "duolingo",
		IntervalMs:         10,
		WebPort:            5873,
		EnableWeb:          false,
		MaxErrorThreshold:  500.0,
		StabilityThreshold: 2,
		AutoPlay:           true,
		Humanize:           true,
		MoveDelayMinMs:     400,
		MoveDelayMaxMs:     1200,
		TapDelayMinMs:      120,
		TapDelayMaxMs:      260,
		Duolingo: DuolingoConfig{
			FiguresDir:    "./Figures/Duolingo",
			YRatio:        0.355,
			YRatioOscar:   0.355,
			YRatioPeople:  0.336,
			GameType:      "People",
			AutoClaimXP:   true,
			AutoStartGame: true,
		},
		ChessCom: ChessComConfig{
			FiguresDir:    "./Figures/ChessCom",
			YRatio:        0.16875,
			AutoStartGame: true,
			AutoDismiss:   true,
		},
	}

	actualPath := path
	if _, err := os.Stat(actualPath); err != nil {
		ext := filepath.Ext(path)
		altPath := strings.TrimSuffix(path, ext) + ".jsonc"
		if _, errAlt := os.Stat(altPath); errAlt == nil {
			actualPath = altPath
		}
	} else if strings.HasSuffix(path, ".json") {
		altPath := strings.TrimSuffix(path, ".json") + ".jsonc"
		if _, errAlt := os.Stat(altPath); errAlt == nil {
			actualPath = altPath
		}
	}

	data, err := os.ReadFile(actualPath)
	if err != nil {
		if os.IsNotExist(err) {
			raw, _ := json.MarshalIndent(cfg, "", "  ")
			_ = os.WriteFile(actualPath, raw, 0644)
			log.Printf("Created default config at %s", actualPath)
		}
		return cfg
	}

	cleanData := stripComments(data)
	if err := json.Unmarshal(cleanData, &cfg); err != nil {
		log.Printf("Config parse error in %s: %v", actualPath, err)
		return cfg
	}

	// Migrate / inherit top-level backwards compatible fields into DuolingoConfig if set
	if cfg.FiguresDir != "" && cfg.Duolingo.FiguresDir == "" {
		cfg.Duolingo.FiguresDir = cfg.FiguresDir
	}
	if cfg.Duolingo.FiguresDir == "" {
		cfg.Duolingo.FiguresDir = "./Figures/Duolingo"
	}
	if cfg.YRatio != 0 && cfg.Duolingo.YRatio == 0 {
		cfg.Duolingo.YRatio = cfg.YRatio
	}
	if cfg.YRatioOscar != 0 && cfg.Duolingo.YRatioOscar == 0 {
		cfg.Duolingo.YRatioOscar = cfg.YRatioOscar
	}
	if cfg.YRatioPeople != 0 && cfg.Duolingo.YRatioPeople == 0 {
		cfg.Duolingo.YRatioPeople = cfg.YRatioPeople
	}
	if cfg.GameType != "" && cfg.Duolingo.GameType == "" {
		cfg.Duolingo.GameType = cfg.GameType
	}
	if cfg.Duolingo.GameType == "" {
		cfg.Duolingo.GameType = "People"
	}
	if cfg.AutoClaimXP != nil {
		cfg.Duolingo.AutoClaimXP = *cfg.AutoClaimXP
	}
	if cfg.AutoStartGame != nil {
		cfg.Duolingo.AutoStartGame = *cfg.AutoStartGame
	}

	return cfg
}
