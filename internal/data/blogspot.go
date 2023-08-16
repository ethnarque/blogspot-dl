package data

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

const (
	BlogspotImageURLRX = `^(?:https?://)?(?:[a-zA-Z0-9.-]*\.?blogger\.googleusercontent\.com/|[0-9a-zA-Z.-]+\.bp\.blogspot\.com/|/wp-content/)`
	BlogspotPostURLRX  = `^(?:https?://)?[^/]+/\d{4}/\d{2}`
	PostListFilename   = "posts_list.json"
	AssetListFilename  = "assets_list.txt"
	UserAgent          = ""
)

type Blog struct {
	name      string
	namespace string
	url       string
	config    *Config
}

type BlogspotAsset struct {
	URL      string
	Filename string
}

type BlogspotImage struct {
	URL       *url.URL
	Namespace string
	Filename  string
}

func NewBlogspot(baseURL, name, workDir string) *Blog {
	return &Blog{
		name:      name,
		namespace: workDir,
		url:       baseURL,
		config: &Config{
			Namespace: workDir,
		},
	}
}

func (blog *Blog) MakePostsLists() error {
	var posts []Post

	seen := make(map[string]bool)

	if _, err := os.Stat(filepath.Join(blog.config.Namespace, "blog.json")); err != nil {
		blog.config.writeConfig()
	}

	config, err := blog.config.readConfig()
	if err != nil {
		return err
	}

	if len(config.Posts) > 0 {
		fmt.Println("using posts from config file ...")
		for _, post := range config.Posts {
			posts = append(posts, post)
			seen[post.URL] = true
		}
	}

	c := colly.NewCollector()
	pattern := regexp.MustCompile(BlogspotPostURLRX)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		hasPrefix := strings.HasPrefix(href, blog.url)
		containsComment := strings.Contains(href, "showComment")

		if pattern.MatchString(href) && hasPrefix && !containsComment {
			parsedHref, err := url.Parse(href)
			if err != nil {
				fmt.Printf("error parsing %s: %s", href, err)
			}
			parsedHref.Fragment = ""

			link := parsedHref.String()
			namespace := strings.Join(strings.Split(parsedHref.Path, ".html"), "")

			if pattern.MatchString(link) && !seen[link] {
				seen[link] = true
				posts = append(posts, Post{
					Title:              "",
					Namespace:          namespace,
					URL:                link,
					IsPendingCompleted: false,
					Assets:             []string{},
				})
				fmt.Println("adding post:", parsedHref.String())
			}

			blog.config.Posts = posts
			err = blog.config.writeConfig()
			if err != nil {
				fmt.Print("error writing", link)
			}

			c.Visit(parsedHref.String())
		}
	})

	if !config.Completed {
		fmt.Printf("fetching new posts from %s ...\n", blog.url)
		err := c.Visit(blog.url)
		if err != nil {
			return err
		}
		blog.config.Posts = posts
		blog.config.Completed = true
		fmt.Println("appending new posts to config file...")
		err = blog.config.writeConfig()
		if err != nil {
			return err
		}
		fmt.Println("finished...")
	}

	return nil
}

func (blog *Blog) MakeImagesList() error {
	config, err := blog.config.readConfig()
	if err != nil {
		return err
	}

	pattern := regexp.MustCompile(BlogspotImageURLRX)
	c := colly.NewCollector()

	for i, post := range config.Posts {

		if post.IsPendingCompleted {
			fmt.Println("skipping:", post.URL)
			continue
		}

		seen := make(map[string]bool)

		if len(post.Assets) > 0 {
			for _, v := range post.Assets {
				seen[v] = true
			}
		}

		var assets []string

		c.OnHTML("img[src]", func(e *colly.HTMLElement) {
			href, exists := e.DOM.Parent().Attr("href")

			if href != "" && strings.HasPrefix(href, "//") {
				href = "http:" + href
			}

			if exists && href != "" && !seen[href] && pattern.MatchString(href) {
				seen[href] = true
				assets = append(assets, href)
			}

			src := e.Attr("src")
			if src != "" && strings.HasPrefix(src, "//") {
				src = "http:" + src
			}

			if !exists && src != "" && !seen[src] && pattern.MatchString(src) {
				seen[src] = true
				assets = append(assets, src)
			}
		})

		err := c.Visit(post.URL)
		if err != nil {
			fmt.Print(err)
			continue
		}

		var filteredAssets []string

		for _, asset := range assets {
			u, err := url.Parse(asset)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				continue
			}

			pathSegments := strings.Split(u.Path, "/")

			for _, segment := range pathSegments {
				if strings.HasPrefix(segment, "s") {
					numberPart := segment[1:]
					number, err := strconv.Atoi(numberPart)
					if err == nil && (number >= 640) {
						fmt.Println(asset)
						filteredAssets = append(filteredAssets, asset)
					}
					break
				}
			}
		}

		config.Posts[i].Assets = append(config.Posts[i].Assets, filteredAssets...)
		config.Posts[i].IsPendingCompleted = true
		config.writeConfig()
		fmt.Println("remaining", len(config.Posts)-i, post.URL)
	}

	return nil
}

func (blog *Blog) DownloadPostAssets() error {
	config, err := blog.config.readConfig()
	if err != nil {
		return err
	}

	client := http.Client{}

	for _, post := range config.Posts {
		for i, link := range post.Assets {
			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				return err
			}
			req.Header.Set("User-Agent", UserAgent)

			dirName := filepath.Join(blog.namespace, post.Namespace)
			if _, err := os.Stat(dirName); err != nil {
				err := os.MkdirAll(dirName, 0755)
				if err != nil {
					return err
				}
			}

			extension := filepath.Ext(link)[1:]

			fileName := filepath.Join(dirName, strconv.Itoa(i)+"."+extension)

			if _, err := os.Stat(fileName); err != nil {
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				file, err := os.Create(fileName)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(file, resp.Body)
				if err != nil {
					fmt.Println("Error writing to file:", err)
				}

				fmt.Println("downloaded:", fileName)
			}
		}
	}

	fmt.Println("finished!")

	return nil

}

func saveURLsToFile(urls []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, url := range urls {
		_, err := file.WriteString(url + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}
