package storage

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// ConflictResolution represents how to resolve a conflict between local and remote data.
type ConflictResolution int

const (
	// ResolutionKeepLocal keeps the local (file) version.
	ResolutionKeepLocal ConflictResolution = iota
	// ResolutionKeepRemote keeps the remote (cloud) version.
	ResolutionKeepRemote
	// ResolutionMerge attempts to merge local and remote data.
	ResolutionMerge
)

// ConflictHandler is a function that resolves conflicts between local and remote data.
// It receives both versions and returns the resolution strategy.
type ConflictHandler func(local, remote interface{}) (ConflictResolution, error)

// StorageCoordinator coordinates between file storage (authoritative) and cloud storage.
// It handles synchronization and conflict resolution.
type StorageCoordinator struct {
	file            Storage
	cloud           Storage
	conflictHandler ConflictHandler
}

// NewStorageCoordinator creates a new StorageCoordinator.
// If cloud is nil, only file storage will be used.
func NewStorageCoordinator(file Storage, cloud Storage, conflictHandler ConflictHandler) *StorageCoordinator {
	return &StorageCoordinator{
		file:            file,
		cloud:           cloud,
		conflictHandler: conflictHandler,
	}
}

// Load loads configuration from file (authoritative) and optionally merges with cloud data.
func (c *StorageCoordinator) Load(filePath string, target interface{}) error {
	// First, load from file (authoritative source)
	if err := c.file.Load(filePath, target); err != nil {
		return fmt.Errorf("failed to load from file: %w", err)
	}

	// If no cloud storage, we're done
	if c.cloud == nil {
		return nil
	}

	// Try to load from cloud
	var cloudData interface{}
	// Create a new instance of the same type as target
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
		cloudData = reflect.New(targetType).Interface()
	} else {
		cloudData = reflect.New(targetType).Elem().Interface()
	}

	cloudErr := c.cloud.Load(filePath, cloudData)
	if cloudErr != nil {
		// Cloud load failed, but file load succeeded - that's okay
		return nil
	}

	// Check if data differs
	if c.dataEqual(target, cloudData) {
		// Data is the same, no conflict
		return nil
	}

	// Data differs - resolve conflict
	if c.conflictHandler == nil {
		// No handler, keep local (file) version
		return nil
	}

	resolution, err := c.conflictHandler(target, cloudData)
	if err != nil {
		return fmt.Errorf("conflict handler error: %w", err)
	}

	switch resolution {
	case ResolutionKeepLocal:
		// Keep local (file) version - already loaded, do nothing
		return nil
	case ResolutionKeepRemote:
		// Keep remote (cloud) version - copy cloud data to target
		return c.copyData(cloudData, target)
	case ResolutionMerge:
		// Attempt to merge
		return c.mergeData(target, cloudData)
	default:
		return fmt.Errorf("unknown conflict resolution: %v", resolution)
	}
}

// Save saves configuration to file first (authoritative), then syncs to cloud if available.
func (c *StorageCoordinator) Save(filePath string, data interface{}) error {
	// Save to file first (authoritative)
	if err := c.file.Save(filePath, data); err != nil {
		return fmt.Errorf("failed to save to file: %w", err)
	}

	// If cloud storage is available, sync to cloud
	if c.cloud != nil {
		if err := c.cloud.Save(filePath, data); err != nil {
			// Cloud save failed, but file save succeeded - log but don't fail
			// In the future, this could be handled by retry logic or error reporting
			return fmt.Errorf("failed to sync to cloud (file saved successfully): %w", err)
		}
	}

	return nil
}

// dataEqual checks if two data structures are equal by comparing their JSON representation.
func (c *StorageCoordinator) dataEqual(a, b interface{}) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}

	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}

// copyData copies data from source to target using JSON marshaling/unmarshaling.
func (c *StorageCoordinator) copyData(source, target interface{}) error {
	data, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %w", err)
	}

	return nil
}

// mergeData attempts to merge local and remote data using MergePatch.
func (c *StorageCoordinator) mergeData(local, remote interface{}) error {
	// Convert remote to map for patching
	remoteJSON, err := json.Marshal(remote)
	if err != nil {
		return fmt.Errorf("failed to marshal remote: %w", err)
	}

	var remoteMap map[string]interface{}
	if err := json.Unmarshal(remoteJSON, &remoteMap); err != nil {
		return fmt.Errorf("failed to unmarshal remote to map: %w", err)
	}

	// Use MergePatch to merge remote changes into local
	return MergePatch(local, remoteMap, local)
}

