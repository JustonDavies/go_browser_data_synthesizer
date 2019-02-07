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
	OPERA_LINUX_DATA_PATH   = fmt.Sprintf(`%s/.mozilla/opera/`, os.Getenv(`HOME`))
	OPERA_DARWIN_DATA_PATH  = fmt.Sprintf(`%s//Library/Application Support/com.operasoftware.Opera/`, os.Getenv(`HOME`))
	OPERA_WINDOWS_DATA_PATH = fmt.Sprintf(`%s\Opera Software\Opera Stable\`, os.Getenv(`LOCALAPPDATA`))
)

//-- Structs -----------------------------------------------------------------------------------------------------------
type opera struct {
	dataPath    string
	history     *gorm.DB
	credentials *gorm.DB
	bookmarks   *chromeBookmarksManifest
}

//-- Exported Functions ------------------------------------------------------------------------------------------------
func (o *opera) InjectHistory(item History) error {
	return nil
}

func (o *opera) InjectCredential(item Credential) error {
	return nil
}

func (o *opera) InjectBookmark(item Bookmark) error {
	return nil
}

//-- Internal Functions ------------------------------------------------------------------------------------------------
func (o *opera) open() error {

	// Determine OS-Specific Data Path
	{
		switch runtime.GOOS {
		case `linux`:
			o.dataPath = OPERA_LINUX_DATA_PATH
		case `darwin`:
			o.dataPath = OPERA_DARWIN_DATA_PATH
		case `windows`:
			o.dataPath = OPERA_WINDOWS_DATA_PATH
		}
	}

	return nil
}

func (o *opera) close() error {
	return nil
}

func (o *opera) purge() error {
	return nil
}
