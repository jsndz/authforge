Use a hybrid approach with refresh tokens and access token

User logs in -> get access token( short lived) and refresh token(long-lived )(stored in server)

client send request -> Bearer ACCESS TOKEN -> token
token expires in 5/10 min

client sends:
Refresh token → /refresh endpoint
Checks refresh token in DB/Redis
Validates:
Exists?
Not expired?
Not revoked?



Access token expires -> Client sends refresh token
Server:
Validates refresh token -> Deletes old refresh token
Creates:
New access token
New refresh token
Sends both back
Every refresh -> rotation