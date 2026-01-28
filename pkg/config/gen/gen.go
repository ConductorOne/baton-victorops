package main

import (
	cfg "github.com/conductorone/baton-victorops/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("victorops", cfg.Config)
}
