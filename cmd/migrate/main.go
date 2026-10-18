package main

import (
	"context"
	"iml/internal/config"
	"iml/internal/database"
	"log"
	"os"
	"time"
)

func main() {
	c, e := config.Load()
	if e != nil {
		log.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p, e := database.Open(ctx, c.DatabaseURL)
	if e != nil {
		log.Fatal(e)
	}
	defer p.Close()
	down := len(os.Args) > 1 && os.Args[1] == "down"
	if e = database.Migrate(ctx, p, down); e != nil {
		log.Fatal(e)
	}
	log.Print("migration selesai")
}
