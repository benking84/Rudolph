package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	deploymentEnvironmentsBaseDir        = "deployments/environments"
	deploymentEnvironmentsFoldersBaseDir = "deployments/environments"
)

type environmentConfiguration struct {
	Prefix              string `json:"prefix" mapstructure:"prefix"`
	ProjectID           string `json:"project_id" mapstructure:"project_id"`
	FirestoreDatabase   string `json:"firestore_database" mapstructure:"firestore_database"`
	FirestoreDatabaseID string `json:"firestore_database_id" mapstructure:"firestore_database_id"`
	Region              string `json:"region" mapstructure:"region"`
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func retrieveConfig(cmd *cobra.Command) (*environmentConfiguration, error) {
	var tfConfigFileFound bool
	p, err := cmd.Flags().GetString("ENV")
	if err != nil {
		return nil, err
	}

	if p == "" {
		// Try to get prefix from env instead
		p = os.Getenv("ENV")
	}
	env = p

	if env == "" {
		return nil, errors.New("missing ENV variable")
	}

	// Check to see if this is running under the git repo root source

	if fileExists(fmt.Sprintf(
		"%s/%s/config.auto.tfvars.json",
		deploymentEnvironmentsBaseDir,
		env,
	)) {
		tfConfigFileFound = true
	}

	if fileExists(fmt.Sprintf(
		"%s/%s/config.auto.tfvars.json",
		deploymentEnvironmentsFoldersBaseDir,
		env,
	)) {
		tfConfigFileFound = true
	}

	v := viper.New()

	// If running under the current git repo, use the TF config.auto.tfvars file
	// Else use a configuration stored under the $HOME/.rudolph-cli or current working directory
	if tfConfigFileFound {
		v.SetConfigName("config.auto.tfvars")
		v.SetConfigType("json")
		v.AddConfigPath(fmt.Sprintf("%s/%s", deploymentEnvironmentsBaseDir, env))
	} else {
		v.SetConfigName(env)
		v.SetConfigType("json")
		v.AddConfigPath("$HOME/.rudolph-cli")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return nil, err
		} else {
			// Config file was found but another error was produced
			return nil, err
		}
	}

	config := &environmentConfiguration{}
	err = v.Unmarshal(config)

	if err != nil {
		return nil, err
	}

	if config.ProjectID == "" {
		config.ProjectID = os.Getenv("GCP_PROJECT")
	}
	if config.FirestoreDatabase == "" {
		config.FirestoreDatabase = config.FirestoreDatabaseID
	}
	if config.FirestoreDatabase == "" {
		config.FirestoreDatabase = "rudolph"
	}
	cmd.Flags().Set("prefix", config.Prefix)
	cmd.Flags().Set("project", config.ProjectID)
	cmd.Flags().Set("firestore_database", config.FirestoreDatabase)

	return config, err

}
