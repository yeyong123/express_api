/*
File Name: redis.go
Created Date: 2022-05-24 10:58:25
Author: yeyong
Last modified: 2022-05-24 10:58:43
*/
package service

import(
    "context"
    "log"
	"senkoo.cn/config"
    "github.com/go-redis/redis/v8"
)
var cache = config.CacheSetting
var Ctx = context.Background()
var Redis = &redis.Client{}
func init() {
    log.Println("连接Redis数据库...")
    Redis = redis.NewClient(&redis.Options{
        Addr: cache.RedisAddr,
        Password: cache.RedisPass,
        DB: cache.RedisDB,
    })
    log.Println("测试连接,发送[PING]")
    pong, err := Redis.Ping(Ctx).Result()
    if err != nil {
        log.Printf("Redis数据库连接失败%v", err)
        log.Fatal(err)
    }
    log.Printf("缓存Redis数据库连接成功, 回复[%s]\n", pong)
}
