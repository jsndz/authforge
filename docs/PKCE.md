PKCE (Proof Key for Code Exchange) prevents authorization code interception.

Client starts login → gets code
That code travels via browser (redirect)
Attacker (malicious app, proxy, injected script) steals the code
Attacker calls /token with that code
Gets access token → account compromise

PKCE adds a one-time secret proof tied to the code.

Client creates a random string called code_verifier

client derives challenge = >code_challenge = SHA256(code_verifier)
/oauth/authorize
send the challenge with method for server

server Stores code_challenge with the auth code
/oauth/token

sends original string random string as code_verifier

recomputes the code_challenge on method

compares with code_challenge

Attacker may steal code
But cannot guess code_verifier


/token POST request comes not as JSON but as form
Content-Type: application/x-www-form-urlencoded
So use ShouldBind which is general instead of shouldBindJSON


OAuth supports multiple flows:

authorization_code
refresh_token
client_credentials

Your /token endpoint must know which flow is being used.
to differentiate we use grant_type=authorization_code

and use proper structuring of redirect URL for avoiding misuses