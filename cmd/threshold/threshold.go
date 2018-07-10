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

const confFileAnno = "config_annotate.toml"
const confFileEval = "config_evaluate.toml"

// e.g. collectd/instance-00000001/virt/if_octets-tapd21acb51-35
const redisKey = "collectd/*/virt/if_octets-*"

func zrangebyscore(client *redis.Client, key string, pool annotate.Pool, config *ThresholdConfig, notifier notify.Notifier) {

	unixNow := int(time.Now().Unix())

	val, err := client.ZRangeByScore(key, redis.ZRangeBy{
		Min: strconv.Itoa(unixNow - config.Interval),
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
		if maxVal > config.Min {

			fmt.Println("kick action")

			item := strings.Split(key, "/")
			ifItem := strings.SplitN(item[3], "-", 2)
			virtName := item[1]
			virtIF := ifItem[1]

			var message bytes.Buffer
			message.WriteString("Value ")
			message.WriteString(strconv.Itoa(maxVal))
			message.WriteString(" exceeded threshold ")
			message.WriteString(strconv.Itoa(config.Min))
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

func checkVirtIF(client *redis.Client, pool annotate.Pool, config *ThresholdConfig, notifier notify.Notifier) {
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
	Annotate AnnotateConfig
	Evaluate EvaluateConfig
}

// AnnotateConfig is ...
type AnnotateConfig struct {
	Redis RedisConfig
}

// EvaluateConfig is ...
type EvaluateConfig struct {
	Redis     RedisConfig
	Threshold ThresholdConfig
	Collectd  CollectdConfig
}

// RedisConfig is ...
type RedisConfig struct {
	Ipaddress string
	Port      string
	Password  string
	DB        int
}

// ThresholdConfig is ...
type ThresholdConfig struct {
	Interval int
	Min      int
}

// CollectdConfig is ...
type CollectdConfig struct {
	Plugin string
	Type   string
}

func main() {
	var config Config
	_, err := toml.DecodeFile(confFileAnno, &config.Annotate)
	if err != nil {
		log.Fatalf("read error of annotate config: %s", err)
	}
	_, err2 := toml.DecodeFile(confFileEval, &config.Evaluate)
	if err2 != nil {
		log.Fatalf("read error of evaluate config: %s", err2)
	}

	redisConfig := config.Annotate.Redis
	log.Printf("annotate redis config : %s:%s XXXXX %d", redisConfig.Ipaddress, redisConfig.Port, redisConfig.DB)
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Ipaddress + ":" + redisConfig.Port,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	infoPool := annotate.RedisPool{Client: client}

	redisConfig = config.Evaluate.Redis
	log.Printf("evaluate redis config : %s:%s XXXXX %d", redisConfig.Ipaddress, redisConfig.Port, redisConfig.DB)
	rawStore := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Ipaddress + ":" + redisConfig.Port,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	collectdConfig := config.Evaluate.Collectd
	notifier := notify.CollectdNotifier{
		PluginName: collectdConfig.Plugin,
		TypeName:   collectdConfig.Type}

	//how to stop after compile...
	ticker := time.NewTicker(time.Duration(config.Evaluate.Threshold.Interval) * time.Second)

	for range ticker.C {
		checkVirtIF(rawStore, infoPool, &config.Evaluate.Threshold, notifier)
	}

	fmt.Println("end")
}
