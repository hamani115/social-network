curl -i -c user3_cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "hasan123@gmail.com",
    "password": "123123"
  }'

curl -b user3_cookies.txt http://localhost:8080/api/me