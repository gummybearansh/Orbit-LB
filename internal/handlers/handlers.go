package handlers

import (
	"dashboard/internal/cluster"
	"dashboard/ui"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
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

	// Optimistic UI Injection (few microseconds till the server launches - till then this works)
	optimisticNode := &cluster.Node{
		Port:            port,
		Status:          true,
		RequestsHandled: 0,
	}

	// 4. Render directly to the DOM
	card := ui.ServerNodeWrapper(optimisticNode)
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

	// Lock for writing and verify the node actually exists in memory
	cluster.RegistryMutex.Lock()
	node, exists := cluster.Registry[target_port]
	if !exists {
		// Zombie poll detected from an old browser session.
		cluster.RegistryMutex.Unlock()
		// Returning an empty response tells HTMX to delete the card from the UI
		return
	}

	if err != nil {
		// Failure - server is dead 
		// update the registry - this server's status 
		node.Status = false
	}  else {
		// SUCCESS: The server is alive. 
		conn.Close() // Immediately close the socket so we don't leak memory
		node.Status = true
	}
	cluster.RegistryMutex.Unlock()

	// Lock for reading and render
	cluster.RegistryMutex.RLock()
	defer cluster.RegistryMutex.RUnlock()

	// isOOB MUST be false here! This is a direct HTMX response.
	card := ui.ServerNodeInner(cluster.Registry[target_port])
	card.Render(r.Context(), w)
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


// this function will tick every second from the frontned
// with the count of the blast in the query param
func HandleBlast (w http.ResponseWriter, r *http.Request){
	// get the count from the query parameter 
	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return;
	}

	// wait group so function doesn't end until all 50 requests have been fired
	var wg sync.WaitGroup

	// instantly spins up 50 go routines
	for _ = range count {

		// goers up 50 times - 50 go routines spin up 
		// expects 50 go routines to decrement it by calling Done()
		wg.Add(1)

		// anonymous go routine 
		go func (){
			// each go routine will say it's done
			defer wg.Done()
			// hit the Load balanced endpoint
			resp, err := http.Get("http://127.0.0.1:3000/proxy")
			if err == nil {
				// manually close the connection if it succeeded
				resp.Body.Close()
			}
		}()
	}

	// don't let this function end until all go routines are done - make the waitgroup wait 
	wg.Wait()

	// Acquire Read Lock to safely pass the map to the template
	cluster.RegistryMutex.RLock()
	defer cluster.RegistryMutex.RUnlock()

	// 2. Render the OOB component directly to the socket
	countStr := r.URL.Query().Get("count")
	res := ui.BlastResult(countStr, cluster.Registry)
	res.Render(r.Context(), w)
}


func HandleHome(w http.ResponseWriter, r *http.Request){
	cluster.RegistryMutex.RLock()
	dashboard := ui.Dashboard(cluster.Registry)

	dashboard.Render(r.Context(), w)
	// unlock after the renderere has rendered it (loops through it so it needs read access too)
	cluster.RegistryMutex.RUnlock()
}
