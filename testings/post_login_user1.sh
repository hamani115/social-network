curl -i -c user1_cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "hamani115@gmail.com",
    "password": "123456"
  }'

curl -b user1_cookies.txt http://localhost:8080/api/me