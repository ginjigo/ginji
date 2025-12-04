# Security Headers

Add comprehensive security headers to protect your application.

## Basic Usage

```go
import "github.com/ginjigo/ginji/middleware"

app.Use(middleware.Secure())
```

## Strict Mode

Maximum security for production:

```go
app.Use(middleware.SecureStrict())
```

## Custom Configuration

```go
config := middleware.SecureConfig{
    XSSProtection:         "1; mode=block",
    ContentTypeNosniff:    "nosniff",
    XFrameOptions:         "DENY",
    HSTSMaxAge:            31536000, // 1 year
    HSTSIncludeSubdomains: true,
    HSTSPreload:           true,
    ContentSecurityPolicy: "default-src 'self'",
    ReferrerPolicy:        "strict-origin-when-cross-origin",
}

app.Use(middleware.SecureWithConfig(config))
```

## CSP Builder

Build Content Security Policy easily:

```go
csp := middleware.NewCSP().
    DefaultSrc("'self'").
    ScriptSrc("'self'", "'unsafe-inline'", "https://cdn.example.com").
    StyleSrc("'self'", "https://fonts.googleapis.com").
    ImgSrc("'self'", "data:", "https:").
    FontSrc("'self'", "https://fonts.gstatic.com").
    ConnectSrc("'self'").
    FrameSrc("'none'").
    ObjectSrc("'none'").
    UpgradeInsecureRequests().
    Build()

config := middleware.SecureConfig{
    ContentSecurityPolicy: csp,
}

app.Use(middleware.SecureWithConfig(config))
```

## Headers Set

### XSS Protection
```
X-XSS-Protection: 1; mode=block
```

### Content Type Nosniff
```
X-Content-Type-Options: nosniff
```

### Frame Options
```
X-Frame-Options: DENY
```

### HSTS
```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

### CSP
```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'
```

### Referrer Policy
```
Referrer-Policy: strict-origin-when-cross-origin
```

### Permissions Policy
```
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### Cross-Origin Policies
```
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
```

## Complete Example

```go
package main

import (
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

func main() {
    app := ginji.New()

    // Build CSP
    csp := middleware.NewCSP().
        DefaultSrc("'self'").
        ScriptSrc("'self'", "https://cdn.jsdelivr.net").
        StyleSrc("'self'", "'unsafe-inline'").
        ImgSrc("'self'", "data:", "https:").
        FontSrc("'self'", "https://fonts.gstatic.com").
        Build()

    // Configure security headers
    app.Use(middleware.SecureWithConfig(middleware.SecureConfig{
        XFrameOptions:         "SAMEORIGIN",
        HSTSMaxAge:            31536000,
        HSTSIncludeSubdomains: true,
        ContentSecurityPolicy: csp,
        ReferrerPolicy:        "no-referrer-when-downgrade",
    }))

    app.Get("/", func(c *ginji.Context) {
        c.HTML(ginji.StatusOK, "<h1>Secure App</h1>")
    })

    app.Listen(":443") // Use HTTPS in production
}
```

## Best Practices

1. **Use HTTPS** - Security headers work best with HTTPS
2. **Start strict** - Use `SecureStrict()` and relax as needed
3. **Test CSP** - Use report-only mode first
4. **Enable HSTS** - Force HTTPS for all requests
5. **Regular updates** - Keep security policies current

## Security Checklist

- ✅ XSS Protection enabled
- ✅ Content Type nosniffing
- ✅ Clickjacking protection (X-Frame-Options)
- ✅ HSTS with preload
- ✅ Content Security Policy
- ✅ Referrer Policy set
- ✅ Cross-Origin policies configured
