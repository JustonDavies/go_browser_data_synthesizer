//-- Package Declaration -----------------------------------------------------------------------------------------------
package browsers

//-- Imports -----------------------------------------------------------------------------------------------------------
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//-- Constants ---------------------------------------------------------------------------------------------------------
var (
	CHROME_STATE_FILE        = `Local State`
	CHROME_BOOKMARK_BUFFER   = 1000
	CHROME_LINUX_DATA_PATH   = fmt.Sprintf(`%s/.config/google-chrome/`, os.Getenv(`HOME`))
	CHROME_DARWIN_DATA_PATH  = fmt.Sprintf(`%s/Library/Application Support/Google/Chrome/`, os.Getenv(`HOME`))
	CHROME_WINDOWS_DATA_PATH = fmt.Sprintf(`%s\Google\Chrome\User Data\`, os.Getenv(`LOCALAPPDATA`))
)

//-- Structs -----------------------------------------------------------------------------------------------------------
type chrome struct {
	dataPath  string
	stateFile *os.File

	state    *chromeState
	profiles []*chromeProfile
}

type chromeState struct {
	Profile struct {
		Info map[string]struct {
			Name string `json:"name"`
		} `json:"info_cache"`
	} `json:"profile"`
}

type chromeProfile struct {
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

func (c *chromeBookmarksManifest) init() *chromeBookmarksManifest {
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
			Name:      `Other Bookmarks`,
			Type:      `folder`,
			CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(24*time.Hour))),
			UpdatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(time.Duration(1*time.Hour))),
			Bookmarks: []*chromeBookmark{},
		},
		`synced`: {
			ID:        `3`,
			Name:      `Mobile Bookmarks`,
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
	//-- Select random profile ----------
	var profile *chromeProfile
	{
		if len(c.profiles) < 1 {
			return errors.New(`no profiles detected, unable to act`)
		} else {
			profile = c.profiles[rand.Intn(len(c.profiles))]
		}
	}

	//-- Create history entry ----------
	{
		var newEntry = &chromeHistoryURL{
			URL:           item.URL,
			Title:         item.Name,
			VisitCount:    item.Visits,
			LastVisitTime: int(randomWebKitTimestamp(item.VisitWindow)),
		}

		//-- Add individual visit data ----------
		for i := 0; i < item.Visits; i++ {
			var visit = &chromeHistoryVisit{
				URL:       int(newEntry.ID),
				VisitTime: int(randomWebKitTimestamp(item.VisitWindow)),

				Transition:    805306374,
				VisitDuration: 60000000,
			}

			newEntry.Visits = append(newEntry.Visits, visit)

		}

		profile.historyItems = append(profile.historyItems, newEntry)
	}

	//-- Return ---------
	return nil
}

func (c *chrome) AddBookmark(item Bookmark) error {
	//-- Select random profile ----------
	var profile *chromeProfile
	{
		if len(c.profiles) < 1 {
			return errors.New(`no profiles detected, unable to act`)
		} else {
			profile = c.profiles[rand.Intn(len(c.profiles))]
		}
	}

	//-- Create new bookmark item ----------
	var newEntry = &chromeBookmark{
		ID:        fmt.Sprintf(`%d`, len(profile.bookmarkManifest.Folders)+profile.bookmarkManifest.bookmarkCount()+CHROME_BOOKMARK_BUFFER),
		Name:      item.Name,
		Type:      `url`,
		URL:       item.URL,
		CreatedAt: fmt.Sprintf(`%d`, randomWebKitTimestamp(item.CreateWindow)),
	}

	//-- Insert into random position ----------
	{
		var bookmarkSets []string
		for set := range profile.bookmarkManifest.Folders {
			bookmarkSets = append(bookmarkSets, set)
		}

		var randomSet = bookmarkSets[rand.Intn(len(bookmarkSets)-1)]
		profile.bookmarkManifest.Folders[randomSet].Bookmarks = append(profile.bookmarkManifest.Folders[randomSet].Bookmarks, newEntry)
	}

	//-- Return ---------
	return nil
}

func (c *chrome) AddCredential(item Credential) error {
	//-- Select random profile ----------
	//var profile *chromeProfile
	//{
	//	if len(c.profiles) < 1 {
	//		return errors.New(`no profiles detected, unable to act`)
	//	} else {
	//		var names []string
	//		for set := range c.profiles {
	//			names = append(names, set)
	//		}
	//
	//		var random = names[rand.Intn(len(names)-1)]
	//		profile =  c.profiles[random]
	//	}
	//}

	//-- Create credential entry ----------
	//TODO: I need to map the values we have to this
	//{
	//	var newEntry = &chromeCredential{}
	//
	//	profile.credentialItems = append(profile.credentialItems, newEntry)
	//}

	//-- Return ---------
	return nil
}

//-- Internal Functions ------------------------------------------------------------------------------------------------
func (c *chrome) open() error {
	//-- Determine OS-specific Data Path ----------
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

	//-- Open/Parse `Local State` file----------
	{
		if file, err := os.Open(c.dataPath + CHROME_STATE_FILE); err != nil {
			return err
		} else {
			c.stateFile = file
		}

		var parser = json.NewDecoder(c.stateFile)

		if err := parser.Decode(&c.state); err != nil {
			return err
		}
	}

	//-- Connect to detected profiles ----------
	{
		var errs []error
		for directory := range c.state.Profile.Info {
			var profile = chromeProfile{dataPath: c.dataPath + directory + `/`}
			if err := profile.open(); err != nil {
				log.Printf(`Chrome: unable to connect to profile %s`, directory) //NOTE: Just doing this as a kindness, though it DOES break convention for the project
				errs = append(errs, err)
			} else {
				c.profiles = append(c.profiles, &profile)
			}
		}

		if len(c.profiles) < 1 {
			return errors.New(`unable to open any profiles`)
		}
	}

	//-- Return ---------
	return nil
}

func (c *chromeProfile) open() error {
	//-- Open history database ----------
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

	//-- Open credential database ----------
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

	//-- Open/Parse Bookmark file ----------
	{
		if file, err := os.Open(fmt.Sprintf(`%sBookmarks`, c.dataPath)); os.IsNotExist(err) {
			c.bookmarkFile = nil //NOTE: Maybe I should create the file or return err to eliminate error handling elsewhere.
		} else if err != nil {
			return err
		} else {
			c.bookmarkFile = file
		}
	}

	//-- Return ---------
	return nil
}

func (c *chrome) load() error {
	//-- Load each profile ----------
	{
		var errs []error
		for _, profile := range c.profiles {
			if err := profile.load(); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors.New(`one or more errors encountered trying to load profiles`)
		}
	}

	//-- Return ---------
	return nil
}

func (c *chromeProfile) load() error {
	//-- Load history ----------
	{
		c.historyItems = []*chromeHistoryURL{}

		if result := c.historyDatabase.Find(&c.historyItems); result.Error != nil {
			return result.Error
		}
	}

	//-- Open credentials ----------
	{
		c.credentialItems = []*chromeCredential{}

		if result := c.credentialDatabase.Find(&c.credentialItems); result.Error != nil {
			return result.Error
		}
	}

	//-- Open/Parse bookmark manifest ----------
	{
		c.bookmarkManifest = new(chromeBookmarksManifest).init()
		if c.bookmarkFile != nil {
			var parser = json.NewDecoder(c.bookmarkFile)
			if err := parser.Decode(&c.bookmarkManifest); err != nil {
				return err
			}
		}
	}

	//-- Return ---------
	return nil
}

func (c *chrome) close() error {
	//-- Close local state file ----------
	{
		if err := c.stateFile.Close(); err != nil {
			return err
		}
	}

	//-- Close detected profiles ----------
	{
		var errs []error
		for _, profile := range c.profiles {
			if err := profile.close(); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors.New(`one or more errors encountered trying to close profiles`)
		}
	}

	//-- Return ---------
	return nil
}

func (c *chromeProfile) close() error {

	//-- Close history database ----------
	{
		if err := c.historyDatabase.Close(); err != nil {
			return err
		}
	}

	//-- Close credential database ----------
	{
		if err := c.credentialDatabase.Close(); err != nil {
			return err
		}
	}

	//-- Close Bookmark file ----------
	{
		if c.bookmarkFile != nil {
			if err := c.bookmarkFile.Close(); err != nil {
				return err
			}
		}
	}

	//-- Return ---------
	return nil
}

func (c *chrome) purge() error {
	//-- Purge detected profiles ----------
	{
		var errs []error
		for _, profile := range c.profiles {
			if err := profile.purge(); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors.New(`one or more errors encountered trying to purge profiles`)
		}
	}

	//-- Return ---------
	return nil
}

func (c *chromeProfile) purge() error {
	//-- Purge history database ----------
	{
		var ctx = c.historyDatabase.Begin()

		//-- Purge flat URL history ----------
		{
			if result := ctx.Exec(`DELETE FROM urls`); result.Error != nil {
				return result.Error
			}
		}

		//-- Purge individual visit history ----------
		{
			if result := ctx.Exec(`DELETE FROM visits`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM visit_source`); result.Error != nil {
				return result.Error
			}
		}

		//-- Purge individual download historyDatabase ----------
		{
			if result := ctx.Exec(`DELETE FROM downloads`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM downloads_slices`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM downloads_url_chains`); result.Error != nil {
				return result.Error
			}
		}

		//-- Purge individual search terms ----------
		{
			if result := ctx.Exec(`DELETE FROM keyword_search_terms`); result.Error != nil {
				return result.Error
			}
		}

		//-- Purge segments ----------
		{
			if result := ctx.Exec(`DELETE FROM segment_usage`); result.Error != nil {
				return result.Error
			} else if result := ctx.Exec(`DELETE FROM segments`); result.Error != nil {
				return result.Error
			}
		}

		//-- Commit ----------
		{
			if result := ctx.Commit(); result.Error != nil {
				return result.Error
			}
		}

		c.historyItems = []*chromeHistoryURL{}
	}

	//-- Purge credential database ----------
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
		c.bookmarkManifest = new(chromeBookmarksManifest).init()
		if err := c.writeBookmarks(); err != nil {
			return err
		}

		c.bookmarkManifest = new(chromeBookmarksManifest).init()
	}

	//-- Return ---------
	return nil
}

func (c *chrome) commit() error {
	//-- Commit detected profiles ----------
	{
		var errs []error
		for _, profile := range c.profiles {
			if err := profile.commit(); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors.New(`one or more errors encountered trying to commit profiles`)
		}
	}

	//-- Return ---------
	return nil
}

func (c *chromeProfile) commit() error {
	//-- Commit pending history to database ----------
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

	//-- Commit pending credentials to database ----------
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

	//-- Return ---------
	return nil
}

func (c *chromeProfile) writeBookmarks() error {

	//-- Re-create and truncate file ----------
	{
		if c.bookmarkFile != nil {
			if err := c.bookmarkFile.Close(); err != nil {
				return err
			}
		}

		if file, err := os.Create(fmt.Sprintf(`%sBookmarks`, c.dataPath)); err != nil {
			return err
		} else {
			c.bookmarkFile = file
		}
	}

	//-- Clear backup file ----------
	{
		if err := os.Remove(fmt.Sprintf(`%sBookmarks.bak`, c.dataPath)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	//-- Write fresh bookmark file ----------
	{
		if output, err := json.Marshal(c.bookmarkManifest); err != nil {
			return err
		} else if _, err := c.bookmarkFile.WriteString(string(output)); err != nil {
			return err
		}
	}

	//-- Return ---------
	return nil
}
