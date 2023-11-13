package logger

import "sync"

var gl LoggerV1
var lMutex sync.RWMutex

func SetGlobalLogger(l LoggerV1) {
	lMutex.Lock()
	defer lMutex.Unlock()
	gl = l
}

func L() LoggerV1 {
	lMutex.RLock()
	g := gl
	lMutex.RUnlock()
	return g
}

var GL LoggerV1 = &NopLogger{}
