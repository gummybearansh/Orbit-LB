package cluster

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	Port string
	Status bool
	RequestsHandled int
	CreatedAt time.Time
	CurrentRPS int
	RequestsThisSec int
	History []int
}

// global variables
var (
	// handles which server request should go to
	NextPort = 8000
	PortMutex sync.Mutex

	// handles which servers are active
	RegistryMutex sync.RWMutex
	Registry = make(map[string]*Node)
	// active array for Round Robin Load balancing 
	ActivePorts []string
)

func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			RegistryMutex.Lock()
			for _, node := range Registry {
				if node.Status {
					// Push to history
					node.History = append(node.History, node.RequestsThisSec)
					// Keep history to 20 data points (for sparkline)
					if len(node.History) > 20 {
						node.History = node.History[1:]
					}
					node.CurrentRPS = node.RequestsThisSec
					node.RequestsThisSec = 0
				}
			}
			RegistryMutex.Unlock()
		}
	}()
}

func GetPort() (string, error) {
	PortMutex.Lock()
	defer PortMutex.Unlock()


	// Read lock the registry to safely check if ports are active 
	RegistryMutex.RLock()
	defer RegistryMutex.RUnlock()

	portString := ":" + strconv.Itoa(NextPort)

	// first check if all servers are live 
	if len(ActivePorts) == 1000 {
		return "", errors.New("Max server limit reached")
	}

	// is this port active - then increment
	for {
		NextPort += 1	
		// last port - but we've already checked if all are active 
		// so there must be a gap
		if (NextPort >= 9000) {
			NextPort = 8000 
		}
		portString = ":" + strconv.Itoa(NextPort)
		node, exists := Registry[portString]

		// If the node doesn't exist, OR it exists but is dead, this port is free!
		if !exists || !node.Status {
			return portString, nil
		}
	}

}


func SpawnServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// lock and grab the value first
		RegistryMutex.Lock()
		var currentReqs int
		if node, exists := Registry[port]; exists {
			node.RequestsHandled++
			node.RequestsThisSec++
			currentReqs = node.RequestsHandled
		}
		RegistryMutex.Unlock() // Unlock instantly!
		// send simple 200 response
		fmt.Fprintf(w, "Node %s | Total Handled: %d", port, currentReqs)
	})

	server := &http.Server{
		Addr:         port,           // Set the network address (TCP)
		Handler: mux,
	}

	// i would cancel this timer / reset the timer if there is a call made to this server
	// time.AfterFunc(1 * time.Minute, func () {
	time.AfterFunc(20 * time.Second, func(){
		// need to make sure this server's node says it's dead
		RegistryMutex.Lock()
		Registry[port].Status = false

		// also remove it from the active ports 
		// find it via a loop, and then use slices to delete it (No new memory allocation but O(N) op)
		for i, v := range ActivePorts {
			if v == port {
				// move everything after 'i' to before i - Go's way of deleting an element
				ActivePorts = append(ActivePorts[:i], ActivePorts[i+1:]...)
				break
			}
		}

		RegistryMutex.Unlock()
		server.Close()
	}) 

	// first lock the mutex 
	RegistryMutex.Lock()
	// get it from the active servers and activate it / create it 
	registryServer, ok := Registry[port]
	if !ok {
		// first time connecting to this server 
		registryServer = &Node{
			Port: port,
			Status: true,	
			RequestsHandled: 0,
			CreatedAt: time.Now(),
			History: make([]int, 0, 20),
		}
		Registry[port] = registryServer
	} else {
		registryServer.CreatedAt = time.Now()
		registryServer.History = make([]int, 0, 20)
		registryServer.CurrentRPS = 0
		registryServer.RequestsThisSec = 0
	}
	// also make sure it's added to the active servers 
	ActivePorts = append(ActivePorts, port)

	registryServer.Status = true

	// unlock manually (this functino never returns)
	RegistryMutex.Unlock()

	// blocking function - will keep the goroutine alive for 10 minutes 
	server.ListenAndServe()
}
