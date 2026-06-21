package server

// auth
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

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
}

// posts
type CreatePostRequest struct {
	Content        string `json:"content"`
	Privacy        string `json:"privacy"`
	AllowedUserIDs []int  `json:"allowed_user_ids"`
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

// comments
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

// followers
type FollowUserResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type UserListItem struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
}

type FollowRequestResponse struct {
	ID            int    `json:"id"`
	RequesterID   int    `json:"requester_id"`
	RequesterName string `json:"requester_name"`
	RequesterNick string `json:"requester_nickname"`
	TargetID      int    `json:"target_id"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type UserWithFollowStatus struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Nickname     string `json:"nickname"`
	IsPublic     bool   `json:"is_public"`
	FollowStatus string `json:"follow_status"`
}

// profile
type ProfileResponse struct {
	ID             int    `json:"id"`
	Email          string `json:"email"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"`
	AvatarPath     string `json:"avatar_path"`
	Nickname       string `json:"nickname"`
	AboutMe        string `json:"about_me"`
	IsPublic       bool   `json:"is_public"`
	IsOwner        bool   `json:"is_owner"`
	CanViewProfile bool   `json:"can_view_profile"`
	FollowStatus   string `json:"follow_status"`
	FollowersCount int    `json:"followers_count"`
	FollowingCount int    `json:"following_count"`
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	AboutMe  string `json:"about_me"`
	IsPublic *bool  `json:"is_public"`
}

// notification
type NotificationResponse struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	RequesterID int    `json:"requester_id"`
	Type        string `json:"type"`
	Message     string `json:"message"`
	LinkPath    string `json:"link_path"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
}

// groups
type CreateGroupRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GroupResponse struct {
	ID               int    `json:"id"`
	CreatorID        int    `json:"creator_id"`
	CreatorName      string `json:"creator_name"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	MemberCount      int    `json:"member_count"`
	MembershipStatus string `json:"membership_status"`
	CreatedAt        string `json:"created_at"`
}

type GroupJoinRequestResponse struct {
	ID                int    `json:"id"`
	GroupID           int    `json:"group_id"`
	RequesterID       int    `json:"requester_id"`
	RequesterName     string `json:"requester_name"`
	RequesterNickname string `json:"requester_nickname"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
}

type CreateGroupInvitationRequest struct {
	InviteeID int `json:"invitee_id"`
}

type GroupInvitationResponse struct {
	ID              int    `json:"id"`
	GroupID         int    `json:"group_id"`
	GroupTitle      string `json:"group_title"`
	InviterID       int    `json:"inviter_id"`
	InviterName     string `json:"inviter_name"`
	InviteeID       int    `json:"invitee_id"`
	InviteeName     string `json:"invitee_name"`
	InviteeNickname string `json:"invitee_nickname"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
}

type CreateGroupPostRequest struct {
	Content string `json:"content"`
}

type GroupPostResponse struct {
	ID             int    `json:"id"`
	GroupID        int    `json:"group_id"`
	UserID         int    `json:"user_id"`
	AuthorName     string `json:"author_name"`
	AuthorNickname string `json:"author_nickname"`
	Content        string `json:"content"`
	ImagePath      string `json:"image_path"`
	CreatedAt      string `json:"created_at"`
}

type CreateGroupCommentRequest struct {
	Content string `json:"content"`
}

type GroupCommentResponse struct {
	ID             int    `json:"id"`
	GroupPostID    int    `json:"group_post_id"`
	UserID         int    `json:"user_id"`
	AuthorName     string `json:"author_name"`
	AuthorNickname string `json:"author_nickname"`
	Content        string `json:"content"`
	ImagePath      string `json:"image_path"`
	CreatedAt      string `json:"created_at"`
}

// group events
type CreateGroupEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventTime   string `json:"event_time"`
}

type GroupEventResponse struct {
	ID            int    `json:"id"`
	GroupID       int    `json:"group_id"`
	CreatorID     int    `json:"creator_id"`
	CreatorName   string `json:"creator_name"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	EventTime     string `json:"event_time"`
	GoingCount    int    `json:"going_count"`
	NotGoingCount int    `json:"not_going_count"`
	MyResponse    string `json:"my_response"`
	CreatedAt     string `json:"created_at"`
}
