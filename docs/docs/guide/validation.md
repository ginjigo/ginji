# Validation

Ginji includes a powerful validation system with 15+ built-in validators and support for custom validators.

## Basic Validation

Use struct tags to define validation rules:

```go
type User struct {
    Email    string `json:"email" ginji:"required,email"`
    Username string `json:"username" ginji:"required,min=3,max=20"`
    Age      int    `json:"age" ginji:"required,gte=18,lte=100"`
}

app.Post("/users", func(c *ginji.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        // Validation errors returned automatically
        return
    }
    
    // Data is valid, proceed
    createUser(user)
})
```

## Built-in Validators

### Required

```go
Field string `ginji:"required"`
```

### String Validators

```go
Email    string `ginji:"email"`       // Valid email
URL      string `ginji:"url"`         // Valid URL
Alpha    string `ginji:"alpha"`       // Letters only
Numeric  string `ginji:"numeric"`     // Numbers only
Alphanum string `ginji:"alphanum"`    // Letters and numbers
```

### Length Validators

```go
Username string `ginji:"min=3,max=20"`      // String length
Password string `ginji:"len=8"`             // Exact length
Tags     []string `ginji:"min=1,max=5"`     // Slice length
```

### Numeric Validators

```go
Age    int `ginji:"gte=18,lte=100"`  // Greater/less than or equal
Score  int `ginji:"gt=0,lt=100"`     // Greater/less than
Rating int `ginji:"gte=1,lte=5"`
```

### Enum Validator

```go
Status string `ginji:"oneof=active inactive pending"`
Role   string `ginji:"oneof=admin user guest"`
```

### Regex Validator

```go
ZipCode string `ginji:"regex=^[0-9]{5}$"`
Phone   string `ginji:"regex=^\\+?[1-9]\\d{1,14}$"`
```

## Nested Struct Validation

```go
type Address struct {
    Street  string `ginji:"required"`
    City    string `ginji:"required"`
    ZipCode string `ginji:"required,regex=^[0-9]{5}$"`
}

type User struct {
    Name    string  `ginji:"required"`
    Email   string  `ginji:"required,email"`
    Address Address `ginji:"required"`
}
```

## Slice Validation

```go
type Post struct {
    Title string   `ginji:"required"`
    Tags  []string `ginji:"required,min=1,max=5"`
}
```

## Custom Validators

Register custom validation functions:

```go
func init() {
    ginji.RegisterValidator("username", func(value any) bool {
        str, ok := value.(string)
        if !ok {
            return false
        }
        // Username must start with letter and contain only alphanumeric
        if len(str) < 3 {
            return false
        }
        // Check first character is letter
        if !unicode.IsLetter(rune(str[0])) {
            return false
        }
        return true
    })
}

type User struct {
    Username string `ginji:"required,username"`
}
```

## Validation Error Response

When validation fails, a structured error is returned:

```json
{
  "error": "Validation failed",
  "status": 400,
  "errors": [
    {
      "field": "email",
      "constraint": "email",
      "message": "email validation failed"
    },
    {
      "field": "age",
      "constraint": "gte",
      "message": "gte validation failed (expected: 18)"
    }
  ]
}
```

## Manual Validation

Validate manually without binding:

```go
user := User{
    Email: "test@example.com",
    Age: 25,
}

if err := ginji.Validate(user); err != nil {
    // Handle validation error
}
```

## Complete Example

```go
package main

import (
    "github.com/ginjigo/ginji"
)

type CreatePostRequest struct {
    Title    string   `json:"title" ginji:"required,min=5,max=100"`
    Content  string   `json:"content" ginji:"required,min=10"`
    Tags     []string `json:"tags" ginji:"required,min=1,max=5"`
    Status   string   `json:"status" ginji:"oneof=draft published archived"`
    AuthorID int      `json:"author_id" ginji:"required,gt=0"`
}

type UpdateProfileRequest struct {
    Username string `json:"username" ginji:"required,min=3,max=20,alphanum"`
    Email    string `json:"email" ginji:"required,email"`
    Bio      string `json:"bio" ginji:"max=500"`
    Age      int    `json:"age" ginji:"gte=13,lte=120"`
    Website  string `json:"website" ginji:"url"`
}

func main() {
    app := ginji.New()
    app.Use(ginji.DefaultErrorHandler())

    app.Post("/posts", func(c *ginji.Context) {
        var req CreatePostRequest
        if err := c.BindJSON(&req); err != nil {
            return // Validation errors sent automatically
        }
        
        // Data is validated, create post
        post := createPost(req)
        c.JSON(ginji.StatusCreated, post)
    })

    app.Put("/profile", func(c *ginji.Context) {
        var req UpdateProfileRequest
        if err := c.BindJSON(&req); err != nil {
            return
        }
        
        // Update profile
        updateProfile(req)
        c.JSON(ginji.StatusOK, ginji.H{"message": "Profile updated"})
    })

    app.Listen(":8080")
}
```

## All Validators

| Validator | Description | Example |
|-----------|-------------|---------|
| `required` | Field must not be empty | `ginji:"required"` |
| `email` | Valid email address | `ginji:"email"` |
| `url` | Valid URL | `ginji:"url"` |
| `alpha` | Letters only | `ginji:"alpha"` |
| `numeric` | Numbers only | `ginji:"numeric"` |
| `alphanum` | Letters and numbers | `ginji:"alphanum"` |
| `min` | Minimum length/value | `ginji:"min=3"` |
| `max` | Maximum length/value | `ginji:"max=100"` |
| `len` | Exact length | `ginji:"len=10"` |
| `gt` | Greater than | `ginji:"gt=0"` |
| `gte` | Greater than or equal | `ginji:"gte=18"` |
| `lt` | Less than | `ginji:"lt=100"` |
| `lte` | Less than or equal | `ginji:"lte=120"` |
| `oneof` | Value in set | `ginji:"oneof=a b c"` |
| `regex` | Matches pattern | `ginji:"regex=^[A-Z]"` |

## Best Practices

1. **Validate at entry points** - API endpoints, not business logic
2. **Use struct tags** - Cleaner than manual validation
3. **Custom validators** - For domain-specific rules
4. **Clear error messages** - Help clients understand what's wrong
5. **Fail fast** - Validate before expensive operations
