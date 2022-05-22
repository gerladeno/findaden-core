# homie-core

#### Authentication
homie-core uses JWT authorization
```
-H 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiZjdlYjVhM2ItZDlkMi0xMWVjLWFiYmQtMDI0MmFjMTUwMDAyIn0.oFHi0hoydk2-KRm5Ph8zBdOM6jx7cuvYV-0r2hYarZpk9ugLm4ThxQKvK67sB3vTrkJVy3FkqXtIV1ROEIzi6G3dwgsz4klPQ0xELhsQ_ClgPD5AimZgUqnlLdyvXLQtfNTNQg6MiyboeUBc-92fmXy04o7FAaF8OgnHTJA3pqHssNRTyE1d-TRbgDU7EQvJEbOX1p0H6_c3MvZrD2FujtWqftNnW_Ky0OYIplcnBYlDRkzi5HMcUwlJHjN-kY5i0eabybjEYfDkFaXtmz0zIy1-RxBOSI6rs_Si9dRnkrDylB1b4EvZBCGlW8cNyr3WXCRfGlZqNhmNg_gWTgQbyQ'
```

### Config
endpoint: /public/v1/config  

GET 
```json
{
  "data": {
    "uuid": "f7eb5a3b-d9d2-11ec-abbd-0242ac150002",
    "personal": {
      "uuid": "f7eb5a3b-d9d2-11ec-abbd-0242ac150002",
      "username": "chuvak",
      "avatar_link": "jopa.ru",
      "gender": 1,
      "age": 26
    },
    "criteria": {
      "uuid": "f7eb5a3b-d9d2-11ec-abbd-0242ac150002",
      "regions": [
        1,
        2,
        3
      ],
      "price_range": {
        "from": 35000,
        "to": 700000
      },
      "gender": 0,
      "age_range": {
        "from": 20,
        "to": 35
      }
    },
    "settings": {
      "uuid": "f7eb5a3b-d9d2-11ec-abbd-0242ac150002",
      "theme": 0
    }
  }
}
```

POST
```json
{
  "personal": {
    "username": "chuvak",
    "avatar_link": "jopa.ru",
    "gender": 1,
    "age": 26
  },
  "criteria": {
    "regions": [
      1,
      2,
      3
    ],
    "price_range": {
      "from": 35000,
      "to": 700000
    },
    "gender": 0,
    "age_range": {
      "from": 20,
      "to": 35
    }
  },
  "settings": {
    "theme": 0
  }
}
```

### Matches
```
GET /public/v1/matches?count=5
```

### Like
```
GET /public/v1/like/{uuid}?super=true
```

### Dislike
```
GET /public/v1/dislike/{uuid}
```

### Liked
```
GET /public/v1/liked?limit=10&offset=0
```

### Disliked
```
GET /public/v1/disliked?limit=10&offset=0
```

### Get list of chats
```
GET /public/v1/chats
```

### Start a chat
```
/public/v1/chat/{uuid}
```
