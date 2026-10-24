package main

import (
	"context"
	"desktop-uploader/api/pb"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: uploader-service <file-path>")
		return
	}
	filePath := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Failed to connect to backend: %v", err)
	}
	defer conn.Close()

	clinet := pb.NewUploadClient(conn)

	for i := 1; i < len(os.Args); i++ {
		filePath := os.Args[i]

		_, err = clinet.UploadFile(ctx, &pb.UploadRequest{
			FilePath: filePath,
		})

		if err != nil {
			log.Printf("Failed to send file %s to backend: %v", filePath, err)
			continue
		}
		log.Printf("Successfullt routed file to backend: %s", filePath)
	}

	if err != nil {
		log.Fatalf("Failed to send file to backend pipeline: %v", err)
	}

	log.Printf("Successfullt routed fiel to backend: %s", filePath)
}
