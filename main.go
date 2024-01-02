package main

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"mime"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func tmpsend(url string) {
	pattern := regexp.MustCompile(`https://tmpsend.com/(\w+)$`)

	if matches := pattern.FindStringSubmatch(url); len(matches) > 1 {
		fileID := matches[1]
		downloadLink := fmt.Sprintf("https://tmpsend.com/download?d=%s", fileID)

		request, err := http.NewRequest("GET", downloadLink, nil)
		if err != nil {
			fmt.Println("ERROR: Unable to create HTTP request:", err)
			os.Exit(1)
		}
		request.Header.Set("Referer", fmt.Sprintf("https://tmpsend.com/thank-you?d=%s", fileID))

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			fmt.Println("ERROR: Unable to download the file:", err)
			os.Exit(1)
		}
		defer response.Body.Close()

		fileSize, _ := strconv.Atoi(response.Header.Get("Content-Length"))
		bar := progressbar.DefaultBytes(int64(fileSize), "Downloading")

		filename := getFilenameFromHeader(response.Header)
		if filename == "" {
			filename = fmt.Sprintf("downloaded_file_%s", fileID)
		}

		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("ERROR: Unable to create file:", err)
			os.Exit(1)
		}
		defer file.Close()

		writer := io.MultiWriter(file, bar)

		_, err = io.Copy(writer, response.Body)
		if err != nil {
			fmt.Println("ERROR: Unable to download the file:", err)
			os.Exit(1)
		}

		fmt.Printf("\nDownload complete. File saved as %s\n", filename)
	} else {
		fmt.Println("ERROR: Invalid URL format")
		os.Exit(1)
	}
}

func getFilenameFromHeader(header http.Header) string {
	contentDisposition := header.Get("Content-Disposition")
	if contentDisposition == "" {
		return ""
	}

	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return ""
	}

	filename, ok := params["filename"]
	if !ok {
		return ""
	}

	return filename
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <URL>\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	tmpsend(url)
}
