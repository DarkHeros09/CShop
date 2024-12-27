package worker

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/util"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendResetPassword = "task:send_reset_password"

type PayloadSendResetPassword struct {
	Email string `json:"email"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendResetPassword(
	ctx context.Context,
	payload *PayloadSendResetPassword,
	opts ...asynq.Option,
) error {
	jsonPayload, err := sonic.ConfigFastest.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendResetPassword, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendResetPassword(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload PayloadSendResetPassword
	if err := sonic.ConfigFastest.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	resetPassword, err := processor.store.CreateResetPassword(ctx, db.CreateResetPasswordParams{
		UserID:     user.ID,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	subject := "Welcome to Classic Shop"
	// TODO: replace this URL with an environment variable that points to a front-end page
	verifyUrl := fmt.Sprintf("http://%s/api/v1/reset_password?email_id=%d&secret_code=%s", processor.config.ServerAddress,
		resetPassword.ID, resetPassword.SecretCode)
	content := fmt.Sprintf(`Dear %s,<br/>
	We received a request to reset your password. To proceed, please click the link below:<br/>
	<a href="%s">Reset Password</a><br/>
	`, user.Username, verifyUrl)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send reset password: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
