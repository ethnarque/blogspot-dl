package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func (app *application) parseConfig() error {
	var dirName string

	fmt.Println(app.config.outDir)

	baseURL, err := url.Parse(app.config.baseURL)
	if err != nil {
		return err
	}

	dirName = strings.Join(strings.Split(baseURL.Hostname(), "www."), "")

	outDir := filepath.Join(app.config.outDir, dirName)
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return err
	}

	app.workDir = outDir
	app.name = dirName

	return nil
}
