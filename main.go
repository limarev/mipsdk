package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type bytes int

func (b bytes) String() string {
	bf := float64(b)
	for _, unit := range []string{"", "K", "M", "G", "T", "P", "E", "Z"} {
		if math.Abs(bf) < 1000.0 {
			return fmt.Sprintf("%3.1f %sB", bf, unit)
		}
		bf /= 1000.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

type FileDescription struct {
	IsPrimary     string `json:"isPrimary"`
	Name          string `json:"name"`
	Url           string `json:"url"`
	Size          string `json:"size"`
	Version       string `json:"version"`
	DatePublished string `json:"datePublished"`
}

type File struct {
	Buffer []byte
	Name   string
	Size   bytes
}

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func (fd *FileDescription) download(results chan File, errors chan error, timeout time.Duration) {
	resch := make(chan File, 1)
	errch := make(chan error, 1)
	go func() {
		resp, err := http.Get(fd.Url)
		if err != nil {
			errch <- err
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			errch <- err
		}

		size, err := strconv.Atoi(fd.Size)
		if err != nil {
			errch <- err
		}
		f := File{Buffer: b, Name: fd.Name, Size: bytes(size)}
		resch <- f
	}()

	select {
	case <-time.After(timeout):
		errors <- fmt.Errorf("%s timed out after %v", fd.Url, timeout)
	case err := <-errch:
		errors <- err
	case res := <-resch:
		results <- res
	}
}

func (app *application) Scrape(url string) ([]FileDescription, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	status := fmt.Sprintf("%s %s %s", resp.Request.Method, resp.Request.URL, resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", status)
	}
	app.infoLog.Print(status)

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := string(b)
	targetString := string("\"downloadFile\":")
	a := strings.Index(s, targetString)
	if a == -1 {
		return nil, fmt.Errorf("failed to find %s field in response body", targetString)
	}

	app.infoLog.Printf("%s was found in responce body", targetString)

	s1 := string(b[a:])
	targetRune := ']'
	a1 := strings.IndexRune(s1, targetRune)
	if a == -1 {
		return nil, fmt.Errorf("%s is invalid json array. Failed to find %c after %s", targetString, targetRune, targetString)
	}

	app.infoLog.Printf("%s is valid json array", targetString)

	files := []FileDescription{}
	err = json.Unmarshal(b[a+len(targetString):a+a1+1], &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (app *application) dumpFiles(results chan File, errors chan error, dir string, files []FileDescription) {
	i := 0
	for range files {
		select {
		case file := <-results:
			filePath := dir + string(os.PathSeparator) + file.Name
			f, err := os.Create(filePath)
			if err != nil {
				app.errorLog.Print(err)
			}
			defer f.Close()

			n, err := f.Write(file.Buffer)
			if err != nil || n != int(file.Size) {
				app.errorLog.Print(err)
			}

			i++

			app.infoLog.Printf("%s is downloaded successfully: %s: %v out %v items", file.Name, file.Size, i, len(files))
		case err := <-errors:
			app.errorLog.Print(err)
		}
	}
}

func dumpVersion(downloadDir string, version string) (string, error) {
	filePath := downloadDir + string(os.PathSeparator) + "version.txt"
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.Write([]byte(version))
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func main() {
	dir := flag.String("dir", ".", "download dir")
	url := flag.String("url", "https://aka.ms/mipsdkbins", "url for scraping")
	timeout := flag.Int("timeout", 600, "downloading timeout per file in seconds")
	isVersionOnlyMode := flag.Bool("version-only", false, "no downloading actually happens, returns mipsdk binaries version if found")
	flag.Parse()

	app := &application{
		infoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}

	downloadDir, err := filepath.Abs(*dir)
	if err != nil {
		app.errorLog.Fatal(err)
	}

	err = os.MkdirAll(downloadDir, 0755)
	if err != nil {
		app.errorLog.Fatal(err)
	}

	app.infoLog.Printf("version only mode: %v", *isVersionOnlyMode)
	app.infoLog.Printf("download dir: %v", downloadDir)
	app.infoLog.Printf("url for scraping: %s", *url)
	if !*isVersionOnlyMode {
		app.infoLog.Printf("downloading timeout per file in seconds: %v", *timeout)
	}

	files, err := app.Scrape(*url)
	if err != nil {
		app.errorLog.Fatal(err)
	}

	version := "unknown"
	if len(files) > 0 {
		version = files[0].Version
	}
	app.infoLog.Printf("Data scraping is successful. MIP SDK binaries version %s. Found %v items", version, len(files))

	for i, file := range files {
		app.infoLog.Printf("%s (%v out of %v items)", file.Url, i+1, len(files))
	}

	if *isVersionOnlyMode {
		versionPath, err := dumpVersion(downloadDir, version)
		if err != nil {
			app.errorLog.Fatal(err)
		}

		app.infoLog.Printf("%s is created successfully", versionPath)
	} else {
		results := make(chan File, len(files))
		errors := make(chan error, len(files))

		for _, file := range files {
			go file.download(results, errors, time.Duration(*timeout)*time.Second)
		}

		app.dumpFiles(results, errors, downloadDir, files)
	}
}
