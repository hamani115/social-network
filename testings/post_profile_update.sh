curl -b user1_cookies.txt http://localhost:8080/api/profile/me

curl -b user1_cookies.txt http://localhost:8080/api/profiles/4

curl -b user1_cookies.txt http://localhost:8080/api/profiles/4/posts

curl -b user1_cookies.txt -X PUT http://localhost:8080/api/profile/me \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "Hamani",
    "about_me": "Computer Engineering and Machine Learning student.",
    "is_public": true
  }'

curl -b user1_cookies.txt http://localhost:8080/api/profile/me
