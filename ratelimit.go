package limit

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Start()
	GetMetrics() map[string]uint
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

type rateLimiter struct {
	droppedBytes    uint
	droppedMessages uint
	successBytes    uint
	successMessages uint
	inChan          chan []byte
	outChan         chan []byte
	limiter         *rate.Limiter
}

func (rl *rateLimiter) GetMetrics() map[string]uint {
	return map[string]uint{
		"droppedBytes":    rl.droppedBytes,
		"droppedMessages": rl.droppedMessages,
		"successBytes":    rl.successBytes,
		"successMessages": rl.successMessages,
	}
}

func (rl *rateLimiter) Start() {
	go func() {
		for data := range rl.inChan {
			if err := rl.limiter.Wait(context.TODO()); err != nil {
				log.Errorf("Error in token bucket: %s", err.Error())
				return
			}
			rl.outChan <- data
		}
	}()
}

func (rl *rateLimiter) Stop() {
	close(rl.inChan)
}

func (rl *rateLimiter) Read(p []byte) (n int, err error) {
	p = <-rl.outChan
	return len(p), nil
}

func (rl *rateLimiter) Write(p []byte) (int, error) {
	select {
	case rl.inChan <- p:
		rl.successMessages = rl.successMessages + 1
		rl.successBytes = rl.successBytes + uint(len(p))
	default:
		rl.droppedMessages = rl.droppedMessages + 1
		rl.droppedBytes = rl.droppedBytes + uint(len(p))
	}
	return len(p), nil
}

func NewRateLimiter(limit rate.Limit, burst, bufSize int) RateLimiter {
	rl := &rateLimiter{
		inChan:  make(chan []byte),
		outChan: make(chan []byte),
		limiter: rate.NewLimiter(rate.Limit(limit), burst),
	}
	return rl

}
