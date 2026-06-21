curl -b user2_cookies.txt -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -d '{
    "content": "This post is for followers only",
    "privacy": "followers"
  }'

curl -b user2_cookies.txt http://localhost:8080/api/posts

curl -b user1_cookies.txt http://localhost:8080/api/posts

curl -b user1_cookies.txt -X POST http://localhost:8080/api/users/4/unfollow

curl -b user1_cookies.txt http://localhost:8080/api/posts
