# curl -b cookies.txt -X POST http://localhost:8080/api/posts/1/comments \
#   -H "Content-Type: application/json" \
#   -d '{
#     "content": "This is my first comment"
#   }'

# curl -b cookies.txt http://localhost:8080/api/posts/1/comments

# INVALID POST ID
# curl -b cookies.txt http://localhost:8080/api/posts/abc/comments

# EMPTY COMMENTS
# curl -b cookies.txt -X POST http://localhost:8080/api/posts/1/comments \
#   -H "Content-Type: application/json" \
#   -d '{
#     "content": "   "
#   }'

# NOT LOGGED IN
# curl http://localhost:8080/api/posts/1/comments