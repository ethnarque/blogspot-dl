package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pmlogist/blogspot-dl/internal/data"
)

type config struct {
	baseURL string
	outDir  string
}

type application struct {
	config  config
	name    string
	workDir string
	Blog    *data.Blog
}

func main() {
	var config config

	flag.StringVar(&config.baseURL, "url", os.Getenv("URL"), "The blogspot URL")
	flag.StringVar(&config.outDir, "outdir", "public", "The output directory for downloading files")

	flag.Parse()

	if config.baseURL == "" {
		fmt.Println("Please, provide an url")
		return
	}
	app := &application{
		config: config,
	}

	err := app.newBlogspot()
	if err != nil {
		log.Fatalln(err)
	}

}
