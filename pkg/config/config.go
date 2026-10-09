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
		field.StringField(
			"removal-replacement-user",
			field.WithDescription("VictorOps username used as the default replacement for removed team members who are on call (rotations or escalation policies). " +
				"Does not need to belong to the team. Required for revoking team membership."),
			field.WithDisplayName("Removal replacement user"),
			field.WithPlaceholder("Enter a VictorOps username"),
		),
		BaseURLField,
	},
	field.WithConnectorDisplayName("VictorOps"),
	field.WithIconUrl("/static/app-icons/victorops.svg"),
	field.WithHelpUrl("/docs/baton/victorops"),
)
