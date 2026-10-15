package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"commit-archaeologist/internal/analyzer"
	"commit-archaeologist/internal/db"
	"commit-archaeologist/internal/git"
	"commit-archaeologist/internal/server"
)

func main() {
	port := flag.Int("port", 3456, "HTTP port for Commit Archaeologist API")
	dbPath := flag.String("db", "", "Path to SQLite database file (default: ~/.archeologist/archeologist.db)")
	demoDir := flag.String("demo-dir", "", "Directory for storing synthetic demo repos (default: ~/.archeologist/demo)")
	initialRepo := flag.String("repo", "", "Optional repository path to analyze immediately on startup")
	flag.Parse()

	// Resolve default directories in user home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	appDir := filepath.Join(homeDir, ".archeologist")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		log.Fatalf("Failed to create application directory: %v", err)
	}

	if *dbPath == "" {
		*dbPath = filepath.Join(appDir, "archeologist.db")
	}
	if *demoDir == "" {
		*demoDir = filepath.Join(appDir, "demo")
	}
	_ = os.MkdirAll(*demoDir, 0755)

	log.Printf("[Commit Archaeologist] Initializing persistence store at: %s", *dbPath)
	store, err := db.NewStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	gitExec, err := git.NewExecutor()
	if err != nil {
		log.Fatalf("Git initialization error: %v", err)
	}

	an := analyzer.NewAnalyzer(store, gitExec)
	srv := server.NewServer(store, gitExec, an, *demoDir)

	if *initialRepo != "" {
		go func() {
			log.Printf("[Commit Archaeologist] Analyzing initial repository: %s", *initialRepo)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			_, err := an.AnalyzeRepository(ctx, *initialRepo, false, func(stage string, current, total int, msg string) {
				log.Printf("  [%d/%d] %s: %s", current, total, stage, msg)
			})
			if err != nil {
				log.Printf("Warning: Initial repository analysis failed: %v", err)
			} else {
				log.Printf("[Commit Archaeologist] Initial repository analysis complete!")
			}
		}()
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srv,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("=====================================================")
		log.Printf("  Commit Archaeologist Backend Server v1.0.0         ")
		log.Printf("  Local API listening on: http://%s                 ", addr)
		log.Printf("  Persisting local metadata in: %s                  ", *dbPath)
		log.Printf("  Local-first | Safe Git Invocation | Zero External   ")
		log.Printf("=====================================================")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-sigChan
	log.Println("\n[Commit Archaeologist] Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
	log.Println("[Commit Archaeologist] Server stopped.")
}
