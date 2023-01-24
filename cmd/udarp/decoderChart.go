package main

/*
This is used for debugging decode and show the individual R values from the fft.
This is mostly used for debugging right now, may be removed later.
*/

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/8ff/udarp/pkg/misc"
	"github.com/gorilla/websocket"
)

// Embed all files from static folder to serve over HTTP
//
//go:embed static
var staticWeb embed.FS

// TODO: THESE ARE FOR DEBUGGING
var clients = make([]*websocket.Conn, 0)
var upgrader = websocket.Upgrader{}
var lastFrame []byte

// Function that handles /ws websocket connections and store clients in global variable
func handleWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		misc.Log("error", fmt.Sprintf("Error upgrading connection: %s", err))
		return
	}

	// Add client to global variable
	clients = append(clients, conn)

	// Send lastFrame to client
	err = conn.WriteMessage(1, lastFrame)
	if err != nil {
		misc.Log("error", fmt.Sprintf("Error sending lastFrame to client: %s", err))
	}

	// Print number of clients
	misc.Log("info", fmt.Sprintf("Clients: %d", len(clients)))
}

// Function that serves http
func (conf *Config) serveHTTP() {
	fsys := fs.FS(staticWeb)
	html, _ := fs.Sub(fsys, "static")
	fs := http.FileServer(http.FS(html))

	http.Handle("/", fs)
	http.HandleFunc("/ws", handleWs)

	misc.Log("info", fmt.Sprintf("Starting http server on %s", conf.HTTP_Listen_Addr))
	http.ListenAndServe(conf.HTTP_Listen_Addr, nil)
}

// Function that broadcasts data to all clients
func bcastWs(data []byte) {
	for _, client := range clients {
		err := client.WriteMessage(1, data)
		if err != nil {
			misc.Log("error", fmt.Sprintf("%s", err))
			client.Close()
			// Delete client from clients slice
			for i, c := range clients {
				if c == client {
					clients = append(clients[:i], clients[i+1:]...)
				}
			}
		}
	}
}
