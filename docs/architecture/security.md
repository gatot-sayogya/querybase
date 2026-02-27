# Security Architecture

QueryBase implements a multi-layered security strategy to ensure safe database exploration and query management.

## 1. Authentication Strategy

QueryBase uses a **Dual-Token Strategy** to maximize security while maintaining a seamless user experience.

### 1.1 Access Tokens

- **Type**: Short-lived JWT (15 minutes).
- **Storage**: In-memory (JavaScript variables).
- **Usage**: Sent in the `Authorization: Bearer <token>` header for all API requests.
- **Security**: In-memory storage prevents persistence-based XSS attacks.

### 1.2 Refresh Tokens

- **Type**: Long-lived opaque tokens (7 days).
- **Storage**: `HttpOnly`, `Secure`, `SameSite=Strict` Cookie.
- **Usage**: Exchanged for new access tokens automatically by the frontend interceptor when the access token expires.
- **Security**: `HttpOnly` prevents JavaScript from accessing the token, mitigating XSS stealing. `SameSite=Strict` protects against CSRF.

## 2. Token Revocation & Blacklisting

Immediate session invalidation is handled via a **Redis-based Blacklist Service**.

### 2.1 Access Token Blacklisting

When a user logs out or a session is revoked, the JTI (JWT ID) of the access token is stored in Redis for the remainder of its TTL. The `AuthMiddleware` checks every request against this blacklist.

### 2.2 Refresh Token Rotation

Refresh tokens are rotated upon use. Old tokens are invalidated immediately. Redis maintains the mapping of valid refresh tokens to prevent replay attacks.

## 3. Middleware Security

The API layer is protected by several security-focused middlewares.

### 3.1 Security Headers

The `SecurityHeadersMiddleware` injects industry-standard headers into every response:

- `Strict-Transport-Security` (HSTS)
- `Content-Security-Policy` (CSP)
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: strict-origin-when-cross-origin`

### 3.2 Input Sanitization

The `SanitizationMiddleware` automatically validates and cleanses input from:

- Query Parameters
- Request Headers
- Request Bodies (where applicable)

### 3.3 Intelligent Rate Limiting

A custom token-bucket rate limiter protects the API from brute-force and DoS attacks:

- **General API**: 300 requests/minute with a burst of 30, using **prefix matching** to avoid throttling during rapid UI navigation.
- **Authentication**: Strict limits (5 attempts/minute) on `/api/v1/auth/login`.

## 4. Browser Integration (CORS)

Cross-Origin Resource Sharing is configured strictly:

- No wildcards (\*) allowed when credentials are enabled.
- Origins are reflected back from the allowed list.
- `Vary: Origin` header ensures proper downstream caching behavior.
