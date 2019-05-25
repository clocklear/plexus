package plex

import (
	"encoding/json"
	"io"
)

// WebhookPayload represents a payload from a plex webhook
type WebhookPayload struct {
	Event   string `json:"event"`
	User    bool   `json:"user"`
	Owner   bool   `json:"owner"`
	Account struct {
		ID    int    `json:"id"`
		Thumb string `json:"thumb"`
		Title string `json:"title"`
	} `json:"Account"`
	Server struct {
		Title string `json:"title"`
		UUID  string `json:"uuid"`
	} `json:"Server"`
	Player struct {
		Local         bool   `json:"local"`
		PublicAddress string `json:"publicAddress"`
		Title         string `json:"title"`
		UUID          string `json:"uuid"`
	} `json:"Player"`
	Metadata struct {
		LibrarySectionType   string `json:"librarySectionType"`
		RatingKey            string `json:"ratingKey"`
		Key                  string `json:"key"`
		ParentRatingKey      string `json:"parentRatingKey"`
		GrandparentRatingKey string `json:"grandparentRatingKey"`
		GUID                 string `json:"guid"`
		LibrarySectionID     int    `json:"librarySectionID"`
		Type                 string `json:"type"`
		Title                string `json:"title"`
		GrandparentKey       string `json:"grandparentKey"`
		ParentKey            string `json:"parentKey"`
		GrandparentTitle     string `json:"grandparentTitle"`
		ParentTitle          string `json:"parentTitle"`
		Summary              string `json:"summary"`
		Index                int    `json:"index"`
		ParentIndex          int    `json:"parentIndex"`
		RatingCount          int    `json:"ratingCount"`
		Thumb                string `json:"thumb"`
		Art                  string `json:"art"`
		ParentThumb          string `json:"parentThumb"`
		GrandparentThumb     string `json:"grandparentThumb"`
		GrandparentArt       string `json:"grandparentArt"`
		AddedAt              int    `json:"addedAt"`
		UpdatedAt            int    `json:"updatedAt"`
	} `json:"Metadata"`
}

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
