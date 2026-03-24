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

Refresh token need to be sent as a cookie


Will blacklist access_token since they are stateless and can't be tracked 
on logput just say that they are blacklisted



Authentication also includes middleware for some endpoints also called as protected routes
Like /me only logged in user should be able to get in.
Made a middleware for authentication



Validation in MICROSERVICES:

the auth service can't be called everytime you need to verify the user
so in microservice
2 ways to do it:

Gateway handles the verification(tightly coupled)
Each service has user authentication middleware(repeated code)
