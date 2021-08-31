package cron

import (
	"context"
	"log"
)

type usecaseProvider interface {
	RefreshProductInformation(ctx context.Context) error
}
type cron struct {
	usecase usecaseProvider
}

func New(usecase usecaseProvider) *cron {
	return &cron{
		usecase: usecase,
	}
}

func (c *cron) CronRefreshProductInformation() error {
	err := c.usecase.RefreshProductInformation(context.Background())
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("cron finished")
	return nil
}
