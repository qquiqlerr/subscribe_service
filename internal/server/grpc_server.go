package server

import (
	"context"
	"errors"
	"github.com/qquiqlerr/subpub"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "subscribe_service/internal/gen/pb/proto"
)

// PubSubServer is a gRPC server that implements the PubSub service.
type pubSubServer struct {
	log *zap.Logger
	pb.UnimplementedPubSubServer
	pubSub subpub.SubPub
}

func (G pubSubServer) Subscribe(request *pb.SubscribeRequest, g grpc.ServerStreamingServer[pb.Event]) error {
	const op = "server.pubSubServer.Subscribe"
	_log := G.log.With(zap.String("op", op))

	_log.Debug("subscribe request", zap.Any("request", request))
	key := request.GetKey()
	if key == "" {
		_log.Warn("request for subscribe without key")
		return status.Error(codes.InvalidArgument, "key is required")
	}

	sub, err := G.pubSub.Subscribe(key, func(msg interface{}) {
		_log.Debug("message received", zap.Any("msg", msg))
		strMsg, ok := msg.(string)
		if !ok {
			_log.Error("message is not string", zap.Any("msg", msg), zap.String("key", key))
			return
		}
		if err := g.Send(&pb.Event{Data: strMsg}); err != nil {
			_log.Error("send message error", zap.Error(err))
			return
		}
	})
	if err != nil {
		_log.Error("subscribe error", zap.Error(err))
		return status.Error(codes.Internal, "subscribe error")
	}

	// Wait for the context to be done before unsubscribing
	<-g.Context().Done()
	_log.Debug("stream closed, unsubscribing", zap.String("key", key))
	sub.Unsubscribe()
	return nil
}

func (G pubSubServer) Publish(ctx context.Context, request *pb.PublishRequest) (*emptypb.Empty, error) {
	const op = "server.pubSubServer.Publish"
	_log := G.log.With(zap.String("op", op))

	key := request.GetKey()
	if key == "" {
		_log.Warn("request for publish without key")
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	msg := request.GetData()
	if msg == "" {
		_log.Warn("request for publish without message")
		return nil, status.Error(codes.InvalidArgument, "message is required")
	}
	if err := G.pubSub.Publish(key, msg); err != nil {
		_log.Error("publish error", zap.Error(err))
		if errors.Is(err, subpub.ErrTopicNotFound) {
			return nil, status.Error(codes.NotFound, "topic not found")
		}
		return nil, status.Error(codes.Internal, "publish error")
	}

	return &emptypb.Empty{}, nil
}

func NewPubSubServer(log *zap.Logger, pubSub subpub.SubPub) pb.PubSubServer {
	return &pubSubServer{
		log:    log,
		pubSub: pubSub,
	}
}
