package server

import (
	"database/sql"
	"log"
	"net/http"

	dbsqlite "social-network/backend/pkg/db/sqlite"
)

var db *sql.DB

func Run(addr string) error {
	var err error

	dbPath := "social.db"
	migrationsPath := "pkg/db/migrations/sqlite"

	err = dbsqlite.RunMigrations(dbPath, migrationsPath)
	if err != nil {
		log.Fatal("migration error:", err)
	}

	db, err = dbsqlite.Open(dbPath)
	if err != nil {
		log.Fatal("database error:", err)
	}
	defer db.Close()

	router := http.NewServeMux()

	router.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	
	router.HandleFunc("/api/register", registerHandler)
	router.HandleFunc("/api/login", loginHandler)
	router.HandleFunc("/api/logout", logoutHandler)
	router.HandleFunc("/api/me", authMiddleware(meHandler))
	router.HandleFunc("/api/posts", authMiddleware(postsHandler))
	router.HandleFunc("/api/posts/", authMiddleware(postSubroutesHandler))

	log.Printf("Backend running on http://localhost%s\n", addr)
	// err = http.ListenAndServe(":8080", corsMiddleware(router))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	return http.ListenAndServe(addr, corsMiddleware(router))
}