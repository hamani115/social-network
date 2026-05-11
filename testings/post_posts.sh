curl -b cookies.txt -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -d '{
    "content": "My first post",
    "privacy": "public"
  }'

curl -b cookies.txt http://localhost:8080/api/posts