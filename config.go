package main

import (
	"log"
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

	coursesDir := os.Getenv("COURSES_DIR")
	if coursesDir == "" {
		coursesDir = "courses"
	}

	viewsDir := os.Getenv("VIEWS_DIR")
	if viewsDir == "" {
		viewsDir = "views"
	}

	config := &Config{
		Port:       port,
		CoursesDir: coursesDir,
		ViewsDir:   viewsDir,
	}

	// Validate directories exist
	if _, err := os.Stat(config.CoursesDir); os.IsNotExist(err) {
		log.Printf("Warning: courses directory '%s' does not exist", config.CoursesDir)
	}

	if _, err := os.Stat(config.ViewsDir); os.IsNotExist(err) {
		log.Printf("Warning: views directory '%s' does not exist", config.ViewsDir)
	}

	return config
}
