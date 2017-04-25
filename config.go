package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sync"
)

// varnamd configurations
type config struct {
	filePath      string
	UpstreamUrl   string                 `json:"upstreamUrl"`
	SchemesToSync map[string]*syncStatus `json:"schemesToSync"`
	mutex         sync.Mutex
}

// Represents sync status for a language
type syncStatus struct {
	Enabled  bool `json:"enabled"`
	Offset   int  `json:"offset"`
	LastPage bool `json:"lastPage"`
}

func defaultConfig() *config {
	return &config{UpstreamUrl: "https://api.varnamproject.com", filePath: path.Join(getConfigDir(), "config.json")}
}

func getConfigDir() string {
	if runtime.GOOS == "windows" {
		return path.Join(os.Getenv("localappdata"), ".varnamd")
	} else {
		return path.Join(os.Getenv("HOME"), ".varnamd")
	}
}

func loadConfig(file string) (*config, error) {
	blob, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var c config
	err = json.Unmarshal(blob, &c)
	if err != nil {
		return nil, err
	}

	c.filePath = file
	return &c, nil
}

func (c *config) save() error {
	blob, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.filePath, blob, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *config) setSyncStatus(langCode string, status *syncStatus) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.SchemesToSync == nil {
		c.SchemesToSync = make(map[string]*syncStatus)
	}

	c.SchemesToSync[langCode] = status
	return c.save()
}
