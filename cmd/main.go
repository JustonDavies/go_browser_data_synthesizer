//-- Package Declaration -----------------------------------------------------------------------------------------------
package main

//-- Imports -----------------------------------------------------------------------------------------------------------
import (
	"log"
	"math/rand"
	"time"

	"github.com/JustonDavies/go_browser_data_synthesizer/configs"
	"github.com/JustonDavies/go_browser_data_synthesizer/pkg/browsers"
)

//-- Constants ---------------------------------------------------------------------------------------------------------

//-- Structs -----------------------------------------------------------------------------------------------------------

//-- Exported Functions ------------------------------------------------------------------------------------------------
func main() {
	//-- Log nice output ----------
	var start = time.Now().Unix()
	log.Println(`Starting task...`)

	//-- Perform task ----------
	var browserz = browsers.Open()

	if len(browserz) < 1 {
		panic(`unable to open any supported browsers, aborting...`)
	} else {
		defer browsers.Close(browserz)
		rand.Seed(time.Now().UnixNano())
	}

	browsers.Load(browserz)
	browsers.Purge(browserz)

	log.Println(`Creating history...`)
	for _, item := range configs.ActivityItems {
		var browser = browserz[rand.Intn(len(browserz))]
		var item = browsers.History{
			Name:        item.Name,
			URL:         item.URL,
			Visits:      rand.Intn(configs.MaximumVisits),
			VisitWindow: configs.DefaultDuration,
		}

		if err := browser.AddHistory(item); err != nil {
			log.Printf("unable to inject history item for: \n\tURL: '%s' \n\tError: '%s'", item.URL, err)
		}
	}

	log.Println(`Creating bookmarks...`)
	for _, item := range configs.ActivityItems {
		if rand.Intn(configs.BookmarkOneInX) == 0 {
			var browser = browserz[rand.Intn(len(browserz))]
			var item = browsers.Bookmark{
				Name:         item.Name,
				URL:          item.URL,
				CreateWindow: configs.DefaultDuration,
			}

			if err := browser.AddBookmark(item); err != nil {
				log.Printf("unable to inject bookmark item for: \n\tURL: '%s' \n\tError: '%s'", item.URL, err)
			}
		}
	}

	log.Println(`Committing changes...`)
	browsers.Commit(browserz)

	//-- Log nice output ----------
	log.Printf(`Task complete! It took %d seconds`, time.Now().Unix()-start)
}

//-- Internal Functions ------------------------------------------------------------------------------------------------
