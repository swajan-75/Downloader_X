package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func shouldSkip(link string) bool {
	skipExtensions := map[string]bool{
		".html": true,
		".css":  true,
		".js":   true,
		".ts":   true,
		".php":  true,
	}

	ext := strings.ToLower(filepath.Ext(link))
	return skipExtensions[ext]
}
func downloadFile(link, dir string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}
	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" {
		fileName = "index.html"
	}

	resp2, err := http.Head(link)
	if err != nil {
		return false
	}
	defer resp2.Body.Close()
	contentLength := resp2.Header.Get("Content-Length")
	if contentLength == "" {
		contentLength = ""
	}
	var file_size float64 = 0.0
	if contentLength != "" {
		file_size_tmp, _ := strconv.ParseInt(contentLength, 10, 64)

		file_size = float64(file_size_tmp) / (1024 * 1024)

	}

	resp, err := http.Get(link)
	if err != nil {

		return false
	}

	defer resp.Body.Close()

	filePath := filepath.Join(dir, fileName)

	file, err := os.Create(filePath)
	if err != nil {

		return false
	}
	defer file.Close()

	buf := make([]byte, 1024) // 1 KB buffer
	totalBytes := 0
	startTime := time.Now()

	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return false
		}
		if n == 0 {
			break // End of file
		}
		totalBytes += n

		// Write the chunk to the file
		if _, err := file.Write(buf[:n]); err != nil {
			return false
		}

		elapsed := time.Since(startTime).Seconds()
		speed := float64(totalBytes) / elapsed

		fmt.Printf("\rDownloading %s ------- %0.3f MB --------- %0.3f mbps size : [%0.3f]", fileName, (float64(totalBytes) / (1024 * 1024)), float64(speed/(1024*1024)), file_size)
	}

	fmt.Printf("\r%*s\r", 3-00, "")
	fmt.Printf("Downloaded: %s\n", fileName)

	return true
}
func main() {

	urlFlag := flag.String("url", "", "URL to fetch and extract links from")
	preFlag := flag.String("pre", "", "Prefix to add before each link")
	sufFlag := flag.String("suf", "", "Suffix to add after each link")
	dirFlag := flag.String("d", "", "Dir")
	flag.Parse()
	if *urlFlag == "" {
		log.Fatal("URL must be provided using -url flag")
	}
	res, err := http.Get(*urlFlag)
	if err != nil {
		fmt.Println((err))
	}
	defer res.Body.Close()

	token := html.NewTokenizer(res.Body)
	for {
		token_type := token.Next()
		switch token_type {
		case html.ErrorToken:
			return
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := token.Token()
			if tok.Data == "a" {
				for _, attr := range tok.Attr {
					if attr.Key == "href" {
						link := *preFlag + attr.Val + *sufFlag
						if !shouldSkip(link) {

							if downloadFile(link, *dirFlag) == false {
								fmt.Printf("Failed to download %s \n", link)
							}

						}
					}
				}

			}
		}

	}

}
