package main

import (
	"github.com/pmlogist/blogspot-dl/internal/data"
)

const (
	BlogspotImageURLRX = `^(?:https?://)?(?:[a-zA-Z0-9.-]*\.?blogger\.googleusercontent\.com/|[0-9a-zA-Z.-]+\.bp\.blogspot\.com/|/wp-content/)`
	BlogspotPostURLRX  = `^(?:https?://)?[^/]+/\d{4}/\d{2}`
)

func (app *application) newBlogspot() error {
	err := app.parseConfig()
	if err != nil {
		return err
	}

	app.Blog = data.NewBlogspot(
		app.config.baseURL,
		app.name,
		app.workDir,
	)

	err = app.fetchBlogspotPosts()
	if err != nil {
		return err
	}

	return nil
}

func (app *application) fetchBlogspotPosts() error {
	err := app.Blog.MakePostsLists()
	if err != nil {
		return err
	}

	err = app.Blog.MakeImagesList()
	if err != nil {
		return err
	}

	err = app.Blog.DownloadPostAssets()
	if err != nil {
		return err
	}

	return nil
}
