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

// Notifier is an interface of notification.
// e.g. Send("Threshold is exceeded.", "warning", [["virt-name", "instance-00000001"], [...]])
type Notifier interface {
	Send(string, string, [][2]string) error // (message, severity, metaDataList)
}
