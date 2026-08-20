package handlers

import (
	"net"
	"net/http"
	"time"
	"dashboard/internal/cluster"
	"dashboard/ui"
)

func HandleSpawn(w http.ResponseWriter, r *http.Request){
	// get the port 
	port := cluster.GetPort()

	// spawn the actual server 
	go cluster.SpawnServer(port)
	// can't return value from goroutine??? 
	// if err != nil {
	// 	fmt.Fprintf(w, "<h2>Failed to spawn server</h2>")
	// }

	card := ui.ServerCard(port, true)
	// context of the request and render directly to responseWriter
	card.Render(r.Context(), w)
}


func HandleHealth(w http.ResponseWriter, r *http.Request){
	// get which port was sent in the url
	port := r.URL.Query().Get("port")	
	// nothing was sent / no port in query
	if port == "" {
		// no port was sent - do nothing 
		return
	}

	// try connecting to the server 
	target_port := ":" + port
	conn, err := net.DialTimeout("tcp", target_port, 1 * time.Second)

	if err != nil {
		// Failure - server is dead - return the dead card
		// Notice it STILL has the hx-get attributes, so it keeps polling! If the server comes back, it turns green again.
		
		failure_card := ui.ServerCard(port, false)
		failure_card.Render(r.Context(), w)
		return
	} 
	// SUCCESS: The server is alive. 
	conn.Close() // Immediately close the socket so we don't leak memory

	// Return the GREEN card
	success_card := ui.ServerCard(port, true)
	success_card.Render(r.Context(), w)
}
