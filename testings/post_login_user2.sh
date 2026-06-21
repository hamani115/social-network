curl -i -c user2_cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "munther123@gmail.com",
    "password": "123123"
  }'

curl -b user2_cookies.txt http://localhost:8080/api/me