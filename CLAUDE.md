# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a generic key-value store library written in Go that uses SQLite as the backend storage. The package provides a simple, type-safe API using Go generics.

## Commands

### Testing
```bash
# Run all tests
go test -v

# Run a specific test
go test -v -run TestForEach

# Run tests with coverage
go test -v -cover
```

### Building
```bash
# Build the module
go build

# Verify module dependencies
go mod verify

# Download dependencies
go mod download
```

## Architecture

The KV store (`store.go`) is a generic type `KV[T1, T2]` where:
- `T1` is the key type
- `T2` is the value type
- Uses modernc.org/sqlite (pure Go SQLite implementation)
- Data is stored in a single table with TEXT columns for key and value

### Core Methods
- `New[T1, T2](name string)` - Creates a new KV store with SQLite database file
- `Set(key T1, value T2)` - Inserts or replaces a key-value pair
- `Get(key T1, value T2)` - Retrieves a value by key (second parameter unused)
- `Delete(key T1)` - Removes a key-value pair
- `ForEach(fn func(key T1, value T2) error)` - Iterates over all entries

### Testing Approach
Tests use the package as an external import (`package kvstore_test`) to validate the public API. Test database files are created temporarily and cleaned up after each test.