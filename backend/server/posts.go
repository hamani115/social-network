package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func postsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		listPostsHandler(w, r)
		return
	}

	if r.Method == http.MethodPost {
		createPostHandler(w, r)
		return
	}

	errorJSON(w, "method not allowed", http.StatusMethodNotAllowed)
}

func listPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	// PAGINATION
	limit := 10

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {

		parsedLimit, err :=
			strconv.Atoi(rawLimit)

		if err != nil ||
			parsedLimit <= 0 ||
			parsedLimit > 50 {

			errorJSON(
				w,
				"invalid post limit",
				http.StatusBadRequest,
			)
			return
		}

		limit = parsedLimit
	}

	offset := 0

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {

		parsedOffset, err :=
			strconv.Atoi(rawOffset)

		if err != nil ||
			parsedOffset < 0 {

			errorJSON(
				w,
				"invalid post offset",
				http.StatusBadRequest,
			)
			return
		}

		offset = parsedOffset
	}

	// SORTING
	sortValue :=
		r.URL.Query().Get("sort")

	if sortValue == "" {
		sortValue = "newest"
	}

	orderDirection := "DESC"

	switch sortValue {
	case "newest":
		orderDirection = "DESC"

	case "oldest":
		orderDirection = "ASC"

	default:
		errorJSON(
			w,
			"invalid post sort",
			http.StatusBadRequest,
		)
		return
	}
	
	query := fmt.Sprintf(`
		SELECT
			posts.id,
			posts.user_id,
			users.first_name || ' ' ||
				users.last_name
				AS author_name,
			COALESCE(
				users.nickname,
				''
			) AS author_nickname,
			COALESCE(
				users.avatar_path,
				''
			) AS author_avatar_path,
			posts.content,
			COALESCE(
				posts.image_path,
				''
			) AS image_path,
			posts.privacy,
			posts.created_at
		FROM posts

		JOIN users
			ON users.id = posts.user_id

		WHERE (
			posts.privacy = 'public'

			OR posts.user_id = ?

			OR (
				posts.privacy = 'followers'
				AND EXISTS (
					SELECT 1
					FROM followers
					WHERE
						followers.follower_id = ?
						AND
						followers.following_id =
							posts.user_id
				)
			)

			OR (
				posts.privacy = 'private'
				AND EXISTS (
					SELECT 1
					FROM post_allowed_users
					WHERE
						post_allowed_users.post_id =
							posts.id
						AND
						post_allowed_users.user_id = ?
				)
			)
		)

		ORDER BY
			posts.created_at %s,
			posts.id %s

		LIMIT ?
		OFFSET ?
	`,
		orderDirection,
		orderDirection,
	)

	rows, err := db.Query(
		query,
		userID,
		userID,
		userID,
		limit+1,
		offset,
	)

	if err != nil {
		errorJSON(
			w,
			"could not load posts",
			http.StatusInternalServerError,
		)
		return
	}

	defer rows.Close()

	posts := []PostResponse{}

	for rows.Next() {
		var post PostResponse

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.AuthorName,
			&post.AuthorNickname,
			&post.AuthorAvatarPath,
			&post.Content,
			&post.ImagePath,
			&post.Privacy,
			&post.CreatedAt,
		)

		if err != nil {
			errorJSON(
				w,
				"could not read post data",
				http.StatusInternalServerError,
			)
			return
		}

		posts = append(
			posts,
			post,
		)
	}

	if err := rows.Err(); err != nil {
		errorJSON(
			w,
			"error while reading posts",
			http.StatusInternalServerError,
		)
		return
	}

	hasMore :=
		len(posts) > limit

	if hasMore {
		posts =
			posts[:limit]
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"posts": posts,

			"has_more": hasMore,

			"next_offset": offset + len(posts),
		},
	)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	var content string
	var privacy string
	var imagePath string
	var allowedUserIDs []int

	keepUploadedImage := false

	defer func() {
		if !keepUploadedImage {
			removeUploadedImage(imagePath)
		}
	}()

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		err := r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			errorJSON(w, "could not read form data", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(r.FormValue("content"))
		privacy = strings.TrimSpace(r.FormValue("privacy"))

		allowedUserRaw := strings.TrimSpace(r.FormValue("allowed_user_ids"))

		if allowedUserRaw != "" {
			err = json.Unmarshal([]byte(allowedUserRaw), &allowedUserIDs)
			if err != nil {
				errorJSON(w, "invalid allowed_user_ids", http.StatusBadRequest)
				return
			}
		}

		imagePath, err = saveUploadedImage(r, "image", "uploads/posts")
		if err != nil {
			errorJSON(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else { // ! WILL REMOVE LATER
		var req CreatePostRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			errorJSON(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		content = strings.TrimSpace(req.Content)
		privacy = strings.TrimSpace(req.Privacy)
		allowedUserIDs = req.AllowedUserIDs
	}

	if content == "" {
		errorJSON(w, "post content is required", http.StatusBadRequest)
		return
	}

	if privacy == "" {
		privacy = "public"
	}

	if privacy != "public" && privacy != "followers" && privacy != "private" {
		errorJSON(w, "invalid privacy value", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		errorJSON(
			w,
			"could not start transaction",
			http.StatusInternalServerError,
		)
		return
	}

	defer tx.Rollback()

	result, err := tx.Exec(`
	INSERT INTO posts (
		user_id,
		content,
		image_path,
		privacy
	)
	VALUES (?, ?, ?, ?)
`,
		userID,
		content,
		imagePath,
		privacy,
	)

	if err != nil {
		errorJSON(
			w,
			"could not create post",
			http.StatusInternalServerError,
		)
		return
	}

	postID, err := result.LastInsertId()
	if err != nil {
		errorJSON(
			w,
			"could not read post id",
			http.StatusInternalServerError,
		)
		return
	}

	if privacy == "private" {
		err = savePostAllowedUsersTx(
			tx,
			int(postID),
			userID,
			allowedUserIDs,
		)

		if err != nil {
			errorJSON(
				w,
				err.Error(),
				http.StatusBadRequest,
			)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		errorJSON(
			w,
			"could not create post",
			http.StatusInternalServerError,
		)
		return
	}

	keepUploadedImage = true

	writeJSON(
		w,
		http.StatusCreated,
		map[string]interface{}{
			"message": "post created successfully",
			"post_id": postID,
		},
	)
}

func savePostAllowedUsersTx(
	tx *sql.Tx,
	postID int,
	ownerID int,
	allowedUserIDs []int,
) error {
	if len(allowedUserIDs) == 0 {
		return nil
	}

	for _, allowedUserID := range allowedUserIDs {

		if allowedUserID == ownerID {
			continue
		}

		var followerCount int

		err := tx.QueryRow(`
			SELECT COUNT(*)
			FROM followers
			WHERE follower_id = ?
			  AND following_id = ?
		`,
			allowedUserID,
			ownerID,
		).Scan(&followerCount)

		if err != nil {
			return err
		}

		if followerCount == 0 {
			return fmt.Errorf(
				"selected user %d is not your follower",
				allowedUserID,
			)
		}

		_, err = tx.Exec(`
			INSERT OR IGNORE INTO post_allowed_users (
				post_id,
				user_id
			)
			VALUES (?, ?)
		`,
			postID,
			allowedUserID,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
