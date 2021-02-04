package grpc

import (
	"context"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Service struct {
	repository *rooms.Repository
}

func NewService(repository *rooms.Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetRoom(ctx context.Context, in *pb.RoomQuery) (*pb.RoomState, error) {
	r, err := s.repository.Get(in.Id)
	if err != nil {
		return nil, err
	}

	return r.ToProto(), nil
}

func (s *Service) RegisterWelcomeRoom(ctx context.Context, in *pb.WelcomeRoomRegisterRequest) (*pb.WelcomeRoomRegisterResponse, error) {

	id := internal.GenerateRoomID()

	return &pb.WelcomeRoomRegisterResponse{Id: id}, nil
}

