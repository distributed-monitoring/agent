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

import (
	"github.com/go-redis/redis"
	"log"
)

const redisLabel = "barometer-localagent"

// RedisPool is an implementation of Pool by redis.
type RedisPool struct {
	Client *redis.Client
}

// Set is to set data in redis.
func (thisPool RedisPool) Set(infoType string, infoName string, data string) error {
	key := redisLabel + "/" + infoType + "/" + infoName
	err := thisPool.Client.Set(key, data, 0).Err()
	if err != nil {
		log.Printf("redis Set error: %s", err)
	}
	return err
}

// Get is to get data in redis.
func (thisPool RedisPool) Get(infoType string, infoName string) (string, error) {
	key := redisLabel + "/" + infoType + "/" + infoName
	value, err := thisPool.Client.Get(key).Result()
	if err != nil {
		log.Printf("redis Get error: %s", err)
	}
	return value, err
}

// Del is to delete data in redis.
func (thisPool RedisPool) Del(infoType string, infoName string) error {
	key := redisLabel + "/" + infoType + "/" + infoName
	err := thisPool.Client.Del(key).Err()
	if err != nil {
		log.Printf("redis Del error: %s", err)
	}
	return err
}
