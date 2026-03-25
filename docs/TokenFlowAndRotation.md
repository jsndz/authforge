User login -> check -> give access token and refresh token 
-> access token valid for 15 min 
-> so 15 min no worries
-> after that needs new token
-> send refresh token from cookie through /refresh
-> create access token and delete  old refresh token 
-> send and repeat


deleting helps in stoping reuse of token


We hash the token cause if redis gets compromised that wont get them tokens since you can get token from hashed token
so we store hashed token
and raw token for the user