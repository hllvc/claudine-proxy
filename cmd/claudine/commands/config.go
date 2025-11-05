package commands

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"

	"github.com/florianilch/claudine-proxy/internal/app"
)

// envPrefix is stripped from environment variables during config loading (e.g., CLAUDINE_SERVER__HOST → server.host)
const envPrefix = "CLAUDINE_"

// loadConfig loads application configuration from various sources with precedence:
// config file → environment variables → CLI flags → defaults
func loadConfig(configPath string, cmd *cli.Command, environFunc func() []string) (*app.Config, error) {
	k := koanf.New(".")

	// 1. Load from config file if provided
	if configPath != "" {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return nil, fmt.Errorf("loading config file: %w", err)
		}
	}

	// 2. Load from environment variables
	envProvider := env.Provider(".", env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(key, value string) (string, any) {
			stripped := strings.TrimPrefix(key, envPrefix)
			nested := strings.ToLower(strings.ReplaceAll(stripped, "__", "."))
			return nested, value
		},
		EnvironFunc: environFunc,
	})
	if err := k.Load(envProvider, nil); err != nil {
		return nil, fmt.Errorf("loading environment variables: %w", err)
	}

	// 3. Load from CLI flags if provided
	if cmd != nil {
		flagValues := extractAndTransformFlags(cmd)
		if err := k.Load(confmap.Provider(flagValues, "."), nil); err != nil {
			return nil, fmt.Errorf("loading CLI flags: %w", err)
		}
	}

	config := &app.Config{}
	if err := k.UnmarshalWithConf("", config, koanf.UnmarshalConf{Tag: "json"}); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := config.ApplyDefaults(); err != nil {
		return nil, fmt.Errorf("applying defaults: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// extractAndTransformFlags transforms CLI flag names to match config structure.
// Includes parent flags. Examples: --server--host → server.host, --log-level → log_level
func extractAndTransformFlags(cmd *cli.Command) map[string]any {
	values := make(map[string]any)

	// FlagNames() includes flags from parent commands (via lineage)
	for _, name := range cmd.FlagNames() {
		// Skip unset flags to preserve precedence from earlier config sources
		if !cmd.IsSet(name) {
			continue
		}

		if value := cmd.Value(name); value != nil {
			key := strings.ReplaceAll(name, "--", ".")
			key = strings.ReplaceAll(key, "-", "_")
			values[key] = value
		}
	}

	return values
}
