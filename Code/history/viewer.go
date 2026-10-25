package history

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"chessbot/Code/assets"
)

type MatchItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Date      string `json:"date"`
	GameType  string `json:"game_type"`
	MyTeam    string `json:"my_team"`
	Result    string `json:"result"`
	Frames    int    `json:"frames"`
	Path      string `json:"path"`
}

type DateGroup struct {
	Date    string      `json:"date"`
	Matches []MatchItem `json:"matches"`
}

func ListHistory(baseDir string) []DateGroup {
	var groups []DateGroup
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return groups
	}

	for _, entry := range entries {
		name := entry.Name()
		fullEntryPath := filepath.Join(baseDir, name)

		if entry.IsDir() {
			date := name
			dateDir := fullEntryPath
			files, err := os.ReadDir(dateDir)
			if err != nil {
				continue
			}

			var items []MatchItem
			for _, f := range files {
				if !strings.HasSuffix(f.Name(), ".json") {
					continue
				}
				fPath := filepath.Join(dateDir, f.Name())
				data, err := os.ReadFile(fPath)
				if err != nil {
					continue
				}

				var match MatchHistory
				if err := json.Unmarshal(data, &match); err != nil {
					continue
				}

				title := match.Title
				if title == "" {
					title = match.ID
				}

				items = append(items, MatchItem{
					ID:       match.ID,
					Title:    title,
					Date:     date,
					GameType: match.GameType,
					MyTeam:   match.MyTeam,
					Result:   match.Result,
					Frames:   len(match.Frames),
					Path:     filepath.ToSlash(filepath.Join(date, f.Name())),
				})
			}

			sort.Slice(items, func(i, j int) bool {
				return items[i].Path > items[j].Path
			})

			if len(items) > 0 {
				groups = append(groups, DateGroup{
					Date:    date,
					Matches: items,
				})
			}
		} else if strings.HasSuffix(name, ".json") {
			// Also support match files saved directly under baseDir
			data, err := os.ReadFile(fullEntryPath)
			if err != nil {
				continue
			}
			var match MatchHistory
			if err := json.Unmarshal(data, &match); err != nil {
				continue
			}
			title := match.Title
			if title == "" {
				title = match.ID
			}
			date := match.Date
			if date == "" {
				date = "Matches"
			}
			// Find existing group for date or create
			found := false
			for idx := range groups {
				if groups[idx].Date == date {
					groups[idx].Matches = append(groups[idx].Matches, MatchItem{
						ID:       match.ID,
						Title:    title,
						Date:     date,
						GameType: match.GameType,
						MyTeam:   match.MyTeam,
						Result:   match.Result,
						Frames:   len(match.Frames),
						Path:     filepath.ToSlash(name),
					})
					found = true
					break
				}
			}
			if !found {
				groups = append(groups, DateGroup{
					Date: date,
					Matches: []MatchItem{
						{
							ID:       match.ID,
							Title:    title,
							Date:     date,
							GameType: match.GameType,
							MyTeam:   match.MyTeam,
							Result:   match.Result,
							Frames:   len(match.Frames),
							Path:     filepath.ToSlash(name),
						},
					},
				})
			}
		}
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Date > groups[j].Date
	})

	for i := range groups {
		sort.Slice(groups[i].Matches, func(a, b int) bool {
			return groups[i].Matches[a].Path > groups[i].Matches[b].Path
		})
	}

	return groups
}

func RegisterHistoryAPI(mux *http.ServeMux, baseDir string) {
	mux.HandleFunc("/api/history/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		groups := ListHistory(baseDir)
		if groups == nil {
			groups = []DateGroup{}
		}
		_ = json.NewEncoder(w).Encode(groups)
	})

	mux.HandleFunc("/api/history/load", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("file")
		if path == "" || strings.Contains(path, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(baseDir, filepath.FromSlash(path))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	mux.HandleFunc("/api/history/rename", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			File     string `json:"file"`
			NewTitle string `json:"new_title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if req.File == "" || strings.Contains(req.File, "..") || strings.TrimSpace(req.NewTitle) == "" {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(baseDir, filepath.FromSlash(req.File))
		data, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		var match MatchHistory
		if err := json.Unmarshal(data, &match); err != nil {
			http.Error(w, "Corrupted file", http.StatusInternalServerError)
			return
		}

		match.Title = strings.TrimSpace(req.NewTitle)
		newData, err := json.MarshalIndent(match, "", "  ")
		if err != nil {
			http.Error(w, "Marshal error", http.StatusInternalServerError)
			return
		}

		if err := os.WriteFile(fullPath, newData, 0644); err != nil {
			http.Error(w, "Write error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
}

func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		next.ServeHTTP(w, r)
	})
}

func StartHistoryViewerServer(port int, figuresDir, historyDir string) {
	mux := http.NewServeMux()

	var fileHandler http.Handler
	webFiguresDir := filepath.Join(figuresDir, "WEB")
	if fi, err := os.Stat(webFiguresDir); err == nil && fi.IsDir() {
		fileHandler = http.FileServer(http.Dir(webFiguresDir))
	} else if figFS, err := fs.Sub(assets.EmbeddedFS, "embedded/Figures/WEB"); err == nil {
		fileHandler = http.FileServer(http.FS(figFS))
	} else if figFS, err := assets.GetFiguresFS(); err == nil {
		fileHandler = http.FileServer(http.FS(figFS))
	} else {
		fileHandler = http.FileServer(http.Dir(figuresDir))
	}

	mux.Handle("/figures/", cacheControlMiddleware(http.StripPrefix("/figures/", fileHandler)))

	RegisterHistoryAPI(mux, historyDir)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, HistoryViewerHTML)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[History Viewer] Server running at: http://localhost%s", addr)
	log.Printf("[History Viewer] Stored matches in: %s", historyDir)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
