package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// InstrumentationConfig provides configuration options for instrumentation.
// It contains a flag to enable instrumentation and lists of strings to specify
// which items to include or exclude during instrumentation.
type InstrumentationConfig struct {
	Enable  bool     `yaml:"enable"`
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// LoggingConfig provides configuration options for logging.
// It includes the log level and the output destination for log messages.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

// TracingConfig provides configuration options for tracing.
// It specifies the output format for traces and a flag to determine whether to dump traces on exit.
type TracingConfig struct {
	OutputFormat string `yaml:"outputFormat"`
	DumpOnExit   bool   `yaml:"dumpOnExit"`
}

// VisualizationConfig provides configuration options for visualization.
// It contains a flag indicating whether to generate a call graph and the output path for the call graph.
type VisualizationConfig struct {
	GenerateCallGraph bool   `yaml:"generateCallGraph"`
	CallGraphOutput   string `yaml:"callGraphOutput"`
}

// Config aggregates all configuration settings including instrumentation, logging,
// tracing, and visualization configurations.
type Config struct {
	Instrumentation InstrumentationConfig `yaml:"instrumentation"`
	Logging         LoggingConfig         `yaml:"logging"`
	Tracing         TracingConfig         `yaml:"tracing"`
	Visualization   VisualizationConfig   `yaml:"visualization"`
}

// LoadConfig reads a YAML configuration file and unmarshals its contents into a Config struct.
// If the filename parameter is an empty string, it defaults to "tracewrap.yaml" in the current working directory.
//
// Parameters:
//   - filename (string): the path to the YAML configuration file.
//
// Returns:
//   - *Config: a pointer to the populated Config struct.
//   - error: an error value if reading or unmarshalling the file fails.
func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		filename = "tracewrap.yaml"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
