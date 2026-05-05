package libs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"potd/models"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	setWallpaperMethod = "org.kde.PlasmaShell.setWallpaper"
	wallpaperMethod    = "org.kde.PlasmaShell.wallpaper"
	wallpaperPlugin    = "org.kde.image"
	screenNum          = uint(0)
)

func VerifyResolution(res int) string {
	switch res {
	case 1366:
		return "1366x768"
	case 1920:
		return "1920x1080"
	case 3840:
		return "3840x2160"
	case 1280:
		return "1280x768"
	case 1024:
		return "1024x768"
	case 800:
		return "800x600"
	case 1080:
		return "1080x1920"
	case 768:
		return "768x1280"
	case 720:
		return "720x1280"
	case 640:
		return "640x480"
	case 480:
		return "480x800"
	case 400:
		return "400x240"
	case 320:
		return "320x240"
	case 240:
		return "240x320"
	default:
		return "ERROR"
	}
}

func createImageDirectory(path string) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create image directory with error %s", err)
			os.Exit(1)
		}
	}
}

func createCacheDir(cacheDir string) {
	_, err := os.Stat(cacheDir)

	if os.IsNotExist(err) {
		err := os.MkdirAll(cacheDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create cache directory with error %s", err)
			os.Exit(1)
		}
	}
}

func readCacheFileData(cacheFile string) *models.Bing {
	_, err := os.Stat(cacheFile)

	if os.IsNotExist(err) {
		return nil
	} else {
		file, err := os.ReadFile(cacheFile)
		if err != nil {
			log.Fatalf("Failed to read cache file with error %s", err)
			return nil
		}
		log.Println("Cache file exists")
		var data models.Bing
		err = json.Unmarshal(file, &data)

		if err != nil {
			log.Fatalf("Failed to process cache file with error %s", err)
			return nil
		}

		return &data
	}
}

func getResponseBody(client *http.Client, url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Request failed with error %s", err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get response with error %s", err)
		os.Exit(1)
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Failed to get read response: %s", err)
			os.Exit(1)
		}

		return body

	} else {
		log.Fatalf("Request failed with status %s", res.Status)
		os.Exit(1)
	}

	return nil
}

func SaveImage(apiURL string, bingURL, path string, resolution string, screenNum uint, client *http.Client, today string, cacheDir string) {

	cacheFile := fmt.Sprintf("%s/potd_%s.json", cacheDir, strings.ReplaceAll(today, "-", ""))
	busObj := GetDbusObject("org.kde.plasmashell", "/PlasmaShell")

	createCacheDir(cacheDir)

	if data := readCacheFileData(cacheFile); data == nil {

		createImageDirectory(path)
		responseBody := getResponseBody(client, apiURL)

		var jsonResponse models.Bing

		err := json.Unmarshal(responseBody, &jsonResponse)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		if len(jsonResponse.Images) > 0 {
			// only 1 result
			if len(jsonResponse.Images[0].Urlbase) > 0 && len(jsonResponse.Images[0].Title) > 0 {
				url := fmt.Sprintf("%s%s_%s.jpg", bingURL, jsonResponse.Images[0].Urlbase, resolution)
				if len(url) > 0 {
					invalidChars := []rune{':', '?', '!', '\'', ' ', '`', '¬', '@', '#', ',', '\'', '\u0027', ';', ' '}

					log.Printf("Title = %s\n", jsonResponse.Images[0].Title)
					log.Printf("Copyright = %s\n", jsonResponse.Images[0].Copyright)

					title := strings.ToLower(jsonResponse.Images[0].Title)

					for _, c := range invalidChars {
						if strings.ContainsRune(title, c) {
							title = strings.ReplaceAll(title, string(c), "_")
						}
					}

					jpegFile := path + "/" + strings.ReplaceAll(today, "-", "") + "_" + strings.ReplaceAll(title, "__", "_") + "_" + resolution + ".jpg"

					jsonResponse.Imagepath = jpegFile

					log.Printf("Image = %s\n", jpegFile)

					_, err := os.Stat(jpegFile)
					if os.IsNotExist(err) {
						//create file
						file, err := os.Create(jpegFile)

						if err != nil {
							panic(err)
						}

						response := getResponseBody(client, url)

						_, err = file.Write(response)

						if err != nil {
							log.Fatalf("Error writing file: %s", err)
							os.Exit(1)
						}

						defer file.Close()

						writeCache(cacheFile, &jsonResponse)

					} else {
						log.Printf("%s already exists\n", jpegFile)
						writeCache(cacheFile, &jsonResponse)
					}

					setWallpaper(busObj, screenNum, fmt.Sprintf("file://%s", jpegFile))

				}
			}
		} else {
			log.Fatalf("The response returned has no images\n")
			log.Fatal(jsonResponse)
			os.Exit(1)

		}
	} else {
		// read cache file
		log.Printf("Image = %s\n", data.Images[0].Title)
		log.Printf("Startdate = %s\n", data.Images[0].Startdate)
		log.Printf("Copyright = %s\n", data.Images[0].Copyright)
		log.Printf("Imagepath = %s\n", data.Imagepath)

		if _, err := os.Stat(data.Imagepath); err == nil {
			setWallpaper(busObj, screenNum, fmt.Sprintf("file://%s", data.Imagepath))
		}

	}
}

func writeCache(cacheFile string, jsonResponse *models.Bing) {
	_, errF := os.Stat(cacheFile)

	if os.IsNotExist(errF) {
		file, err := os.Create(cacheFile)

		if err != nil {
			log.Fatalf("Failed to create cache file with error %s", err)
			os.Exit(1)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if err := encoder.Encode(jsonResponse); err != nil {
			log.Fatal(err)
		}

	}
}

func GetDbusObject(dbusInterface string, dbusMethod string) dbus.BusObject {
	conn := getSessionBus()
	return conn.Object(dbusInterface, dbus.ObjectPath(dbusMethod))
}

func GetCurrentWallpaper(busObj dbus.BusObject, screenNum uint) string {
	dbusCall := busObj.Call(wallpaperMethod, 0, screenNum)

	if dbusCall.Err != nil {
		log.Fatal(dbusCall.Err.Error())
		os.Exit(1)
	}

	wallpaper := dbusCall.Body[screenNum].(map[string]dbus.Variant)

	if wallpaper["Image"].Value() != nil {
		return wallpaper["Image"].Value().(string)
	}

	return ""

}

func setWallpaper(busObj dbus.BusObject, screenNum uint, file string) {

	log.Println("Setting wallpaper")
	parameters := map[string]dbus.Variant{}

	parameters["Image"] = dbus.MakeVariant(file)
	dbusCall := busObj.Call(setWallpaperMethod, 0, wallpaperPlugin, parameters, screenNum)

	if dbusCall.Err != nil {
		log.Fatal(dbusCall.Err.Error())
		os.Exit(1)
	}
}

func getSessionBus() *dbus.Conn {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	return conn
}

func WaitForConnection() bool {
	timeout := 100
	count := 0
	for {
		count += 1

		_, err := net.LookupIP("one.one.one.one")
		if err != nil {
			time.Sleep(time.Millisecond * 250)
		} else if count == timeout {
			return false
		} else {
			return true
		}

	}
}
