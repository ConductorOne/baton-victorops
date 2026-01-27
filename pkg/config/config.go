package config

//go:generate go run ./gen

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	VictorOpsApiIdField = field.StringField(
		"victorops-api-id",
		field.WithRequired(true),
		field.WithDescription("The client ID for the VictorOps API"),
	)

	VictorOpsApiKeyField = field.StringField(
		"victorops-api-key",
		field.WithRequired(true),
		field.WithDescription("The API key for the VictorOps API"),
	)

	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{
		VictorOpsApiIdField,
		VictorOpsApiKeyField,
	}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}

	// Config is the configuration schema for the connector.
	Config = field.Configuration{
		Fields:      ConfigurationFields,
		Constraints: FieldRelationships,
	}
)
