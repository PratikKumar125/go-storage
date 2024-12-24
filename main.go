package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/PratikKumar125/go-storage/storage"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var templates = template.Must(template.ParseFiles("public/upload.html"))

func display(w http.ResponseWriter, page string, data interface{}) {
	templates.ExecuteTemplate(w, page+".html", data)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create(handler.Filename); if err != nil {
		panic(err)
	}

	defer dst.Close()

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//compression

	go transformLarge(handler.Filename)
	go transformMedium(handler.Filename)
	go transformSmall(handler.Filename)

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		display(w, "upload", nil)
	case "POST":
		uploadFile(w, r)
	}
}

func transformLarge(path string) (error) {
	ffmpeg.Input(path).Output("./regi_large.png",
			ffmpeg.KwArgs{"vf": "scale=3840x2160:flags=lanczos"},
		).
		Run()
	time.Sleep(10 * time.Second)
	return nil
}
func transformMedium(path string) (error) {
	ffmpeg.Input(path).Output("./regi_medium.png",
	ffmpeg.KwArgs{"vf": "scale=1920x1080:flags=lanczos"},
	).
	Run()
	time.Sleep(5 * time.Second)
	return nil
}

func transformSmall(path string) (error) {
	ffmpeg.Input(path).Output("./regi_small.png",
			ffmpeg.KwArgs{"vf": "scale=1280x720:flags=lanczos"},
		).
		Run()
	time.Sleep(3 * time.Second)
	return nil
}

func main() {
	originals := storage.DiskStruct{
		Bucket: "testing-media-services",
		Region: "us-east-1",
		Profile: "ankush_prasoon_account",
	}
	medium := storage.DiskStruct{
		Bucket: "pdf",
		Region: "eu-north-1",
		Profile: "chetaaaaaan",
	}

	disks := map[string]storage.DiskStruct{
        "originals": originals,
        "medium":   medium,
    }

	// Initializing the available disks and the current disks
	st := storage.InitStorage("originals", disks)

	items, err := st.GetBucketItems()
	fmt.Println(items,err, "<<<<<<BUCKET ITEMS")

	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":8080", nil)
	
	// url, err := st.SignedURL("/originals/testing3.mp3")
	// fmt.Println(url, err)

	// st.SetCurrentDisk("small")
	// st.SetCurrentDisk("medium")
}