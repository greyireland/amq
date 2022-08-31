# amq

simple async mq

## example

```go
package amq

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/greyireland/log"
	"testing"
	"time"
)

type Msg struct {
	Tp   string `json:"tp"`
	Data string `json:"data"`
}

func TestNewSender(t *testing.T) {
	r := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	queue := "test"

	s := NewSender(r, queue)
	go func() {
		for i := 0; i < 100000; i++ {

			s.Send(Msg{
				Tp:   "1",
				Data: fmt.Sprintf("hello %d", i),
			})
		}
	}()
	re := NewReceiver(r, queue)
	go func() {
		for {
			var data Msg
			err := re.Receive(&data)
			if err != nil {
				log.Warn("dequeue error", "err", err)
			}
			fmt.Println("data", data)
		}
	}()
	time.Sleep(10 * time.Second)
}

```