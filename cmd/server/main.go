/*
 * Copyright 2017 NEC Corporation
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

package main

import (
	"flag"
	"log"
)

var serverTypeOpt = flag.String("type", "pubsub", "server type: pubsub, api")

const confDirPath = "/etc/collectd/collectd.conf.d/"

func main() {
	flag.Parse()

	switch *serverTypeOpt {
	case "pubsub":
		runSubscriber()
	case "api":
		runAPIServer()
	default:
		log.Fatalln("server type is wrong, see help.")
	}
}
