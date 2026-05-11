curl -i -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "hamani115@gmail.com",
    "password": "123456"
  }'

curl -b cookies.txt http://localhost:8080/api/me