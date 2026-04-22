# AuthForge

AuthForge is a comprehensive authentication and authorization service built with Go, featuring OAuth2.0, OpenID Connect (OIDC), and traditional JWT-based authentication. It provides secure session management, PKCE support, and multi-client support.

## Implemented Features

### Authentication & User Management
- User registration with unique email and username validation
- Password hashing with Argon2id with salt
- Password strength validation
- Email verification flow with secure one-time tokens
- Email verification via token-based verification endpoint
- Login with short-lived JWT access tokens
- Refresh token-based session management
- Refresh token rotation on every token refresh
- Refresh token storage hashed in Redis
- Protected routes with JWT authentication middleware
- Logout from the current session
- Logout from all sessions
- Username update for authenticated users
- Password reset request flow with email verification
- Last login tracking
- Failed login tracking by email and IP address
- Temporary login blocking after repeated failed attempts
- Session tracking per user in Redis
- User deactivation support
- Database auto-migration for user records

### OAuth2.0 & OpenID Connect
- Full OAuth2.0 authorization code flow support
- PKCE (Proof Key for Code Exchange) for enhanced security
- OAuth client management
- Scope-based authorization
- Authorization code generation with TTL
- Token endpoint for exchanging codes for tokens
- OpenID Connect (OIDC) ID token generation
- Support for `openid` scope to retrieve user identity
- Session-based authorization with secure session cookies

### Infrastructure & Monitoring
- Health check endpoint (`/ping`)
- CORS configuration for frontend integration
- RESTful API design with consistent error responses

## Authentication Flow

### Traditional JWT Authentication Flow
1. User signs up with email, username, and password
2. Password is hashed using Argon2id with salt
3. Verification email with one-time token is sent
4. User verifies email via token link
5. Access token and refresh token are issued upon verification
6. Access token is used for protected API routes
7. Refresh token (stored as HTTP-only cookie) can be used to obtain new access tokens
8. Tokens are rotated on each refresh request
9. Logout revokes current refresh token and blacklists access token
10. Logout all removes every active session for the user

### OAuth2.0 with PKCE Flow
1. Client initiates authorization request with `client_id`, `redirect_uri`, `code_challenge`, and `scopes`
2. User must be authenticated (session_id cookie required)
3. AuthForge generates authorization code and redirects with code + state
4. Client exchanges code for tokens via `/oauth/token` endpoint
5. Client provides `code_verifier` which is verified against stored `code_challenge`
6. Upon successful verification, access token, refresh token, and optionally ID token are returned
7. ID token is generated when `openid` scope is requested

## Security Features

- **Argon2id Hashing**: Password hashing with salt for secure storage
- **Password Complexity**: Password strength validation requirements
- **One-Time Tokens**: Secure email verification and reset tokens
- **Hashed Token Storage**: Refresh tokens are hashed before storage in Redis
- **JWT Authentication**: Stateless authentication middleware
- **Access Token Blacklist**: Logout blacklists access tokens
- **Brute-Force Protection**: Rate limiting using Redis counters by email and IP
- **PKCE Support**: Protection against authorization code interception attacks
- **Secure Cookies**: HTTP-only, secure, and SameSite cookies for refresh tokens
- **CORS Configuration**: Strict origin validation
- **Session Isolation**: Per-user session management with Redis

## API Routes

### Authentication Endpoints
Base path: `/api/v1/auth`

**User Management:**
- `POST /signup` - Register a new user
- `POST /login` - Authenticate user and return tokens
- `GET /email/verify?token=...` - Verify user email with token
- `PATCH /update/username` - Update authenticated user's username
- `POST /reset/password` - Request password reset
- `GET /logout` - Logout current session
- `GET /logout/all` - Logout from all sessions
- `POST /refresh` - Refresh access token

**OAuth2.0 Endpoints:**
- `GET /oauth/authorize` - OAuth authorization endpoint (requires user authentication)
- `POST /oauth/token` - OAuth token endpoint (exchange authorization code for tokens)

### Health Check
- `GET /ping` - Service health check

## Tech Stack

- **Language**: Go 1.25.4
- **Web Framework**: Gin
- **Database**: PostgreSQL with GORM ORM
- **Cache/Session Store**: Redis
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: Argon2id
- **CORS**: gin-contrib/cors
- **API Documentation**: OpenAPI 3.0

## Configuration

The application reads environment variables. A `.env` file can be used to configure the following:

### Required Environment Variables
- `DB_CONNECT_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for signing JWT tokens

### Optional Configuration
- CORS allowed origins and methods
- Server port (default: 8080)

## Project Structure

```
auth/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── bootstrap/              # Application initialization
│   ├── config/                 # Configuration management
│   ├── handler/                # HTTP request handlers
│   ├── middleware/             # Authentication middleware
│   ├── model/                  # Data models
│   ├── repository/             # Data access layer
│   ├── routes/                 # Route definitions
│   ├── security/               # Security utilities (hashing, tokens)
│   └── services/               # Business logic & services
└── pkg/
    ├── db/                     # Database initialization & migration
    ├── email/                  # Email service
    ├── redis/                  # Redis client
    └── util/                   # Utility functions

client/                         # Next.js frontend client
docs/                          # API and feature documentation
```

## Notes

- Refresh tokens are stored in secure, HTTP-only cookies
- Login attempts are tracked separately by email address and IP address
- Active sessions are grouped per user in Redis for efficient session management
- Authorization codes expire after 5 minutes
- PKCE is mandatory for OAuth2.0 token endpoint
- ID tokens are only generated when the `openid` scope is requested
- All passwords are hashed with Argon2id before storage
