package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

// Embeds the contents of the "public" directory
//
//go:embed public
var staticFiles embed.FS

// main entry point
func main() {
	env, port := GetEnv()
	// Serve static files based on environment
	// In development, serve from local filesystem
	// In production, serve from embedded files
	if env == Dev {
		log.Println("⚙️  Running in development mode")
		http.Handle("/", http.FileServer(http.Dir("./public")))
	} else if env == Prod {
		log.Println("🚀 Running in production mode")
		publicFS, _ := fs.Sub(staticFiles, "public")
		http.Handle("/", http.FileServer(http.FS(publicFS)))
	} else {
		log.Fatal("Please set ENV to 'development' or 'production'")
	}

	http.HandleFunc("/api/dry-run", func(w http.ResponseWriter, r *http.Request) {
		run(w, true)
	})

	http.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		run(w, false)
	})

	log.Printf("✅ UI available at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
