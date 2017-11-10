# 基于redis的cache cluster系统
## 系统架构
基于Client的Key一致性哈希分片

## SDK使用说明
### 使用流程
* 初始化package:func InitPackage(confPath string)
* new出来一个CacheClient:func NewCacheClient() (*CacheClient, error) 
* 通过new出来的CacheClient进行对CacheCluster的访问
* 定期通过GetStats()接口获取这个CacheClient的统计

备注：
* 建议一个APP使用一个CacheClient
* CacheClient线程安全
* 对于热点数据，应用需要在设计时精心考虑key的设计
* Cache Server不提供持久化功能
* 业务需要自己处理Cache引入后引起的逻辑变化（如：需要先读取Cache的内容，如果miss则回DB读取数据）

### API
* func InitPackage(confPath string)
* func NewCacheClient() (*CacheClient, error) 
* func (cc *CacheClient) Set(key string, value interface{}, expire int) error
* func (cc *CacheClient) Get(key string) *redis.StringCmd
* func (cc *CacheClient) GetString(key string) (string, error)
* func (cc *CacheClient) SetString(key string, value string, expire int) error
* func (cc *CacheClient) GetObject(key string, object interface{}) error
* func (cc *CacheClient) SetObject(key string, object interface{}, expire int) error
* func (cc *CacheClient) Del(key string) (int64, error)
* func (cc *CacheClient) GetStats() string

## redis部署说明
### IDC内部
redis cluster的机器最好部署在同一个机架

### 集群case:机器
3.redis-r630.bjs.p1staff.com 10.191.161.140
4.redis-r630.bjs.p1staff.com 10.191.161.141

### 配置
* 单实例内存大小16G内存
* 一台机器部署6个实例(6379-6384)
* 配置文件：disable持久化（/etc/redis/6379.conf  ...6384.conf）
* 启动服务文件：/usr/lib/systemd/system/redis.6379.service（...6384.service）
* redis version: 3.2.10 (yum install)

## Cache策略
### 应用自定义策略
应用根据自己的数据特性，确定Cache的过期时间。

### 统一的基础策略
redis cluster会设置基础的LRU策略。

## 监控系统
### 实例监控
* 进程监控
    - 同一有Systemd进行管理，Ops统一Owner
    - 系统异常退出后，服务自动重启
* 实例监控
    - Prometheus的通用模块单独监控这个机器上不同的实例
    
### Cluster监控
* ClusterInfo
    - 通过在一个page中显示cluster所有实例的方式进行监控

### 应用监控
各个应用方通过调用API:GetStats()得到命中率信息，返回JSON数据：
`{"StartTime":1506500144361743000,"EndTime":1506500156389818000,"HitRatio":100,"Rt":25,"QPS":33}`

## 可运维性支持
* 进程管理
    - 进程统一由systemd管理，宕机、进程重启等。
* 配置管理
    - 每个实例对应各自的配置文件，不同的端口区分
* 系统扩容
    - 系统扩容时，需要把新扩容的实例地址信息同步至应用的配置文件（redis.json），再重启应用服务



