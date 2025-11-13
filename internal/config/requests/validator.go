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
	// This combines reference validation and root level validation in a single pass
	if err := validateReferencesAndRootLevel(config.Values); err != nil {
		return err
	}

	// Validate maximum nesting depth (3 folders: root -> nested -> nested -> request)
	if err := validateMaxNestingDepth(config.Values); err != nil {
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

// validateReferencesAndRootLevel validates references and root level items efficiently
// Time complexity: O(n*m) where n is number of items, m is average number of children
// Space complexity: O(n) for the referencedIDs map
func validateReferencesAndRootLevel(allItems map[string]Item) error {
	// Collect all referenced child IDs in a single pass
	referencedIDs := make(map[string]bool)

	// First pass: collect all referenced IDs and check for circular references
	for id, item := range allItems {
		if item.Children != nil {
			for _, childID := range item.Children {
				// Check for circular reference (item cannot reference itself)
				if childID == id {
					return fmt.Errorf("circular reference detected: item '%s' references itself", id)
				}
				referencedIDs[childID] = true
			}
		}
	}

	// Second pass: check root level items (must be folders)
	for id, item := range allItems {
		if !referencedIDs[id] {
			// This is a root level item - must be a folder
			if item.Type != ItemTypeFolder {
				return fmt.Errorf("root level item '%s' must be a folder, but got type '%s'", id, item.Type)
			}
		}
	}

	// Third pass: verify all referenced IDs exist
	// This catches cases where a child ID is referenced but doesn't exist in allItems
	for childID := range referencedIDs {
		if _, exists := allItems[childID]; !exists {
			return fmt.Errorf("child reference '%s' does not exist", childID)
		}
	}

	return nil
}

const (
	// MaxFolderDepth is the maximum allowed depth of folder nesting
	// Structure: root (level 0) -> nested (level 1) -> nested (level 2) -> request
	// This means maximum 3 folders in the chain
	MaxFolderDepth = 3
)

// validateMaxNestingDepth validates that folder nesting doesn't exceed MaxFolderDepth
// Time complexity: O(n*m) where n is number of items, m is average number of children
// Space complexity: O(n) for recursion stack and visited map
func validateMaxNestingDepth(allItems map[string]Item) error {
	// Find all root level items (not referenced as children)
	referencedIDs := make(map[string]bool)
	for _, item := range allItems {
		if item.Children != nil {
			for _, childID := range item.Children {
				referencedIDs[childID] = true
			}
		}
	}

	// Track visited items to avoid infinite loops (defensive check)
	visited := make(map[string]bool)

	// Validate depth starting from each root level folder
	for id, item := range allItems {
		if !referencedIDs[id] && item.Type == ItemTypeFolder {
			if err := validateFolderDepth(id, allItems, 0, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateFolderDepth recursively validates folder depth
func validateFolderDepth(itemID string, allItems map[string]Item, currentDepth int, visited map[string]bool) error {
	// Check for cycles (defensive check)
	if visited[itemID] {
		return fmt.Errorf("circular reference detected in folder depth validation for item '%s'", itemID)
	}
	visited[itemID] = true
	defer delete(visited, itemID)

	item, exists := allItems[itemID]
	if !exists {
		return fmt.Errorf("item '%s' not found during depth validation", itemID)
	}

	// Only folders contribute to depth
	if item.Type == ItemTypeFolder {
		// Check if current depth exceeds maximum (depth 0, 1, 2 are allowed, 3+ is not)
		// MaxFolderDepth = 3 means: root(0) -> nested(1) -> nested(2) -> request
		if currentDepth >= MaxFolderDepth {
			return fmt.Errorf("folder '%s' exceeds maximum nesting depth of %d levels", itemID, MaxFolderDepth)
		}

		// Recursively check children
		if item.Children != nil {
			for _, childID := range item.Children {
				childItem, exists := allItems[childID]
				if !exists {
					continue // Already validated in validateReferencesAndRootLevel
				}

				// If child is a folder, increment depth; if it's a request, depth stays the same
				nextDepth := currentDepth
				if childItem.Type == ItemTypeFolder {
					nextDepth = currentDepth + 1
					// Check before recursing: if nextDepth would exceed max, this folder can't have folder children
					if nextDepth >= MaxFolderDepth {
						return fmt.Errorf("folder '%s' at depth %d cannot contain nested folders (maximum depth is %d)", itemID, currentDepth, MaxFolderDepth)
					}
				}

				if err := validateFolderDepth(childID, allItems, nextDepth, visited); err != nil {
					return err
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
