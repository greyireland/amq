package amq

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/greyireland/log"
	"time"
)

type Sender struct {
	redis *redis.Client
	ch    chan string
	queue string
}

func NewSender(r *redis.Client, queue string) *Sender {
	s := &Sender{
		redis: r,
		ch:    make(chan string, 10000),
		queue: queue,
	}
	go s.proc()
	return s
}
func (a *Sender) proc() {
	for {
		select {
		case data := <-a.ch:
			err := a.redis.LPush(context.TODO(), a.queue, data).Err()
			if err != nil {
				log.Warn("LPush error", "err", err)
			}

		}
	}
}
func (a *Sender) Send(v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case a.ch <- string(buf):
		return nil
	case <-time.After(time.Millisecond * 100):
		return errors.New("enqueue timeout")
	}
}

type Receiver struct {
	redis *redis.Client
	ch    chan string
	queue string
}

func NewReceiver(r *redis.Client, queue string) *Receiver {
	re := &Receiver{
		redis: r,
		ch:    make(chan string, 10000),
		queue: queue,
	}
	go re.proc()
	return re
}
func (a *Receiver) proc() {
	for {
		data, err := a.redis.RPop(context.TODO(), a.queue).Result()
		if err != nil || len(data) == 0 {
			log.Warn("RPop error", "err", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}
		select {
		case a.ch <- data:
		case <-time.After(time.Millisecond * 100):
			log.Warn("drop data", "data", data)
		}
	}
}
func (a *Receiver) Receive(v interface{}) error {
	data := <-a.ch
	err := json.Unmarshal([]byte(data), &v)
	if err != nil {
		return err
	}
	return nil
}
