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
	"context"
	"flag"
	"github.com/BurntSushi/toml"
	"log"
	"sync"
)

var serverTypeOpt = flag.String("type", "pubsub", "server type: pubsub, api")

// Config is ...
type Config struct {
	Amqp     AmqpConfig
	Collectd CollectdConfig
}

// AmqpConfig is ...
type AmqpConfig struct {
	Host     string
	User     string
	Password string
	Port     string
}

// CollectdConfig is ...
type CollectdConfig struct {
	ConfDir string
}

func main() {

	var config Config
	_, err := toml.DecodeFile("../../config/config.toml", &config)
	if err != nil {
		log.Println("read error of config file")
	}

	var waitgroup sync.WaitGroup

	flag.Parse()

	switch *serverTypeOpt {
	case "pubsub":
		ctx := context.Background()
		waitgroup.Add(1)
		go func() {
			defer waitgroup.Done()
			runSubscriber(ctx, &config)
		}()
		log.Printf("Waiting for publish.")
	case "api":
		ctx := context.Background()
		waitgroup.Add(1)
		go func() {
			defer waitgroup.Done()
			runAPIServer(ctx, &config.Collectd)
		}()
		log.Printf("Waiting for REST.")
	default:
		log.Fatalln("server type is wrong, see help.")
	}

	waitgroup.Wait()
}
