package bridges

import (
	"time"

	"github.com/mmcdole/gofeed/atom"
)

type Bridge interface {
	Entries() []atom.Entry
	UpdateEntries() error
	LastFetched() time.Time
	CacheTime() time.Duration
}
