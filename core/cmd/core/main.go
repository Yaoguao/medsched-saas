package main

import (
	"base/internal/config"
	"base/internal/lib/logger/sl"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("config loaded", "cfg", cfg)

}
