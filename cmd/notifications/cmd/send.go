package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/notifications/pb"
	"github.com/soapboxsocial/soapbox/pkg/sql"
)

var send = &cobra.Command{
	Use:   "send",
	Short: "sends a notification",
	RunE:  runSend,
}

var (
	addr string

	// Related to who to send the notification to
	targets []int64
	query   string

	// The actual notification data
	body     string
	category string
)

func init() {
	send.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:50053", "grpc address")
	send.Flags().Int64SliceVarP(&targets, "targets", "t", []int64{}, "target user IDs")
	send.Flags().StringVarP(&query, "query", "q", "", "a query for target users")

	send.Flags().StringVarP(&body, "body", "", "", "notification body")
	send.Flags().StringVarP(&category, "category", "", "", "notification category")
}

func runSend(*cobra.Command, []string) error {
	if body == "" {
		return errors.New("body cannot be empty")
	}

	if category == "" {
		return errors.New("category cannot be empty")
	}

	notification := &pb.Notification{
		Category: category,
		Alert: &pb.Notification_Alert{
			Body: body,
		},
	}

	ids, err := getTargets()
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return errors.New("no targets")
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	defer conn.Close()

	client := pb.NewNotificationServiceClient(conn)

	_, err = client.SendNotification(context.TODO(), &pb.SendNotificationRequest{Targets: ids, Notification: notification})
	return err
}

func getTargets() ([]int64, error) {
	if len(targets) > 0 {
		return targets, nil
	}

	if query == "" {
		return nil, errors.New("query not supplied")
	}

	db, err := sql.Open(config.DB)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	result := make([]int64, 0)

	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			continue
		}

		result = append(result, id)
	}

	return result, nil
}
