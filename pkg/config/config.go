package config

//go:generate go run ./gen

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	BaseURLField = field.StringField(
		"base-url",
		field.WithDescription("Override the VictorOps API URL (for testing)"),
		field.WithHidden(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	)
)

var Config = field.NewConfiguration(
	[]field.SchemaField{
		field.StringField(
			"victorops-api-id",
			field.WithRequired(true),
			field.WithDescription("The client ID for the VictorOps API"),
			field.WithDisplayName("API ID"),
			field.WithPlaceholder("Enter your VictorOps API ID"),
		),
		field.StringField(
			"victorops-api-key",
			field.WithRequired(true),
			field.WithDescription("The API key for the VictorOps API"),
			field.WithDisplayName("API key"),
			field.WithPlaceholder("Enter your VictorOps API key"),
			field.WithIsSecret(true),
		),
		BaseURLField,
	},
	field.WithConnectorDisplayName("VictorOps"),
	field.WithIconUrl("/static/app-icons/victorops.svg"),
	field.WithHelpUrl("/docs/baton/victorops"),
)

func ValidateConfig(c *Victorops) error {
	return nil
}
