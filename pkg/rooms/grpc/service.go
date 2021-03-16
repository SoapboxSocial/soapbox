package grpc

import (
	"context"
	"errors"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Service struct {
	pb.UnsafeRoomServiceServer

	repository *rooms.Repository
	ws         *rooms.WelcomeStore
}

func NewService(repository *rooms.Repository, ws *rooms.WelcomeStore) *Service {
	return &Service{
		repository: repository,
		ws:         ws,
	}
}

func (s *Service) GetRoom(ctx context.Context, request *pb.GetRoomRequest) (*pb.GetRoomResponse, error) {
	r, err := s.repository.Get(request.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetRoomResponse{State: r.ToProto()}, nil
}

func (s *Service) RegisterWelcomeRoom(ctx context.Context, request *pb.RegisterWelcomeRoomRequest) (*pb.RegisterWelcomeRoomResponse, error) {
	id := internal.GenerateRoomID()

	if request == nil || request.UserId == 0 {
		return nil, errors.New("no message")
	}

	err := s.ws.StoreWelcomeRoomID(id, request.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterWelcomeRoomResponse{Id: id}, nil
}
