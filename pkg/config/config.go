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

var Config = field.NewConfiguration([]field.SchemaField{
	field.StringField(
		"victorops-api-id",
		field.WithRequired(true),
		field.WithDescription("The client ID for the VictorOps API"),
	),
	field.StringField(
		"victorops-api-key",
		field.WithRequired(true),
		field.WithDescription("The API key for the VictorOps API"),
		field.WithIsSecret(true),
	),
	BaseURLField,
})

func ValidateConfig(c *Victorops) error {
	return nil
}
