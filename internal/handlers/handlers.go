package handlers

import (
	"dashboard/internal/cluster"
	"dashboard/ui"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

var (
	proxyCounter int
	proxyMutex sync.Mutex
)

func HandleSpawn(w http.ResponseWriter, r *http.Request){
	// get the port 
	port, err := cluster.GetPort()
	if err != nil {
		// can't create another server
		println(err)
		return
	}

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
	target_port := port // it's already a string
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


func HandleProxy(w http.ResponseWriter, r *http.Request){
	cluster.RegistryMutex.RLock()
	// first make sure that atleast one server is running 
	if len(cluster.ActivePorts) == 0 {
		cluster.RegistryMutex.RUnlock()
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// proxy counter math and unlock it immediately
	proxyMutex.Lock()
	targetIdx := proxyCounter % len(cluster.ActivePorts)
	proxyCounter++;
	proxyMutex.Unlock()

	// this has the ":8080" string
	targetPort := cluster.ActivePorts[targetIdx]

	// unlock the readlock
	cluster.RegistryMutex.RUnlock()

	// it spawns servers on server's own machine 
	serverIp := "http://127.0.0.1"

	targetUrl, err := url.Parse(serverIp + targetPort)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return;
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.ServeHTTP(w, r)
}
