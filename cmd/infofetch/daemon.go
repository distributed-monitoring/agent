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

package main

import (
	"github.com/go-redis/redis"
	"time"
)

/*
  e.g. set("server", "instance-00000001", "{"os-name": "testvm1"}")
*/
type pool interface {
	set(string, string, string) error   // (infoType, libvirtValue, JsonData)
	get(string, string) (string, error) // (infoType, libvirtValue)
	del(string, string) error           // (infoType, libvirtValue)
}

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	infoPool := redisPool{client: redisClient}

	forever := make(chan bool)

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			writeInfo(infoPool)
		}
	}()

	<-forever
}
