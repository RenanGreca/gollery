package main

import (
	"context"
	"gollery/model"
	"gollery/templates"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/joho/godotenv"
)

var imageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("Please specify a PORT environment variable")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", listHandler)
	mux.HandleFunc("GET /book/{title}", bookHandler)
	mux.HandleFunc("GET /book/{series}/{title}", bookHandler)

	assetsPath := os.Getenv("MEDIADIR")
	if assetsPath == "" {
		log.Fatalln("Please specify a MEDIADIR environmetn variable")
	}
	media := http.FileServer(http.Dir(assetsPath))
	mux.Handle("GET /media/", http.StripPrefix("/media", media))

	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("GET /assets/", http.StripPrefix("/assets", fs))

	log.Println("Server starting on port: ", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalln("Error starting server: ", err)
	}
	log.Println("Server started on port: ", port)
}

func bookHandler(w http.ResponseWriter, req *http.Request) {
	assetsPath := os.Getenv("MEDIADIR")
	if assetsPath == "" {
		log.Fatalln("Please specify a MEDIADIR environmetn variable")
	}

	series := req.PathValue("series")
	title := req.PathValue("title")
	if series != "" {
		title = series + "/" + title
	}
	bookPath := assetsPath + "/" + title
	bookPath, err := url.QueryUnescape(bookPath)
	if err != nil {
		log.Fatalln("Error unescaping book path: ", err)
	}

	images := getImages(bookPath, title)
	body := templates.Index(images)
	template := templates.Body(body)
	template.Render(context.Background(), w)
}

func listHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("Listing books")
	assetsPath := os.Getenv("MEDIADIR")
	if assetsPath == "" {
		log.Fatalln("Please specify a MEDIADIR environmetn variable")
	}

	dirs := listDirs(assetsPath, "/media", "")

	body := templates.List(dirs)
	template := templates.Body(body)
	template.Render(context.Background(), w)
}

func listDirs(assetsPath string, mediaPath string, relativePath string) []model.Directory {
	entries, err := os.ReadDir(assetsPath)
	if err != nil {
		log.Fatalln("Error reading contents of directory: ", assetsPath, err)
	}

	var dirs []model.Directory
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := assetsPath + "/" + e.Name()
		mediapath := mediaPath + "/" + url.PathEscape(e.Name())
		relativepath := relativePath + "/" + e.Name()
		log.Printf("Scanning directory: %s\n", path)
		log.Printf("Media directory: %s\n", mediapath)

		contents, err := os.ReadDir(path)
		if err != nil {
			log.Fatalln("Error reading contents of directory: ", path, err)
		}
		var image string
		for _, c := range contents {
			if c.IsDir() {
				log.Printf("Found directory: %s\n", c.Name())
				dirs2 := listDirs(path, mediapath, relativepath)
				dirs = append(dirs, dirs2...)
				break
			}
			if isImage(c.Name()) {
				log.Println("Found image: ", image)
				name := url.PathEscape(c.Name())
				image = mediapath + "/" + name
				break
			}
		}

		if image == "" {
			continue
		}

		dir := model.Directory{
			Name:  url.PathEscape(relativepath),
			Path:  url.PathEscape(path),
			Image: image,
		}
		log.Printf("Adding directory: %q\n", dir)
		dirs = append(dirs, dir)
	}

	return dirs
}

func getImages(path, title string) []string {
	contents, err := os.ReadDir(path)
	if err != nil {
		log.Fatalln("Error reading contents of directory: ", path, err)
	}

	var images []string
	for _, c := range contents {
		if c.IsDir() {
			continue
		}
		image := "/media" + title + "/" + c.Name()
		if isImage(c.Name()) {
			images = append(images, image)
			log.Println("Found image: ", image)
		}
	}

	return images
}

func isImage(file string) bool {
	extension := filepath.Ext(file)
	return slices.Contains(imageExtensions, extension) && !strings.HasPrefix(file, ".")
}
