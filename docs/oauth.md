It lets a client (app) access user data without sharing the user’s password.
Like get data of google user account without sharing google password to the application
Or without any external provider oauth works with your own system

How does oauth work?
Before that let's understand the current system.

Client -> connects to backend with user name and password. -> gets tokens -> send to client

this is totally ok.

But this cannot handle mutiple clients (web, mobile, CLI)
third party integration
token delegation like giving limited access
Security separation by separating authentication and token delegation

So new flow
and in oauth
there is a client app that redirect to authforge and authforge logs in and get code that code is sent to client app client app then sends the code to the authforge to get the data of the user

Your current system already follows good practices:

password only sent to backend
JWT + refresh tokens

So it's not wrong, just not extensible

Use OAuth if:

multiple apps/services
public clients (mobile, SPA)
third-party integrations
centralized auth service

## Why current approach is limited

Your flow:

```text
client → /login → access_token
```

Works, but has structural limits.

---

## 1. No client isolation

- Every client uses same login
- No identity for apps (no `client_id`)

Problem:

- Cannot control which app is accessing
- Cannot revoke access per app

---

## 2. No safe third-party support

- Only your frontend can call `/login`
- Third-party apps would need password

Problem:

- insecure
- no standard integration

---

## 3. No delegation (all-or-nothing access)

- Access token = full user access

Problem:

```text
no way to say → read only
```

---

## 4. No standardized flow

- Each client implements login differently

Problem:

- mobile, web, CLI → inconsistent
- harder to scale

---

## 5. Tight coupling (auth + client)

- Frontend directly tied to `/login`

Problem:

- cannot separate auth into independent service
- harder to reuse across systems

---

## 6. Limited security model

- No PKCE
- No authorization code step

Problem:

- weaker protection for public clients (SPA/mobile)

---

## 7. No consent layer

- User cannot control:

```text
which app can access what
```

---

Now Regular flow exist but with that also there is oauth flow which can be used for external API or multi-client

there are 2 api's
/ oauth/authorize

where you pass the session id get verified and Server generates authorization code
and you get redirected to client page with code where the code get send to
/oauth/token

where it is verified and it provides tokens for the external API system

this is basic setup

Add state:
state helps in CSRF protection (primary reason)
Prevents attackers from forging authorization responses.
with oauth/authorize
and redirect uri pass state code
You must
accept:
/oauth/authorize?...&state=xyz
Return: redirect_uri?code=...&state=xyz
