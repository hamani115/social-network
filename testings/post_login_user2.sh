curl -i -c user2_cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rashedbucheery@gmail.com",
    "password": "123456"
  }'

curl -b user2_cookies.txt http://localhost:8080/api/me