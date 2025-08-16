package config

import "github.com/urfave/cli/v3"

// Firestore contains configuration for Google Cloud Firestore
type Firestore struct {
	ProjectID  string
	DatabaseID string
}

// Flags returns CLI flags for Firestore configuration
func (f *Firestore) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "firestore-project-id",
			Usage:       "Google Cloud Project ID for Firestore",
			Sources:     cli.EnvVars("TAMAMO_FIRESTORE_PROJECT_ID"),
			Destination: &f.ProjectID,
		},
		&cli.StringFlag{
			Name:        "firestore-database-id",
			Usage:       "Firestore Database ID (default: (default))",
			Sources:     cli.EnvVars("TAMAMO_FIRESTORE_DATABASE_ID"),
			Value:       "(default)",
			Destination: &f.DatabaseID,
		},
	}
}

// SetDefaults sets default values for Firestore configuration
func (f *Firestore) SetDefaults() {
	if f.DatabaseID == "" {
		f.DatabaseID = "(default)"
	}
}

// IsValid checks if the Firestore configuration is valid
func (f *Firestore) IsValid() bool {
	return f.ProjectID != "" && f.DatabaseID != ""
}
