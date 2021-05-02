package grpc

import (
	"context"
	"errors"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/pb"
	"github.com/soapboxsocial/soapbox/pkg/notifications/worker"
)

type Service struct {
	pb.UnimplementedNotificationServiceServer

	dispatch *worker.Dispatcher
	settings *notifications.Settings
}

func NewService(dispatch *worker.Dispatcher, settings *notifications.Settings) *Service {
	return &Service{
		dispatch: dispatch,
		settings: settings,
	}
}

func (s *Service) SendNotification(_ context.Context, request *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	notification := request.Notification
	if notification == nil {
		return nil, errors.New("empty notification")
	}

	push := notification.ToPushNotification()
	ids := request.Targets

	targets, err := s.settings.GetSettingsForUsers(ids)
	if err != nil {
		return nil, errors.New("failed to get targets")
	}

	s.dispatch.Dispatch(targets, push)

	return &pb.SendNotificationResponse{Success: true}, nil
}
