package web

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chessbot/Code/assets"
	"chessbot/Code/history"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		next.ServeHTTP(w, r)
	})
}

func StartWebServer(hub *WSHub, port int, figuresDir string) {
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

	http.Handle("/figures/", cacheControlMiddleware(http.StripPrefix("/figures/", fileHandler)))

	// Serve static assets (CSS, etc.) from embedded staticFS or local folder
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err == nil {
		http.Handle("/static/", cacheControlMiddleware(http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS)))))
	} else {
		http.Handle("/static/", cacheControlMiddleware(http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join("Code", "web", "static"))))))
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
			http.Error(w, "WebSocket required", http.StatusBadRequest)
			return
		}

		key := r.Header.Get("Sec-WebSocket-Key")
		h := sha1.New()
		h.Write([]byte(key + wsGUID))
		acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijack failed", http.StatusInternalServerError)
			return
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			return
		}
		defer conn.Close()

		res := "HTTP/1.1 101 Switching Protocols\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
		bufrw.WriteString(res)
		bufrw.Flush()

		msgChan := make(chan []byte, 2)
		hub.Register(msgChan)
		defer hub.Unregister(msgChan)

		go func() {
			buf := make([]byte, 512)
			for {
				if _, err := conn.Read(buf); err != nil {
					return
				}
			}
		}()

		for msg := range msgChan {
			if _, err := bufrw.Write(msg); err != nil {
				break
			}
			if err := bufrw.Flush(); err != nil {
				break
			}
		}
	})

	history.RegisterHistoryAPI(http.DefaultServeMux, filepath.Join("DATA", "History"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/history" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, history.HistoryViewerHTML)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, DuolingoBoardHTML)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[Web UI] Dashboard available at: http://localhost%s", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()
}
