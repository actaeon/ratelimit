package limit

import (
	"io/ioutil"
	"reflect"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/time/rate"
	"testing"
)

func TestNewRateLimiter(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	out, in := NewRateLimiter(rate.Every(1*time.Second), 1, 100)
	collector := [][]byte{}
	ticker := time.NewTicker(10 * time.Millisecond)
	count := 100

	expected := [][]uint8{
		[]uint8{
			0x30, 0x31, 0x32, 0x33,
			0x34, 0x35, 0x36, 0x37,
			0x38, 0x39, 0x61, 0x62,
			0x63, 0x64, 0x65, 0x66,
		},
	}

	go func() {
		for {
			data := make([]byte, 4096)
			n, err := out.Read(data)
			if err != nil {
				t.Fatalf("Error reading from pipe: %s", err.Error())
			}
			collector = append(collector, data[:n])
		}
	}()

	for i := 0; i < count; i++ {
		data := []byte("0123456789abcdef")

		_, err := in.Write(data)
		if err != nil {
			t.Fatalf("Error writing to pipe: %s", err.Error())
		}
		<-ticker.C
	}

	if !reflect.DeepEqual(collector, expected) {
		t.Fail()
	}
}
