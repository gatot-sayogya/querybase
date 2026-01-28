# Encrypted Frontend-Backend Communication

**Status:** TODO - Planned Feature
**Priority:** Medium
**Category:** Security & Frontend Integration

---

## Overview

Implement end-to-end encrypted communication between the QueryBase backend (Go/Gin) and the frontend (Next.js) with comprehensive request validation to ensure secure data transmission and prevent malicious payloads.

## Business Value

### Security Benefits
1. **Data Protection**: Encrypt sensitive query results in transit
2. **Request Integrity**: Prevent tampering with requests/responses
3. **Validation Layer**: Add defense-in-depth against malicious payloads
4. **Compliance**: Meet security standards for handling database queries

### User Trust
- Users trust that their queries and results are secure
- Audit trail shows encryption was used
- Professional-grade security posture

## Architecture

### Current State (Backend Only)

```
Frontend (Not Implemented)           Backend (Go/Gin)
─────────────────────                ──────────────────
                                        │
                                    [JWT Auth Middleware]
                                        │
                                    [RBAC Middleware]
                                        │
                                    [Route Handlers]
                                        │
                                    [Service Layer]
                                        │
                                 [PostgreSQL Database]
```

### Proposed State (With Encryption)

```
Frontend (Next.js)                    Backend (Go/Gin)
────────────────────                ──────────────────
      │                                      │
      │ 1. Request                            │
      ├───────────────────────────────────>│
      │                                      │
      │ [Encrypt Request Payload]           │
      │ - AES-256-GCM encryption           │
      │ - Nonce for uniqueness            │
      │ - HMAC for integrity              │
      │                                      │
      │ 2. Encrypted Request                │
      ├───────────────────────────────────>│
      │                                      │
      │                          [Decryption Middleware]
      │                                      │
      │                          [Request Validation Middleware]
      │                          - Schema validation       │
      │                          - SQL injection check    │
      │                          - Size limits            │
      │                          - Rate limiting           │
      │                                      │
      │                          [JWT Auth Middleware]
      │                                      │
      │                          [RBAC Middleware]
      │                                      │
      │                          [Route Handlers]
      │                                      │
      │                          [Service Layer]
      │                                      │
      │                          [Response Processing]
      │                                      │
      │ 3. Encrypted Response                 │
      │<────────────────────────────────────┤
      │                                      │
      │ [Decrypt Response]                    │
      │ - Verify HMAC                         │
      │ - Decrypt payload                     │
      │                                      │
      │ 4. Display Results                    │
      └──────────────────────────────────────┘
```

## Implementation Plan

### Phase 1: Backend Encryption Infrastructure

#### 1.1 Encryption Service
**File:** `internal/service/encryption.go`

```go
package service

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "io"
)

type EncryptionService struct {
    key []byte
}

func NewEncryptionService(key string) *EncryptionService {
    // In production, load from environment variable
    hash := sha256.Sum256([]byte(key))
    return &EncryptionService{key: hash[:]}
}

// Encrypt encrypts data using AES-256-GCM
func (s *EncryptionService) Encrypt(plaintext []byte) (string, error) {
    block, err := aes.NewCipher(s.key)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

    // Return: base64(nonce + ciphertext)
    return base64.StdEncoding.EncodeToString(nonce), nil
}

// Decrypt decrypts data encrypted with Encrypt()
func (s *EncryptionService) Decrypt(ciphertext string) ([]byte, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return nil, err
    }

    block, err := aes.NewCipher(s.key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, errors.New("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]

    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return plaintext, nil
}

// GenerateHMAC creates HMAC for message integrity
func (s *EncryptionService) GenerateHMAC(message []byte) string {
    h := hmac.New(sha256.New, s.key)
    h.Write(message)
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifyHMAC verifies HMAC signature
func (s *EncryptionService) VerifyHMAC(message []byte, signature string) bool {
    expectedMAC := s.GenerateHMAC(message)
    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
```

#### 1.2 Encrypted Request DTOs
**File:** `internal/api/dto/encryption.go`

```go
package dto

// EncryptedRequest wraps encrypted payloads from frontend
type EncryptedRequest struct {
    Payload   string `json:"payload" binding:"required"`   // Encrypted data
    HMAC      string `json:"hmac" binding:"required"`      // HMAC signature
    Timestamp int64  `json:"timestamp" binding:"required"` // Request timestamp
}

// EncryptionHeaders contains encryption metadata
type EncryptionHeaders struct {
    KeyID      string `json:"key_id"`      // Key identifier (for key rotation)
    Algorithm  string `json:"algorithm"`   // Encryption algorithm (e.g., "AES-256-GCM")
    Version    string `json:"version"`     // Protocol version
}
```

### Phase 2: Encryption Middleware

#### 2.1 Decryption Middleware
**File:** `internal/api/middleware/encryption.go`

```go
package middleware

import (
    "bytes"
    "encoding/json"
    "errors"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/yourorg/querybase/internal/service"
)

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
    Service           *service.EncryptionService
    RequiredEndpoints []string              // Endpoints requiring encryption
    MaxRequestSize    int64                 // Max encrypted request size (1MB)
    MaxAgeSeconds     int64                 // Max request age (30 seconds)
}

// EncryptionMiddleware decrypts and validates encrypted requests
func EncryptionMiddleware(config *EncryptionConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip if endpoint doesn't require encryption
        if !requiresEncryption(c.Request.URL.Path, config.RequiredEndpoints) {
            c.Next()
            return
        }

        // Validate Content-Type
        if c.Request.Header.Get("Content-Type") != "application/json" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type must be application/json"})
            c.Abort()
            return
        }

        // Check encryption headers
        algorithm := c.Request.Header.Get("X-Encryption-Algorithm")
        version := c.Request.Header.Get("X-Encryption-Version")

        if algorithm != "AES-256-GCM" || version != "1.0" {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Unsupported encryption version or algorithm",
                "supported": "AES-256-GCM, version 1.0",
            })
            c.Abort()
            return
        }

        // Read request body
        var encryptedReq dto.EncryptedRequest
        if err := c.ShouldBindJSON(&encryptedReq); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            c.Abort()
            return
        }

        // Validate request age (replay attack prevention)
        requestAge := time.Now().Unix() - encryptedReq.Timestamp
        if requestAge > config.MaxAgeSeconds || requestAge < 0 {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Request expired or timestamp invalid",
                "max_age_seconds": config.MaxAgeSeconds,
            })
            c.Abort()
            return
        }

        // Validate request size
        if c.Request.ContentLength > config.MaxRequestSize {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Request too large",
                "max_size_mb": config.MaxRequestSize / 1024 / 1024,
            })
            c.Abort()
            return
        }

        // Verify HMAC
        payloadBytes := []byte(encryptedReq.Payload)
        if !config.Service.VerifyHMAC(payloadBytes, encryptedReq.HMAC) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid HMAC signature"})
            c.Abort()
            return
        }

        // Decrypt payload
        decryptedBytes, err := config.Service.Decrypt(encryptedReq.Payload)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Decryption failed"})
            c.Abort()
            return
        }

        // Parse decrypted JSON
        var decryptedData map[string]interface{}
        if err := json.Unmarshal(decryptedBytes, &decryptedData); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON in payload"})
            c.Abort()
            return
        }

        // Store decrypted data in context for handlers
        c.Set("decrypted_payload", decryptedData)
        c.Set("encrypted_request", true)

        c.Next()
    }
}

// requiresEncryption checks if endpoint requires encryption
func requiresEncryption(path string, endpoints []string) bool {
    for _, endpoint := range endpoints {
        if matchPath(path, endpoint) {
            return true
        }
    }
    return false
}

func matchPath(path, pattern string) bool {
    // Simple pattern matching (can be enhanced)
    return path == pattern || len(path) > len(pattern) && path[:len(pattern)] == pattern
}
```

### Phase 3: Request Validation Middleware

#### 3.1 Comprehensive Request Validator
**File:** `internal/api/middleware/validation.go`

```go
package middleware

import (
    "encoding/json"
    "net/http"
    "regexp"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/yourorg/querybase/internal/service"
)

// ValidationConfig holds validation configuration
type ValidationConfig struct {
    MaxQueryLength      int     // Maximum SQL query length (10,000 chars)
    MaxBatchSize        int     // Maximum batch operations
    AllowedPatterns     []string // Regex patterns for allowed operations
    BlockedPatterns     []string // Regex patterns for blocked operations
}

// RequestValidationMiddleware validates request payloads
func RequestValidationMiddleware(config *ValidationConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if request has decrypted payload
        decryptedPayload, exists := c.Get("decrypted_payload")
        if !exists {
            c.Next()
            return
        }

        data, ok := decryptedPayload.(map[string]interface{})
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid payload type"})
            c.Abort()
            return
        }

        // Validate query_text if present
        if queryText, ok := data["query_text"].(string); ok {
            if err := validateQueryText(queryText, config); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Invalid query text",
                    "details": err.Error(),
                })
                c.Abort()
                return
            }
        }

        // Validate data_source_id if present
        if dataSourceID, ok := data["data_source_id"].(string); ok {
            if err := validateUUID(dataSourceID); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Invalid data source ID",
                    "details": err.Error(),
                })
                c.Abort()
                return
            }
        }

        // Validate parameters for queries
        if params, ok := data["parameters"].([]interface{}); ok {
            if len(params) > config.MaxBatchSize {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Too many parameters",
                    "max_size": config.MaxBatchSize,
                })
                c.Abort()
                return
            }
        }

        c.Next()
    }
}

// validateQueryText validates SQL query text
func validateQueryText(query string, config *ValidationConfig) error {
    // Check length
    if len(query) > config.MaxQueryLength {
        return errors.New("query too long")
    }

    // Trim and check empty
    query = strings.TrimSpace(query)
    if query == "" {
        return errors.New("query cannot be empty")
    }

    // Check for blocked patterns (e.g., multiple statements)
    for _, pattern := range config.BlockedPatterns {
        matched, _ := regexp.MatchString(pattern, query)
        if matched {
            return errors.New("query contains blocked pattern")
        }
    }

    // Validate SQL syntax (reuse existing parser)
    if err := service.ValidateSQL(query); err != nil {
        return err
    }

    return nil
}

// validateUUID validates UUID format
func validateUUID(id string) error {
    uuidRegex := regexp.MustCompile(
        `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
    )
    if !uuidRegex.MatchString(id) {
        return errors.New("invalid UUID format")
    }
    return nil
}
```

### Phase 4: Encrypted Response Middleware

#### 4.1 Response Encryption
**File:** `internal/api/middleware/response_encryption.go`

```go
package middleware

import (
    "bytes"
    "crypto/rand"
    "encoding/json"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/yourorg/querybase/internal/service"
)

// ResponseEncryptionMiddleware encrypts successful responses
func ResponseEncryptionMiddleware(encryptionService *service.EncryptionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Write response writer that intercepts response
        w := &responseWriter{
            ResponseWriter: c.Writer,
            shouldEncrypt: shouldEncryptResponse(c),
        }

        c.Writer = w

        c.Next()

        // Encrypt response if needed
        if w.shouldEncrypt && w.Status() == http.StatusOK {
            if err := encryptResponse(c, encryptionService); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "Failed to encrypt response",
                })
            }
        }
    }
}

func shouldEncryptResponse(c *gin.Context) bool {
    // Check if client requested encryption
    return c.Request.Header.Get("X-Request-Encryption") == "true"
}

func encryptResponse(c *gin.Context, svc *service.EncryptionService) error {
    // Get response body
    body := c.Writer.(*responseWriter).body

    // Encrypt
    encrypted, err := svc.Encrypt(body)
    if err != nil {
        return err
    }

    // Generate HMAC
    hmac := svc.GenerateHMAC(body)

    // Build encrypted response
    response := map[string]interface{}{
        "payload":   encrypted,
        "hmac":      hmac,
        "timestamp": time.Now().Unix(),
    }

    // Set encryption headers
    c.Header("X-Encryption-Algorithm", "AES-256-GCM")
    c.Header("X-Encryption-Version", "1.0")

    // Write encrypted response
    c.JSON(http.StatusOK, response)
    return nil
}

// responseWriter intercepts response body
type responseWriter struct {
    gin.ResponseWriter
    shouldEncrypt bool
    body         []byte
    status       int
}

func (w *responseWriter) Write(data []byte) (int, error) {
    if w.shouldEncrypt {
        w.body = append(w.body, data...)
        return len(data), nil
    }
    return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
    w.status = statusCode
    if !w.shouldEncrypt {
        w.ResponseWriter.WriteHeader(statusCode)
    }
}

func (w *responseWriter) Status() int {
    return w.status
}
```

### Phase 5: Frontend Implementation

#### 5.1 Encryption Utility (Frontend)
**File:** `web/lib/encryption.ts` (Next.js)

```typescript
import { Buffer } from 'buffer';

// Encryption configuration
const ENCRYPTION_CONFIG = {
  algorithm: 'AES-256-GCM',
  version: '1.0',
  keySize: 256, // bits
  nonceSize: 12, // bytes (for GCM)
};

// Client-side encryption (for request payloads)
export async function encryptPayload(
  payload: any,
  encryptionKey: string
): Promise<{ payload: string; hmac: string; timestamp: number }> {
  // Get encryption key from backend (via secure endpoint)
  const key = await deriveEncryptionKey(encryptionKey);

  // Convert payload to JSON
  const plaintext = JSON.stringify(payload);

  // Encrypt using Web Crypto API
  const nonce = crypto.getRandomValues(new Uint8Array(ENCRYPTION_CONFIG.nonceSize));

  const encoder = new TextEncoder();
  const data = encoder.encode(plaintext);

  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    key,
    { name: 'AES-GCM' },
    false,
    ['encrypt']
  );

  const encryptedData = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv: nonce },
    cryptoKey,
    data
  );

  // Combine nonce + ciphertext
  const combined = new Uint8Array(nonce.length + encryptedData.byteLength);
  combined.set(nonce);
  combined.set(new Uint8Array(encryptedData));

  // Encode to base64
  const payloadBase64 = btoa(String.fromCharCode(...combined));

  // Generate HMAC
  const hmac = await generateHMAC(combined, key);

  return {
    payload: payloadBase64,
    hmac: hmac,
    timestamp: Math.floor(Date.now() / 1000),
  };
}

// Decrypt response from backend
export async function decryptResponse<T>(
  encryptedResponse: any,
  encryptionKey: string
): Promise<T> {
  const key = await deriveEncryptionKey(encryptionKey);

  // Decode base64
  const combined = Uint8Array.from(atob(encryptedResponse.payload), c => c.charCodeAt(0));

  // Extract nonce and ciphertext
  const nonce = combined.slice(0, ENCRYPTION_CONFIG.nonceSize);
  const ciphertext = combined.slice(ENCRYPTION_CONFIG.nonceSize);

  // Verify HMAC
  const isValid = await verifyHMAC(combined, encryptedResponse.hmac, key);
  if (!isValid) {
    throw new Error('Invalid HMAC signature');
  }

  // Decrypt
  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    key,
    { name: 'AES-GCM' },
    false,
    ['decrypt']
  );

  const decryptedData = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: nonce },
    cryptoKey,
    ciphertext
  );

  // Decode to string
  const decoder = new TextDecoder();
  const plaintext = decoder.decode(decryptedData);

  return JSON.parse(plaintext) as T;
}

// Derive encryption key from server-provided key
async function deriveEncryptionKey(serverKey: string): Promise<Uint8Array> {
  // Use PBKDF2 to derive key from server key
  const encoder = new TextEncoder();
  const keyMaterial = await crypto.subtle.importKey(
    'raw',
    encoder.encode(serverKey),
    { name: 'PBKDF2' },
    false,
    ['deriveKey']
  );

  // In production, use proper salt from environment
  const salt = crypto.getRandomValues(new Uint8Array(16));

  return crypto.subtle.deriveKey(
    {
      name: 'PBKDF2',
      salt: salt,
      iterations: 100000,
      hash: 'SHA-256',
    },
    keyMaterial,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  );
}

// Generate HMAC using Web Crypto API
async function generateHMAC(data: Uint8Array, key: Uint8Array): Promise<string> {
  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    key,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  );

  const signature = await crypto.subtle.sign(
    { name: 'HMAC', hash: 'SHA-256' },
    cryptoKey,
    data
  );

  return btoa(String.fromCharCode(...new Uint8Array(signature)));
}

// Verify HMAC
async function verifyHMAC(
  data: Uint8Array,
  signature: string,
  key: Uint8Array
): Promise<boolean> {
  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    key,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['verify']
  );

  const sig = Uint8Array.from(atob(signature), c => c.charCodeAt(0));

  return await crypto.subtle.verify(
    { name: 'HMAC', hash: 'SHA-256' },
    cryptoKey,
    sig,
    data
  );
}
```

#### 5.2 API Client with Encryption (Frontend)
**File:** `web/lib/api-client.ts`

```typescript
import { encryptPayload, decryptResponse } from './encryption';

class QueryBaseAPI {
  private encryptionKey: string;

  constructor(apiURL: string, encryptionKey: string) {
    this.apiURL = apiURL;
    this.encryptionKey = encryptionKey;
  }

  private async request(
    endpoint: string,
    payload: any,
    options: RequestInit = {}
  ): Promise<any> {
    // Encrypt payload
    const encrypted = await encryptPayload(payload, this.encryptionKey);

    // Add encryption headers
    const headers = {
      'Content-Type': 'application/json',
      'X-Encryption-Algorithm': 'AES-256-GCM',
      'X-Encryption-Version': '1.0',
      'X-Request-Encryption': 'true',
      ...options.headers,
    };

    const response = await fetch(`${this.apiURL}${endpoint}`, {
      ...options,
      method: 'POST',
      headers,
      body: JSON.stringify(encrypted),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    // Decrypt response if encrypted
    const data = await response.json();

    if (data.payload && data.hmac) {
      return await decryptResponse(data, this.encryptionKey);
    }

    return data;
  }

  // Query execution example
  async executeQuery(params: {
    dataSourceId: string;
    queryText: string;
  }): Promise<any> {
    return this.request('/api/v1/queries', params);
  }

  // Other API methods...
}

export default QueryBaseAPI;
```

### Phase 6: Key Management

#### 6.1 Key Distribution Endpoint
**File:** `internal/api/handlers/encryption.go`

```go
package handlers

import (
    "crypto/rand"
    "encoding/base64"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/yourorg/querybase/internal/service"
)

// EncryptionHandler handles encryption key operations
type EncryptionHandler struct {
    encryptionService *service.EncryptionService
}

func NewEncryptionHandler(encryptionService *service.EncryptionService) *EncryptionHandler {
    return &EncryptionHandler{
        encryptionService: encryptionService,
    }
}

// GetEncryptionKey returns a public key for frontend encryption
// In production, use asymmetric encryption (RSA) for key exchange
func (h *EncryptionHandler) GetEncryptionKey(c *gin.Context) {
    // Generate session key (for demo purposes, use symmetric key)
    // In production: Generate RSA key pair, return public key

    key := make([]byte, 32) // 256-bit key
    if _, err := rand.Read(key); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate key"})
        return
    }

    // Encode key (in production, encrypt with user's public key)
    encodedKey := base64.StdEncoding.EncodeToString(key)

    // Store key in session/cache with expiry
    // Return to frontend

    c.JSON(http.StatusOK, gin.H{
        "key_id": "session-key-id",
        "key": encodedKey,
        "algorithm": "AES-256-GCM",
        "expires_in": 3600, // 1 hour
    })
}
```

#### 6.2 Route Registration
**File:** `internal/api/routes/routes.go`

```go
// Add to routes setup
encryptionHandler := handlers.NewEncryptionHandler(encryptionService)

protected.GET("/encryption/key", encryptionHandler.GetEncryptionKey)
```

## Configuration

### Environment Variables
```yaml
# config/config.yaml
encryption:
  enabled: true                    # Enable encryption
  key_rotation_interval: 86400      # 24 hours
  max_request_age: 30               # seconds
  max_request_size: 1048576         # 1MB
  required_endpoints:               # Endpoints requiring encryption
    - /api/v1/queries
    - /api/v1/queries/save
    - /api/v1/approvals
    - /api/v1/datasources/:id/test
```

### Validation Rules
```go
const ValidationConfig = {
    MaxQueryLength: 10000,              // 10,000 characters
    MaxBatchSize: 100,                   // Max 100 operations
    AllowedPatterns: [
        `^SELECT.*FROM.*`,
        `^INSERT INTO.*`,
        `^UPDATE.*SET.*`,
        `^DELETE FROM.*`,
    ],
    BlockedPatterns: [
        `;.*DROP`,                       // Prevent SQL injection
        `;.*EXEC`,                       // Prevent command injection
        `/\*.*\*/`,                       // Prevent comment-based injection
    ],
}
```

## Security Considerations

### ✅ Benefits
1. **Confidentiality**: Encrypts sensitive query results
2. **Integrity**: HMAC prevents tampering
3. **Replay Protection**: Timestamp validation
4. **Rate Limiting**: Prevents abuse
5. **Input Validation**: Defense-in-depth

### ⚠️ Challenges
1. **Performance**: Encryption/decryption overhead (~10-20ms)
2. **Key Management**: Secure key distribution
3. **Frontend Complexity**: More complex frontend code
4. **Debugging**: Harder to debug encrypted traffic

### 🔒 Best Practices
1. Use HTTPS (TLS) in addition to payload encryption
2. Implement key rotation (every 24 hours)
3. Store encryption keys securely (environment variables, KMS)
4. Log all encryption operations (audit trail)
5. Use asymmetric encryption for key exchange (RSA/ECDH)
6. Implement fallback mechanism for compatibility

## Implementation Checklist

### Backend (Go)
- [ ] Create encryption service
- [ ] Implement decryption middleware
- [ ] Implement request validation middleware
- [ ] Implement response encryption middleware
- [ ] Add key distribution endpoint
- [ ] Configure encryption settings
- [ ] Add encryption to unit tests
- [ ] Add encryption to integration tests

### Frontend (Next.js)
- [ ] Create encryption utility
- [ ] Implement API client with encryption
- [ ] Add error handling for decryption failures
- [ ] Implement key fetching
- [ ] Add TypeScript types
- [ ] Add unit tests for encryption

### Testing
- [ ] Test encryption/decryption round-trip
- [ ] Test HMAC validation
- [ ] Test replay attack prevention
- [ ] Test request validation
- [ ] Performance benchmarks
- [ ] Security penetration testing

## Performance Impact

### Estimated Overhead
- **Encryption**: ~5-10ms per request
- **Decryption**: ~5-10ms per request
- **HMAC Generation**: ~1-2ms per request
- **HMAC Verification**: ~1-2ms per request
- **Total**: ~12-24ms per request (acceptable for most use cases)

### Optimization Strategies
1. **Cache Derived Keys**: Frontend caches derived encryption keys
2. **Selective Encryption**: Only encrypt sensitive endpoints
3. **Compression**: Compress data before encryption
4. **Async Operations**: Use Web Crypto API asynchronously

## Migration Path

### Phase 1: Backend (Week 1-2)
- Implement encryption service
- Add middleware infrastructure
- Unit tests
- Internal testing

### Phase 2: Frontend (Week 3-4)
- Encryption utilities
- API client integration
- Error handling
- Frontend testing

### Phase 3: Integration (Week 5)
- End-to-end testing
- Performance optimization
- Security audit
- Documentation

### Phase 4: Rollout (Week 6)
- Gradual rollout (feature flag)
- Monitor performance
- Gather feedback
- Full production release

## Alternatives Considered

### Option 1: TLS Only
**Pros:** Simple, standard
**Cons:** Doesn't protect against compromised TLS termination

### Option 2: Application-Level Encryption ✅ CHOSEN
**Pros:** End-to-end security, works even if TLS terminated
**Cons:** More complex, performance overhead

### Option 3: Hybrid (TLS + Selective Encryption)
**Pros:** Balance of security and performance
**Cons:** More complex configuration

## Dependencies

### Go Packages
```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
)
```

### Frontend
```typescript
// Uses Web Crypto API (built into modern browsers)
// No external dependencies needed
```

## Related Documentation

- **[Architecture](../architecture/detailed-flow.md)** - System flow
- **[Development](../development/testing.md)** - Testing strategies
- **[CLAUDE.md](../../CLAUDE.md)** - Complete project guide

---

**Status:** TODO - Planned for Frontend Development Phase
**Priority:** Medium
**Estimated Effort:** 4-6 weeks (full implementation)
**Risk Level:** Medium (requires careful security implementation)
