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
	libvirt "github.com/libvirt/libvirt-go"
	"sync"
	"log"
	"github.com/distributed-monitoring/agent/pkg/annotate"
	"github.com/go-redis/redis"
)

var InfoPool annotate.RedisPool

func main() {
	var waitgroup sync.WaitGroup
	ctx := context.Background()
	libvirt.EventRegisterDefaultImpl()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	InfoPool = annotate.RedisPool{Client: redisClient}
	// Initialize redis db...
	InfoPool.DelAll()

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Fatalln("libvirt connect error")
	}
	defer conn.Close()

	//Get active VM info
	GetActiveDomain(conn)

	waitgroup.Add(1)
	go func() {
		RunVirshEventLoop(ctx, conn)
		waitgroup.Done()
	}()

	waitgroup.Wait()
}
