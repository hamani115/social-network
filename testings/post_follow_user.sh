# sqlite3 social.db "UPDATE users SET is_public = 1 WHERE id = 4;"

curl -b user1_cookies.txt -X POST http://localhost:8080/api/users/4/follow