# add user
curl -X POST http://localhost:8080/api/v1/spenders \
     -H "Content-Type: application/json" \
     -d '{"name": "HongJot", "email": "hong@jot.ok"}' \
     -u user:secret

# get user
curl -X GET http://localhost:8080/api/v1/spenders \
     -u user:secret

# get user
curl -X GET http://localhost:8080/api/v1/spenders/2 \
     -u user:secret

# get categories
curl -X GET http://localhost:8080/api/v1/categories \
     -u user:secret

# curl -X POST http://localhost:8080/api/v1/upload \
# 	-H "Content-Type: multipart/form-data" \
# 	-F "images=@e-slip1.png" \
# 	-F "images=@e-slip2.png"

# post transaction
curl -X POST http://localhost:8080/api/v1/transactions \
     -H "Content-Type: application/json" \
     -d '{"date": "2024-04-30T09:00:00.000Z", "amount": 1000, "category": "Travel", "transaction_type": "expense", "note": "Lunch", "image_url": "https://example.com/image1.jpg", "spender_id": 1}' \
     -u user:secret
