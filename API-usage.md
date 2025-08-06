### API Usage 

Search non existing stream (should be rejected)
```curl
curl http://localhost:8080/search/1
```

Create a flight search request
```curl
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{
    "from": "CGK",
    "to": "DPS",
    "date": "2025-07-10",
    "passengers": 2
}'
```
