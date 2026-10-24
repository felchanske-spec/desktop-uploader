package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	googleGrpc "google.golang.org/grpc"

	"desktop-uploader/api/pb"
	"desktop-uploader/internal/config"
	"desktop-uploader/internal/queue"
	"desktop-uploader/internal/telegram"
	transportGrpc "desktop-uploader/internal/transport/grpc"
	"desktop-uploader/internal/usecase"
)

func main() {
	log.Println("[Main] Initializing desktop-uploader backend...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Main] Configuration initialization failed: %v", err)
	}
	log.Println("[Main] Configuration successfully loaded.")

	tgClient, err := telegram.NewClient(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("[Main] Telegram client initialization failed: %v", err)
	}
	log.Println("[Main] Telegram Bot connection successfully authorized.")

	taskQueue := queue.NewTaskChannelQueue(100, tgClient, cfg.TelegramChatID)

	shutdownCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	var wg sync.WaitGroup
	workerCount := 3
	taskQueue.StartWorkers(shutdownCtx, workerCount, &wg)

	uploadInteractor := usecase.NewUploadInteractor(taskQueue)
	grpcHandler := transportGrpc.NewUploadGRPCServer(uploadInteractor)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("[Main] Failed listen in port 50051: %v", err)
	}
	log.Println("[Main] Listener successfully bound on port: 50051")

	grpcServer := googleGrpc.NewServer()
	pb.RegisterUploadServer(grpcServer, grpcHandler)

	go func() {
		log.Println("[Main] gRPC server successfully running...")
		if err := grpcServer.Serve(listener); err != nil && err != googleGrpc.ErrServerStopped {
			log.Fatalf("[Main] gRPC server crashed: %v", err)
		}
	}()

	<-shutdownCtx.Done()
	log.Println("[Main] Shutdown signal captured. Initiating graceful shutdown sequence...")

	log.Println("[Main] Stopping gRPC transport server...")
	grpcServer.GracefulStop()

	log.Println("[Main] Securing and closing the internal task queue...")
	taskQueue.Close()

	log.Println("[Main] Awaiting background workers termination (Max 5s)...")
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		log.Println("[Main] All background workers terminated cleanly.")
	case <-timeoutCtx.Done():
		log.Println("[Main] Shutdown timeout reached. Forcing application exit.")
	}

	log.Println("[Main] Service desktop-uploader successfully stopped.")
}
