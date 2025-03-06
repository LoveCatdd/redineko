package redineko

import (
	"fmt"
	"time"

	"github.com/LoveCatdd/util/pkg/lib/core/ids"
	"github.com/LoveCatdd/util/pkg/lib/core/log"
	"github.com/LoveCatdd/util/pkg/lib/core/viper"
	"github.com/gomodule/redigo/redis"
)

type RediNekoDoFunc func(args ...any) (reply any, err error)
type RediNekoLoadFunc func(key any) (reply any, err error)
type RediNekoBoolFunc func(key any) (reply bool, err error)
type RediNekoTranFunc func(conn redis.Conn) (err error)

// 创建 Redis 连接池
var pool *redis.Pool

// Redis 连接配置
func RediNekoConn() {
	viper.Yaml(RediConf)

	conf := RediConf.Redis
	if !conf.Enable {
		return
	}

	pool = &redis.Pool{
		MaxIdle:     conf.MaxIdle,                                       // 最大空闲连接数
		MaxActive:   conf.MaxActive,                                     // 最大活跃连接数（0=无限制）
		IdleTimeout: time.Duration(conf.IdleTimeout) * time.Millisecond, // 空闲连接超时时间
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%v:%v", conf.Ip, conf.Port), redis.DialPassword(conf.Password))
		},
	}

	if pool == nil {
		log.Errorf("Redis pool failed to initialize at conf:%+v", conf)
		return
	}

	log.Infof("Redis pool success to initialize at conf:%+v", conf)

}

func Do(commandName string) RediNekoDoFunc {
	return func(args ...any) (reply any, err error) {
		conn := pool.Get()
		defer conn.Close()

		reply, err = conn.Do(commandName, args...)

		log.SetTraceId(ids.UUIDV1())
		log.Infof("Redis do [%v %v] at reply:%v error:%v", commandName, args, reply, err)

		return
	}
}

func Load(resultType string, commandName string) RediNekoLoadFunc {
	return func(key any) (reply any, err error) {
		conn := pool.Get()
		defer conn.Close()

		switch resultType {
		case "string":
			reply, err = redis.String(conn.Do(commandName, key))
		case "int":
			reply, err = redis.Int(conn.Do(commandName, key))
		case "int64":
			reply, err = redis.Int64(conn.Do(commandName, key))
		case "float64":
			reply, err = redis.Float64(conn.Do(commandName, key))
		case "bool":
			reply, err = redis.Bool(conn.Do(commandName, key))
		case "bytes":
			reply, err = redis.Bytes(conn.Do(commandName, key))
		case "stringMap":
			reply, err = redis.StringMap(conn.Do(commandName, key))
		case "stringSlice":
			reply, err = redis.Strings(conn.Do(commandName, key))
		case "intSlice":
			reply, err = redis.Ints(conn.Do(commandName, key))
		case "int64Slice":
			reply, err = redis.Int64s(conn.Do(commandName, key))
		case "float64Slice":
			reply, err = redis.Float64s(conn.Do(commandName, key))
		case "bytesSlice":
			reply, err = redis.ByteSlices(conn.Do(commandName, key))
		case "stringSliceMap":
			reply, err = redis.StringMap(conn.Do(commandName, key))
		case "bytesSliceSlice":
			reply, err = redis.ByteSlices(conn.Do(commandName, key))
		case "uint64":
			reply, err = redis.Uint64(conn.Do(commandName, key))
		case "intMap":
			reply, err = redis.IntMap(conn.Do(commandName, key))
		case "int64Map":
			reply, err = redis.Int64Map(conn.Do(commandName, key))
		case "float64Map":
			reply, err = redis.Float64Map(conn.Do(commandName, key))
		case "uint64Slice":
			reply, err = redis.Uint64s(conn.Do(commandName, key))
		case "uint64Map":
			reply, err = redis.Uint64Map(conn.Do(commandName, key))
		default:
			reply, err = conn.Do(commandName, key)
		}

		log.SetTraceId(ids.UUIDV1())
		log.Infof("Redis Load [%v %v %v] at reply:%v error:%v", resultType, commandName, key, reply, err)
		return
	}
}

func Exists() RediNekoBoolFunc {
	return func(key any) (reply bool, err error) {
		conn := pool.Get()
		defer conn.Close()

		exists, err := redis.Int(conn.Do("EXISTS", key))

		log.SetTraceId(ids.UUIDV1())
		log.Infof("Redis Load [EXISTS %v] at exists:%v error:%v", key, exists, err)

		return exists == 1, err
	}
}

// 事务
func RediNekoTran(do ...RediNekoTranFunc) (err error) {
	conn := pool.Get()
	defer conn.Close()

	// 开启事务
	conn.Send("MULTI")

	for _, doFunc := range do {
		if err := doFunc(conn); err != nil {
			conn.Do("DISCARD") // 回滚
			return err
		}
	}
	_, err = conn.Do("EXEC")
	return
}
