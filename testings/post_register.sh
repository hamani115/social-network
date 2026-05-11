curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "hamani115@gmail.com",
    "password": "123456",
    "first_name": "Abdulrahman",
    "last_name": "Almarzouqi",
    "date_of_birth": "2002-01-26"
  }'