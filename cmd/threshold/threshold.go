package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"strings"
	"time"
)

const interval = 10 //collectd interval

func zrangebyscore(client *redis.Client) {

	unixNow := int(time.Now().Unix())

	val, err := client.ZRangeByScore("collectd/wanted-filly.maas/memory/memory-used", redis.ZRangeBy{
		Min: strconv.Itoa(unixNow - interval),
		Max: strconv.Itoa(unixNow),
	}).Result()

	if err == redis.Nil {
		fmt.Println("this key is not exist")
		os.Exit(1)
	} else if err != nil {
		panic(err)
	} else {
		split := strings.Split(val[0], ":")
		val := split[1]
		intVal, err := strconv.Atoi(val)
		if err != nil {
			os.Exit(1)
		}
		fmt.Println(intVal)
		threshold(intVal)
	}
}

func threshold(val int) {

	thresholdVal := 270540800
	if val > thresholdVal {
		action()
	}
}

func action() {
	fmt.Println("kick action")
}

func main() {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	//how to stop after compile...
	ticker := time.NewTicker(interval * time.Second)

	for range ticker.C {
		zrangebyscore(client)
	}

	fmt.Println("end")
}
