//go:build k8s

package config

// k8s环境
var Config = config{
	DCfg: DBCfg{
		Addr: "webook-mysql",
		Port: "3308",
		Pass: "root",
	},
	RCg: RedisCfg{
		Addr: "webook-redis",
		Port: "6380",
	},
	LoginCheckType: CheckSession,
	ServerPort:     "8081",
	ValidTime:      1,
}
