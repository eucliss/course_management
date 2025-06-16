package main

import (
	"os"
)

type Config struct {
	Port       string
	CoursesDir string
	ViewsDir   string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:       port,
		CoursesDir: "courses",
		ViewsDir:   "views",
	}
}
