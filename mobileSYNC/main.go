package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
)

// Define amount of horizontal and vertical screens
// NOTE: Make sure the image/video ratio matches the file ratio
var horizontal_ScreenCount = 3
var vertical_ScreenCount = 3

//	Defines what the devices should display
// 'image' for image
// 'video' for video
var media_type = "video"

// Amount of seconds videos should resync
var reSync = 30

type ID struct {
	HorScreens int
	VerScreens int
	Nid        int
	ReSync	   int64
}

type indx struct {
	Value int
}

// Stores html templates
var templates *template.Template

func main() {

	// Display settings
	fmt.Println("")
	color.HiWhite("Program running")
	if media_type == "image" {
		fmt.Println("Showing image")
	} else {
		fmt.Println("Showing video")
	}

	// Load templates
	load_templates()

	// Start serving
	ServeHTTP()
}

// Serving HTML
func ServeHTTP() {
	// Run sync when showing a video
	if media_type != "image" {
		go recalc()
	}

	// HTML Pages
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", index)
	r.HandleFunc("/show_media", show_media)

	// Static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
	http.Handle("/", r)

	// Start server
	color.HiGreen("ServeHTTP running -> :80")
	//err := http.ListenAndServe(":80", handlers.LoggingHandler(os.Stdout, r)) // DEBUG version
	err := http.ListenAndServe(":80", nil) // RELEASE version
	checkError(err)
}


func recalc() {
	// Easily prevent main() race condition
	time.Sleep(3 * time.Second)

	// Get current second
	now := time.Now()
	sec := now.Unix()

	// Calculate waiting time
	t := sec % int64(reSync)

	// Print waiting time
	color.HiYellow(fmt.Sprintf("Starting in %ds", t))
	time.Sleep( time.Duration(t) * time.Second)

	// Loop Sync timer
	for {
		// Calculate waiting time
		now := time.Now()
		sec := now.Unix()
		waitSec := int64(reSync) - (sec % int64(reSync))

		for i := waitSec; i > 0; i-- {
			if i > 3 && i%10 == 0 {
				color.HiMagenta(fmt.Sprintf("%s%d%s", "Syncing in: ", i, "s"))
			} else if i <= 3 {
				color.HiWhite(fmt.Sprintf("%s%d%s", "Syncing in: ", i, "s"))
			}
			time.Sleep(1 * time.Second)
		}
		color.HiYellow("Syncing devices")
	}
}


func index(w http.ResponseWriter, r *http.Request) {

	// Building page in parts
	row := templates.Lookup("index_header.html")
	row.ExecuteTemplate(w, "index_header", nil)
	row.Execute(w, nil)

	var button_id = 0
	for x := 0; x < horizontal_ScreenCount; x++ {
		for y := 0; y < vertical_ScreenCount; y++ {
			pageData := indx{
				Value: button_id,
			}
			row = templates.Lookup("button.html")
			row.ExecuteTemplate(w, "button", pageData)
			row.Execute(w, pageData)
			button_id++
		}

		row = templates.Lookup("break.html")
		row.ExecuteTemplate(w, "break", nil)
		row.Execute(w, nil)
	}

	row = templates.Lookup("index_footer.html")
	row.ExecuteTemplate(w, "index_footer", nil)
	row.Execute(w, nil)

	r.ParseForm()

}

func show_media(w http.ResponseWriter, r *http.Request) {

	// Load all templates
	var allFiles []string
	files, err := ioutil.ReadDir("./template")
	checkError(err)

	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".html") {
			allFiles = append(allFiles, "./template/"+filename)
		}
	}

	// Building page in parts
	templates, err = template.ParseFiles(allFiles...)
	checkError(err)

	// Receive form and button id
	r.ParseForm()
	giveID := r.Form["niddevice"][0]

	// Log new device
	color.HiCyan(fmt.Sprintf("New device with ID: %s", giveID))

	// Convert ID type
	giveIDint, err := strconv.Atoi(giveID)
	if err != nil {
		log.Fatal(err)
	}

	// Setup device data in struct
	pageData := ID{
		HorScreens: horizontal_ScreenCount,
		VerScreens: vertical_ScreenCount,
		Nid:        giveIDint,
		ReSync:  	int64(reSync),
	}

	// Load and execute correct templates
	if media_type == "image" {
		row := templates.Lookup("image.html")
		row.ExecuteTemplate(w, "image", pageData)
		row.Execute(w, pageData)
	} else {
		row := templates.Lookup("video.html")
		row.ExecuteTemplate(w, "video", pageData)
		row.Execute(w, pageData)
	}
}

func load_templates() {
	// Load all template files
	var allFiles []string
	files, err := ioutil.ReadDir("./template")
	checkError(err)
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".html") {
			allFiles = append(allFiles, "./template/"+filename)
		}
	}
	templates, err = template.ParseFiles(allFiles...)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}