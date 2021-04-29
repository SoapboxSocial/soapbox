package grpc

import (
	"context"
	"errors"

	"github.com/soapboxsocial/soapbox/pkg/notifications/pb"
)

type Service struct {
	pb.UnimplementedNotificationServiceServer
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SendNotification(_ context.Context, request *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	notification := request.GetNotification()
	if notification == nil {
		return nil, errors.New("empty notification")
	}

	return &pb.SendNotificationResponse{Success: true}, nil
}
