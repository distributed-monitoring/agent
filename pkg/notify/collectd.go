/*
 * Copyright 2018 NEC Corporation
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package notify

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// CollectdNotifier is an implementation of Notifier by collectd.
// Process using this package should be started by collectd-exec.
type CollectdNotifier struct {
	PluginName string
	TypeName   string
}

// Send is to notify message with collectd notification.
func (thisNotify CollectdNotifier) Send(message string, severity string, metaData [][2]string) error {
	unixNow := float64(time.Now().UnixNano()) / 1000000000

	var metaDataStr bytes.Buffer
	for _, data := range metaData {
		metaDataStr.WriteString(" s:")
		metaDataStr.WriteString(data[0])
		metaDataStr.WriteString("=\"")
		metaDataStr.WriteString(strings.Replace(data[1], "\"", "\\\"", -1))
		metaDataStr.WriteString("\"")
	}

	fmt.Printf("PUTNOTIF message=\"%s\" severity=%s time=%f "+
		"host=localhost plugin=%s type=%s %s\n",
		message, severity, unixNow, thisNotify.PluginName, thisNotify.TypeName, metaDataStr.String())

	return nil
}
