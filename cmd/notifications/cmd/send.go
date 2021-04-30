package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var send = &cobra.Command{
	Use:   "send",
	Short: "sends a notification",
	RunE:  runSend,
}

var (
	addr string

	// Related to who to send the notification to
	targets []int
	query   string

	// The actual notification data
	body     string
	category string
)

func init() {
	send.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:50053", "grpc address")
	send.Flags().IntSliceVarP(&targets, "targets", "t", []int{}, "target user IDs")
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

	return nil
}
