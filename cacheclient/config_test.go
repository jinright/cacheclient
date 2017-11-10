package cacheclient

import (
	"fmt"
	"testing"
)

var debug = false

func Test_parseConfig(t *testing.T) {
	err := parseConf("redis.json")
	if err != nil {
		t.Error("parse config format is error", err)
	}

	if debug == true {
		fmt.Println("Test_parseConfig:", conf)
	}

	if conf.Pool.PoolSize != 4 {
		t.Error("parse config PoolSize error")
	}

	if conf.ConnTimeout.DialTimeout != 30 {
		t.Error("parse config DialTimeout error")
	}
}

func Test_parseStringsToMap_1(t *testing.T) {
	var nodes = []string{"server1:192.168.4.41:6379"}

	addrs := make(map[string]string)
	parseStringsToMap(nodes, addrs)

	if len(addrs) != 1 {
		t.Error("map len is error", len(addrs))
	}

	if addrs["server1"] != "192.168.4.41:6379" {
		t.Error("parse is error", addrs)
	}
}

func Test_parseStringsToMap_2(t *testing.T) {
	var nodes = []string{"server1:192.168.4.41:6379", "server2:192.168.4.41:6380"}

	addrs := make(map[string]string)
	parseStringsToMap(nodes, addrs)

	if debug == true {
		fmt.Println("Test_parseStringsToMap_2", addrs)
		for k, v := range addrs {
			fmt.Println(k, v)
		}
	}

	if len(addrs) != 2 {
		t.Error("map len is error", len(addrs))
	}

	if addrs["server2"] != "192.168.4.41:6380" {
		t.Error("parse is error", addrs["server2"])
	}
}
