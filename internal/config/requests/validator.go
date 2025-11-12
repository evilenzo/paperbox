package requests

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate validates the requests configuration
func Validate(config *RequestsConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// Validate basic structure using validator
	if err := validate.Struct(config); err != nil {
		return formatValidationError(err)
	}

	// Validate type-specific rules first (before checking references)
	for id, item := range config.Values {
		if err := validateItemTypeSpecificRules(item); err != nil {
			return fmt.Errorf("item %s: %w", id, err)
		}
	}

	// Validate business logic that can't be expressed in tags
	if err := validateChildrenReferences(config.Values); err != nil {
		return err
	}

	return nil
}

// validateHTTPMethod validates that the method is a valid HTTP method
func validateHTTPMethod(fl validator.FieldLevel) bool {
	method := fl.Field().String()
	if method == "" {
		return true // Empty is allowed (omitempty handles this)
	}

	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"PATCH":   true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
		"CONNECT": true,
		"TRACE":   true,
	}

	return validMethods[strings.ToUpper(method)]
}

// validateItemTypeSpecificRules validates rules that depend on item type
func validateItemTypeSpecificRules(item Item) error {
	switch item.Type {
	case ItemTypeRequest:
		// Request must have method
		if item.Method == "" {
			return fmt.Errorf("request must have a method")
		}

		// Request must have path
		if item.Path == "" {
			return fmt.Errorf("request must have a path")
		}

		// Request must not have children
		if len(item.Children) > 0 {
			return fmt.Errorf("request cannot have children")
		}

	case ItemTypeFolder:
		// Folder must not have method
		if item.Method != "" {
			return fmt.Errorf("folder cannot have a method")
		}

		// Folder must not have path
		if item.Path != "" {
			return fmt.Errorf("folder cannot have a path")
		}
	}

	return nil
}

// validateChildrenReferences validates that all children references exist
func validateChildrenReferences(allItems map[string]Item) error {
	// Collect all referenced child IDs
	referencedIDs := make(map[string]bool)
	for _, item := range allItems {
		if item.Children != nil {
			for _, childID := range item.Children {
				referencedIDs[childID] = true
			}
		}
	}

	// Check that all referenced IDs exist
	for childID := range referencedIDs {
		if _, exists := allItems[childID]; !exists {
			return fmt.Errorf("child reference '%s' does not exist", childID)
		}
	}

	// Check for circular references (simple check - item cannot reference itself)
	for id, item := range allItems {
		if item.Children != nil {
			for _, childID := range item.Children {
				if childID == id {
					return fmt.Errorf("circular reference detected: item '%s' references itself", id)
				}
			}
		}
	}

	return nil
}

// formatValidationError formats validator errors into a readable string
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, validationError := range validationErrors {
			field := validationError.Field()
			tag := validationError.Tag()
			param := validationError.Param()

			var message string
			switch tag {
			case "required":
				message = fmt.Sprintf("%s is required", field)
			case "min":
				message = fmt.Sprintf("%s must be at least %s characters", field, param)
			case "oneof":
				message = fmt.Sprintf("%s must be one of: %s", field, param)
			case "http_method":
				message = fmt.Sprintf("%s must be a valid HTTP method", field)
			default:
				message = fmt.Sprintf("%s failed validation for tag '%s'", field, tag)
			}
			messages = append(messages, message)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}
	return err
}
