package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromEmptyFile(t *testing.T) {
	_, err := loadConfig(path.Join(os.TempDir(), "emptyfile.json"))
	assert.NotNil(t, err, "err should exist")
}

func TestDefaultConfig(t *testing.T) {
	c := defaultConfig()
	assert.Equal(t, c.UpstreamUrl, "https://api.varnamproject.com")
	assert.Equal(t, c.filePath, path.Join(os.Getenv("HOME"), ".varnamd", "config.json"))
	assert.Nil(t, c.SchemesToSync, "schemesToSync should be nil")
}

func TestSaveConfig(t *testing.T) {
	c := defaultConfig()
	c.filePath = path.Join(os.TempDir(), fmt.Sprintf("%d-config.json", rand.Int()))
	err := c.save()
	assert.Nil(t, err, "err should be nil")
	stat, err := os.Stat(c.filePath)
	assert.Nil(t, err, "err should be nil")
	mode := stat.Mode()
	assert.True(t, mode.IsRegular())

	c1, err := loadConfig(c.filePath)
	assert.Nil(t, err, "err should be nil")
	assert.Equal(t, c, c1)
}

func TestSetSyncStatus(t *testing.T) {
	c := defaultConfig()
	c.filePath = path.Join(os.TempDir(), fmt.Sprintf("%d-config.json", rand.Int()))
	c.setSyncStatus("ml", &syncStatus{Offset: 100, Enabled: true, LastPage: false})

	c1, err := loadConfig(c.filePath)
	assert.Nil(t, err, "err should be nil")
	assert.Equal(t, c, c1)
}
