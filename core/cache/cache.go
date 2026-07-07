// Package cache provides a bbolt-backed offline read-only cache
// for sessions and messages. Used when the device is offline or
// the server is unreachable.
package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Leathal1/hermey-android/core/models"
	bolt "go.etcd.io/bbolt"
)

var (
	bucketSessions  = []byte("sessions")
	bucketMessages  = []byte("messages")
)

// Store is a bbolt-backed cache for sessions and messages.
type Store struct {
	db *bolt.DB
}

// Open opens or creates a cache database at the given path.
func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("cache: open: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketSessions, bucketMessages} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("cache: init buckets: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the cache database.
func (s *Store) Close() error {
	return s.db.Close()
}

// PutSession stores a session in the cache.
func (s *Store) PutSession(session *models.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("cache: marshal session: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSessions).Put([]byte(session.ID), data)
	})
}

// GetSession retrieves a session from the cache.
func (s *Store) GetSession(id string) (*models.Session, error) {
	var session models.Session
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketSessions).Get([]byte(id))
		if data == nil {
			return fmt.Errorf("cache: session %s not found", id)
		}
		return json.Unmarshal(data, &session)
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessions returns all cached sessions.
func (s *Store) ListSessions() ([]models.Session, error) {
	var sessions []models.Session
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSessions).ForEach(func(k, v []byte) error {
			var s models.Session
			if err := json.Unmarshal(v, &s); err != nil {
				return nil // skip corrupt entries
			}
			sessions = append(sessions, s)
			return nil
		})
	})
	return sessions, err
}

// PutMessages stores messages for a session.
func (s *Store) PutMessages(sessionID string, messages []models.ChatMessage) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("cache: marshal messages: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMessages).Put([]byte(sessionID), data)
	})
}

// GetMessages retrieves cached messages for a session.
func (s *Store) GetMessages(sessionID string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketMessages).Get([]byte(sessionID))
		if data == nil {
			return nil // no cached messages
		}
		return json.Unmarshal(data, &messages)
	})
	return messages, err
}

// EvictOldest removes sessions older than the given duration.
func (s *Store) EvictOldest(maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge)
	var evicted int
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSessions)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var session models.Session
			if err := json.Unmarshal(v, &session); err != nil {
				continue
			}
			if session.UpdatedAt.Before(cutoff) {
				b.Delete(k)
				tx.Bucket(bucketMessages).Delete(k)
				evicted++
			}
		}
		return nil
	})
	return evicted, err
}
