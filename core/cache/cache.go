package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketSessions  = []byte("sessions")
	bucketMessages  = []byte("messages")
	bucketMeta      = []byte("meta")
	keyMessageBytes = []byte("messageBytes")
)

// Session is a cached conversation container.
type Session struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	LastMessageAt time.Time `json:"lastMessageAt"`
	MessageCount  int       `json:"messageCount"`
}

// Message is a cached chat message.
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Config controls cache size and eviction.
type Config struct {
	MaxMessages int
	MaxBytes    int64
}

// Cache is a persistent bbolt-backed read-only offline cache.
type Cache struct {
	db     *bolt.DB
	config Config
	mu     sync.Mutex
}

// Open opens (or creates) a cache at path.
func Open(path string, cfg Config) (*Cache, error) {
	if cfg.MaxMessages <= 0 {
		cfg.MaxMessages = 10000
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = 50 * 1024 * 1024 // 50 MB
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open cache db: %w", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketSessions, bucketMessages, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return &Cache{db: db, config: cfg}, nil
}

// Close closes the cache.
func (c *Cache) Close() error {
	return c.db.Close()
}

// PutSession stores a session.
func (c *Cache) PutSession(s Session) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSessions)
		data, err := json.Marshal(s)
		if err != nil {
			return err
		}
		return b.Put([]byte(s.ID), data)
	})
}

// GetSession retrieves a session by ID.
func (c *Cache) GetSession(id string) (Session, error) {
	var s Session
	err := c.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketSessions).Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &s)
	})
	return s, err
}

// ListSessions returns sessions ordered by last activity descending.
func (c *Cache) ListSessions() ([]Session, error) {
	var sessions []Session
	err := c.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSessions).ForEach(func(k, v []byte) error {
			var s Session
			if err := json.Unmarshal(v, &s); err != nil {
				return err
			}
			sessions = append(sessions, s)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastMessageAt.After(sessions[j].LastMessageAt)
	})
	return sessions, nil
}

// PutMessage stores a message and updates session metadata.
func (c *Cache) PutMessage(m Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMessages)
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		if err := b.Put([]byte(m.ID), data); err != nil {
			return err
		}

		// Update message bytes.
		meta := tx.Bucket(bucketMeta)
		var total int64
		if v := meta.Get(keyMessageBytes); v != nil {
			_ = json.Unmarshal(v, &total)
		}
		total += int64(len(data))
		v, _ := json.Marshal(total)
		if err := meta.Put(keyMessageBytes, v); err != nil {
			return err
		}

		// Update session count and last message time.
		sb := tx.Bucket(bucketSessions)
		var s Session
		if sv := sb.Get([]byte(m.SessionID)); sv != nil {
			_ = json.Unmarshal(sv, &s)
		}
		s.ID = m.SessionID
		s.LastMessageAt = m.Timestamp
		s.MessageCount++
		sv, _ := json.Marshal(s)
		return sb.Put([]byte(m.SessionID), sv)
	})
}

// GetMessagesBySession returns messages for a session, oldest first.
func (c *Cache) GetMessagesBySession(sessionID string) ([]Message, error) {
	var msgs []Message
	err := c.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMessages).ForEach(func(k, v []byte) error {
			var m Message
			if err := json.Unmarshal(v, &m); err != nil {
				return err
			}
			if m.SessionID == sessionID {
				msgs = append(msgs, m)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Timestamp.Before(msgs[j].Timestamp)
	})
	return msgs, nil
}

// DeleteSessionAndMessages removes a session and all its messages.
func (c *Cache) DeleteSessionAndMessages(sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.db.Update(func(tx *bolt.Tx) error {
		sb := tx.Bucket(bucketSessions)
		mb := tx.Bucket(bucketMessages)
		meta := tx.Bucket(bucketMeta)

		var freed int64
		if err := mb.ForEach(func(k, v []byte) error {
			var m Message
			if err := json.Unmarshal(v, &m); err != nil {
				return err
			}
			if m.SessionID == sessionID {
				freed += int64(len(v))
				return mb.Delete(k)
			}
			return nil
		}); err != nil {
			return err
		}

		var total int64
		if v := meta.Get(keyMessageBytes); v != nil {
			_ = json.Unmarshal(v, &total)
		}
		total -= freed
		if total < 0 {
			total = 0
		}
		v, _ := json.Marshal(total)
		if err := meta.Put(keyMessageBytes, v); err != nil {
			return err
		}

		return sb.Delete([]byte(sessionID))
	})
}

// Evict removes oldest sessions/messages until limits are satisfied.
func (c *Cache) Evict() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		var totalBytes int64
		var totalCount int
		var oldest Session
		var hasOldest bool

		err := c.db.View(func(tx *bolt.Tx) error {
			if v := tx.Bucket(bucketMeta).Get(keyMessageBytes); v != nil {
				_ = json.Unmarshal(v, &totalBytes)
			}
			return tx.Bucket(bucketSessions).ForEach(func(k, v []byte) error {
				var s Session
				if err := json.Unmarshal(v, &s); err != nil {
					return err
				}
				totalCount += s.MessageCount
				if !hasOldest || s.LastMessageAt.Before(oldest.LastMessageAt) {
					oldest = s
					hasOldest = true
				}
				return nil
			})
		})
		if err != nil {
			return err
		}

		if (totalCount <= c.config.MaxMessages && totalBytes <= c.config.MaxBytes) || !hasOldest {
			return nil
		}

		if err := c.deleteSessionAndMessagesNoLock(oldest.ID); err != nil {
			return err
		}
	}
}

func (c *Cache) deleteSessionAndMessagesNoLock(sessionID string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		sb := tx.Bucket(bucketSessions)
		mb := tx.Bucket(bucketMessages)
		meta := tx.Bucket(bucketMeta)

		var freed int64
		_ = mb.ForEach(func(k, v []byte) error {
			var m Message
			if err := json.Unmarshal(v, &m); err == nil && m.SessionID == sessionID {
				freed += int64(len(v))
				_ = mb.Delete(k)
			}
			return nil
		})

		var total int64
		if v := meta.Get(keyMessageBytes); v != nil {
			_ = json.Unmarshal(v, &total)
		}
		total -= freed
		if total < 0 {
			total = 0
		}
		v, _ := json.Marshal(total)
		_ = meta.Put(keyMessageBytes, v)
		return sb.Delete([]byte(sessionID))
	})
}

// ErrNotFound is returned when a record is missing.
var ErrNotFound = errors.New("cache: not found")
