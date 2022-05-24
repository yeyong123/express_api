module senkoo.cn

go 1.16

replace (
	senkoo.cn/config => ./config
	senkoo.cn/express => ./express
	senkoo.cn/service => ./service
)

require (
	github.com/go-redis/redis/v8 v8.11.5
	gopkg.in/ini.v1 v1.66.4 // indirect
	senkoo.cn/config v0.0.0-00010101000000-000000000000
	senkoo.cn/express v0.0.0-00010101000000-000000000000
)
