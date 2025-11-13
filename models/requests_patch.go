package models

import "paperbox/internal/config/requests"

// RequestsPatch represents a partial update to the requests configuration
// All fields are optional - only provided fields will be updated
type RequestsPatch struct {
	Values map[string]requests.Item `json:"values,omitempty"`
}
