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
	"bytes"
	"fmt"
	"github.com/distributed-monitoring/agent/pkg/annotate"
	"github.com/distributed-monitoring/agent/pkg/notify"
	"github.com/go-redis/redis"
	"strconv"
	"strings"
)

func transmit(config *Config, edlist []evalData) {
	annoConfig := config.Common

	client := redis.NewClient(&redis.Options{
		Addr:     annoConfig.RedisHost + ":" + annoConfig.RedisPort,
		Password: annoConfig.RedisPassword,
		DB:       annoConfig.RedisDB,
	})
	pool := annotate.RedisPool{Client: client}

	notifier := notify.CollectdNotifier{
		PluginName: config.Threshold.CollectdPlugin,
		TypeName:   config.Threshold.CollectdType}

	for _, ed := range edlist {
		if ed.label == 1 {

			fmt.Println("kick action")

			item := strings.Split(ed.key, "/")
			ifItem := strings.SplitN(item[3], "-", 2)
			virtName := item[1]
			virtIF := ifItem[1]

			var message bytes.Buffer
			message.WriteString("Value exceeded threshold ")
			message.WriteString(strconv.Itoa(config.Threshold.Min))
			message.WriteString(".")

			nameVal, _ := pool.Get(fmt.Sprintf("%s/%s/vminfo", "vm", virtName))
			ifVal, _ := pool.Get(fmt.Sprintf("%s/%s/neutron_network", "if", virtIF))

			nameInfo := fmt.Sprintf("{\"%s\": %s}", virtName, nameVal)
			ifInfo := fmt.Sprintf("{\"%s\": %s}", virtIF, ifVal)

			fmt.Println(nameInfo)
			fmt.Println(ifInfo)

			notifier.Send(message.String(),
				"warning",
				[][2]string{{"vminfo", nameInfo}, {"neutron_network", ifInfo}})

		}
	}
}
