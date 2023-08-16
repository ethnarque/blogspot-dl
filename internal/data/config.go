package data

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Namespace string `json:"namespace"`
	Completed bool   `json:"completed"`
	Posts     []Post `json:"posts"`
}

type Post struct {
	Title              string   `json:"title"`
	Namespace          string   `json:"namespace"`
	URL                string   `json:"url"`
	IsPendingCompleted bool     `json:"isPendingCompleted"`
	Assets             []string `json:"assets"`
}

func (c *Config) writeConfig() error {
	filename := filepath.Join(c.Namespace, "blog.json")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) readConfig() (*Config, error) {
	filename := filepath.Join(c.Namespace, "blog.json")

	file, err := os.OpenFile(filename, os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	decoder.Decode(&config)

	return &config, nil
}
