# Config Package

This package manages all application configurations.

## Structure

- `manager.go` - Central manager for all configs
- `requests/` - Requests configuration (HTTP requests and folders)

## Adding New Configs

To add a new config (e.g., user config):

1. Create a new subdirectory (e.g., `user/`)
2. Implement load, save, validate, and migrate functions
3. Add the config to `Manager` in `manager.go`
4. Ensure versioning is implemented

All configs should:
- Have versioning support
- Implement validation
- Support migration from older versions
- Be stored in separate JSON files

