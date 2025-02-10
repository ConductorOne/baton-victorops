package main

import (
	"errors"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
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
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	if v.GetString(VictorOpsApiIdField.FieldName) == "" {
		return errors.New("victorops-api-id is required")
	}

	if v.GetString(VictorOpsApiKeyField.FieldName) == "" {
		return errors.New("victorops-api-key is required")
	}

	return nil
}
