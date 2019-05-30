package plex

import (
	"encoding/json"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
)

type Activity struct {
	ReceivedAt time.Time      `json:"receivedAt"`
	RequestID  string         `json:"requestId"`
	Payload    WebhookPayload `json:"payload"`
}

// Store is used to read/write data relevant to the application.  I acknowledge that this may be an unnecessary abstraction.
type Store struct {
	db *scribble.Driver
}

// NewStore creates a JSON store instance
func NewStore(dbPath string) (*Store, error) {
	db, err := scribble.New(dbPath, nil)
	if err != nil {
		return nil, err
	}
	s := Store{
		db: db,
	}
	return &s, nil
}

// GetAllActivity returns all Activity items in the Store
func (s *Store) GetAllActivity() ([]Activity, error) {
	acts := []Activity{}
	recs, err := s.db.ReadAll("activity")
	if err != nil {
		for _, a := range recs {
			recFound := Activity{}
			if err := json.Unmarshal([]byte(a), &recFound); err != nil {
				return acts, err
			}
			acts = append(acts, recFound)
		}
	}
	return acts, nil
}

// AddActivity appends the given Activity to the Store.
func (s *Store) AddActivity(act Activity) error {
	return s.db.Write("activity", act.RequestID, act)
}
