package limit

import (
	"fmt"
	"io/ioutil"
	// "reflect"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/time/rate"
	"testing"
)

func TestNewRateLimiter(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	rl := NewRateLimiter(rate.Every(1*time.Second), 1, 100)
	rl.Start()
	collector := [][]byte{}
	ticker := time.NewTicker(1 * time.Millisecond)
	count := 5500

	go func() {
		for {
			data := make([]byte, 4096)
			n, err := rl.Read(data)
			if err != nil {
				t.Fatalf("Error reading from pipe: %s", err.Error())
			}
			collector = append(collector, data[:n])
		}
	}()

	for i := 0; i < count; i++ {
		data := []byte("0123456789abcdef")

		_, err := rl.Write(data)
		if err != nil {
			t.Fatalf("Error writing to pipe: %s", err.Error())
		}
		<-ticker.C
	}

	for k, v := range rl.GetMetrics() {
		fmt.Printf("%s: %d\n", k, v)
	}

	for _, row := range collector {
		fmt.Println(string(row))
	}
}
