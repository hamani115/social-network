curl -b user1_cookies.txt -X POST http://localhost:8080/api/posts/4/comments \
  -H "Content-Type: application/json" \
  -d '{
    "content": "I can comment on this post because I follow you"
  }'

curl -b user1_cookies.txt -X POST http://localhost:8080/api/users/4/unfollow

curl -b user1_cookies.txt http://localhost:8080/api/posts/4/comments