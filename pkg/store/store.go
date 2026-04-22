package store

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
	boltErrors "go.etcd.io/bbolt/errors"

	"github.com/flowexec/flow/pkg/filesystem"
)

const (
	cacheBucket          = "cache"
	historyBucket        = "history"
	processBucketName    = "process"
	backgroundBucketName = "background"
	storeFileName        = "store.db"

	// BucketEnv is the environment variable used to identify the current process bucket.
	BucketEnv = "FLOW_PROCESS_BUCKET"
	// RootBucket is the name of the global (root) process bucket.
	RootBucket = "root"

	openTimeout = 3 * time.Second
)

// DataStore manages structured internal state (cache, execution history, and process env vars).
// It is intentionally in pkg/ so external consumers (e.g. pro wrapper) can import it.
//
//go:generate mockgen -destination=mocks/mock_data_store.go -package=mocks github.com/flowexec/flow/pkg/store DataStore
type DataStore interface { //nolint:interfacebloat // single backing store with cache, history, and process var concerns
	SetCacheEntry(key string, value []byte) error
	GetCacheEntry(key string) ([]byte, error)
	DeleteCacheEntry(key string) error

	RecordExecution(record ExecutionRecord) error
	GetExecutionHistory(ref string, limit int) ([]ExecutionRecord, error)
	ListExecutionRefs() ([]string, error)
	DeleteExecutionHistory(ref string) error

	// Background run management (detached process tracking).
	SaveBackgroundRun(run BackgroundRun) error
	GetBackgroundRun(id string) (BackgroundRun, error)
	ListBackgroundRuns() ([]BackgroundRun, error)
	DeleteBackgroundRun(id string) error

	// Process env var management (per-execution scoped key-value storage).
	// bucketID identifies the execution scope; use EnvironmentBucket() to get the current scope.
	CreateProcessBucket(id string) error
	DeleteProcessBucket(id string) error
	SetProcessVar(bucketID, key, value string) error
	GetProcessVar(bucketID, key string) (string, error)
	GetAllProcessVars(bucketID string) (map[string]string, error)
	GetProcessVarKeys(bucketID string) ([]string, error)
	DeleteProcessVar(bucketID, key string) error

	Close() error
}

// ExecutionRecord holds metadata about a single executable run.
type ExecutionRecord struct {
	Ref       string        `json:"ref"`
	StartedAt time.Time     `json:"startedAt"`
	Duration  time.Duration `json:"duration"`
	ExitCode  int           `json:"exitCode"`
	Error     string        `json:"error,omitempty"`
	// LogArchiveID links this record to a tuikit log archive entry for cross-referencing.
	LogArchiveID string `json:"logArchiveId,omitempty"`
}

// BackgroundRunStatus represents the state of a background run.
type BackgroundRunStatus string

const (
	BackgroundRunning   BackgroundRunStatus = "running"
	BackgroundCompleted BackgroundRunStatus = "completed"
	BackgroundFailed    BackgroundRunStatus = "failed"
)

// BackgroundRun holds metadata about a detached background execution.
type BackgroundRun struct {
	ID           string              `json:"id"`
	PID          int                 `json:"pid"`
	Ref          string              `json:"ref"`
	StartedAt    time.Time           `json:"startedAt"`
	Status       BackgroundRunStatus `json:"status"`
	LogArchiveID string              `json:"logArchiveId,omitempty"`
	Error        string              `json:"error,omitempty"`
	CompletedAt  *time.Time          `json:"completedAt,omitempty"`
}

// BoltDataStore opens and closes the BBolt database for each operation, so the
// exclusive file lock is held only for the duration of a single transaction.
// This allows multiple flow processes to share the same store file safely.
type BoltDataStore struct {
	dbPath string
}

// NewDataStore creates a DataStore backed by the BBolt file at dbPath.
// If dbPath is empty, the default path is used.
// The database file is opened and closed per operation; no persistent lock is held.
func NewDataStore(dbPath string) (DataStore, error) {
	if dbPath == "" {
		dbPath = Path()
	}

	if err := filesystem.EnsureCachedDataDir(); err != nil {
		return nil, fmt.Errorf("failed to ensure cache directory: %w", err)
	}

	// Verify the path is usable by opening and immediately closing.
	db, err := bolt.Open(dbPath, 0666, &bolt.Options{Timeout: openTimeout})
	if err != nil {
		return nil, fmt.Errorf("failed to open data store: %w", err)
	}
	_ = db.Close()

	return &BoltDataStore{dbPath: dbPath}, nil
}

// open acquires the BBolt file lock, runs fn, and releases the lock.
func (s *BoltDataStore) open(fn func(db *bolt.DB) error) error {
	db, err := bolt.Open(s.dbPath, 0666, &bolt.Options{Timeout: openTimeout})
	if err != nil {
		return fmt.Errorf("failed to open data store: %w", err)
	}
	defer db.Close()
	return fn(db)
}

// Path returns the default store database path.
func Path() string {
	return filepath.Join(filesystem.CachedDataDirPath(), storeFileName)
}

// EnvironmentBucket returns the process bucket ID for the current execution scope,
// determined by the FLOW_PROCESS_BUCKET environment variable. Falls back to RootBucket.
func EnvironmentBucket() string {
	id := RootBucket
	if val, set := os.LookupEnv(BucketEnv); set && val != "" {
		id = val
	}
	replacer := strings.NewReplacer(":", "_", "/", "_", " ", "_")
	return replacer.Replace(id)
}

// DestroyStore removes the store database file entirely.
func DestroyStore() error {
	err := os.Remove(Path())
	if err != nil && !isNotExist(err) {
		return fmt.Errorf("failed to destroy store: %w", err)
	}
	return nil
}

// ---- cache bucket ----

func (s *BoltDataStore) SetCacheEntry(key string, value []byte) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(cacheBucket))
			if err != nil {
				return fmt.Errorf("failed to open cache bucket: %w", err)
			}
			if err := b.Put([]byte(key), value); err != nil {
				return fmt.Errorf("failed to set cache entry %s: %w", key, err)
			}
			return nil
		})
	})
}

func (s *BoltDataStore) GetCacheEntry(key string) ([]byte, error) {
	var value []byte
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(cacheBucket))
			if b == nil {
				return nil
			}
			v := b.Get([]byte(key))
			if v != nil {
				value = make([]byte, len(v))
				copy(value, v)
			}
			return nil
		})
	})
	return value, err
}

func (s *BoltDataStore) DeleteCacheEntry(key string) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(cacheBucket))
			if b == nil {
				return nil
			}
			return b.Delete([]byte(key))
		})
	})
}

// ---- history bucket ----

func (s *BoltDataStore) RecordExecution(record ExecutionRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal execution record: %w", err)
	}
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			parent, err := tx.CreateBucketIfNotExists([]byte(historyBucket))
			if err != nil {
				return fmt.Errorf("failed to open history bucket: %w", err)
			}
			refBucket, err := parent.CreateBucketIfNotExists([]byte(record.Ref))
			if err != nil {
				return fmt.Errorf("failed to open ref bucket for %s: %w", record.Ref, err)
			}
			seq, err := refBucket.NextSequence()
			if err != nil {
				return fmt.Errorf("failed to generate sequence for %s: %w", record.Ref, err)
			}
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, seq)
			return refBucket.Put(key, data)
		})
	})
}

func (s *BoltDataStore) GetExecutionHistory(ref string, limit int) ([]ExecutionRecord, error) {
	var records []ExecutionRecord
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			parent := tx.Bucket([]byte(historyBucket))
			if parent == nil {
				return nil
			}
			refBucket := parent.Bucket([]byte(ref))
			if refBucket == nil {
				return nil
			}
			var all [][]byte
			_ = refBucket.ForEach(func(_, v []byte) error {
				cp := make([]byte, len(v))
				copy(cp, v)
				all = append(all, cp)
				return nil
			})
			start := 0
			if limit > 0 && len(all) > limit {
				start = len(all) - limit
			}
			for _, v := range all[start:] {
				var rec ExecutionRecord
				if err := json.Unmarshal(v, &rec); err != nil {
					return fmt.Errorf("failed to unmarshal execution record: %w", err)
				}
				records = append(records, rec)
			}
			return nil
		})
	})
	return records, err
}

func (s *BoltDataStore) ListExecutionRefs() ([]string, error) {
	var refs []string
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			parent := tx.Bucket([]byte(historyBucket))
			if parent == nil {
				return nil
			}
			return parent.ForEach(func(k, v []byte) error {
				if v == nil { // sub-bucket, not a key-value pair
					refs = append(refs, string(k))
				}
				return nil
			})
		})
	})
	return refs, err
}

func (s *BoltDataStore) DeleteExecutionHistory(ref string) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			parent := tx.Bucket([]byte(historyBucket))
			if parent == nil {
				return nil
			}
			err := parent.DeleteBucket([]byte(ref))
			if err != nil && !isNotFound(err) {
				return fmt.Errorf("failed to delete history for ref %s: %w", ref, err)
			}
			return nil
		})
	})
}

// ---- process bucket ----

// processBucketFor returns the sub-bucket for id nested under the top-level "process" bucket.
func processBucketFor(tx *bolt.Tx, id string, create bool) (*bolt.Bucket, error) {
	var parent *bolt.Bucket
	var err error
	if create {
		parent, err = tx.CreateBucketIfNotExists([]byte(processBucketName))
		if err != nil {
			return nil, fmt.Errorf("failed to open process bucket: %w", err)
		}
		b, err := parent.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return nil, fmt.Errorf("failed to create process sub-bucket %s: %w", id, err)
		}
		return b, nil
	}
	parent = tx.Bucket([]byte(processBucketName))
	if parent == nil {
		return nil, nil //nolint:nilnil // nil bucket signals "not found" — callers handle it
	}
	return parent.Bucket([]byte(id)), nil
}

func (s *BoltDataStore) CreateProcessBucket(id string) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			_, err := processBucketFor(tx, id, true)
			return err
		})
	})
}

func (s *BoltDataStore) DeleteProcessBucket(id string) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			parent := tx.Bucket([]byte(processBucketName))
			if parent == nil {
				return nil
			}
			err := parent.DeleteBucket([]byte(id))
			if err != nil && !isNotFound(err) {
				return fmt.Errorf("failed to delete process bucket %s: %w", id, err)
			}
			return nil
		})
	})
}

func (s *BoltDataStore) SetProcessVar(bucketID, key, value string) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			bucket, err := processBucketFor(tx, bucketID, true)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(key), []byte(value)); err != nil {
				return fmt.Errorf("failed to set var %s in bucket %s: %w", key, bucketID, err)
			}
			return nil
		})
	})
}

// GetProcessVar retrieves the value for key in bucketID, falling back to RootBucket if not found.
func (s *BoltDataStore) GetProcessVar(bucketID, key string) (string, error) {
	var value []byte
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			bucket, _ := processBucketFor(tx, bucketID, false)
			if bucket != nil {
				value = bucket.Get([]byte(key))
			}
			if value == nil && bucketID != RootBucket {
				rootBucket, _ := processBucketFor(tx, RootBucket, false)
				if rootBucket != nil {
					value = rootBucket.Get([]byte(key))
				}
			}
			if value == nil {
				return fmt.Errorf("key %s not found in bucket %s", key, bucketID)
			}
			return nil
		})
	})
	return string(value), err
}

// GetAllProcessVars returns all key-value pairs from bucketID merged with RootBucket (bucketID wins on conflict).
func (s *BoltDataStore) GetAllProcessVars(bucketID string) (map[string]string, error) {
	m := make(map[string]string)
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			bucket, _ := processBucketFor(tx, bucketID, false)
			if bucket != nil {
				_ = bucket.ForEach(func(k, v []byte) error {
					m[string(k)] = string(v)
					return nil
				})
			}
			if bucketID != RootBucket {
				rootBucket, _ := processBucketFor(tx, RootBucket, false)
				if rootBucket != nil {
					_ = rootBucket.ForEach(func(k, v []byte) error {
						if _, exists := m[string(k)]; !exists {
							m[string(k)] = string(v)
						}
						return nil
					})
				}
			}
			return nil
		})
	})
	return m, err
}

// GetProcessVarKeys returns all keys from bucketID merged with RootBucket.
func (s *BoltDataStore) GetProcessVarKeys(bucketID string) ([]string, error) {
	var keys []string
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			processKeys := make(map[string]bool)
			bucket, _ := processBucketFor(tx, bucketID, false)
			if bucket != nil {
				_ = bucket.ForEach(func(k, _ []byte) error {
					key := string(k)
					keys = append(keys, key)
					processKeys[key] = true
					return nil
				})
			}
			if bucketID != RootBucket {
				rootBucket, _ := processBucketFor(tx, RootBucket, false)
				if rootBucket != nil {
					_ = rootBucket.ForEach(func(k, _ []byte) error {
						if key := string(k); !processKeys[key] {
							keys = append(keys, key)
						}
						return nil
					})
				}
			}
			return nil
		})
	})
	return keys, err
}

func (s *BoltDataStore) DeleteProcessVar(bucketID, key string) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			bucket, err := processBucketFor(tx, bucketID, true)
			if err != nil {
				return err
			}
			return bucket.Delete([]byte(key))
		})
	})
}

// ---- background bucket ----

func (s *BoltDataStore) SaveBackgroundRun(run BackgroundRun) error {
	data, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("failed to marshal background run: %w", err)
	}
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(backgroundBucketName))
			if err != nil {
				return fmt.Errorf("failed to open background bucket: %w", err)
			}
			return b.Put([]byte(run.ID), data)
		})
	})
}

func (s *BoltDataStore) GetBackgroundRun(id string) (BackgroundRun, error) {
	var run BackgroundRun
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(backgroundBucketName))
			if b == nil {
				return fmt.Errorf("background run %s not found", id)
			}
			v := b.Get([]byte(id))
			if v == nil {
				return fmt.Errorf("background run %s not found", id)
			}
			return json.Unmarshal(v, &run)
		})
	})
	return run, err
}

func (s *BoltDataStore) ListBackgroundRuns() ([]BackgroundRun, error) {
	var runs []BackgroundRun
	err := s.open(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(backgroundBucketName))
			if b == nil {
				return nil
			}
			return b.ForEach(func(_, v []byte) error {
				var run BackgroundRun
				if err := json.Unmarshal(v, &run); err != nil {
					return fmt.Errorf("failed to unmarshal background run: %w", err)
				}
				runs = append(runs, run)
				return nil
			})
		})
	})
	return runs, err
}

func (s *BoltDataStore) DeleteBackgroundRun(id string) error {
	return s.open(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(backgroundBucketName))
			if b == nil {
				return nil
			}
			return b.Delete([]byte(id))
		})
	})
}

// Close is a no-op for the per-operation store — each operation opens and closes the DB itself.
func (s *BoltDataStore) Close() error {
	return nil
}

// MigrateProcessBuckets moves any legacy top-level exec-ref buckets (created before the "process"
// parent bucket was introduced) into the nested structure under the "process" bucket.
// This is safe to call multiple times; already-migrated buckets are skipped.
func MigrateProcessBuckets(dbPath string) error {
	if dbPath == "" {
		dbPath = Path()
	}
	db, err := bolt.Open(dbPath, 0666, &bolt.Options{Timeout: openTimeout})
	if err != nil {
		return fmt.Errorf("failed to open db for migration: %w", err)
	}
	defer db.Close()

	reserved := map[string]bool{
		processBucketName: true,
		cacheBucket:       true,
		historyBucket:     true,
	}

	var legacy []string
	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			n := string(name)
			if !reserved[n] {
				legacy = append(legacy, n)
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("failed to enumerate buckets for migration: %w", err)
	}
	if len(legacy) == 0 {
		return nil
	}

	return db.Update(func(tx *bolt.Tx) error {
		parent, err := tx.CreateBucketIfNotExists([]byte(processBucketName))
		if err != nil {
			return fmt.Errorf("failed to create process bucket: %w", err)
		}
		for _, name := range legacy {
			src := tx.Bucket([]byte(name))
			if src == nil {
				continue
			}
			if parent.Bucket([]byte(name)) != nil {
				continue
			}
			dst, err := parent.CreateBucket([]byte(name))
			if err != nil {
				return fmt.Errorf("failed to create nested bucket %s: %w", name, err)
			}
			if err := src.ForEach(func(k, v []byte) error {
				return dst.Put(k, v)
			}); err != nil {
				return fmt.Errorf("failed to migrate bucket %s: %w", name, err)
			}
			if err := tx.DeleteBucket([]byte(name)); err != nil {
				return fmt.Errorf("failed to delete legacy bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

func isNotFound(err error) bool {
	return err != nil && err.Error() == boltErrors.ErrBucketNotFound.Error()
}

func isNotExist(err error) bool {
	return os.IsNotExist(err)
}
