package models

import (
	"paperbox/internal/config/requests"
)

// Item is re-exported from requests for Wails bindings
type Item = requests.Item

// Requests represents the requests structure for Wails bindings
type Requests struct {
	Values    map[string]Item `json:"values"`
	RootOrder []string        `json:"rootOrder,omitempty"`
}

// NewRequests creates a new empty Requests structure
func NewRequests() Requests {
	return Requests{
		Values: make(map[string]Item),
	}
}

// MarshalJSON implements json.Marshaler interface
func (r Requests) MarshalJSON() ([]byte, error) {
	return requests.MarshalRequests(r.Values)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (r *Requests) UnmarshalJSON(data []byte) error {
	values, err := requests.UnmarshalRequests(data)
	if err != nil {
		return err
	}
	r.Values = values
	if r.Values == nil {
		r.Values = make(map[string]Item)
	}
	return nil
}
