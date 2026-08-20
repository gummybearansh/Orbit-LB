package handlers

import (
	"net"
	"net/http"
	"fmt"
	"time"
	"dashboard/internal/cluster"
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

	// return the string that the server has been spawned 
	html := `<div class="bg-slate-800 border-l-4 border-emerald-500 text-emerald-400 p-4 mt-4 rounded shadow-md font-mono" hx-get="/health?port=%s" hx-trigger="every 2s" hx-swap="outerHTML">
		Server :%s [ALIVE]
	</div>`
	fmt.Fprintf(w, html, port, port)
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
		html := `<div class="bg-slate-800 border-l-4 border-red-500 text-red-400 p-4 mt-4 rounded shadow-md font-mono" hx-get="/health?port=%s" hx-trigger="every 2s" hx-swap="outerHTML">
			Server :%s [DEAD]
		</div>`
		fmt.Fprintf(w, html, port, port)
		return
	} 
	// SUCCESS: The server is alive. 
	conn.Close() // Immediately close the socket so we don't leak memory

	// Return the GREEN card
	html := `<div class="bg-slate-800 border-l-4 border-emerald-500 text-emerald-400 p-4 mt-4 rounded shadow-md font-mono" hx-get="/health?port=%s" hx-trigger="every 2s" hx-swap="outerHTML">
		Server :%s [ALIVE]
	</div>`
	fmt.Fprintf(w, html, port, port)
}
