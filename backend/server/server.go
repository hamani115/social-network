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
	// auth
	router.HandleFunc("/api/register", registerHandler)     // POST
	router.HandleFunc("/api/login", loginHandler)           // POST
	router.HandleFunc("/api/logout", logoutHandler)         // GET (or any)
	router.HandleFunc("/api/me", authMiddleware(meHandler)) // GET (or any)
	// posts
	router.HandleFunc("/api/posts", authMiddleware(postsHandler)) // GET + POST
	// comments
	router.HandleFunc("/api/posts/", authMiddleware(postSubroutesHandler)) // GET + POST
	// followers
	router.HandleFunc("/api/users/", authMiddleware(userSubroutesHandler))            // GET + 2 POST
	router.HandleFunc("/api/follow-requests", authMiddleware(followRequestsHandler))  // GET
	router.HandleFunc("/api/follow-requests/", authMiddleware(followRequestsHandler)) // 2 POST
	router.HandleFunc("/api/users", authMiddleware(usersHandler))                     // 2 GET
	// profile
	router.HandleFunc("/api/profile/me", authMiddleware(myProfileHandler))       // GET + PUT
	router.HandleFunc("/api/profiles/", authMiddleware(profileSubroutesHandler)) // 2 GET
	// notification
	router.HandleFunc("/api/notifications", authMiddleware(notificationsHandler))          // GET
	router.HandleFunc("/api/notifications/", authMiddleware(notificationSubroutesHandler)) // 2 POSTS
	// groups
	router.HandleFunc("/api/groups", authMiddleware(groupsHandler))                      // GET + POST
	router.HandleFunc("/api/groups/", authMiddleware(groupSubroutesHandler))             // GET + POST, GET + POST, GET + POST
	router.HandleFunc("/api/group-invitations", authMiddleware(groupInvitationsHandler)) //
	router.HandleFunc("/api/group-invitations/", authMiddleware(groupInvitationsSubroutesHandler))

	log.Printf("Backend running on http://localhost%s\n", addr)
	// err = http.ListenAndServe(":8080", corsMiddleware(router))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	return http.ListenAndServe(addr, corsMiddleware(router))
}
