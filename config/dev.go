//go:build dev

package config

// 生产环境
var Config = config{
	DCfg: DBCfg{
		Addr: "localhost",
		Port: "3306",
		Pass: "your_password",
	},
	RCg: RedisCfg{
		Addr: "localhost",
		Port: "6379",
	},
	LoginCheckType: JWT,
	ServerPort:     "8080",
	ValidTime:      1,
	GormDebug:      false,
}
