package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/99designs/gqlgen/handler"
	"github.com/Earthmark/Motley/server/config"
	"github.com/Earthmark/Motley/server/core"
	"github.com/Earthmark/Motley/server/gen"
)

func main() {
	configPath := flag.String("config", "motley.yaml", "The configuration file to load, or the path to create the config file at.")

	flag.Parse()

	path, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatalf("Failed to find config file path, given %s", *configPath)
	}
	log.Printf("Loading config %s", path)

	config, err := config.LoadOrInit(path)
	if err != nil {
		log.Fatal(err)
	}
	port := config.Port
	resolver, err := core.CreateResolver(path, *config)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/playground", handler.Playground("GraphQL playground", "/api"))
	http.Handle("/api", handler.GraphQL(gen.NewExecutableSchema(gen.Config{Resolvers: resolver})))
	http.Handle("/", http.FileServer(core.SpaFileSystem(core.Client)))

	log.Printf("http://localhost:%d for Motley client", port)
	log.Printf("http://localhost:%d/playground for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
