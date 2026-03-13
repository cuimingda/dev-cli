package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/xdg"
	configtemplate "github.com/cuimingda/dev-cli/config"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

const (
	developerIdentifier = "mingda.dev"
	cliName             = "dev"
	configFileName      = "config.yaml"
)

type ConfigInitializer struct {
	configHome   string
	templateYAML string
}

func newDefaultConfigInitializer() *ConfigInitializer {
	return &ConfigInitializer{
		configHome:   xdg.ConfigHome,
		templateYAML: configtemplate.TemplateYAML(),
	}
}

func (c *ConfigInitializer) DefaultPath() string {
	return filepath.Join(c.configHome, developerIdentifier, cliName, configFileName)
}

func (c *ConfigInitializer) Init() (string, error) {
	configPath, err := c.configPath()
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(c.templateYAML) == "" {
		return "", fmt.Errorf("config template is empty")
	}

	if _, err := os.Stat(configPath); err == nil {
		return "", fmt.Errorf("config file already exists: %s", configPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat config file: %w", err)
	}

	validator := viper.New()
	validator.SetConfigType("yaml")
	if err := validator.ReadConfig(strings.NewReader(c.templateYAML)); err != nil {
		return "", fmt.Errorf("parse config template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(c.templateYAML), 0o644); err != nil {
		return "", fmt.Errorf("write config file: %w", err)
	}

	if _, err := c.loadConfig(); err != nil {
		return "", err
	}

	return configPath, nil
}

func (c *ConfigInitializer) ListDotPaths() ([]string, error) {
	configPath, err := c.existingConfigPath()
	if err != nil {
		return nil, err
	}

	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(configContent, &rootNode); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	paths := flattenYAMLNode("", &rootNode, nil)
	sort.Strings(paths)

	return paths, nil
}

func (c *ConfigInitializer) loadConfig() (*viper.Viper, error) {
	configPath, err := c.existingConfigPath()
	if err != nil {
		return nil, err
	}

	loadedConfig := viper.New()
	loadedConfig.SetConfigFile(configPath)
	if err := loadedConfig.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	return loadedConfig, nil
}

func (c *ConfigInitializer) configPath() (string, error) {
	if strings.TrimSpace(c.configHome) == "" {
		return "", fmt.Errorf("config home is empty")
	}

	configPath := c.DefaultPath()
	if strings.TrimSpace(configPath) == "" {
		return "", fmt.Errorf("default config path is empty")
	}

	return configPath, nil
}

func (c *ConfigInitializer) existingConfigPath() (string, error) {
	configPath, err := c.configPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("config file does not exist: %s", configPath)
	} else if err != nil {
		return "", fmt.Errorf("stat config file: %w", err)
	}

	return configPath, nil
}

func flattenYAMLNode(prefix string, node *yaml.Node, paths []string) []string {
	if node == nil {
		return paths
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			paths = flattenYAMLNode(prefix, child, paths)
		}
	case yaml.MappingNode:
		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode := node.Content[index]
			valueNode := node.Content[index+1]

			nextPrefix := keyNode.Value
			if prefix != "" {
				nextPrefix = prefix + "." + keyNode.Value
			}

			paths = flattenYAMLNode(nextPrefix, valueNode, paths)
		}
	case yaml.SequenceNode:
		for index, item := range node.Content {
			nextPrefix := fmt.Sprintf("%d", index)
			if prefix != "" {
				nextPrefix = fmt.Sprintf("%s.%d", prefix, index)
			}

			paths = flattenYAMLNode(nextPrefix, item, paths)
		}
	default:
		if prefix != "" {
			paths = append(paths, prefix)
		}
	}

	return paths
}
