package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"potd/libs"
	"time"
)

const (
	bingAPIURL = "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1"
	bingURL    = "https://bing.com"
	screenNum  = 0
)

var (
	today = time.Now().Format(time.DateOnly)
)

func parseArgs() (string, int) {
	path := flag.String("path", "", "Specify path to save image")
	resolution := flag.Int("res", 0, "Specify the resolution [3840, 1366, 1920, 1280, 1080, 1024, 800, 768, 720, 640, 400, 320, 240]")

	flag.Parse()

	return *path, *resolution
}

func main() {
	path, resolution := parseArgs()

	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal("Failed to locate home directory for the user")
		os.Exit(1)
	}

	cacheDir := homeDir + "/.cache/potd"

	connected := libs.WaitForConnection()

	if connected {
		if res := libs.VerifyResolution(resolution); res != "ERROR" && len(path) > 0 {

			client := http.Client{}
			libs.SaveImage(bingAPIURL, bingURL, path, res, screenNum, &client, today, cacheDir)

		} else {
			flag.PrintDefaults()

		}
	} else {
		log.Fatal("Failed to set wallpaper, connection timed out.")
		os.Exit(1)
	}

}
