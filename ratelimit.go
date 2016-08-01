package limit

import (
	"io"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

func NewRateLimiter(limit rate.Limit, burst, bufSize int) (io.Reader, io.Writer) {

	inRd, inWr := io.Pipe()
	outRd, outWr := io.Pipe()
	limiter := rate.NewLimiter(rate.Limit(limit), burst)

	outChan := make(chan []byte)

	go func() {
		for {
			data := make([]byte, bufSize)
			var n int
			var err error
			if n, err = inRd.Read(data); err != nil {
				log.Errorf("Error Reading from input pipe: %s", err.Error())
				return
			}
			log.Debugf("RateLimiter read %d bytes", n)
			select {
			case outChan <- data[:n]:
				log.Debugf("RateLimiter wrote %d bytes", len(data))
			default:
				log.Warnf("No token for message, dropping %d bytes", len(data))
			}
		}
	}()

	go func() {
		for {
			if err := limiter.Wait(context.TODO()); err != nil {
				log.Errorf("Error in token bucket: %s", err.Error())
				return
			}
			data := <-outChan
			log.Debugf("On Token, read %d bytes from channel", len(data))
			_, err := outWr.Write(data)
			log.Debugf("On Token, wrote %d bytes to output pipe", len(data))
			if err != nil {
				log.Errorf("Error writing to output pipe: %s", err.Error())
				return
			}
		}
	}()

	return outRd, inWr
}
