package media

import "github.com/hibiken/asynq"

// WithTaskOptions applies producer settings at enqueue time.
func (uc *UseCase) WithTaskOptions(options ...asynq.Option) *UseCase {
	uc.taskOptions = append([]asynq.Option(nil), options...)
	return uc
}
