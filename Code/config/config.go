package config

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	IntervalMs         int     `json:"interval_ms"`
	WebPort            int     `json:"web_port"`
	EnableWeb          bool    `json:"enable_web"`
	YRatio             float64 `json:"y_ratio"`
	YRatioOscar        float64 `json:"y_ratio_oscar"`
	YRatioPeople       float64 `json:"y_ratio_people"`
	FiguresDir         string  `json:"figures_dir"`
	MaxErrorThreshold  float64 `json:"max_error_threshold"`
	StabilityThreshold int     `json:"stability_threshold"`
	AutoPlay           bool    `json:"auto_play"`
	AutoClaimXP        bool    `json:"auto_claim_xp"`
	AutoStartGame      bool    `json:"auto_start_game"`
	GameType           string  `json:"game_type"`
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
		IntervalMs:         10,
		WebPort:            5873,
		EnableWeb:          false,
		YRatio:             0.355,
		YRatioOscar:        0.355,
		YRatioPeople:       0.336,
		FiguresDir:         "./Figures",
		MaxErrorThreshold:  500.0,
		StabilityThreshold: 2,
		AutoPlay:           true,
		AutoClaimXP:        true,
		AutoStartGame:      true,
		GameType:           "People",
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

	return cfg
}
