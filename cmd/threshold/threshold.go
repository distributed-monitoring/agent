package main

import (
	"bytes"
	"fmt"
	"github.com/distributed-monitoring/agent/pkg/notify"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"strings"
	"time"
)

const interval = 5 //collectd interval

const redisKey = "collectd/localhost/interface-eth0/if_octets"
const minThresh = 5000000

// exceed by exec `sudo ping <IP> -i 0.00005 -c 1000 -s 1000`

func zrangebyscore(client *redis.Client, notifier notify.Notifier) {

	unixNow := int(time.Now().Unix())

	val, err := client.ZRangeByScore(redisKey, redis.ZRangeBy{
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
			var message bytes.Buffer
			message.WriteString("Value ")
			message.WriteString(strconv.Itoa(maxVal))
			message.WriteString(" exceeded threshold ")
			message.WriteString(strconv.Itoa(minThresh))
			message.WriteString(".")
			notifier.Send(message.String(),
				"warning",
				[][2]string{{"add_info", "some value"}})
		}
	}
}

/*
func action(val int) {
	fmt.Println("kick action")
}
*/

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	notifier := notify.CollectdNotifier{
		PluginName: "barometer-localhost",
		TypeName:   "if-octet-threshold"}

	//how to stop after compile...
	ticker := time.NewTicker(interval * time.Second)

	for range ticker.C {
		zrangebyscore(client, notifier)
	}

	fmt.Println("end")
}
