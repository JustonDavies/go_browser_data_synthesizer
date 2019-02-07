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
	dataPath    string
	history     *gorm.DB
	credentials *gorm.DB
	bookmarks   *chromeBookmarksManifest
}

type chromeHistoryURL struct {
	//-- Primary Key ----------
	ID uint `gorm:"primary_key"`

	//-- User Variables ----------
	URL           string
	Title         string
	VisitCount    int `gorm:"default:0;not null"`
	LastVisitTime int `gorm:"not null"`

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

type chromeLogin struct {
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

func (chromeLogin) TableName() string {
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

	Roots map[string]*chromeBookmarkSet `json:"roots"`

	Version int `json:"version"`
}

//-- Exported Functions ------------------------------------------------------------------------------------------------
func (c *chrome) InjectHistory(item History) error {
	var history = &chromeHistoryURL{
		URL:           item.URL,
		Title:         item.Name,
		VisitCount:    item.Visits,
		LastVisitTime: int(randomWebKitTimestamp(item.VisitWindow)),
	}

	// Inject flat URL summary
	{
		var ctx = c.history.Begin()

		if result := ctx.Create(history); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	// Inject individual visits
	{
		var ctx = c.history.Begin()

		for i := 0; i < item.Visits; i++ {
			var visit = &chromeHistoryVisit{
				URL:       int(history.ID),
				VisitTime: int(randomWebKitTimestamp(item.VisitWindow)),

				Transition:    805306374,
				VisitDuration: 60000000,
			}

			if result := ctx.Create(visit); result.Error != nil {
				return result.Error
			}
		}

		if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func (c *chrome) InjectCredential(item Credential) error {
	//So this is using keyrings in most cases, GnomeKeyring/kWallet,
	return nil
}

func (c *chrome) InjectBookmark(item Bookmark) error {
	// Create new bookmark item
	var bookmark = &chromeBookmark{
		ID:        fmt.Sprintf(`%d`, len(c.bookmarks.Roots[`bookmark_bar`].Bookmarks)+5),
		Name:      item.Name,
		Type:      `url`,
		URL:       item.URL,
		CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(item.CreateWindow)),
	}

	// Insert into random bookmark set
	{
		var bookmarkSets []string
		for set := range c.bookmarks.Roots {
			bookmarkSets = append(bookmarkSets, set)
		}

		var randomSet = bookmarkSets[rand.Intn(len(bookmarkSets)-1)]
		c.bookmarks.Roots[randomSet].Bookmarks = append(c.bookmarks.Roots[randomSet].Bookmarks, bookmark)
	}

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
			c.history = orm
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
			c.credentials = orm
		}
	}

	// Open/Read/Close Bookmark manifest
	{
		if file, err := os.Open(c.dataPath + `Bookmarks`); os.IsNotExist(err) {
			c.bookmarks = c.newBookmarkManifest()
		} else if err != nil {
			return err
		} else if parser := json.NewDecoder(file); parser == nil {
			return err
		} else if err := parser.Decode(&c.bookmarks); err != nil {
			c.bookmarks = c.newBookmarkManifest()
		} else if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chrome) close() error {

	if err := c.history.Close(); err != nil {
		return err
	} else if err := c.credentials.Close(); err != nil {
		return err
	} else if err := c.writeBookmarkManifest(); err != nil {
		return err
	}

	return nil
}

func (c *chrome) purge() error {

	// Purge flat URL history
	{
		var ctx = c.history.Begin()
		if result := ctx.Exec(`DELETE FROM urls`); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	// Purge individual visits
	{
		var ctx = c.history.Begin()
		if result := ctx.Exec(`DELETE FROM visits`); result.Error != nil {
			return result.Error
		} else if result := ctx.Exec(`DELETE FROM visit_source`); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	// Purge individual download history
	{
		var ctx = c.history.Begin()

		if result := ctx.Exec(`DELETE FROM downloads`); result.Error != nil {
			return result.Error
		} else if result := ctx.Exec(`DELETE FROM downloads_slices`); result.Error != nil {
			return result.Error
		} else if result := ctx.Exec(`DELETE FROM downloads_url_chains`); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	// Purge individual search terms
	{
		var ctx = c.history.Begin()

		if result := ctx.Exec(`DELETE FROM keyword_search_terms`); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	// Purge segments
	{
		var ctx = c.history.Begin()

		if result := ctx.Exec(`DELETE FROM segment_usage`); result.Error != nil {
			return result.Error
		} else if result := ctx.Exec(`DELETE FROM segments`); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	// Purge credentials
	{
		var ctx = c.credentials.Begin()

		if result := ctx.Exec(`DELETE FROM logins`); result.Error != nil {
			return result.Error
		} else if result := ctx.Exec(`DELETE FROM stats`); result.Error != nil {
			return result.Error
		} else if result := ctx.Commit(); result.Error != nil {
			return result.Error
		}
	}

	// Purge Bookmarks
	{
		c.bookmarks = c.newBookmarkManifest()
		if err := c.writeBookmarkManifest(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chrome) newBookmarkManifest() *chromeBookmarksManifest {
	var mainifest = new(chromeBookmarksManifest)

	mainifest.Roots = map[string]*chromeBookmarkSet{
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
			Name:      `Other bookmarks`,
			Type:      `folder`,
			CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(24*time.Hour))),
			UpdatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(1*time.Hour))),
			Bookmarks: []*chromeBookmark{},
		},
		`synced`: {
			ID:        `3`,
			Name:      `Mobile bookmarks`,
			Type:      `folder`,
			CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(24*time.Hour))),
			UpdatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(1*time.Hour))),
			Bookmarks: []*chromeBookmark{},
		},
	}

	mainifest.Version = 1

	return mainifest
}

func (c *chrome) writeBookmarkManifest() error {
	{
		if err := os.Remove(c.dataPath + `Bookmarks`); err != nil && !os.IsNotExist(err) {
			return err
		} else if err := os.Remove(c.dataPath + `Bookmarks.bak`); err != nil && !os.IsNotExist(err) {
			return err
		} else if file, err := os.Create(c.dataPath + `Bookmarks`); err != nil {
			return err
		} else if err := file.Truncate(0); err != nil {
			return err
		} else if output, err := json.Marshal(c.bookmarks); err != nil {
			return err
		} else if _, err := file.WriteString(string(output)); err != nil {
			return err
		} else if err := file.Close(); err != nil {
			return err
		}

	}
	return nil
}
