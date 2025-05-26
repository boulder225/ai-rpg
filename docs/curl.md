# 1. Check your current status
curl "http://localhost:8080/api/game/status?session_id=ebefdb45-a0e0-4007-8c2a-59f62ffa2181"

# 2. Look around your current location
curl -X POST http://localhost:8080/api/game/action \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "ebefdb45-a0e0-4007-8c2a-59f62ffa2181",
    "command": "/look around"
  }'

# 3. Move to the village
curl -X POST http://localhost:8080/api/game/action \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "ebefdb45-a0e0-4007-8c2a-59f62ffa2181",
    "command": "/move thornwick_village"
  }'

# 4. Talk to the tavern keeper
curl -X POST http://localhost:8080/api/game/action \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "ebefdb45-a0e0-4007-8c2a-59f62ffa2181",
    "command": "/talk tavern_keeper"
  }'