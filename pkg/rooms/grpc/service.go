package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
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

func (s *Service) RegisterWelcomeRoom(ctx context.Context, in *pb.WelcomeRoomRegisterRequest, opts ...grpc.CallOption) (*pb.WelcomeRoomRegisterResponse, error) {
	return nil, nil
}

