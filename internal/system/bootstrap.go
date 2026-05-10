package system

import (
	"log"

	"github.com/robfig/cron/v3"
)

type Bootstrap struct {
	cron *cron.Cron
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{
		cron: cron.New(),
	}
}

func (b *Bootstrap) Init() {
	log.Println("[bootstrap] init config/cache/payment/channel placeholders")
	b.cron.AddFunc("@every 1m", func() {
		log.Println("[cron] heartbeat task")
	})
	b.cron.Start()
}

func (b *Bootstrap) Stop() {
	if b.cron != nil {
		ctx := b.cron.Stop()
		<-ctx.Done()
	}
}
