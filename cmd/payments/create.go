package payments

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/SirWaithaka/gorequest"
	"github.com/SirWaithaka/payments-api/pkg/sdk"
)

func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources in the payments system",
	}

	cmd.AddCommand(NewCreateShortCodeCmd())

	return cmd
}

func NewCreateShortCodeCmd() *cobra.Command {
	var (
		endpoint          string
		environment       string
		service           string
		typeField         string
		shortcode         string
		initiatorName     string
		initiatorPassword string
		key               string
		secret            string
		passphrase        string
	)

	cmd := &cobra.Command{
		Use:   "shortcode",
		Short: "Create a new shortcode configuration",
		Long: `Create a new shortcode configuration in the payments system.

This command configures a new shortcode for processing payments through
the specified service (daraja or quikk) in the given environment.`,
		Example: `  # Create a Daraja shortcode for sandbox
  payments create shortcode \
    --endpoint https://api.payments.example.com \
    --environment sandbox \
    --service daraja \
    --type payout \
    --shortcode 174379 \
    --initiator-name "api_user" \
    --initiator-password "secret123" \
    --key "consumer_key" \
    --secret "consumer_secret"

  # Create a Quikk shortcode with passphrase
  payments create shortcode \
    --endpoint https://api.payments.example.com \
    --environment production \
    --service quikk \
    --type charge \
    --shortcode 600000 \
    --initiator-name "initiator" \
    --initiator-password "pass" \
    --key "key123" \
    --secret "secret123" \
    --passphrase "optional_phrase"`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate enum fields
			if environment != "sandbox" && environment != "production" {
				return fmt.Errorf("environment must be either 'sandbox' or 'production'")
			}
			if service != "daraja" && service != "quikk" {
				return fmt.Errorf("service must be either 'daraja' or 'quikk'")
			}
			if typeField != "charge" && typeField != "payout" && typeField != "transfer" {
				return fmt.Errorf("type must be one of: 'charge', 'payout', 'transfer'")
			}

			// Initialize SDK client
			client := sdk.New(sdk.Config{
				Endpoint: endpoint,
				LogLevel: gorequest.LogDebug,
			})

			// Build request
			request := sdk.RequestAddShortCode{
				Environment:       environment,
				Service:           service,
				Type:              typeField,
				ShortCode:         shortcode,
				InitiatorName:     initiatorName,
				InitiatorPassword: initiatorPassword,
				Key:               key,
				Secret:            secret,
				Passphrase:        passphrase,
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			// Execute request
			err := client.AddShortCode(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to create shortcode: %w", err)
			}

			fmt.Printf("✓ Successfully created shortcode %s\n", shortcode)
			return nil
		},
	}

	// Required flags
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint URL (required)")
	cmd.Flags().StringVar(&environment, "environment", "", "Environment: sandbox or production (required)")
	cmd.Flags().StringVar(&service, "service", "", "Service: daraja or quikk (required)")
	cmd.Flags().StringVar(&typeField, "type", "", "Type: charge, payout, or transfer (required)")
	cmd.Flags().StringVar(&shortcode, "shortcode", "", "Shortcode number (required)")
	cmd.Flags().StringVar(&initiatorName, "initiator-name", "", "Initiator name (required)")
	cmd.Flags().StringVar(&initiatorPassword, "initiator-password", "", "Initiator password (required)")
	cmd.Flags().StringVar(&key, "key", "", "Consumer key/API key (required)")
	cmd.Flags().StringVar(&secret, "secret", "", "Consumer secret/API secret (required)")

	// Optional flags
	cmd.Flags().StringVar(&passphrase, "passphrase", "", "Passphrase (optional)")

	// Mark required flags
	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("environment")
	cmd.MarkFlagRequired("service")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("shortcode")
	cmd.MarkFlagRequired("initiator-name")
	cmd.MarkFlagRequired("initiator-password")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("secret")

	return cmd
}
