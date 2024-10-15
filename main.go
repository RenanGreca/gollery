package main

import (
	"context"
	"gollery/model"
	"gollery/templates"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("Please specify a PORT environment variable")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", listHandler)
	mux.HandleFunc("GET /book/{title}", bookHandler)

	assetsPath := os.Getenv("MEDIADIR")
	if assetsPath == "" {
		log.Fatalln("Please specify a MEDIADIR environmetn variable")
	}
	// prefix := assetsPath + "/"
	// assets := http.FileServer(http.Dir(prefix))
	media := http.FileServer(http.Dir(assetsPath))
	// mux.Handle("/media", assets)
	mux.Handle("GET /media/", http.StripPrefix("/media", media))
	// mux.Handle("/", http.FileServer(http.Dir(assetsPath)))

	// scriptsPath := "/assets/"
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("GET /assets/", http.StripPrefix("/assets", fs))

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalln("Error starting server: ", err)
	}
}

//	func mainHandler(w http.ResponseWriter, req *http.Request) {
//		log.Println("Main page")
//	}
func bookHandler(w http.ResponseWriter, req *http.Request) {
	assetsPath := os.Getenv("MEDIADIR")
	if assetsPath == "" {
		log.Fatalln("Please specify a MEDIADIR environmetn variable")
	}

	title := req.PathValue("title")
	bookPath := assetsPath + "/" + title

	images := getImages(bookPath, title)
	body := templates.Index(images)
	template := templates.Body(body)
	template.Render(context.Background(), w)
}

func listHandler(w http.ResponseWriter, req *http.Request) {
	assetsPath := os.Getenv("MEDIADIR")
	if assetsPath == "" {
		log.Fatalln("Please specify a MEDIADIR environmetn variable")
	}

	entries, err := os.ReadDir(assetsPath)
	if err != nil {
		log.Fatalln("Error reading contents of directory: ", assetsPath, err)
	}

	var dirs []model.Directory
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// log.Printf("Found file: %s\n", e.Name())
		path := assetsPath + "/" + e.Name()
		mediapath := "/media/" + e.Name()

		contents, err := os.ReadDir(path)
		if err != nil {
			log.Fatalln("Error reading contents of directory: ", path, err)
		}
		var image string
		for _, c := range contents {
			if c.IsDir() {
				continue
			}
			p := path + "/" + c.Name()
			b, err := os.ReadFile(p)
			if err != nil {
				log.Fatalln("Error reading file: ", err)
			}
			contentType := http.DetectContentType(b)
			if strings.Contains(contentType, "image") {
				image = mediapath + "/" + c.Name()
				// log.Println("Found image: ", image)
				break
			}
		}

		if image == "" {
			continue
		}

		dir := model.Directory{
			Name:  e.Name(),
			Path:  path,
			Image: image,
		}
		dirs = append(dirs, dir)
	}

	body := templates.List(dirs)
	template := templates.Body(body)
	template.Render(context.Background(), w)
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
		p := path + "/" + c.Name()
		image := "/media/" + title + "/" + c.Name()
		b, err := os.ReadFile(p)
		if err != nil {
			log.Fatalln("Error reading file: ", err)
		}
		contentType := http.DetectContentType(b)
		if strings.Contains(contentType, "image") {
			images = append(images, image)
		}
	}

	return images
}
