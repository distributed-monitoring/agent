package notify

import (
	"fmt"
	"time"
)

func NotifyCollectd(message string,
	colPlugin string, colType string, metaKey string, metaData string) {
	unixNow := float64(time.Now().UnixNano()) / 1000000000
	fmt.Printf("PUTNOTIF message=\"%s\" severity=warning time=%f "+
		"host=localhost plugin=%s type=%s s:%s=\"%s\"\n",
		message, unixNow, colPlugin, colType, metaKey, metaData)

}
