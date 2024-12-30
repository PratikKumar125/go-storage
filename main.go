package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/PratikKumar125/go-storage/storage"
)

func uploadFileToS3(signedURL, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    fileInfo, err := file.Stat()
    if err != nil {
        return fmt.Errorf("failed to stat file: %w", err)
    }
    contentLength := fileInfo.Size()

    req, err := http.NewRequest(http.MethodPut, signedURL, file)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "image/jpeg")
    req.ContentLength = contentLength

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to upload file: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to upload file, status code: %d", resp.StatusCode)
    }

    return nil
}


func main() {
	originals := storage.DiskStruct{
		Bucket: "prateek-dev",
		Region: "ap-south-1",
		Profile: "chetan",
	}
	medium := storage.DiskStruct{
		Bucket: "pdf",
		Region: "eu-north-1",
		Profile: "prateek",
	}

	disks := map[string]storage.DiskStruct{
        "originals": originals,
        "medium":   medium,
    }

	// initializing the available disks and the current disks
	st := storage.InitStorage("originals", disks)

	items, err := st.GetBucketItems()
	fmt.Println(items,err, "<<<<<<BUCKET ITEMS")
	
	resMap, err := st.SignedURL("/prateek-dev/PratikKumar125.jpeg", 10)
	if err != nil {
		panic(err)
	}
	fmt.Println(resMap["signedUrl"], err)

	// Open the local file
	filePath := "/Users/pratikkumar/desktop/golang/go-storage/PratikKumar125.jpeg"
	uploadFileToS3(resMap["signedUrl"] ,filePath)

	st.SetCurrentDisk("medium")
}