package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
)

// Winsize represents the size of the terminal window
type Winsize struct {
	Rows   uint16
	Cols   uint16
	Xpixel uint16
	Ypixel uint16
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./wbfy [command] [args...]")
		os.Exit(1)
	}

	// Create and start the command with PTY
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatal("Failed to start PTY:", err)
	}
	defer func() { _ = ptmx.Close() }()

	// Setup HTTP server
	http.Handle("/", http.FileServer(http.Dir("web")))

	// WebSocket endpoint for terminal I/O
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("New WebSocket connection from:", r.RemoteAddr)

		// Upgrade HTTP connection to WebSocket
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer ws.Close()

		// Create channel to signal when PTY session ends
		done := make(chan struct{})

		// PTY → WebSocket: Read from PTY and send to WebSocket
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := ptmx.Read(buf)
				if err != nil {
					log.Println("PTY read error:", err)
					break
				}

				if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
					log.Println("WebSocket write error:", err)
					break
				}
			}

			// Signal that PTY session has ended
			close(done)
			log.Println("PTY read loop ended for client:", r.RemoteAddr)
		}()

		// WebSocket → PTY: Read from WebSocket and write to PTY
		go func() {
			for {
				msgType, msg, err := ws.ReadMessage()
				if err != nil {
					log.Println("WebSocket read error:", err)
					break
				}

				// Debug input content
				if len(msg) > 0 {
					log.Printf("Received input: %d bytes", len(msg))
				}

				// Handle resize messages sent as special string format: "RESIZE:cols,rows"
				// This is a simple approach that avoids needing json parsing
				if msgType == websocket.TextMessage && len(msg) > 7 && string(msg[:7]) == "RESIZE:" {
					// Extract cols and rows
					var cols, rows uint16
					_, err := fmt.Sscanf(string(msg[7:]), "%d,%d", &cols, &rows)
					if err == nil && cols > 0 && rows > 0 {
						ws := &Winsize{
							Rows: rows,
							Cols: cols,
						}

						// TIOCSWINSZ is the ioctl request code for setting window size
						const TIOCSWINSZ = 0x5414
						syscall.Syscall(
							syscall.SYS_IOCTL,
							ptmx.Fd(),
							uintptr(TIOCSWINSZ),
							uintptr(unsafe.Pointer(ws)),
						)
						log.Printf("Resized terminal to %dx%d", cols, rows)
						continue
					}
				}

				// Otherwise, treat as normal input
				if _, err := ptmx.Write(msg); err != nil {
					log.Println("PTY write error:", err)
					break
				}
			}
		}()

		// We'll handle the resize in the main WebSocket input loop to avoid conflicts

		// Wait for PTY session to end
		<-done

		// Send terminal closed message
		ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[33m[ Terminal session ended ]\x1b[0m\r\n"))
	})

	// Open browser automatically
	go func() {
		url := "http://localhost:8080"
		log.Println("Opening browser at:", url)
		if err := browser.OpenURL(url); err != nil {
			log.Println("Failed to open browser:", err)
		}
	}()

	// Start HTTP server
	fmt.Println("Serving at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
