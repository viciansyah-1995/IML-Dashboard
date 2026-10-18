package main

import (
	"context"
	"iml/internal/config"
	"iml/internal/database"
	"iml/internal/demo"
	"log"
)

func main() {
	c, e := config.Load()
	if e != nil {
		log.Fatal(e)
	}
	if !c.Demo {
		log.Fatal("seed demo opt-in: DEMO_MODE=true required")
	}
	ctx := context.Background()
	p, e := database.Open(ctx, c.DatabaseURL)
	if e != nil {
		log.Fatal(e)
	}
	defer p.Close()
	if e = demo.Seed(ctx, p, c.UploadDir, ".local/demo-credentials.json"); e != nil {
		log.Fatal(e)
	}
	log.Print("seed demo selesai; login lokal tersimpan di .local/demo-credentials.json (0600)")
}
