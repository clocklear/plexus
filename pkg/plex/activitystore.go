package plex

import (
	"io"
	"encoding/json"
)

// ActivityStore is used to read/write webhook activity to/from the given ReadWriter
type ActivityStore struct {
	maxItems int
	store    io.ReadWriter
	items    []WebhookPayload
}

// NewActivityStore creates an instance backed by the given ReadWriter
func NewActivityStore(store io.ReadWriter, maxItems int) (*ActivityStore, error) {
	s := ActivityStore{
		maxItems: maxItems,
		store:    store,
	}
	err := s.reload()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *ActivityStore) reload() error {
	// Load the contents from the store
	return json.NewDecoder(s.store).Decode(&s.items)
}

func (s *ActivityStore) commit() error {
	return json.NewEncoder(s.store).Encode(s.items)
}

// GetAll returns all items in the ActivityStore
func (s *ActivityStore) GetAll() []WebhookPayload {
	return s.items
}

// Add appends give given WebhookPayload to the ActivityStore.  If this causes the size of the store to exceed maxItems, the oldest item is removed.
func (s *ActivityStore) Add(wh WebhookPayload) error {
	s.items = append(s.items, wh)
	if len(s.items) > s.maxItems {
		s.items = s.items[(len(s.items) - s.maxItems):]
	}
	return s.commit()
}
