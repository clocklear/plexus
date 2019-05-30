package plex

import (
	"io"
	"encoding/json"
	"net/http"
	"fmt"

	"github.com/Jeffail/gabs"
	"github.com/go-kit/kit/log"
)

// NewConfig returns an instance of config loaded from the given io.Reader
func NewConfig(r io.Reader) (Config, error) {
	cfg := Config{}
	err := json.NewDecoder(r).Decode(&cfg)
	// The store might be empty, which is ok
	if err != nil {
		return cfg, err
	}
	for i, t := range cfg.Triggers {
		cfg.Triggers[i].ParsedActions = []Action{}
		for _, ra := range t.RawActions {
			switch ra.Type {
			case "webhook":
				c, err := gabs.Consume(ra.Config)
				if err != nil {
					return cfg, fmt.Errorf("invalid webhook action configuration specified: %v", err)
				}
				url, ok := c.Path("url").Data().(string)
				if !ok {
					return cfg, fmt.Errorf("invalid webhook action specified; missing URL")
				}
				act, ok := c.Path("action").Data().(string)
				if !ok {
					act = "GET"
				}
				cfg.Triggers[i].ParsedActions = append(cfg.Triggers[i].ParsedActions, WebhookAction{
					URL: url,
					Action: act,
				})
			default:
				// Nothing, this is something we don't know how to handle
			}
		}
	}
	return cfg, nil
	
}

// Config represents a plexus config
type Config struct {
	Triggers []Trigger `json:"triggers"`
}

// Handle uses the current configuration to transact the given webhookpayload
func (c Config) Handle(logger log.Logger, pl WebhookPayload, raw []byte) error {
	m := false
	for _, t := range c.Triggers {
		if !t.IsMatch(raw) {
			continue
		}
		m = true
		logger.Log("msg", "matched trigger, executing actions")
		// Must be a match
		for _, a := range t.ParsedActions {
			err := a.Execute(logger, pl)
			if err != nil {
				return err
			}
		}
	}
	if !m {
		logger.Log("msg", "received hook, but did not match any configured triggers")
	}
	return nil
}

// Trigger is a configuration for tying a specific Plex webhook to a set of desired actions
type Trigger struct {
	Properties map[string]interface{} `json:"properties"`
	RawActions []RawAction `json:"actions"`
	ParsedActions []Action `json:"-"`
}

// IsMatch determines if the Trigger matches the given webhook payload
func (t Trigger) IsMatch(payload []byte) bool {
	cnt, err := gabs.ParseJSON(payload)
	if err != nil {
		return false
	}
	// Iterate properties and desired values
	for k, v := range t.Properties {
		// If we encounter a non-match, short-circuit and return false
		if cnt.Path(k).Data() != v {
			return false
		}
	}
	// If we get here, must be a match
	return true
}

// RawAction is the definition of a thing that should occur when a Trigger matches a Plex webhook
type RawAction struct {
	Type string `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// Action represents a type of thing to be done for a webhook payload
type Action interface {
	Execute(logger log.Logger, payload WebhookPayload) error
}

type WebhookAction struct {
	URL string
	Action string
}

func (w WebhookAction) Execute(logger log.Logger, payload WebhookPayload) error {
	// Simple, make a web request to the desired URL
	c := &http.Client{}
	req, err := http.NewRequest(w.Action, w.URL, nil)
	if err != nil {
		return err
	}
	logger.Log("action", "webhook", "msg", "firing webhook", "verb", w.Action, "url", w.URL)
	// Don't care about response for now
	_, err = c.Do(req)
	return err
}