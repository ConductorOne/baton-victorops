package main

import (
	"errors"

	cfg "github.com/conductorone/baton-victorops/pkg/config"
	"github.com/spf13/viper"
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	if v.GetString(cfg.VictorOpsApiIdField.FieldName) == "" {
		return errors.New("victorops-api-id is required")
	}

	if v.GetString(cfg.VictorOpsApiKeyField.FieldName) == "" {
		return errors.New("victorops-api-key is required")
	}

	return nil
}
