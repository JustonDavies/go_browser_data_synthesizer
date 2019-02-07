//-- Package Declaration -----------------------------------------------------------------------------------------------
package browsers

//-- Imports -----------------------------------------------------------------------------------------------------------
import (
	"fmt"
	"os"
	"runtime"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//-- Constants ---------------------------------------------------------------------------------------------------------
var (
	FIREFOX_DEFAULT_PROFILE   = `default` //TODO: Firefox assignes a random 4 character predix to this, need to emulate or detect an existing profile
	FIREFOX_LINUX_DATA_PATH   = fmt.Sprintf(`%s/.mozilla/firefox/%s/`, os.Getenv(`HOME`), FIREFOX_DEFAULT_PROFILE)
	FIREFOX_DARWIN_DATA_PATH  = fmt.Sprintf(`%s/Library/Application Support/Firefox/Profiles/%s/`, os.Getenv(`HOME`), FIREFOX_DEFAULT_PROFILE)
	FIREFOX_WINDOWS_DATA_PATH = fmt.Sprintf(`%s\Mozilla\Firefox\Profiles\%s\`, os.Getenv(`LOCALAPPDATA`), FIREFOX_DEFAULT_PROFILE)
)

//-- Structs -----------------------------------------------------------------------------------------------------------
type firefox struct {
	dataPath    string
	history     *gorm.DB
	credentials *gorm.DB
	bookmarks   *chromeBookmarksManifest
}

//-- Exported Functions ------------------------------------------------------------------------------------------------
func (f *firefox) InjectHistory(item History) error {
	return nil
}

func (f *firefox) InjectCredential(item Credential) error {
	return nil
}

func (f *firefox) InjectBookmark(item Bookmark) error {
	return nil
}

//-- Internal Functions ------------------------------------------------------------------------------------------------
func (f *firefox) open() error {

	// Determine OS-Specific Data Path
	{
		switch runtime.GOOS {
		case `linux`:
			f.dataPath = FIREFOX_LINUX_DATA_PATH
		case `darwin`:
			f.dataPath = FIREFOX_DARWIN_DATA_PATH
		case `windows`:
			f.dataPath = FIREFOX_WINDOWS_DATA_PATH
		}
	}

	return nil
}

func (f *firefox) close() error {
	return nil
}

func (f *firefox) purge() error {
	return nil
}
