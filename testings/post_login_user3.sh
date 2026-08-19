curl -i -c user3_cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "yusuf9196@gmail.com",
    "password": "123456"
  }'

curl -b user3_cookies.txt http://localhost:8080/api/me