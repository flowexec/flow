package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	boltErrors "go.etcd.io/bbolt/errors"
)

const (
	BucketEnv  = "FLOW_PROCESS_BUCKET"
	RootBucket = "root"

	storeFileName = "store.db"
)

//go:generate mockgen -destination=mocks/mock_store.go -package=mocks github.com/flowexec/flow/internal/services/store Store
type Store interface {
	CreateBucket(id string) error
	CreateAndSetBucket(id string) (string, error)
	DeleteBucket(id string) error

	Set(key, value string) error
	Get(key string) (string, error)
	GetAll() (map[string]string, error)
	GetKeys() ([]string, error)
	Delete(key string) error

	Close() error
}

type BoltStore struct {
	db            *bolt.DB
	processBucket string
}

// NewStore creates a new store with a given db path
// If dbPath is empty, it will use the default path
func NewStore(dbPath string) (Store, error) {
	if dbPath == "" {
		dbPath = Path()
	}
	db, err := bolt.Open(dbPath, 0666, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	return &BoltStore{db: db}, nil
}

// CreateBucket creates a bucket with a given id if it doesn't exist
func (s *BoltStore) CreateBucket(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", id, err)
		}
		return nil
	})
}

// DeleteBucket deletes a bucket by its id
func (s *BoltStore) DeleteBucket(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(id))
		if err != nil {
			if errors.Is(err, boltErrors.ErrBucketNotFound) {
				return nil
			}
			return fmt.Errorf("failed to delete bucket %s: %w", id, err)
		}
		return nil
	})
}

// CreateAndSetBucket creates a temporary bucket for the process and returns the bucket's name
func (s *BoltStore) CreateAndSetBucket(id string) (string, error) {
	if err := s.CreateBucket(id); err != nil {
		return "", fmt.Errorf("failed to create bucket %s: %w", id, err)
	}
	s.processBucket = id
	if id != RootBucket {
		_ = os.Setenv(BucketEnv, id)
	}
	return id, nil
}

func EnvironmentBucket() string {
	id := RootBucket
	if val, set := os.LookupEnv(BucketEnv); set && val != "" {
		id = val
	}
	replacer := strings.NewReplacer(":", "_", "/", "_", " ", "_")
	id = replacer.Replace(id)
	return id
}

// Set stores a key-value pair in the process bucket
func (s *BoltStore) Set(key, value string) error {
	if s.processBucket == "" {
		if _, err := s.CreateAndSetBucket(RootBucket); err != nil {
			return fmt.Errorf("failed to create process bucket: %w", err)
		}
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.processBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", s.processBucket)
		}
		err := bucket.Put([]byte(key), []byte(value))
		if err != nil {
			return fmt.Errorf("failed to put key-value pair for key %s in bucket %s: %w", key, s.processBucket, err)
		}
		return nil
	})
}

// Get retrieves a value for a key from the process bucket
func (s *BoltStore) Get(key string) (string, error) {
	if s.processBucket == "" {
		if _, err := s.CreateAndSetBucket(RootBucket); err != nil {
			return "", fmt.Errorf("failed to create process bucket: %w", err)
		}
	}
	var value []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.processBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", s.processBucket)
		}
		value = bucket.Get([]byte(key))
		if value == nil && s.processBucket != RootBucket {
			rBucket := tx.Bucket([]byte(RootBucket))
			if rBucket != nil {
				value = rBucket.Get([]byte(key))
			}
		}
		if value == nil {
			return fmt.Errorf("key %s not found in bucket %s", key, s.processBucket)
		}
		return nil
	})
	return string(value), err
}

// Keys returns all keys in the process bucket
func (s *BoltStore) GetKeys() ([]string, error) {
	if s.processBucket == "" {
		if _, err := s.CreateAndSetBucket(RootBucket); err != nil {
			return nil, fmt.Errorf("failed to create process bucket: %w", err)
		}
	}

	var keys []string
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.processBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", s.processBucket)
		}

		// Collect keys from process bucket
		processKeys := make(map[string]bool)
		err := bucket.ForEach(func(k, _ []byte) error {
			key := string(k)
			keys = append(keys, key)
			processKeys[key] = true
			return nil
		})
		if err != nil {
			return err
		}

		// If not using root bucket, also collect keys from root bucket
		if s.processBucket != RootBucket {
			rBucket := tx.Bucket([]byte(RootBucket))
			if rBucket == nil {
				return nil
			}
			err := rBucket.ForEach(func(k, _ []byte) error {
				// Only add keys that aren't already in the process bucket
				if key := string(k); !processKeys[key] {
					keys = append(keys, key)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	return keys, err
}

// BucketMap returns a map of all keys and values in the process bucket
func (s *BoltStore) GetAll() (map[string]string, error) {
	if s.processBucket == "" {
		if _, err := s.CreateAndSetBucket(RootBucket); err != nil {
			return nil, fmt.Errorf("failed to create process bucket: %w", err)
		}
	}

	m := make(map[string]string)
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.processBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", s.processBucket)
		}

		// First, collect all key-value pairs from process bucket
		err := bucket.ForEach(func(k, v []byte) error {
			m[string(k)] = string(v)
			return nil
		})
		if err != nil {
			return err
		}

		// If not using root bucket, also collect from root bucket
		if s.processBucket != RootBucket {
			rBucket := tx.Bucket([]byte(RootBucket))
			if rBucket == nil {
				return nil
			}
			err := rBucket.ForEach(func(k, v []byte) error {
				key := string(k)
				// Only add if key doesn't already exist in process bucket
				if _, exists := m[key]; !exists {
					m[key] = string(v)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	return m, err
}

// Delete removes a key from the process bucket
func (s *BoltStore) Delete(key string) error {
	if s.processBucket == "" {
		if _, err := s.CreateAndSetBucket(RootBucket); err != nil {
			return fmt.Errorf("failed to create process bucket: %w", err)
		}
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.processBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", s.processBucket)
		}
		err := bucket.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("failed to delete key %s from bucket %s: %w", key, s.processBucket, err)
		}
		return nil
	})
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}

func DestroyStore() error {
	err := os.Remove(Path())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to destroy store: %w", err)
	}
	return nil
}

func Path() string {
	cacheDir := filesystem.CachedDataDirPath()
	return filepath.Join(cacheDir, storeFileName)
}
