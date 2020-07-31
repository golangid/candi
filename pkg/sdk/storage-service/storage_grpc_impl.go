package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	pb "agungdwiprasetyo.com/backend-microservices/api/storage-service/proto"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var defaultStreamLimitSize = 50 * helper.MByte

type storageGRPCImpl struct {
	client          pb.StorageServiceClient
	authKey         string
	streamLimitSize int64
}

// NewStorageServiceGRPC constructor for storage service GRPC stream
func NewStorageServiceGRPC(host string, authKey string, streamLimitSize int64) (Storage, error) {
	conn, err := grpc.Dial(host,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(100*helper.MByte)), grpc.MaxCallSendMsgSize(int(100*helper.MByte))),
	)
	if err != nil {
		return nil, err
	}

	if streamLimitSize <= 0 {
		streamLimitSize = int64(defaultStreamLimitSize)
	}

	return &storageGRPCImpl{
		client:          pb.NewStorageServiceClient(conn),
		authKey:         authKey,
		streamLimitSize: streamLimitSize,
	}, nil
}

func (u *storageGRPCImpl) Upload(ctx context.Context, param *UploadParam) (res Response, err error) {
	trace := tracer.StartTrace(ctx, "StorageGRPCClient-Upload")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.Finish()
	}()
	ctx = trace.Context()

	md := metadata.Pairs("authorization", u.authKey,
		"filename", param.Filename,
		"folder", param.Folder,
		"contentType", param.ContentType,
		"size", strconv.Itoa(int(param.Size)))
	trace.InjectGRPCMetadata(md)

	ctx = metadata.NewOutgoingContext(ctx, md)
	stream, err := u.client.Upload(ctx)
	if err != nil {
		logger.LogE(err.Error())
		panic(err)
	}
	defer stream.CloseSend()

	// stream send file with grpc
	fileSize := param.Size
	buff := make([]byte, fileSize)
	param.File.Read(buff)
	var i int64
	for i = 0; i < fileSize; i += u.streamLimitSize {
		lastOffset := i + u.streamLimitSize
		if lastOffset > fileSize {
			lastOffset = fileSize
		}
		fmt.Println("stream", i, lastOffset)
		err = stream.Send(&pb.Chunk{
			Content:   buff[i:lastOffset],
			TotalSize: fileSize,
			Received:  lastOffset,
		})
		if err != nil {
			logger.LogE(err.Error())
			panic(err)
		}
	}

	status, err := stream.CloseAndRecv()
	if err != nil {
		logger.LogE(err.Error())
		panic(err)
	}

	if status.Code != pb.StatusCode_Ok {
		logger.LogE("not success")
		panic(errors.New("Status code is not success"))
	}

	res = Response{
		Location: status.File, Size: status.Size,
	}

	return
}
