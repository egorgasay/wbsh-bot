package config

import (
	"bot/internal/storage"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	// ErrKeyNotSet error when the key is not set.
	ErrKeyNotSet = errors.New("key not set")
)

// Config contains all the settings for configuring the application.
type Config struct {
	PathToSchedule string         `json:"path_to_schedule"`
	SheetName      string         `json:"sheet_name"`
	MaxPairPerDay  int            `json:"max_pair_per_day"`
	Key            string         `json:"key"`
	StorageConfig  storage.Config `json:"storage"`
}

// New initializing the config for the application.
func New() (Config, error) {
	flag.Parse()

	var c = Config{}
	f, err := os.OpenFile("config/config.json", os.O_RDONLY, 0644)
	if err != nil {
		return c, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return c, fmt.Errorf("read json: %w", err)
	}

	return c, nil
}
