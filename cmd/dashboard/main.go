package main

import (
	"fmt"
	"net/http"
	"dashboard/internal/handlers"
)


func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.HandleHome)

	mux.HandleFunc("/spawn", handlers.HandleSpawn)

	mux.HandleFunc("/health", handlers.HandleHealth)

	mux.HandleFunc("/blast", handlers.HandleBlast)

	mux.HandleFunc("/proxy", handlers.HandleProxy)

	// actual binding call 
	if err := http.ListenAndServe(":3000", mux); err != nil {
		fmt.Println("server failed: ", err)
	}
}
