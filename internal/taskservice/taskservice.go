package taskservice

import (
	"context"
	"time"
)

type TaskService interface {
	Connect(ctx *context.Context)
	ScheduleMessage(ctx *context.Context, scheduledMessage ScheduledMessage, timeOffset time.Time)
	ScheduleTransaction(ctx *context.Context, scheduledTransaction ScheduledTransaction, timeOffset time.Time)
	Close()
}
