package config

type config struct {
	RCg            RedisCfg
	DCfg           DBCfg
	LoginCheckType int
	ServerPort     string

	ValidTime int // 单位：分钟
}

const (
	CheckSession = 1
	JWT          = 2
)

type RedisCfg struct {
	Addr string
	Port string
}

type DBCfg struct {
	Addr string
	Port string
	Pass string
}
