package cacheclient

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

type object struct {
	Str string
	Num int
}

func Test_NewCacheClient(t *testing.T) {
	InitPackage("redis.json")

	cc, err := NewCacheClient()
	if err != nil {
		t.Error("NewCacheClient error:", err)
	}

	err = cc.ring.Ping().Err()
	if err != nil {
		t.Error("ping server error:", err)
	}
}

func Test_SetJson(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	var in = &object{Str: "test",
		Num: 1}

	err := cc.SetObject("object", in, 0)
	if err != nil {
		t.Error("Test_setObject err", err)
	}
}

func Test_GetJson(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	var out object
	err := cc.GetObject("object", &out)
	if err != nil || out.Str != "test" || out.Num != 1 {
		t.Error("Test_setObject  err", err)
	}
}

// server1:key2, server2:key1, server3:key4
func Test_SetString(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	err := cc.SetString("key1", "v1", 0)
	if err != nil {
		t.Error("Test_SetString", err)
	}

	err = cc.SetString("key2", "v2", 0)
	if err != nil {
		t.Error("Test_SetString", err)
	}

	err = cc.SetString("key4", "v4", 0)
	if err != nil {
		t.Error("Test_SetString", err)
	}
}

func Test_SetString_Exist(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	err := cc.SetString("key1", "v1", 0)
	if err != nil {
		t.Error("Test_SetString_Exist", err)
	}
}

// test client split function
func Test_CheckHashStore(t *testing.T) {
	// key2:server1
	client1 := redis.NewClient(&redis.Options{
		Addr:     "192.168.4.41:6379",
		Password: "", // no password set
	})
	client2 := redis.NewClient(&redis.Options{
		Addr:     "192.168.4.41:6380",
		Password: "", // no password set
	})
	client3 := redis.NewClient(&redis.Options{
		Addr:     "192.168.4.41:6381",
		Password: "", // no password set
	})

	// server1:key2, server2:key1, server3:key4
	val, _ := client1.Get("key2").Result()
	if val != "v2" {
		t.Error("Test_CheckHash", val)
	}
	val, _ = client2.Get("key1").Result()
	if val != "v1" {
		t.Error("Test_CheckHash", val)
	}
	val, _ = client3.Get("key4").Result()
	if val != "v4" {
		t.Error("Test_CheckHash", val)
	}
}

func Test_GetString(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	// key2:server1, key1:server2, key4:server3
	v, err := cc.GetString("key1")

	if err != nil || v != "v1" {
		t.Error("Test_GetString", v)
	}
}

func Test_GetString_NoExist(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	_, err := cc.GetString("key111")
	if err.Error() != "redis: nil" {
		t.Error("Test_GetString_NoExist", err)
	}
}

func Test_Exists(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	result, err := cc.exists("key1")
	if err != nil || result != 1 {
		t.Error("Test_Exists", result, err)
	}
}

func Test_Del(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	n, err := cc.Del("key1")
	if err != nil || n != 1 {
		t.Error("Test_Del", err)
	}
}

func Test_Del_NoExist(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	n, err := cc.Del("key1")
	if err != nil || n != 0 {
		t.Error("Test_Del_NoExist", err)
	}
}

// server1:key2, server2:key1, server3:key4
func Test_CheckSrvFailover(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	client2 := redis.NewClient(&redis.Options{
		Addr:     "192.168.4.41:6380",
		Password: "", // no password set
	})
	client3 := redis.NewClient(&redis.Options{
		Addr:     "192.168.4.41:6381",
		Password: "", // no password set
	})

	_, err := client2.FlushAll().Result()
	if err != nil {
		t.Error("Flush redis server2 error", err)
	}

	// shutdown server3
	err = client3.ShutdownNoSave().Err()
	if err != nil {
		t.Error("ShutdownNoSave server3 error", err)
	}
	time.Sleep(20 * time.Second)

	cc.SetString("key1", "v1", 0)
	cc.SetString("key4", "v4", 0)

	// check server2:key1
	val, _ := client2.Get("key1").Result()
	if val != "v1" {
		t.Error("key1 rehash error", val)
	}

	// rehash key4 to server2
	val, _ = client2.Get("key4").Result()
	if val != "v4" {
		t.Error("key4 rehash error", val)
	}
}

func Test_GetStats(t *testing.T) {
	InitPackage("redis.json")
	cc, _ := NewCacheClient()

	for i := 0; i < 1000; i++ {
		cc.GetString("key1")
		cc.GetString("key11")
	}
	time.Sleep(2 * time.Second)

	st := cc.GetStats()

	for i := 0; i < 1000; i++ {
		cc.GetString("key1")
		cc.GetString("key11")
	}
	time.Sleep(2 * time.Second)

	st = cc.GetStats()

	var obj Stats

	json.Unmarshal([]byte(st), &obj)
	if int(obj.HitRatio) != 50 || obj.QPS == 0 || obj.Rt == 0 {
		t.Error(st)
	}
}
