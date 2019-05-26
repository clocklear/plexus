package plex

import (
	"encoding/json"
	"io"
	"time"
)

// ActivityStore is used to read/write webhook activity to/from the given ReadWriter
type ActivityStore struct {
	maxItems int
	store    io.ReadWriter
	items    ActivityLog
}

// NewActivityStore creates an instance backed by the given ReadWriter
func NewActivityStore(store io.ReadWriter, maxItems int) (*ActivityStore, error) {
	s := ActivityStore{
		maxItems: maxItems,
		store:    store,
		items:    ActivityLog{},
	}
	err := s.reload()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *ActivityStore) reload() error {
	// Load the contents from the store
	err := json.NewDecoder(s.store).Decode(&s.items)
	// The store might be empty, which is ok
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (s *ActivityStore) commit() error {
	return json.NewEncoder(s.store).Encode(s.items)
}

// GetAll returns all items in the ActivityStore
func (s *ActivityStore) GetAll() ActivityLog {
	return s.items
}

// Add appends give given WebhookPayload to the ActivityStore.  If this causes the size of the store to exceed maxItems, the oldest item is removed.
func (s *ActivityStore) Add(wh WebhookPayload) error {
	s.items = append(s.items, LogEntry{
		ReceivedAt: time.Now(),
		Payload:    wh,
	})
	if len(s.items) > s.maxItems {
		s.items = s.items[(len(s.items) - s.maxItems):]
	}
	return s.commit()
}
