curl -b cookies.txt -X POST http://localhost:8080/api/posts \
  -F "content=Post with image" \
  -F "privacy=public" \
  -F "image=@/path/to/image.png"

curl -b cookies.txt http://localhost:8080/api/posts