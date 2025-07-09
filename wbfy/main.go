package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // allow all origins (for dev)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./wbfy [command] [args...]")
	}
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}
	defer ptmx.Close()

	http.Handle("/", http.FileServer(http.Dir("web")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket error:", err)
			return
		}
		defer ws.Close()

		// PTY → WebSocket
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := ptmx.Read(buf)
				if err != nil {
					break
				}
				ws.WriteMessage(websocket.TextMessage, buf[:n])
			}
		}()

		// WebSocket → PTY
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				break
			}
			ptmx.Write(msg)
		}
	})

	fmt.Println("Open http://localhost:8080 in your browser")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
