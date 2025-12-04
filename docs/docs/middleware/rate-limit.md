# Rate Limiting

Prevent API abuse with token bucket-based rate limiting.

## Basic Usage

```go
import "github.com/ginjigo/ginji/middleware"

app.Use(middleware.RateLimit(100, time.Minute)) // 100 req/min
```

## Helper Functions

```go
// Per second
app.Use(middleware.RateLimitPerSecond(10))

// Per minute
app.Use(middleware.RateLimitPerMinute(100))

// Per hour
app.Use(middleware.RateLimitPerHour(1000))
```

## By User ID

Rate limit by authenticated user:

```go
app.Use(middleware.RateLimitByUser(50, time.Minute, "user_id"))
```

The middleware will look for `user_id` in the context (set by your auth middleware).

## By API Key

Rate limit by API key:

```go
app.Use(middleware.RateLimitByAPIKey(1000, time.Hour, "X-API-Key"))
```

## Custom Configuration

```go
config := middleware.RateLimiterConfig{
    Max:    1000,
    Window: time.Hour,
    
    // Custom key function
    KeyFunc: func(c *ginji.Context) string {
        // Rate limit by tenant
        return c.GetString("tenant_id")
    },
    
    // Skip for certain requests
    SkipFunc: func(c *ginji.Context) bool {
        // Admin users bypass rate limit
        return c.GetString("role") == "admin"
    },
    
    ErrorMessage: "Too many requests from your IP",
    StatusCode:   ginji.StatusTooManyRequests,
    Headers:      true, // Add rate limit headers
}

app.Use(middleware.RateLimitWithConfig(config))
```

## Response Headers

When rate limiting is active, these headers are added:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1701691200
Retry-After: 30
```

## Error Response

When rate limit is exceeded:

```json
{
  "error": "Rate limit exceeded. Maximum 100 requests per 1m0s",
  "limit": 100,
  "window": "1m0s",
  "retryAt": "2024-12-04T12:30:00Z"
}
```

## Advanced Example

```go
// Different limits for different endpoints
public := app.Group("/api/public")
public.Use(middleware.RateLimitPerMinute(10))

authenticated := app.Group("/api")
authenticated.Use(middleware.BasicAuth(users))
authenticated.Use(middleware.RateLimitPerMinute(100))

premium := app.Group("/api/premium")
premium.Use(middleware.BearerAuth(validateToken))
premium.Use(middleware.RateLimitPerMinute(1000))
```

## Features

- ✅ Token bucket algorithm
- ✅ Automatic bucket cleanup
- ✅ Thread-safe
- ✅ Custom key functions (IP, user, API key, custom)
- ✅ Skip function for exemptions
- ✅ Configurable error messages
- ✅ Rate limit headers
- ✅ Zero dependencies

## Best Practices

1. **Choose appropriate limits** - Too strict frustrates users, too loose allows abuse
2. **Use different limits** - Public < authenticated < premium tiers
3. **Consider your infrastructure** - Don't exceed your server capacity
4. **Monitor metrics** - Track rate limit hits to adjust limits
5. **Inform users** - Use rate limit headers so clients can back off gracefully
