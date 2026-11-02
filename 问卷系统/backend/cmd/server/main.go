package main

import (
	"log"

	"lineng/configs"
	"lineng/db"
	"lineng/internal/handler"
	"lineng/internal/router"
)

func main() {
	cfg := configs.Load()

	database := db.Init(cfg)
	db.SeedAdmin(database)

	r := router.Setup(handler.New(database, cfg.H5Secret))

	log.Printf("[lineng] 后端已启动: http://127.0.0.1%s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
