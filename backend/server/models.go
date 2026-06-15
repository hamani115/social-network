package server

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Nickname    string `json:"nickname"`
	AboutMe     string `json:"about_me"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreatePostRequest struct {
	Content string `json:"content"`
	Privacy string `json:"privacy"`
}

type PostResponse struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	AuthorName     string `json:"author_name"`
	AuthorNickname string `json:"author_nickname"`
	Content        string `json:"content"`
	ImagePath      string `json:"image_path"`
	Privacy        string `json:"privacy"`
	CreatedAt      string `json:"created_at"`
}

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type CommentResponse struct {
	ID             int    `json:"id"`
	PostID         int    `json:"post_id"`
	UserID         int    `json:"user_id"`
	AuthorName     string `json:"author_name"`
	AuthorNickname string `json:"author_nickname"`
	Content        string `json:"content"`
	ImagePath      string `json:"image_path"`
	CreatedAt      string `json:"created_at"`
}
