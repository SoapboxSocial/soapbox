package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows"
)

type FollowRecommendationsNotificationHandler struct {
	targets *notifications.Settings
	backend *follows.Backend
}

func NewFollowRecommendationsNotificationHandler(targets *notifications.Settings) *FollowRecommendationsNotificationHandler {
	return &FollowRecommendationsNotificationHandler{
		targets: targets,
	}
}

func (f FollowRecommendationsNotificationHandler) Type() pubsub.EventType {
	return pubsub.EventTypeFollowRecommendations
}

func (f FollowRecommendationsNotificationHandler) Origin(*pubsub.Event) (int, error) {
	return 0, errors.New("no origin for event")
}

func (f FollowRecommendationsNotificationHandler) Targets(event *pubsub.Event) ([]notifications.Target, error) {
	targetID, err := event.GetInt("id")
	if err != nil {
		return nil, err
	}

	target, err := f.targets.GetSettingsFor(targetID)
	if err != nil {
		return nil, err
	}

	return []notifications.Target{*target}, nil
}

func (f FollowRecommendationsNotificationHandler) Build(event *pubsub.Event) (*notifications.PushNotification, error) {
	targetID, err := event.GetInt("id")
	if err != nil {
		return nil, err
	}

	recommendations, err := f.backend.RecommendationsFor(targetID)
	if err != nil {
		return nil, err
	}

	count := len(recommendations)
	if count == 0 {
		return nil, errors.New("no recommendations")
	}

	translation := ""
	body := ""
	args := make([]string, 0)

	switch count {
	case 1:
		translation += "1"
		body = fmt.Sprintf("%s who you may know is on Soapbox, why not follow them?", recommendations[0].DisplayName)
		args = append(args, recommendations[0].DisplayName)
	case 2:
		translation += "2"
		body = fmt.Sprintf(
			"%s and %s who you may know are on Soapbox, why not follow them?",
			recommendations[0].DisplayName, recommendations[1].DisplayName,
		)
		args = append(args, recommendations[0].DisplayName, recommendations[1].DisplayName)
	case 3:
		translation += "3"
		body = fmt.Sprintf(
			"%s, %s and %s who you may know are on Soapbox, why not follow them?",
			recommendations[0].DisplayName, recommendations[1].DisplayName, recommendations[2].DisplayName,
		)

		args = append(args, recommendations[0].DisplayName, recommendations[1].DisplayName, recommendations[2].DisplayName)
	default:
		translation += "3_and_more"
		body = fmt.Sprintf(
			"%s, %s, %s and %d others who you may know are on Soapbox, why not follow them?",
			recommendations[0].DisplayName, recommendations[1].DisplayName, recommendations[2].DisplayName, count-3,
		)

		args = append(args, recommendations[0].DisplayName, recommendations[1].DisplayName, recommendations[2].DisplayName, strconv.Itoa(count-3))
	}

	translation += "_follow_recommendations_notification"

	return &notifications.PushNotification{
		Category: notifications.FOLLOW_RECOMMENDATIONS,
		Alert: notifications.Alert{
			Body:      body,
			Key:       translation,
			Arguments: args,
		},
	}, nil
}
