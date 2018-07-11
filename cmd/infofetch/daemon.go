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
	"context"
	"github.com/BurntSushi/toml"
	"github.com/distributed-monitoring/agent/pkg/annotate"
	"github.com/go-redis/redis"
	libvirt "github.com/libvirt/libvirt-go"
	"log"
	"sync"
)

var InfoPool annotate.RedisPool

// Config is ...
type Config struct {
	Redis RedisConfig
}

// RedisConfig is ...
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func main() {

	var config Config
	_, err := toml.DecodeFile("../../config/config.toml", &config)
	if err != nil {
		log.Println("read error of config file")
	}

	var waitgroup sync.WaitGroup
	libvirt.EventRegisterDefaultImpl()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisConfig.Host + ":" + config.RedisConfig.Port,
		Password: config.RedisConfig.Passwrd,
		DB:       config.RedisConfig.DB,
	})
	InfoPool = annotate.RedisPool{Client: redisClient}
	// Initialize redis db...
	InfoPool.DelAll()

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Fatalln("libvirt connect error")
	}
	defer conn.Close()

	vmIfInfoChan := make(chan string)
	{
		ctx := context.Background()
		waitgroup.Add(1)
		go func() {
			RunNeutronInfoFetch(ctx, vmIfInfoChan)
			waitgroup.Done()
		}()
	}

	//Get active VM info
	GetActiveDomain(conn, vmIfInfoChan)
	{
		ctx := context.Background()
		waitgroup.Add(1)
		go func() {
			RunVirshEventLoop(ctx, conn, vmIfInfoChan)
			waitgroup.Done()
		}()
	}

	waitgroup.Wait()
}
