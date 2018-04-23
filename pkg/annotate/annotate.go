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

package annotate

/*
  e.g. Set("virt_name", "instance-00000001", "{"OS-name": "testvm1"}")
  e.g. Set("virt_if", "tap1e793b2b-8e", "{"OS-uuid": "df846647-c16a-4d8a-842a-ac39bd4a971e"}")
*/
type Pool interface {
	Set(string, string, string) error   // (infoType, infoName, JsonData)
	Get(string, string) (string, error) // (infoType, infoName)
	Del(string, string) error           // (infoType, infoName)
}
