# Base Repository Pattern

Base repository pattern provides a generic, reusable data access layer for all services. It abstracts common CRUD operations and provides a consistent interface for database interactions.

## Features

- **Generic Type Support**: Works with any GORM model type
- **CRUD Operations**: Create, Read, Update, Delete operations
- **Soft Delete Support**: Automatic soft delete handling
- **Pagination**: Built-in pagination support
- **Filtering**: Flexible filtering capabilities
- **Query Building**: Access to underlying GORM DB for custom queries

## Usage

### Basic Setup

```go
package repository

import (
    "github.com/writdev-alt/portal-api-shared/repository"
    "gorm.io/gorm"
)

// In your service, initialize the repository
type WalletRepository struct {
    repository.BaseRepository[models.Wallet]
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
    return &WalletRepository{
        BaseRepository: repository.NewBaseRepository[models.Wallet](db),
    }
}
```

### Available Methods

#### Create
```go
wallet := &models.Wallet{
    UserID: 123,
    CurrencyCode: "IDR",
    // ... other fields
}

err := walletRepo.Create(wallet)
```

#### Find by ID
```go
wallet, err := walletRepo.FindByID(1)
if wallet == nil {
    // Record not found
}
```

#### Find by UUID
```go
wallet, err := walletRepo.FindByUUID("550e8400-e29b-41d4-a716-446655440000")
```

#### Find All with Pagination
```go
pagination := &utils.Pagination{
    Page:    1,
    PerPage: 20,
}

filters := map[string]interface{}{
    "user_id": 123,
    "status":  1,
}

wallets, paginationInfo, err := walletRepo.FindAll(pagination, filters, "created_at DESC")
```

#### Find One with Conditions
```go
conditions := map[string]interface{}{
    "user_id":       123,
    "currency_code": "IDR",
}

wallet, err := walletRepo.FindOne(conditions)
```

#### Find Many with Conditions and Pagination
```go
conditions := map[string]interface{}{
    "status": 1,
}

pagination := &utils.Pagination{
    Page:    1,
    PerPage: 20,
}

wallets, paginationInfo, err := walletRepo.FindMany(conditions, pagination, "created_at DESC")
```

#### Update
```go
wallet.Balance = 100000
err := walletRepo.Update(wallet)
```

#### Update by ID
```go
updates := map[string]interface{}{
    "balance": 100000,
    "status":  1,
}

err := walletRepo.UpdateByID(1, updates)
```

#### Delete (Soft Delete)
```go
err := walletRepo.Delete(1)
```

#### Hard Delete
```go
err := walletRepo.HardDelete(1)
```

#### Count
```go
conditions := map[string]interface{}{
    "user_id": 123,
}

count, err := walletRepo.Count(conditions)
```

#### Exists
```go
conditions := map[string]interface{}{
    "user_id":       123,
    "currency_code": "IDR",
}

exists, err := walletRepo.Exists(conditions)
```

#### Custom Queries
```go
// Get underlying GORM DB for custom queries
db := walletRepo.GetDB()

var wallets []models.Wallet
err := db.Where("balance > ?", 1000).
    Where("status = ?", 1).
    Order("created_at DESC").
    Find(&wallets).Error
```

## Advanced Usage

### Extending Base Repository

You can extend the base repository with custom methods:

```go
type WalletRepository struct {
    repository.BaseRepository[models.Wallet]
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
    return &WalletRepository{
        BaseRepository: repository.NewBaseRepository[models.Wallet](db),
    }
}

// Custom method
func (r *WalletRepository) FindByUserAndCurrency(userID uint64, currencyCode string) (*models.Wallet, error) {
    conditions := map[string]interface{}{
        "user_id":       userID,
        "currency_code": currencyCode,
    }
    return r.FindOne(conditions)
}

// Custom method with complex query
func (r *WalletRepository) GetTotalBalanceByUser(userID uint64) (float64, error) {
    var total float64
    db := r.GetDB()
    
    err := db.Model(&models.Wallet{}).
        Where("user_id = ?", userID).
        Select("COALESCE(SUM(balance), 0)").
        Scan(&total).Error
    
    return total, err
}
```

### Using in Handlers

```go
type WalletHandler struct {
    walletRepo *repository.WalletRepository
}

func NewWalletHandler(db *gorm.DB) *WalletHandler {
    return &WalletHandler{
        walletRepo: repository.NewWalletRepository(db),
    }
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
    id := c.Param("id")
    
    wallet, err := h.walletRepo.FindByID(id)
    if err != nil {
        sharedResponses.FailWithMessage(c, "Failed to fetch wallet")
        return
    }
    
    if wallet == nil {
        sharedResponses.NotFoundError(c, sharedResponses.ServiceCodeWallet, sharedResponses.CaseCodeNotFound, "Wallet not found")
        return
    }
    
    sharedResponses.OkWithData(c, wallet)
}
```

## Benefits

1. **Consistency**: All services use the same data access pattern
2. **DRY Principle**: No code duplication for common operations
3. **Type Safety**: Generic types ensure compile-time safety
4. **Maintainability**: Changes to base repository affect all services
5. **Testability**: Easy to mock and test
6. **Flexibility**: Can extend with custom methods as needed

## Notes

- The repository automatically handles soft deletes if your model uses `gorm.DeletedAt`
- All methods return `nil` for the entity when record is not found (not an error)
- Use `GetDB()` for complex queries that don't fit standard operations
- Pagination is optional - pass `nil` to get all records
