# Health Checks

Kubernetes-style health checks for monitoring and orchestration.

## Basic Usage

```go
import "github.com/ginjigo/ginji/middleware"

app.Use(middleware.Health())
```

This creates two endpoints:
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

## Custom Paths

```go
app.Use(middleware.SimpleHealthCheck("/live", "/ready"))
```

## Custom Health Checkers

Add checks for dependencies:

```go
config := middleware.DefaultHealthCheckConfig()

// Add database checker
config.AddHealthChecker("database", func() error {
    return db.Ping()
})

// Add Redis checker
config.AddHealthChecker("redis", func() error {
    return redis.Ping().Err()
})

// Add external API checker
config.AddHealthChecker("api", func() error {
    resp, err := http.Get("https://api.example.com/status")
    if err != nil {
        return err
    }
    if resp.StatusCode != 200 {
        return errors.New("API unhealthy")
    }
    return nil
})

app.Use(middleware.HealthWithConfig(config))
```

## Configuration

```go
config := middleware.HealthCheckConfig{
    LivenessPath:  "/health/live",
    ReadinessPath: "/health/ready",
    Timeout:       5 * time.Second,
}

config.AddHealthChecker("database", dbHealthCheck)

app.Use(middleware.HealthWithConfig(config))
```

## Response Format

### Healthy Response (200 OK)

```json
{
  "status": "UP",
  "checks": {
    "database": "UP",
    "redis": "UP",
    "api": "UP"
  },
  "time": "2024-12-04T12:30:00Z"
}
```

### Unhealthy Response (503 Service Unavailable)

```json
{
  "status": "DOWN",
  "checks": {
    "database": "UP",
    "redis": "DOWN",
    "api": "UP"
  },
  "time": "2024-12-04T12:30:00Z"
}
```

## Liveness vs Readiness

### Liveness Probe
- Checks if application is running
- Kubernetes restarts pod if fails
- Should be simple and fast
- No external dependency checks

### Readiness Probe
- Checks if application can serve traffic
- Kubernetes removes from load balancer if fails
- Can include dependency checks
- May fail temporarily during startup

## Kubernetes Integration

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: myapp
    image: myapp:latest
    ports:
    - containerPort: 8080
    livenessProbe:
      httpGet:
        path: /health/live
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
```

## Complete Example

```go
package main

import (
    "database/sql"
    "errors"
    "time"
    
    "github.com/ginjigo/ginji"
    "github.com/ginjigo/ginji/middleware"
)

var db *sql.DB

func main() {
    app := ginji.New()

    // Configure health checks
    config := middleware.DefaultHealthCheckConfig()
    config.Timeout = 5 * time.Second

    // Database health checker
    config.AddHealthChecker("database", func() error {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        return db.PingContext(ctx)
    })

    // Redis health checker
    config.AddHealthChecker("redis", func() error {
        return redisClient.Ping(context.Background()).Err()
    })

    // External API health checker
    config.AddHealthChecker("payment-api", func() error {
        resp, err := http.Get("https://api.stripe.com/healthcheck")
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != 200 {
            return errors.New("payment API unhealthy")
        }
        return nil
    })

    app.Use(middleware.HealthWithConfig(config))

    // Your routes
    app.Get("/", home)

    app.Listen(":8080")
}
```

## Best Practices

1. **Keep liveness simple** - Just check if app is running
2. **Readiness for dependencies** - Check database, cache, APIs
3. **Set timeouts** - Don't let health checks hang
4. **Fast checks** - Health checks run frequently
5. **Graceful degradation** - Consider partial health states

## Testing Health Checks

```bash
# Check liveness
curl http://localhost:8080/health/live

# Check readiness
curl http://localhost:8080/health/ready
```
