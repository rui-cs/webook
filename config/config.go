package config

type config struct {
	RCg            RedisCfg
	DCfg           DBCfg
	LoginCheckType int
	ServerPort     string
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
