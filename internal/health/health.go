package health

import (
	"fmt"
	"log"
	"net/http"
)


func Start(port string) {
	log.Print("starting health...")
	http.HandleFunc("/", handler)

	// Determine port for HTTP service.
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP health.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, request *http.Request) {
	fmt.Fprint(w, "OK!\n")
}
