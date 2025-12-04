# Type-Safe Routing Example

This example demonstrates Ginji's type-safe routing capabilities using Go generics.

## Features Demonstrated

- ✅ Type-safe request/response handling
- ✅ Automatic request binding and validation
- ✅ Path parameter binding
- ✅ Compile-time type checking
- ✅ Error handling with HTTPError
- ✅ Empty request/response handling
- ✅ Mixing typed and regular handlers

## Running the Example

```bash
cd examples/typed-routing
go run main.go
```

## API Endpoints

### Create User
```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"John Doe","email":"john@example.com","age":30}'
```

### List All Users
```bash
curl http://localhost:3000/api/v1/users
```

### Get User by ID
```bash
curl http://localhost:3000/api/v1/users/1
```

### Delete User
```bash
curl -X DELETE http://localhost:3000/api/v1/users/1
```

## Type-Safe Handler Benefits

1. **Compile-Time Safety**: Type mismatches are caught at compile time
2. **Automatic Binding**: Request data is automatically bound to your types
3. **Built-in Validation**: Validation tags are automatically enforced
4. **Clean Code**: Handler signatures clearly show input/output types
5. **No Manual Parsing**: No need to manually parse JSON or path params

## Example Handler

```go
typed.Post("/users", func(c *ginji.Context, req CreateUserRequest) (User, error) {
    // req is already validated and bound
    user := User{
        ID:    nextID,
        Name:  req.Name,
        Email: req.Email,
        Age:   req.Age,
    }
    return user, nil
})
```

Compare this to a traditional handler:

```go
app.Post("/users", func(c *ginji.Context) {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.AbortWithError(400, err)
        return
    }
    // ... rest of code
    c.JSON(200, user)
})
```

The typed version is cleaner, safer, and more maintainable!
