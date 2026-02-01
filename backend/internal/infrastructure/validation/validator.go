package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/onurerdog4n/search-engine/internal/domain/entity"
	"github.com/onurerdog4n/search-engine/internal/domain/errors"
	"github.com/onurerdog4n/search-engine/internal/domain/port"
)

// Validator wraps go-playground/validator
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateSearchParams validates search parameters
func (v *Validator) ValidateSearchParams(params *port.SearchParams) error {
	// Query length check
	if len(params.Query) > 100 {
		return errors.NewValidationError("query", "query too long (max 100 characters)", params.Query)
	}

	// Page number check
	if params.Page < 1 {
		return errors.NewValidationError("page", "page must be >= 1", params.Page)
	}
	if params.Page > 1000 {
		return errors.NewValidationError("page", "page too large (max 1000)", params.Page)
	}

	// Page size check
	if params.PageSize < 1 {
		return errors.NewValidationError("page_size", "page_size must be >= 1", params.PageSize)
	}
	if params.PageSize > 50 {
		return errors.NewValidationError("page_size", "page_size too large (max 50)", params.PageSize)
	}

	// Sort by check
	if params.SortBy != "" && params.SortBy != "popularity" && params.SortBy != "relevance" {
		return errors.NewValidationError("sort_by", "invalid sort_by (must be 'popularity' or 'relevance')", params.SortBy)
	}

	// Content type check
	if params.ContentType != "" &&
		params.ContentType != entity.ContentTypeVideo &&
		params.ContentType != entity.ContentTypeArticle {
		return errors.NewValidationError("content_type", "invalid content_type (must be 'video' or 'article')", params.ContentType)
	}

	return nil
}

// SanitizeQuery sanitizes search query
func (v *Validator) SanitizeQuery(query string) string {
	// Remove leading/trailing whitespace
	// In production, add more sanitization as needed
	return query
}

// ValidateStruct validates any struct with validation tags
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Return first validation error
			for _, e := range validationErrors {
				return fmt.Errorf("validation failed on field '%s': %s", e.Field(), e.Tag())
			}
		}
		return err
	}
	return nil
}
