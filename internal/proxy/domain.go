package proxy

import (
	"fmt"
	"time"
)

func SuggestRandomSubdomain(site string) string {
	ts := time.Now().UTC().Unix()
	return fmt.Sprintf("%s-%d.gogal.local", site, ts)
}
