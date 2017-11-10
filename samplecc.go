package main

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"samplecc/cacheclient"
)

func pushStats(st string) {
	log.Printf(st)
}

func main() {
	cacheclient.InitPackage("cacheclient/redis.json")

	cc, err := cacheclient.NewCacheClient()
	if err != nil {
		log.Printf("NewCacheClient error:%s", err.Error())
	}

	start := time.Now().UnixNano() / 1e6

	for true {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 100; i++ {
			cc.SetString(strconv.Itoa(r.Intn(1000)), strconv.Itoa(r.Intn(1000)), 0)
			cc.GetString(strconv.Itoa(r.Intn(1000)))
		}
		time.Sleep(1 * time.Second)

		if time.Now().UnixNano()/1e6-start > 10*1000 {
			start = time.Now().UnixNano() / 1e6
			pushStats(cc.GetStats())
		}
	}
}
