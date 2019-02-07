//-- Package Declaration -----------------------------------------------------------------------------------------------
package browsers

//-- Imports -----------------------------------------------------------------------------------------------------------
import (
	"log"
	"math/rand"
	"net/url"
	"time"
)

//-- Constants ---------------------------------------------------------------------------------------------------------
var webkitEpoch = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

//-- Structs -----------------------------------------------------------------------------------------------------------
type Browser interface {
	InjectHistory(History) error
	InjectCredential(Credential) error
	InjectBookmark(Bookmark) error

	open() error
	close() error
}

type History struct {
	Name        string
	URL         string
	Visits      int
	VisitWindow time.Duration
}

type Credential struct {
	URL          url.URL
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
		var subject = new(chrome)
		if err := subject.open(); err != nil {
			log.Println(`Error connecting to Chrome databases: `, err)
		} else if err := subject.purge(); err != nil {
			log.Println(`Error initializing Chrome databases: `, err)
		} else {
			browsers = append(browsers, subject)
		}
	}

	return browsers
}

func Close(browsers []Browser) {
	for _, brow := range browsers {
		if err := brow.close(); err != nil {
			log.Println(`error closing browser: `, err)
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
