package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	firebase "firebase.google.com/go"
	"github.com/cshop/v3/api"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/image"
	"github.com/cshop/v3/mail"
	"github.com/cshop/v3/util"
	"github.com/cshop/v3/worker"
	"github.com/hibiken/asynq"
	"github.com/imagekit-developer/imagekit-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {

	config, err := util.LoadVault() // we use . because app.env is on the same level with main.go
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	conn, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	opt := option.WithCredentialsFile("serviceAccountKey.json")

	fb, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Fatal("error initializing firebase:", err)
	}

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	ik := image.NewImageKit(imagekit.NewParams{
		PrivateKey:  config.ImageKitPrivateKey,
		PublicKey:   config.ImageKitPublicKey,
		UrlEndpoint: config.ImageKitUrlEndPoint,
	})

	sender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, *config, redisOpt, store)
	runFiberServer(*config, store, fb, taskDistributor, ik, sender)

	err = waitGroup.Wait()
	if err != nil {
		// log.Fatal().Err(err).Msg("error from wait group")
	}

}

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer, config)

	// log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		// log.Fatal().Err(err).Msg("failed to start task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		// log.Info().Msg("graceful shutdown task processor")

		taskProcessor.Shutdown()
		// log.Info().Msg("task processor is stopped")

		return nil
	})
}

func runFiberServer(
	// ctx context.Context,
	// waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	fb *firebase.App,
	taskDistributor worker.TaskDistributor,
	ik image.ImageKitManagement,
	sender mail.EmailSender,
) {
	server, err := api.NewServer(config, store, fb, taskDistributor, ik, sender)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

	// waitGroup.Go(func() error {
	// 	<-ctx.Done()
	// 	// log.Info().Msg("graceful shutdown gRPC server")

	// 	server.GracefulShutdown()
	// 	// log.Info().Msg("gRPC server is stopped")

	// 	return nil
	// })
}
