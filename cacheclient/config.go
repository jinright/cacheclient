package cacheclient

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"
)

type config struct {
	HashType           string
	Addrs              []string
	HeartbeatFrequency time.Duration
	DB                 int
	Password           string
	MaxRetries         int
	Stats              struct {
		Interval time.Duration
	}
	ConnTimeout struct {
		DialTimeout  time.Duration
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}
	Pool struct {
		PoolSize           int
		PoolTimeout        time.Duration
		IdleTimeout        time.Duration
		IdleCheckFrequency time.Duration
	}
}

var conf config

const defaultConfPath = "/etc/putong/redis/redis.json"

// parse conf to struct Config
func parseConf(path string) error {
	var cPath string

	if path == "" {
		cPath = defaultConfPath
	} else {
		cPath = path
	}
	data, err := ioutil.ReadFile(cPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		return err
	}
	return nil
}

func parseStringsToMap(nodes []string, addr map[string]string) {
	for k := range nodes {
		sl := strings.SplitN(nodes[k], ":", 2)
		addr[sl[0]] = sl[1]
	}
}

func initConf(confPath string) error {
	return parseConf(confPath)
}
