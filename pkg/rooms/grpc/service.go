package grpc

import (
	"context"
	"errors"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Service struct {
	repository *rooms.Repository
	ws         *rooms.WelcomeStore
}

func NewService(repository *rooms.Repository, ws *rooms.WelcomeStore) *Service {
	return &Service{
		repository: repository,
		ws:         ws,
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

	if in == nil || in.UserId == 0 {
		return nil, errors.New("no message")
	}

	err := s.ws.StoreWelcomeRoomID(id, in.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.WelcomeRoomRegisterResponse{Id: id}, nil
}
