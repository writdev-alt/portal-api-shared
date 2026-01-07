# Shared Package

Shared utilities and common code for all microservices.

## 📦 Installation

Add to your service's `go.mod`:

```go
require github.com/writdev-alt/portal-api/shared v0.0.0

replace github.com/writdev-alt/portal-api/shared => ../shared
```

Or if using Git:

```go
require github.com/writdev-alt/portal-api/shared v0.1.0
```

## 📁 Structure

```
shared/
├── utils/          # Utility functions (pagination, etc.)
├── responses/      # Common response types
├── middleware/     # Shared middleware (CORS, Auth, IP Whitelist, etc.)
├── database/       # Database connection utilities
└── README.md
```

## 🚀 Usage

### Pagination

```go
import "github.com/writdev-alt/portal-api/shared/utils"

type ListRequest struct {
    Filter string `form:"filter"`
    utils.Pagination
}

func (h *Handler) List(c *gin.Context) {
    var req ListRequest
    c.ShouldBindQuery(&req)
    
    req.Pagination.Validate()
    
    query := db.Model(&Model{})
    query.Offset(req.Pagination.Offset())
    query.Limit(req.Pagination.Limit())
    
    paginationInfo := utils.NewPaginationInfo(&req.Pagination, total)
}
```

### Responses

```go
import "github.com/writdev-alt/portal-api/shared/responses"

// Error response
c.JSON(400, responses.NewErrorResponse(err))

// Message response
c.JSON(200, responses.NewMessageResponse("Success"))
```

### Middleware

```go
import "github.com/writdev-alt/portal-api/shared/middleware"

router.Use(middleware.CORS())
router.Use(middleware.Logger())
router.Use(middleware.Recovery())
router.Use(middleware.AuthMiddleware())
```

### Database

```go
import "github.com/writdev-alt/portal-api/shared/database"

config := database.GetConfigFromEnv()
db, err := database.Initialize(config)
```

## 📝 Modules

### utils
- `Pagination` - Pagination utility
- `PaginationInfo` - Pagination metadata

### responses
- `ErrorResponse` - Standard error response
- `MessageResponse` - Simple message response

### middleware
- `CORS()` - CORS middleware
- `Logger()` - Request logger
- `Recovery()` - Panic recovery
- `AuthMiddleware()` - JWT authentication
- `AdminMiddleware()` - Admin role check
- `APIKeyMiddleware()` - API key validation
- `IPWhitelist()` - IP whitelisting
- `CloudflareIPWhitelist()` - Cloudflare-only access

### database
- `Initialize()` - Database connection
- `GetConfigFromEnv()` - Load config from environment
