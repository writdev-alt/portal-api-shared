package repository

import (
	"errors"
	"fmt"

	"github.com/writdev-alt/portal-api-shared/utils"
	"gorm.io/gorm"
)

// BaseRepository defines the interface for base repository operations
type BaseRepository[T any] interface {
	// Create creates a new record
	Create(entity *T) error

	// FindByID finds a record by ID
	FindByID(id interface{}) (*T, error)

	// FindByUUID finds a record by UUID
	FindByUUID(uuid string) (*T, error)

	// FindAll retrieves all records with pagination
	FindAll(pagination *utils.Pagination, filters map[string]interface{}, orderBy string) ([]T, *utils.PaginationInfo, error)

	// FindOne finds a single record matching the conditions
	FindOne(conditions map[string]interface{}) (*T, error)

	// FindMany finds multiple records matching the conditions
	FindMany(conditions map[string]interface{}, pagination *utils.Pagination, orderBy string) ([]T, *utils.PaginationInfo, error)

	// Update updates a record
	Update(entity *T) error

	// UpdateByID updates a record by ID
	UpdateByID(id interface{}, updates map[string]interface{}) error

	// Delete soft deletes a record (if model supports soft delete)
	Delete(id interface{}) error

	// HardDelete permanently deletes a record
	HardDelete(id interface{}) error

	// Count counts records matching conditions
	Count(conditions map[string]interface{}) (int64, error)

	// Exists checks if a record exists
	Exists(conditions map[string]interface{}) (bool, error)

	// GetDB returns the underlying GORM DB instance for custom queries
	GetDB() *gorm.DB
}

// baseRepository implements BaseRepository interface
type baseRepository[T any] struct {
	db    *gorm.DB
	model T
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository[T any](db *gorm.DB) BaseRepository[T] {
	var model T
	return &baseRepository[T]{
		db:    db,
		model: model,
	}
}

// Create creates a new record
func (r *baseRepository[T]) Create(entity *T) error {
	if err := r.db.Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	return nil
}

// FindByID finds a record by ID
func (r *baseRepository[T]) FindByID(id interface{}) (*T, error) {
	var entity T
	if err := r.db.First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find record by ID: %w", err)
	}
	return &entity, nil
}

// FindByUUID finds a record by UUID
func (r *baseRepository[T]) FindByUUID(uuid string) (*T, error) {
	var entity T
	if err := r.db.First(&entity, "uuid = ?", uuid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find record by UUID: %w", err)
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *baseRepository[T]) FindAll(pagination *utils.Pagination, filters map[string]interface{}, orderBy string) ([]T, *utils.PaginationInfo, error) {
	var entities []T
	query := r.db.Model(&r.model)

	// Apply filters
	for key, value := range filters {
		query = query.Where(key, value)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Apply ordering
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if pagination != nil {
		pagination.Validate()
		query = query.Offset(pagination.Offset()).Limit(pagination.Limit())
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find records: %w", err)
	}

	// Build pagination info
	var paginationInfo *utils.PaginationInfo
	if pagination != nil {
		info := utils.NewPaginationInfo(pagination, total)
		paginationInfo = &info
	}

	return entities, paginationInfo, nil
}

// FindOne finds a single record matching the conditions
func (r *baseRepository[T]) FindOne(conditions map[string]interface{}) (*T, error) {
	var entity T
	query := r.db.Model(&r.model)

	// Apply conditions
	for key, value := range conditions {
		query = query.Where(key, value)
	}

	if err := query.First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find record: %w", err)
	}

	return &entity, nil
}

// FindMany finds multiple records matching the conditions
func (r *baseRepository[T]) FindMany(conditions map[string]interface{}, pagination *utils.Pagination, orderBy string) ([]T, *utils.PaginationInfo, error) {
	var entities []T
	query := r.db.Model(&r.model)

	// Apply conditions
	for key, value := range conditions {
		query = query.Where(key, value)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Apply ordering
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if pagination != nil {
		pagination.Validate()
		query = query.Offset(pagination.Offset()).Limit(pagination.Limit())
	}

	// Execute query
	if err := query.Find(&entities).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find records: %w", err)
	}

	// Build pagination info
	var paginationInfo *utils.PaginationInfo
	if pagination != nil {
		info := utils.NewPaginationInfo(pagination, total)
		paginationInfo = &info
	}

	return entities, paginationInfo, nil
}

// Update updates a record
func (r *baseRepository[T]) Update(entity *T) error {
	if err := r.db.Save(entity).Error; err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	return nil
}

// UpdateByID updates a record by ID
func (r *baseRepository[T]) UpdateByID(id interface{}, updates map[string]interface{}) error {
	if err := r.db.Model(&r.model).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update record by ID: %w", err)
	}
	return nil
}

// Delete soft deletes a record (if model supports soft delete)
func (r *baseRepository[T]) Delete(id interface{}) error {
	if err := r.db.Delete(&r.model, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}

// HardDelete permanently deletes a record
func (r *baseRepository[T]) HardDelete(id interface{}) error {
	if err := r.db.Unscoped().Delete(&r.model, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to hard delete record: %w", err)
	}
	return nil
}

// Count counts records matching conditions
func (r *baseRepository[T]) Count(conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&r.model)

	// Apply conditions
	for key, value := range conditions {
		query = query.Where(key, value)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// Exists checks if a record exists
func (r *baseRepository[T]) Exists(conditions map[string]interface{}) (bool, error) {
	count, err := r.Count(conditions)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetDB returns the underlying GORM DB instance for custom queries
func (r *baseRepository[T]) GetDB() *gorm.DB {
	return r.db
}
