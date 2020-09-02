package rooms

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type peer struct {
	stream pb.RoomService_SignalServer
	rtc   *sfu.WebRTCTransport
}

type Room struct {
	mux sync.RWMutex

	members map[int]*peer
}

func (r *Room) Handle(id int, stream pb.RoomService_SignalServer, rtc *sfu.WebRTCTransport) error {
	r.mux.Lock()
	r.members[id] = &peer{
		stream,
		rtc,
	}
	r.mux.Unlock()

	for {
		in, err := stream.Recv()
		if err != nil {
			_ = rtc.Close()

			if err == io.EOF {
				return nil
			}

			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Canceled {
				return nil
			}

			log.Printf("signal error %v %v\n", errStatus.Message(), errStatus.Code())
			return err
		}

		switch payload := in.Payload.(type) {
		case *pb.SignalRequest_Negotiate:
			if payload.Negotiate.Type != "answer" {
				// @todo
				continue
			}

			err = rtc.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  string(payload.Negotiate.Sdp),
			})

			if err != nil {
				return status.Errorf(codes.Internal, "%s", err)
			}
		case *pb.SignalRequest_Trickle:
			var candidate webrtc.ICECandidateInit
			err := json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
			if err != nil {
				log.Printf("error parsing ice candidate: %v", err)
			}

			err = rtc.AddICECandidate(candidate)
			if err != nil {
				return status.Errorf(codes.Internal, "error adding ice candidate")
			}
		case *pb.SignalRequest_Command_:
			r.onCommand(id, payload.Command)
		}
	}
}

func (r *Room) onCommand(from int, cmd *pb.SignalRequest_Command) {
	switch cmd.Type {
	case pb.SignalRequest_Command_ADD_SPEAKER:
		// @TODO IN DEVELOPMENT
		break
	case pb.SignalRequest_Command_REMOVE_SPEAKER:
		// @TODO IN DEVELOPMENT
		break
	case pb.SignalRequest_Command_MUTE_SPEAKER:
	case pb.SignalRequest_Command_UNMUTE_SPEAKER:
	case pb.SignalRequest_Command_REACTION:
		go r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_REACTED,
			From: int64(from),
			Data: cmd.Data,
		})
	}
}

func (r *Room) notify(event *pb.SignalReply_Event) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	for id, p := range r.members {
		if int64(id) == event.From {
			continue
		}

		err := p.stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Event_{
				Event: event,
			},
		})

		if err != nil {
			// @todo
			log.Printf("failed to write to data channel: %s\n", err.Error())
		}
	}
}
