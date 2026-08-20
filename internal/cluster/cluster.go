package cluster

import (
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
}

var (
	NextPort = 8000
	PortMutex sync.Mutex

	Registry = make(map[string]*Node)
	RegistryMutex sync.RWMutex
)

func GetPort() string {
	PortMutex.Lock()
	defer PortMutex.Unlock()

	NextPort += 1 
	return strconv.Itoa(NextPort)
}


func SpawnServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
	})

	port = ":" + port

	server := &http.Server{
		Addr:         port,           // Set the network address (TCP)
		Handler: mux,
	}

	// i would cancel this timer / reset the timer if there is a call made to this server
	time.AfterFunc(1 * time.Minute, func () {
		// need to make sure this server's node says it's dead
		RegistryMutex.Lock()
		Registry[port].Status = false
		RegistryMutex.Unlock()
		server.Close()
	}) 

	// first lock the mutex 
	RegistryMutex.Lock()
	// get it from the active servers and activate it / create it 
	registryServer, ok := Registry[port]
	if !ok {
		// first time connecting to this server 
		Registry[port] = &Node{
			Port: port,
			Status: true,	
			RequestsHandled: 0,
		}
	}

	registryServer.Status = true

	// unlock manually (this functino never returns)
	RegistryMutex.Unlock()

	// blocking function - will keep the goroutine alive for 10 minutes 
	server.ListenAndServe()
}
