package grpc

import (
	"context"
	"errors"

	"github.com/soapboxsocial/soapbox/pkg/notifications/pb"
	"github.com/soapboxsocial/soapbox/pkg/notifications/worker"
)

type Service struct {
	pb.UnimplementedNotificationServiceServer

	dispatch *worker.Dispatcher
}

func NewService(dispatch *worker.Dispatcher) *Service {
	return &Service{
		dispatch: dispatch,
	}
}

func (s *Service) SendNotification(_ context.Context, request *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	notification := request.GetNotification()
	if notification == nil {
		return nil, errors.New("empty notification")
	}

	return &pb.SendNotificationResponse{Success: true}, nil
}
