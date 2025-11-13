package storage

// Storage is the interface for reading and writing configuration data.
// Different implementations can provide file-based, cloud-based, or other storage mechanisms.
type Storage interface {
	// Load reads configuration data from storage into the target.
	Load(filePath string, target interface{}) error

	// Save writes configuration data to storage.
	Save(filePath string, data interface{}) error
}

