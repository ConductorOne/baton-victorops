package main

import (
	"testing"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/test"
	cfg "github.com/conductorone/baton-victorops/pkg/config"
)

func TestConfigs(t *testing.T) {
	configurationSchema := field.NewConfiguration(
		cfg.ConfigurationFields,
		field.WithConstraints(cfg.FieldRelationships...),
	)

	testCases := []test.TestCase{
		// Add test cases here.
	}

	test.ExerciseTestCases(t, configurationSchema, ValidateConfig, testCases)
}
