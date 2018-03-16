package cacheclient

import (
	"encoding/json"
	"log"
	"sync/atomic"
	"time"
	"errors"
	"github.com/go-redis/redis"
)

// CacheClient ...
type CacheClient struct {
	ring  *redis.Ring
	stats struct {
		hits      uint64
		misses    uint64
		request   int64
		elapse    int64
		timeStart int64
	}
}

// InitPackage init all handling about package
func InitPackage(confPath string) {
	initConf(confPath)
}

// NewCacheClient create new cache client for caller
func NewCacheClient() (*CacheClient, error) {
	cc := &CacheClient{}

	addrs := make(map[string]string)
	parseStringsToMap(conf.Addrs, addrs)

	cc.ring = redis.NewRing(&redis.RingOptions{
		Addrs:              addrs,
		HeartbeatFrequency: conf.HeartbeatFrequency * time.Second,
		DB:                 conf.DB,
		Password:           conf.Password,
		MaxRetries:         conf.MaxRetries,

		DialTimeout:  conf.ConnTimeout.DialTimeout * time.Second,
		ReadTimeout:  conf.ConnTimeout.ReadTimeout * time.Second,
		WriteTimeout: conf.ConnTimeout.WriteTimeout * time.Second,

		PoolSize:           conf.Pool.PoolSize,
		PoolTimeout:        conf.Pool.PoolTimeout * time.Second,
		IdleTimeout:        conf.Pool.IdleTimeout * time.Second,
		IdleCheckFrequency: conf.Pool.IdleCheckFrequency * time.Second,
	})

	cc.stats.timeStart = time.Now().UnixNano()

	return cc, nil
}

// Get get string from cache
func (cc *CacheClient) Get(key string) *redis.StringCmd {
	start := time.Now().UnixNano()
	b := cc.ring.Get(key)
	atomic.AddInt64(&cc.stats.elapse, time.Now().UnixNano()-start)
	if b.Err() != nil {
		atomic.AddUint64(&cc.stats.misses, 1)
	} else {
		atomic.AddUint64(&cc.stats.hits, 1)
	}
	atomic.AddInt64(&cc.stats.request, 1)
	return b
}

// Set set string to cache
func (cc *CacheClient) Set(key string, value interface{}, expire int) error {
	start := time.Now().UnixNano()
	err := cc.ring.Set(key, value, time.Duration(expire)).Err()
	if err != nil {
		log.Printf("cache: Set key=%q failed: %s", key, err)
	}
	atomic.AddInt64(&cc.stats.elapse, time.Now().UnixNano()-start)
	atomic.AddInt64(&cc.stats.request, 1)
	return err
}

// GetString get string from cache
func (cc *CacheClient) GetString(key string) (string, error) {
	return cc.Get(key).Result()
}

// SetString set string to cache
func (cc *CacheClient) SetString(key string, value string, expire int) error {
	err := cc.Set(key, value, expire)
	return err
}

// GetObject get object from cache
func (cc *CacheClient) GetObject(key string, object interface{}) error {
	b, err := cc.Get(key).Bytes()
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, object); err != nil {
		log.Printf("cache: key=%q Unmarshal(%T) failed: %s", key, object, err)
		return err
	}

	return nil
}

// SetObject set object to cache, object
func (cc *CacheClient) SetObject(key string, object interface{}, expire int) error {
	b, err := json.Marshal(object)
	if err != nil {
		log.Printf("cache: Marshal key=%q failed: %s", key, err)
		return err
	}

	err = cc.Set(key, b, expire)
	return err
}

// Del by key
func (cc *CacheClient) Del(key string) (int64, error) {
	start := time.Now().UnixNano()
	result, err := cc.ring.Del(key).Result()
	atomic.AddInt64(&cc.stats.request, 1)
	atomic.AddInt64(&cc.stats.elapse, time.Now().UnixNano()-start)
	if err != nil {
		log.Printf("cache: Del key=%q failed: %s", key, err)
		return 0, err
	}
	return result, nil
}

// Test key is exist
func (cc *CacheClient) exists(key string) (int64, error) {
	result, err := cc.ring.Exists(key).Result()
	if err != nil {
		log.Printf("cache: Del key=%q failed: %s", key, err)
		return result, err
	}
	return result, nil
}

// Batch processing
// Gets get strings from cache
func (cc *CacheClient) Gets(keys []string) (map[string]*redis.StringCmd, error) {
	if len(keys) <= 0 {
		return nil, errors.New("keys is empty")
	}
	pipe := cc.ring.Pipeline()
	pipelineCmds := make(map[string]*redis.StringCmd)
	for _, key := range keys {
		pipelineCmds[key] = pipe.Get(key)
	}
	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	res := make(map[string]*redis.StringCmd)
	for key, pcmd := range pipelineCmds {
		res[key] = pcmd
	}

	return res, nil
}

// Sets set strings to cache
func (cc *CacheClient) Sets(kvs map[string]interface{},expire int) (error) {
	if len(kvs) <= 0 {
		return errors.New("kvs is empty")
	}
	pipe := cc.ring.Pipeline()
	pipelineCmds := make([]*redis.StatusCmd, 0, len(kvs))
	for key,value  := range kvs {
		pipelineCmds = append(pipelineCmds, pipe.Set(key,value,time.Duration(expire)))
	}
	_, err := pipe.Exec()
	if err != nil {
		return err
	}

	return nil
}

// GetStrings get strings from cache
func (cc *CacheClient) GetStrings(keys []string) (map[string]string, error) {
	kvs, err := cc.Gets(keys)
	if(err != nil){
		return nil,err
	}
	rets := make(map[string]string)
	for key,value := range kvs{
		value_real,err_real := value.Result();
		if (err_real != nil){
			return rets,errors.New("error:" + err_real.Error() + "; key:" + key)
		}
		rets[key] = value_real;
	}

	return rets,nil
}

// SetStrings set strings to cache
func (cc *CacheClient) SetStrings(kvs map[string]string, expire int) error {
	kvs_real := make(map[string]interface{})
	for key,value := range kvs{
		kvs_real[key] = value;
	}
	err := cc.Sets(kvs_real, expire)
	return err
}

// GetObjects get objects from cache
func (cc *CacheClient) GetObjects(keys []string,value_type interface{})(map[string]interface{},error) {
	if len(keys) <= 0 {
		return nil,errors.New("keys is empty")
	}
	kvs_real,err := cc.Gets(keys);
	if (err != nil){
		return nil,err
	}
	kvs := make(map[string]interface{})
	for key,value :=range kvs_real{
		value_byte,err_real := value.Bytes()
		if (err_real != nil){
			return kvs,err_real
		}
		err_real = json.Unmarshal(value_byte,&value_type)
		kvs[key] = value_type
		if (err_real != nil){
			log.Printf("cache: Unmarshal key=%q failed: %s", key, err_real)
			return kvs,err_real
		}
	}
	return kvs,nil
}

// SetObjects set objects to cache, object
func (cc *CacheClient) SetObjects(kvs map[string]interface{}, expire int) error {
	for key,value := range kvs{
		value_real, err := json.Marshal(value)
		if err != nil {
			log.Printf("cache: Marshal key=%q failed: %s", key, err)
			return err
		}
		kvs[key] = value_real
	}
	err := cc.Sets(kvs, expire)
	return err
}

// Stats info
type Stats struct {
	StartTime int64
	EndTime   int64
	HitRatio  float64
	Rt        float64
	QPS       int
}

// GetStats return stats info
func (cc *CacheClient) GetStats() string {
	st := &Stats{}

	hits := atomic.LoadUint64(&cc.stats.hits)
	misses := atomic.LoadUint64(&cc.stats.misses)
	request := atomic.LoadInt64(&cc.stats.request)
	elapse := atomic.LoadInt64(&cc.stats.elapse)
	st.StartTime = atomic.LoadInt64(&cc.stats.timeStart)
	st.EndTime = time.Now().UnixNano()
	// ms
	interval := (st.EndTime - st.StartTime) / 1e6

	if hits == 0 && misses == 0 {
		b, _ := json.Marshal(st)
		return string(b[:])
	}
	if interval == 0 {
		b, _ := json.Marshal(st)
		return string(b[:])
	}

	st.HitRatio = float64(hits * 100.0 / (hits + misses))
	st.QPS = int(request * 1000 / interval)
	// ms
	st.Rt = float64(elapse / request / 1e6)

	atomic.AddInt64(&cc.stats.request, -request)
	atomic.AddInt64(&cc.stats.elapse, -elapse)
	atomic.AddInt64(&cc.stats.timeStart, -st.StartTime+time.Now().UnixNano())

	b, _ := json.Marshal(st)
	return string(b[:])
}
