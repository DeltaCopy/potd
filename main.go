package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"potd/libs"
	"strings"
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

	libs.CreateCacheDir(cacheDir)

	cacheFile := fmt.Sprintf("%s/potd_%s.json", cacheDir, strings.ReplaceAll(today, "-", ""))

	data := libs.ReadCacheFileData(cacheFile)

	if data != nil {

		// read cache file
		log.Printf("Image = %s\n", data.Images[0].Title)
		log.Printf("Startdate = %s\n", data.Images[0].Startdate)
		log.Printf("Copyright = %s\n", data.Images[0].Copyright)
		log.Printf("Imagepath = %s\n", data.Imagepath)

		if _, err := os.Stat(data.Imagepath); err == nil {
			busObj := libs.GetDbusObject("org.kde.plasmashell", "/PlasmaShell")

			libs.SetWallpaper(busObj, screenNum, fmt.Sprintf("file://%s", data.Imagepath))
		}
	} else {
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

}
