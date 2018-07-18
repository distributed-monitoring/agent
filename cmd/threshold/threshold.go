package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/distributed-monitoring/agent/pkg/annotate"
	"github.com/distributed-monitoring/agent/pkg/notify"
	"github.com/go-redis/redis"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// e.g. collectd/instance-00000001/virt/if_octets-tapd21acb51-35
const redisKey = "collectd/*/virt/if_octets-*"

func zrangebyscore(client *redis.Client, key string, pool annotate.Pool, config *Config, notifier notify.Notifier) {

	unixNow := int(time.Now().Unix())

	val, err := client.ZRangeByScore(key, redis.ZRangeBy{
		Min: strconv.Itoa(unixNow - config.Threshold.Interval),
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
		if maxVal > config.Threshold.Min {

			fmt.Println("kick action")

			item := strings.Split(key, "/")
			ifItem := strings.SplitN(item[3], "-", 2)
			virtName := item[1]
			virtIF := ifItem[1]

			var message bytes.Buffer
			message.WriteString("Value ")
			message.WriteString(strconv.Itoa(maxVal))
			message.WriteString(" exceeded threshold ")
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

/*
func action(val int) {
	fmt.Println("kick action")
}
*/

func checkVirtIF(client *redis.Client, pool annotate.Pool, config *Config, notifier notify.Notifier) {
	keys, err := client.Keys(redisKey).Result()
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		zrangebyscore(client, key, pool, config, notifier)
	}
}

// Config is ...
type Config struct {
	Common    CommonConfig
	Threshold ThresholdConfig
}

// CommonConfig is ...
type CommonConfig struct {
	RedisHost     string `toml:"redis_host"`
	RedisPort     string `toml:"redis_port"`
	RedisPassword string `toml:"redis_password"`
	RedisDB       int    `toml:"redis_db"`
}

// ThresholdConfig is ...
type ThresholdConfig struct {
	RedisHost     string `toml:"redis_host"`
	RedisPort     string `toml:"redis_port"`
	RedisPassword string `toml:"redis_password"`
	RedisDB       int    `toml:"redis_db"`

	Interval int `toml:"interval"`
	Min      int `toml:"min"`

	CollectdPlugin string `toml:"collectd_plugin"`
	CollectdType   string `toml:"collectd_type"`
}

func main() {
	var config Config
	_, err := toml.DecodeFile("/etc/barometer-localagent/config.toml", &config)
	if err != nil {
		log.Fatalf("read error of config: %s", err)
	}

	annoConfig := config.Common
	log.Printf("annotate redis config Addr:%s:%s DB:%d", annoConfig.RedisHost, annoConfig.RedisPort, annoConfig.RedisDB)
	if annoConfig.RedisPassword == "" {
		log.Printf("annotate redis password is not set")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     annoConfig.RedisHost + ":" + annoConfig.RedisPort,
		Password: annoConfig.RedisPassword,
		DB:       annoConfig.RedisDB,
	})
	infoPool := annotate.RedisPool{Client: client}

	thresConfig := config.Threshold

	log.Printf("raw data redis config Addr:%s:%s DB:%d", thresConfig.RedisHost, thresConfig.RedisPort, thresConfig.RedisDB)
	if thresConfig.RedisPassword == "" {
		log.Printf("raw data redis password is not set")
	}
	rawStore := redis.NewClient(&redis.Options{
		Addr:     thresConfig.RedisHost + ":" + thresConfig.RedisPort,
		Password: thresConfig.RedisPassword,
		DB:       thresConfig.RedisDB,
	})

	notifier := notify.CollectdNotifier{
		PluginName: thresConfig.CollectdPlugin,
		TypeName:   thresConfig.CollectdType}

	//how to stop after compile...
	ticker := time.NewTicker(time.Duration(thresConfig.Interval) * time.Second)

	for range ticker.C {
		checkVirtIF(rawStore, infoPool, &config, notifier)
	}

	fmt.Println("end")
}
