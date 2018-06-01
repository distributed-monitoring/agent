package main

import (
	"bytes"
	"fmt"
	"github.com/distributed-monitoring/agent/pkg/annotate"
	"github.com/distributed-monitoring/agent/pkg/notify"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"strings"
	"time"
)

const interval = 5 //collectd interval

// e.g. collectd/instance-00000001/virt/if_octets-tapd21acb51-35
const redisKey = "collectd/*/virt/if_octets-*"
const minThresh = 1000000

// exceed by exec `sudo ping <IP> -i 0.00005 -c 1000 -s 1000`

func zrangebyscore(client *redis.Client, key string, pool annotate.Pool, notifier notify.Notifier) {

	unixNow := int(time.Now().Unix())

	val, err := client.ZRangeByScore(key, redis.ZRangeBy{
		Min: strconv.Itoa(unixNow - interval),
		Max: strconv.Itoa(unixNow),
	}).Result()

	if err == redis.Nil {
		fmt.Println("this key is not exist")
		os.Exit(1)
	} else if err != nil {
		panic(err)
	} else {
		maxVal := 0
		for _, strVal := range val {
			split := strings.Split(strVal, ":")
			txVal := split[2]
			floatVal, err := strconv.ParseFloat(txVal, 64)
			if err != nil {
				os.Exit(1)
			}
			intVal := int(floatVal)
			if maxVal < intVal {
				maxVal = intVal
			}
		}
		if maxVal > minThresh {

			fmt.Println("kick action")

			item := strings.Split(key, "/")
			ifItem := strings.SplitN(item[3], "-", 2)
			virtName := item[1]
			virtIF := ifItem[1]

			var message bytes.Buffer
			message.WriteString("Value ")
			message.WriteString(strconv.Itoa(maxVal))
			message.WriteString(" exceeded threshold ")
			message.WriteString(strconv.Itoa(minThresh))
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

/*
func action(val int) {
	fmt.Println("kick action")
}
*/

func checkVirtIF(client *redis.Client, pool annotate.Pool, notifier notify.Notifier) {
	keys, err := client.Keys(redisKey).Result()
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		zrangebyscore(client, key, pool, notifier)
	}
}

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	infoPool := annotate.RedisPool{Client: client}

	notifier := notify.CollectdNotifier{
		PluginName: "barometer-localagent",
		TypeName:   "if-octets-threshold"}

	//how to stop after compile...
	ticker := time.NewTicker(interval * time.Second)

	for range ticker.C {
		checkVirtIF(client, infoPool, notifier)
	}

	fmt.Println("end")
}
