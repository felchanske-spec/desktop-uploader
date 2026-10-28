package grpc

import (
	"desktop-uploader/api/pb"
	"desktop-uploader/internal/usecase"
	"fmt"
)

type UploadHandler struct {
	pb.UnimplementedUploadServer
	usecase *usecase.UploadInteractor
}

func NewUploadGRPCServer(usecase *usecase.UploadInteractor) *UploadHandler {
	return &UploadHandler{
		usecase: usecase,
	}
}

func (h *UploadHandler) UploadFile(req *pb.UploadRequest, stream pb.Upload_UploadFileServer) error {
	ctx := stream.Context()

	err := h.usecase.ValidateAndQueue(ctx, req.GetFilePath(), req.GetTarget())
	if err != nil {
		failResponse := &pb.UploadResponse{
			State:   pb.UploadResponse_FAILED,
			Message: fmt.Sprintf("Validation failed: %v", err),
		}
		_ = stream.Send(failResponse)
		return err
	}

	successResponse := &pb.UploadResponse{
		State:           pb.UploadResponse_QUEUED,
		ProgressPercent: 0,
		Message:         "File successfully validated and queued for upload",
	}

	if err := stream.Send(successResponse); err != nil {
		return fmt.Errorf("failed t osend grpc stream update: %w", err)
	}
	return nil
}
