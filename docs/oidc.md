authentication layer on top of OAuth2
OAuth → access (permissions)
OIDC → identity (who is the user)


OIDC solves the problem of identity of user by giving a standard, verifiable identity (ID token).

If you only use OAuth:

Client gets access token
Must call API to get user info
No standard identity format
Different implementations everywhere

✔ Clear identity
Know exactly which user
✔ Standard format
No custom user response
✔ Faster login
No extra API calls
✔ Secure
Signed JWT with claims
✔ Industry standard
Used everywhere


ID token is a part of oauth 
which helps to identify the user in oauth

 When ID token is needed

Only when:

client is using OAuth
or third-party integration
or “Login with X” flow


Generate a signed JWT (id_token) with user identity in /oauth/token and return it when openid scope is present.