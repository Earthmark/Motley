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
	"github.com/Earthmark/Motley/server/core/spa"
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

	conf, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config file, %v", err)
	}

	resolver, err := core.CreateResolver(conf)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/playground", handler.Playground("GraphQL playground", "/api"))
	http.Handle("/api", handler.GraphQL(gen.NewExecutableSchema(gen.Config{Resolvers: resolver})))
	http.Handle("/", http.FileServer(spa.SpaFileSystem(core.Client)))

	log.Printf("http://localhost:%d for Motley client", conf.Port)
	log.Printf("http://localhost:%d/playground for GraphQL playground", conf.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil))
}
