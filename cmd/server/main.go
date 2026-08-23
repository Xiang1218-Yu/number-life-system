package main

import (
	"flag"
	"log"
	"number-life-system/api"
	"number-life-system/config"
	"number-life-system/internal/store"
)

func main() {
	configPath := flag.String("config", "./config/config.yaml", "配置文件路径")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	db, err := store.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close(db)
	router := api.NewRouter(db, cfg.Server.JWTSecret, cfg.Server.CORSOrigin)
	log.Printf("digital life server listening on :%s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
