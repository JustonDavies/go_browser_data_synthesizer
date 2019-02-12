//-- Package Declaration -----------------------------------------------------------------------------------------------
package browsers

//-- Imports -----------------------------------------------------------------------------------------------------------
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//-- Constants ---------------------------------------------------------------------------------------------------------
var (
	CHROME_DEFAULT_PROFILE   = `Default`
	CHROME_LINUX_DATA_PATH   = fmt.Sprintf(`%s/.config/google-chrome/%s/`, os.Getenv(`HOME`), CHROME_DEFAULT_PROFILE)
	CHROME_DARWIN_DATA_PATH  = fmt.Sprintf(`%s/Library/Application Support/Google/Chrome/%s/`, os.Getenv(`HOME`), CHROME_DEFAULT_PROFILE)
	CHROME_WINDOWS_DATA_PATH = fmt.Sprintf(`%s\Google\Chrome\User Data\%s\`, os.Getenv(`LOCALAPPDATA`), CHROME_DEFAULT_PROFILE)
)

//-- Structs -----------------------------------------------------------------------------------------------------------
type chrome struct {
	dataPath string

	historyDatabase    *gorm.DB
	credentialDatabase *gorm.DB
	bookmarkFile       *os.File

	historyItems     []*chromeHistoryURL
	credentialItems  []*chromeCredential
	bookmarkManifest *chromeBookmarksManifest
}

type chromeHistoryURL struct {
	//-- Primary Key ----------
	ID uint `gorm:"primary_key"`

	//-- User Variables ----------
	URL           string
	Title         string
	VisitCount    int `gorm:"default:0;not null"`
	LastVisitTime int `gorm:"not null"`

	//-- Relations ----------
	Visits []*chromeHistoryVisit `gorm:"foreignkey:URL"`

	//-- System Variables ----------
	TypedCount int `gorm:"default:0;not null"`
	Hidden     int `gorm:"default:0"`
}

func (chromeHistoryURL) TableName() string {
	return `urls`
}

type chromeHistoryVisit struct {
	//-- Primary Key ----------
	ID uint `gorm:"primary_key"`

	//-- User Variables ----------
	URL       int `gorm:"not null"`
	VisitTime int `gorm:"not null"`

	//-- Relations ----------

	//-- System Variables ----------
	FromVisit                    int
	Transition                   int `gorm:"default:0;not null"`
	SegmentID                    int
	VisitDuration                int  `gorm:"default:0;not null"`
	IncrementedOmniboxTypedScore bool `gorm:"default:false;not null"`
}

func (chromeHistoryVisit) TableName() string {
	return `visits`
}

type chromeCredential struct {
	//-- Primary Key ----------

	//-- User Variables ----------
	OriginURL         string `gorm:"not null"`
	ActionURL         string
	SignonRealm       string `gorm:"not null"`
	UsernameValue     string
	PasswordValue     []byte
	DateCreated       int `gorm:"not null"`
	BlacklistedByUser int `gorm:"not null"`
	Scheme            int `gorm:"not null"`
	PasswordType      int
	DisplayName       string

	//-- System Variables ----------
	UsernameElement string
	PasswordElement string
	SubmitElement   string

	Preferred int `gorm:"not null"`

	TimesUsed  int
	FormData   []byte
	DateSynced int

	IconURL                string
	FederationURL          string
	SkipZeroClick          int
	GenerationUploadStatus int
	PossibleUsernamePairs  []byte
}

func (chromeCredential) TableName() string {
	return `logins`
}

type chromeBookmark struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`

	CreatedAt string `json:"date_added"`
}

type chromeBookmarkSet struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`

	CreatedAt string `json:"date_added"`
	UpdatedAt string `json:"date_modified"`

	Bookmarks []*chromeBookmark `json:"children"`
}

type chromeBookmarksManifest struct {
	//Checksum string `json:"checksum"`

	Folders map[string]*chromeBookmarkSet `json:"roots"`

	Version int `json:"version"`
}

func (c *chromeBookmarksManifest) defaults() *chromeBookmarksManifest {
	c.Folders = map[string]*chromeBookmarkSet{
		`bookmark_bar`: {
			ID:        `1`,
			Name:      `Bookmarks bar`,
			Type:      `folder`,
			CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(24*time.Hour))),
			UpdatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(1*time.Hour))),
			Bookmarks: []*chromeBookmark{},
		},
		`other`: {
			ID:        `2`,
			Name:      `Other bookmarkManifest`,
			Type:      `folder`,
			CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(24*time.Hour))),
			UpdatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(1*time.Hour))),
			Bookmarks: []*chromeBookmark{},
		},
		`synced`: {
			ID:        `3`,
			Name:      `Mobile bookmarkManifest`,
			Type:      `folder`,
			CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(24*time.Hour))),
			UpdatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(1*time.Hour))),
			Bookmarks: []*chromeBookmark{},
		},
	}

	c.Version = 1

	return c
}

func (c *chromeBookmarksManifest) bookmarkCount() int {
	var count = 0

	for _, set := range c.Folders {
		count = count + len(set.Bookmarks)
	}

	return count
}

//-- Exported Functions ------------------------------------------------------------------------------------------------
func (c *chrome) AddHistory(item History) error {
	var newEntry = &chromeHistoryURL{
		URL:           item.URL,
		Title:         item.Name,
		VisitCount:    item.Visits,
		LastVisitTime: int(randomWebKitTimestamp(item.VisitWindow)),
	}

	// Add individual visit data
	{
		for i := 0; i < item.Visits; i++ {
			var visit = &chromeHistoryVisit{
				URL:       int(newEntry.ID),
				VisitTime: int(randomWebKitTimestamp(item.VisitWindow)),

				Transition:    805306374,
				VisitDuration: 60000000,
			}

			newEntry.Visits = append(newEntry.Visits, visit)

		}
	}

	c.historyItems = append(c.historyItems, newEntry)

	return nil
}

func (c *chrome) AddBookmark(item Bookmark) error {
	// Create new bookmark item
	var bookmark = &chromeBookmark{
		ID:        fmt.Sprintf(`%d`, len(c.bookmarkManifest.Folders)+c.bookmarkManifest.bookmarkCount()+100), //NOTE: If you add more than 100 folders IDs could overlap and this might make Chrome mad
		Name:      item.Name,
		Type:      `url`,
		URL:       item.URL,
		CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(item.CreateWindow)),
	}

	// Insert into random bookmark set
	{
		var bookmarkSets []string
		for set := range c.bookmarkManifest.Folders {
			bookmarkSets = append(bookmarkSets, set)
		}

		var randomSet = bookmarkSets[rand.Intn(len(bookmarkSets)-1)]
		c.bookmarkManifest.Folders[randomSet].Bookmarks = append(c.bookmarkManifest.Folders[randomSet].Bookmarks, bookmark)
	}

	return nil
}

func (c *chrome) AddCredential(item Credential) error {
	//So this is using keyrings in most cases, GnomeKeyring/kWallet,
	return nil
}

//-- Internal Functions ------------------------------------------------------------------------------------------------
func (c *chrome) open() error {

	// Determine OS-Specific Data Path
	{
		switch runtime.GOOS {
		case `linux`:
			c.dataPath = CHROME_LINUX_DATA_PATH
		case `darwin`:
			c.dataPath = CHROME_DARWIN_DATA_PATH
		case `windows`:
			c.dataPath = CHROME_WINDOWS_DATA_PATH
		}
	}

	// Open History database
	{
		var dataSourceName = fmt.Sprintf(`file:%sHistory`, c.dataPath)
		if orm, err := gorm.Open(`sqlite3`, dataSourceName); err != nil {
			return err
		} else if err := orm.DB().Ping(); err != nil {
			return err
		} else {
			c.historyDatabase = orm
		}
	}

	// Open Credential database
	{
		var dataSourceName = fmt.Sprintf(`file:%sLogin Data`, c.dataPath)
		if orm, err := gorm.Open(`sqlite3`, dataSourceName); err != nil {
			return err
		} else if err := orm.DB().Ping(); err != nil {
			return err
		} else {
			c.credentialDatabase = orm
		}
	}

	// Open/Read/Close Bookmark manifest
	{
		if file, err := os.Open(c.dataPath + `Bookmarks`); os.IsNotExist(err) {
			c.bookmarkFile = nil //TODO:Maybe I should create the file?
		} else if err != nil {
			return err
		} else {
			c.bookmarkFile = file
		}
	}

	return nil
}

func (c *chrome) load() error {
	// Load history
	{
		c.historyItems = []*chromeHistoryURL{}

		if result := c.historyDatabase.Find(&c.historyItems); result.Error != nil {
			return result.Error
		}
	}

	// Open Credential database
	{
		c.credentialItems = []*chromeCredential{}

		if result := c.credentialDatabase.Find(&c.credentialItems); result.Error != nil {
			return result.Error
		}
	}

	// Open/Read/Close Bookmark manifest
	{
		var parser = json.NewDecoder(c.bookmarkFile)
		if err := parser.Decode(&c.bookmarkManifest); err != nil {
			c.bookmarkManifest = new(chromeBookmarksManifest).defaults()
		}
	}

	return nil
}

func (c *chrome) close() error {

	if err := c.historyDatabase.Close(); err != nil {
		return err
	} else if err := c.credentialDatabase.Close(); err != nil {
		return err
	} else if err := c.bookmarkFile.Close(); err != nil {
		return err
	}

	return nil
}

func (c *chrome) purge() error {

	// Purge history
	{
		var ctx = c.historyDatabase.Begin()
		// Purge flat URL history
		{
			if result := ctx.Exec(`DELETE FROM urls`); result.Error != nil {
				return result.Error
			}
		}

		// Purge individual visit history
		{
			if result := ctx.Exec(`DELETE FROM visits`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM visit_source`); result.Error != nil {
				return result.Error
			}
		}

		// Purge individual download historyDatabase
		{
			if result := ctx.Exec(`DELETE FROM downloads`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM downloads_slices`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM downloads_url_chains`); result.Error != nil {
				return result.Error
			}
		}

		// Purge individual search terms
		{
			if result := ctx.Exec(`DELETE FROM keyword_search_terms`); result.Error != nil {
				return result.Error
			}
		}

		// Purge segments
		{
			if result := ctx.Exec(`DELETE FROM segment_usage`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM segments`); result.Error != nil {
				return result.Error
			}
		}

		// Commit
		{
			if result := ctx.Commit(); result.Error != nil {
				return result.Error
			}
		}

		c.historyItems = []*chromeHistoryURL{}
	}

	// Purge credentialDatabase
	{
		var ctx = c.credentialDatabase.Begin()

		if result := ctx.Exec(`DELETE FROM logins`); result.Error != nil {
			return result.Error
		} else if result := ctx.Exec(`DELETE FROM stats`); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}

		c.credentialItems = []*chromeCredential{}
	}

	// Purge Bookmarks
	{
		c.bookmarkManifest = new(chromeBookmarksManifest).defaults()
		if err := c.writeBookmarks(); err != nil {
			return err
		}

		c.bookmarkManifest = new(chromeBookmarksManifest).defaults()
	}

	return nil
}

func (c *chrome) commit() error {
	//-- Commit pending historyDatabase ----------
	{
		var ctx = c.historyDatabase.Begin()

		for _, history := range c.historyItems {

			if result := ctx.Save(history); result.Error != nil {
				return result.Error
			}
		}

		if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	//-- Commit pending credentialDatabase ----------
	{
		var ctx = c.credentialDatabase.Begin()

		for _, credential := range c.credentialItems {

			if result := ctx.Save(credential); result.Error != nil {
				return result.Error
			}
		}

		if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	//-- Commit pending bookmarkManifest ----------
	{
		if err := c.writeBookmarks(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chrome) writeBookmarks() error {
	{
		if err := os.Remove(c.dataPath + `Bookmarks.bak`); err != nil && !os.IsNotExist(err) {
			return err
		}

		if err := c.bookmarkFile.Truncate(0); err != nil {
			return err
		} else if output, err := json.Marshal(c.bookmarkManifest); err != nil {
			return err
		} else if _, err := c.bookmarkFile.WriteString(string(output)); err != nil {
			return err
		}
	}

	return nil
}
