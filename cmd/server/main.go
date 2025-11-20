package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	httpserver "ecommerce/internal/http"
	"ecommerce/internal/store"
)

func main() {
	nthOrder := 3
	if fromEnv := os.Getenv("NTH_ORDER_DISCOUNT"); fromEnv != "" {
		if v, err := strconv.Atoi(fromEnv); err == nil && v > 0 {
			nthOrder = v
		}
	}

	store := store.NewMemoryStore(nthOrder)
	srv := httpserver.NewServer(store)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}

	log.Printf("Starting server on %s (nth order discount: %d)\n", addr, nthOrder)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
