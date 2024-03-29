package rooms

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/rooms/signal"
)

type Member struct {
	mux sync.RWMutex

	id        int
	name      string
	username  string
	image     string
	muted     bool
	connected bool
	role      pb.RoomState_RoomMember_Role

	joined time.Time

	// @TODO MIGHT MAKE SENSE TO MOVE THIS INTO A CLASS THAT MANAGES CONNECTION STUFF SIMILAR TO HOW IT WORKS ON CLIENT.
	peer   *sfu.PeerLocal
	signal signal.Transport

	dataChannel *BufferedDataChannel
}

func NewMember(id int, name, username, image string, peer *sfu.PeerLocal, signal signal.Transport) *Member {
	m := &Member{
		id:          id,
		name:        name,
		username:    username,
		image:       image,
		muted:       true,
		connected:   false,
		peer:        peer,
		signal:      signal,
		role:        pb.RoomState_RoomMember_ROLE_REGULAR,
		dataChannel: NewBufferedDataChannel(),
		joined:      time.Now(),
	}

	m.setup()
	return m
}

func (m *Member) IsConnected() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.connected
}

func (m *Member) Joined() time.Time {
	return m.joined
}

func (m *Member) Mute() {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.muted = true
}

func (m *Member) Unmute() {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.muted = false
}

func (m *Member) MarkConnected() {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.connected = true
}

func (m *Member) SetRole(role pb.RoomState_RoomMember_Role) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.role = role
}

func (m *Member) Notify(data []byte) error {
	return m.dataChannel.Write(data)
}

func (m *Member) Role() pb.RoomState_RoomMember_Role {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.role
}

func (m *Member) ReceiveMsg() (*pb.SignalRequest, error) {
	msg, err := m.signal.ReadMsg()
	if err != nil {
		_ = m.Close()
		return nil, err
	}

	return msg, nil
}

func (m *Member) StartChannel(label string) {
	m.dataChannel.Start(m.peer.Subscriber().DataChannel(label))
}

func (m *Member) RunSignal() error {
	for {

		// @TODO probably close through a channel

		in, err := m.ReceiveMsg()
		if err != nil {
			return err
		}

		switch in.Payload.(type) {
		case *pb.SignalRequest_Description:
			payload := in.GetDescription()
			sdp := webrtc.SessionDescription{
				Type: webrtc.NewSDPType(strings.ToLower(payload.Type)),
				SDP:  payload.Sdp,
			}

			if sdp.Type == webrtc.SDPTypeOffer {
				answer, err := m.peer.Answer(sdp)
				if err != nil {
					if err == sfu.ErrNoTransportEstablished || err == sfu.ErrOfferIgnored {
						continue
					}

					return fmt.Errorf("negotatie err: %v", err)
				}

				err = m.signal.Write(&pb.SignalReply{
					Id: in.Id,
					Payload: &pb.SignalReply_Description{
						Description: &pb.SessionDescription{
							Type: answer.Type.String(),
							Sdp:  answer.SDP,
						},
					},
				})

				if err != nil {
					log.Printf("conn.Write failed: %v", err)
					return err
				}

			} else if sdp.Type == webrtc.SDPTypeAnswer {
				err := m.peer.SetRemoteDescription(sdp)
				if err != nil && err != sfu.ErrNoTransportEstablished {
					return err
				}
			}
		case *pb.SignalRequest_Trickle:
			payload := in.GetTrickle()

			midLine := uint16(payload.IceCandidate.SdpMLineIndex)
			candidate := webrtc.ICECandidateInit{
				Candidate:     payload.IceCandidate.Candidate,
				SDPMid:        &payload.IceCandidate.SdpMid,
				SDPMLineIndex: &midLine,
			}

			err := m.peer.Trickle(candidate, int(payload.Target))
			if err != nil && err != sfu.ErrNoTransportEstablished {
				return fmt.Errorf("negotatie err: %v", err)
			}
		}
	}
}

func (m *Member) Close() error {
	_ = m.signal.Close()
	return m.peer.Close()
}

func (m *Member) ToProto() *pb.RoomState_RoomMember {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return &pb.RoomState_RoomMember{
		Id:          int64(m.id),
		DisplayName: m.name,
		Username:    m.username,
		Image:       m.image,
		Muted:       m.muted,
		Role:        m.role,
	}
}

func (m *Member) setup() {
	m.peer.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
		candidateProto := &pb.ICECandidate{
			Candidate: candidate.Candidate,
		}

		if candidate.SDPMid != nil {
			candidateProto.SdpMid = *candidate.SDPMid
		}

		if candidate.SDPMLineIndex != nil {
			candidateProto.SdpMLineIndex = int64(*candidate.SDPMLineIndex)
		}

		if candidate.UsernameFragment != nil {
			candidateProto.UsernameFragment = *candidate.UsernameFragment
		}

		err := m.signal.Write(&pb.SignalReply{
			Payload: &pb.SignalReply_Trickle{
				Trickle: &pb.Trickle{
					IceCandidate: candidateProto,
					Target:       pb.Trickle_Target(target),
				},
			},
		})

		if err != nil {
			log.Printf("OnIceCandidate send error %v ", err)
		}
	}

	// Notify user of new offer
	m.peer.OnOffer = func(o *webrtc.SessionDescription) {
		err := m.signal.Write(&pb.SignalReply{
			Payload: &pb.SignalReply_Description{
				Description: &pb.SessionDescription{
					Type: o.Type.String(),
					Sdp:  o.SDP,
				},
			},
		})

		if err != nil {
			log.Printf("negotiation error %s", err)
		}
	}
}
