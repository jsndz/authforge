# AuthForge

AuthForge is a Go-based authentication service built with Gin, GORM, PostgreSQL, and Redis.

## Implemented features

- User registration with unique email and username
- Password hashing with Argon2id
- Password strength validation
- Email verification flow with secure one-time tokens
- Login with short-lived JWT access tokens
- Refresh token-based session management
- Refresh token rotation on every refresh
- Refresh token storage hashed in Redis
- Protected routes with JWT authentication middleware
- Logout from the current session
- Logout from all sessions
- Username update for authenticated users
- Password reset request flow
- Password reset verification flow
- Last login tracking
- Failed login tracking by email and IP
- Temporary login blocking after repeated failed attempts
- Session tracking per user in Redis
- User deactivation support
- Database auto-migration for user records
- Health check endpoint

## Authentication flow

1. A user signs up with email, username, and password.
2. The password is hashed with Argon2id.
3. A verification token is generated and sent by email.
4. After email verification, the user receives an access token and refresh token.
5. Access tokens are used for protected routes.
6. Refresh tokens are stored as secure cookies and rotated on refresh.
7. Logout revokes the current refresh token and blacklists the access token.
8. Logout all removes every active session for the user.

## Security features

- Argon2id password hashing with salt
- Password complexity checks
- One-time verification and reset tokens
- Hashed refresh token storage in Redis
- JWT authentication middleware
- Access token blacklist on logout
- Brute-force login protection using Redis counters

## API routes

Base path: /api/v1/auth

- POST /signup
- POST /login
- GET /email/verify?token=...
- PATCH /update/username
- POST /reset/password
- GET /logout
- GET /logout/all
- POST /refresh

## Tech stack

- Go
- Gin
- GORM
- PostgreSQL
- Redis
- JWT
- Argon2id

## Configuration

The application reads environment variables from .env when available.

Required variables:

- DB_CONNECT_URL
- JWT_SECRET

## Notes

- Refresh tokens are stored in secure cookies.
- Login attempts are tracked separately by email and IP.
- Active sessions are grouped per user in Redis for efficient global logout.
