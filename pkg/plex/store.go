package plex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
)

type Activity struct {
	ReceivedAt time.Time      `json:"receivedAt"`
	RequestID  string         `json:"requestId"`
	Payload    WebhookPayload `json:"payload"`
	ThumbPath  string         `json:"thumbPath,omitempty"`
}

// Store is used to read/write data relevant to the application.  I acknowledge that this may be an unnecessary abstraction.
type Store struct {
	db     *scribble.Driver
	dbPath string
}

// NewStore creates a JSON store instance
func NewStore(dbPath string) (*Store, error) {
	db, err := scribble.New(dbPath, nil)
	if err != nil {
		return nil, err
	}
	s := Store{
		db:     db,
		dbPath: dbPath,
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

// AddThumb saves the given thumb bytes into the Store
func (s *Store) AddThumb(reqID string, origFilename string, thumb []byte) (string, error) {
	// Determine extension from original file
	ext := filepath.Ext(origFilename)
	fp := path.Join(s.dbPath, fmt.Sprintf("%s%s", reqID, ext))
	err := ioutil.WriteFile(fp, thumb, 0644)
	return fp, err
}
