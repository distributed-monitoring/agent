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
	"github.com/BurntSushi/toml"
	"log"
)

var serverTypeOpt = flag.String("type", "pubsub", "server type: pubsub, api")

const confDirPath = "/etc/collectd/collectd.conf.d/"

// Config is ...
type Config struct {
	Amqp AmqpConfig
}

// AmqpConfig is ...
type AmqpConfig struct {
	Host     string
	User     string
	Password string
	Port     string
}

func main() {

	var config Config
	_, err := toml.DecodeFile("../../config/config.toml", &config)
	if err != nil {
		log.Println("read error of amqp config")
	}

	flag.Parse()

	switch *serverTypeOpt {
	case "pubsub":
		runSubscriber(&config.Amqp)
	case "api":
		runAPIServer()
	default:
		log.Fatalln("server type is wrong, see help.")
	}
}
