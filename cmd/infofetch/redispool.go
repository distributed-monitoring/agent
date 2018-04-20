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
	"log"
)

const collectdLabel = "barometer-localagent"

type redisPool struct {
	client *redis.Client
}

func (thisPool redisPool) set(infoType string, virtVal string, data string) error {
	key := collectdLabel + "/" + infoType + "/" + virtVal
	err := thisPool.client.Set(key, data, 0).Err()
	if err != nil {
		log.Printf("redis Set error: %s", err)
	}
	return err
}

func (thisPool redisPool) get(infoType string, virtVal string) (string, error) {
	key := collectdLabel + "/" + infoType + "/" + virtVal
	value, err := thisPool.client.Get(key).Result()
	if err != nil {
		log.Printf("redis Get error: %s", err)
	}
	return value, err
}

func (thisPool redisPool) del(infoType string, virtVal string) error {
	key := collectdLabel + "/" + infoType + "/" + virtVal
	err := thisPool.client.Del(key).Err()
	if err != nil {
		log.Printf("redis Del error: %s", err)
	}
	return err
}
