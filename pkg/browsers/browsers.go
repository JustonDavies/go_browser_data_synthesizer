//-- Package Declaration -----------------------------------------------------------------------------------------------
package browsers

//-- Imports -----------------------------------------------------------------------------------------------------------
import (
	"log"
	"math/rand"
	"time"
)

//-- Constants ---------------------------------------------------------------------------------------------------------
var webkitEpoch = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

//-- Structs -----------------------------------------------------------------------------------------------------------
type Browser interface {
	AddHistory(History) error
	AddBookmark(Bookmark) error
	AddCredential(Credential) error

	open() error
	load() error
	close() error
	purge() error
	commit() error
}

type History struct {
	Name        string
	URL         string
	Visits      int
	VisitWindow time.Duration
}

type Credential struct {
	URL          string
	UserName     string
	Password     string
	CreateWindow time.Duration
}

type Bookmark struct {
	Name         string
	URL          string
	CreateWindow time.Duration
}

//-- Exported Functions ------------------------------------------------------------------------------------------------
func Open() []Browser {
	var browsers []Browser

	{
		var browser = new(chrome)
		if err := browser.open(); err != nil {
			log.Println(`error connecting to chrome data sets: `, err)
		} else {
			browsers = append(browsers, browser)
		}
	}

	return browsers
}

func Load(browsers []Browser) {
	for _, browser := range browsers {
		if err := browser.load(); err != nil {
			log.Println(`error committing browser: `, err)
		}
	}
}

func Close(browsers []Browser) {
	for _, browser := range browsers {
		if err := browser.close(); err != nil {
			log.Println(`error closing browser: `, err)
		}
	}
}

func Purge(browsers []Browser) {
	for _, browser := range browsers {
		if err := browser.purge(); err != nil {
			log.Println(`error purging browser: `, err)
		}
	}
}

func Commit(browsers []Browser) {
	for _, browser := range browsers {
		if err := browser.commit(); err != nil {
			log.Println(`error committing browser: `, err)
		}
	}
}

//-- Internal Functions ------------------------------------------------------------------------------------------------
func randomWebKitTimestamp(duration time.Duration) int64 {
	rand.Seed(time.Now().UnixNano())

	var microMultiplier = int64(1000000)
	var randomUnix = time.Now().Unix() - rand.Int63n(int64(duration.Seconds())) - webkitEpoch.Unix()
	return randomUnix * microMultiplier
}
