package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	defaultYAML  string
}

func newDefaultConfigInitializer() *ConfigInitializer {
	return &ConfigInitializer{
		configHome:   xdg.ConfigHome,
		templateYAML: configtemplate.TemplateYAML(),
		defaultYAML:  configtemplate.DefaultYAML(),
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

	templateYAML := c.templateContent()
	if strings.TrimSpace(templateYAML) == "" {
		return "", fmt.Errorf("config template is empty")
	}

	if _, err := os.Stat(configPath); err == nil {
		return "", fmt.Errorf("config file already exists: %s", configPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat config file: %w", err)
	}

	validator := viper.New()
	validator.SetConfigType("yaml")
	if err := validator.ReadConfig(strings.NewReader(templateYAML)); err != nil {
		return "", fmt.Errorf("parse config template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(templateYAML), 0o644); err != nil {
		return "", fmt.Errorf("write config file: %w", err)
	}

	if _, err := c.loadConfig(); err != nil {
		return "", err
	}

	return configPath, nil
}

func (c *ConfigInitializer) ListKeyValues() ([]string, error) {
	rootNode, err := c.loadConfigNode()
	if err != nil {
		return nil, err
	}

	return flattenKeyValues(rootNode), nil
}

func (c *ConfigInitializer) ListDefaultKeyValues() ([]string, error) {
	rootNode, err := c.loadDefaultConfigNode()
	if err != nil {
		return nil, err
	}

	return flattenKeyValues(rootNode), nil
}

func (c *ConfigInitializer) ListResolvedKeyValues() ([]string, error) {
	rootNode, err := c.loadResolvedConfigNode()
	if err != nil {
		return nil, err
	}

	return flattenKeyValues(rootNode), nil
}

func (c *ConfigInitializer) GetValue(key string) (string, error) {
	rootNode, err := c.loadConfigNode()
	if err != nil {
		return "", err
	}

	return scalarValueFromNode(rootNode, key)
}

func (c *ConfigInitializer) GetResolvedValue(key string) (string, error) {
	rootNode, err := c.loadResolvedConfigNode()
	if err != nil {
		return "", err
	}

	return scalarValueFromNode(rootNode, key)
}

func (c *ConfigInitializer) SetValue(key string, value string) error {
	segments, err := splitDotPath(key)
	if err != nil {
		return err
	}

	rootNode, err := c.loadConfigNode()
	if err != nil {
		return err
	}

	targetNode, err := ensureYAMLValueNode(rootNode, segments, nil)
	if err != nil {
		return err
	}

	setScalarNodeValue(targetNode, value)

	if err := c.writeConfigNode(rootNode); err != nil {
		return err
	}

	if _, err := c.loadConfig(); err != nil {
		return err
	}

	return nil
}

func (c *ConfigInitializer) UnsetValue(key string) error {
	segments, err := splitDotPath(key)
	if err != nil {
		return err
	}

	rootNode, err := c.loadConfigNode()
	if err != nil {
		return err
	}

	deleted, _, err := deleteYAMLValueNode(rootNode, segments, nil)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("config key not found: %s", key)
	}

	if err := c.writeConfigNode(rootNode); err != nil {
		return err
	}

	if _, err := c.loadConfig(); err != nil {
		return err
	}

	return nil
}

func flattenKeyValues(rootNode *yaml.Node) []string {
	entries := flattenYAMLNode("", rootNode, nil)
	sort.Strings(entries)

	return entries
}

func scalarValueFromNode(rootNode *yaml.Node, key string) (string, error) {
	segments, err := splitDotPath(key)
	if err != nil {
		return "", err
	}

	valueNode, found := findYAMLNode(rootNode, segments)
	if !found {
		return "", fmt.Errorf("config key not found: %s", key)
	}

	switch valueNode.Kind {
	case yaml.MappingNode, yaml.SequenceNode:
		return "", fmt.Errorf("config key is not a scalar value: %s", key)
	default:
		return valueNode.Value, nil
	}
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

func (c *ConfigInitializer) templateContent() string {
	if c.templateYAML != "" {
		return c.templateYAML
	}

	return configtemplate.TemplateYAML()
}

func (c *ConfigInitializer) defaultContent() string {
	if c.defaultYAML != "" {
		return c.defaultYAML
	}

	return configtemplate.DefaultYAML()
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

func (c *ConfigInitializer) loadConfigNode() (*yaml.Node, error) {
	configPath, err := c.existingConfigPath()
	if err != nil {
		return nil, err
	}

	return c.loadConfigNodeFromPath(configPath)
}

func (c *ConfigInitializer) loadOptionalConfigNode() (*yaml.Node, error) {
	configPath, err := c.configPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return newDocumentNode(newMappingNode()), nil
	} else if err != nil {
		return nil, fmt.Errorf("stat config file: %w", err)
	}

	return c.loadConfigNodeFromPath(configPath)
}

func (c *ConfigInitializer) loadTemplateConfigNode() (*yaml.Node, error) {
	templateYAML := c.templateContent()
	if strings.TrimSpace(templateYAML) == "" {
		return nil, fmt.Errorf("config template is empty")
	}

	return loadEmbeddedConfigNode(templateYAML, "config template")
}

func (c *ConfigInitializer) loadDefaultConfigNode() (*yaml.Node, error) {
	defaultYAML := c.defaultContent()
	if strings.TrimSpace(defaultYAML) == "" {
		return nil, fmt.Errorf("default config is empty")
	}

	return loadEmbeddedConfigNode(defaultYAML, "default config")
}

func (c *ConfigInitializer) loadResolvedConfigNode() (*yaml.Node, error) {
	templateNode, err := c.loadTemplateConfigNode()
	if err != nil {
		return nil, err
	}

	defaultNode, err := c.loadDefaultConfigNode()
	if err != nil {
		return nil, err
	}

	userNode, err := c.loadOptionalConfigNode()
	if err != nil {
		return nil, err
	}

	resolvedNode := mergeYAMLNodes(cloneYAMLNode(templateNode), defaultNode)
	return mergeYAMLNodes(resolvedNode, userNode), nil
}

func (c *ConfigInitializer) loadConfigNodeFromPath(configPath string) (*yaml.Node, error) {
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	return parseYAMLNode(configContent, "config file")
}

func (c *ConfigInitializer) writeConfigNode(rootNode *yaml.Node) error {
	configPath, err := c.existingConfigPath()
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(rootNode); err != nil {
		return fmt.Errorf("encode config file: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("close config encoder: %w", err)
	}

	if err := os.WriteFile(configPath, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func loadEmbeddedConfigNode(content string, source string) (*yaml.Node, error) {
	return parseYAMLNode([]byte(content), source)
}

func parseYAMLNode(content []byte, source string) (*yaml.Node, error) {
	var rootNode yaml.Node
	if err := yaml.Unmarshal(content, &rootNode); err != nil {
		return nil, fmt.Errorf("parse %s: %w", source, err)
	}
	if rootNode.Kind == 0 {
		return newDocumentNode(newMappingNode()), nil
	}

	return &rootNode, nil
}

func flattenYAMLNode(prefix string, node *yaml.Node, entries []string) []string {
	if node == nil {
		return entries
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			entries = flattenYAMLNode(prefix, child, entries)
		}
	case yaml.MappingNode:
		if len(node.Content) == 0 && prefix != "" {
			entries = append(entries, prefix+"=")
			return entries
		}

		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode := node.Content[index]
			valueNode := node.Content[index+1]

			nextPrefix := keyNode.Value
			if prefix != "" {
				nextPrefix = prefix + "." + keyNode.Value
			}

			entries = flattenYAMLNode(nextPrefix, valueNode, entries)
		}
	case yaml.SequenceNode:
		if len(node.Content) == 0 && prefix != "" {
			entries = append(entries, prefix+"=")
			return entries
		}

		for index, item := range node.Content {
			nextPrefix := fmt.Sprintf("%d", index)
			if prefix != "" {
				nextPrefix = fmt.Sprintf("%s.%d", prefix, index)
			}

			entries = flattenYAMLNode(nextPrefix, item, entries)
		}
	default:
		if prefix != "" {
			entries = append(entries, prefix+"="+node.Value)
		}
	}

	return entries
}

func splitDotPath(key string) ([]string, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("config key is empty")
	}

	segments := strings.Split(key, ".")
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			return nil, fmt.Errorf("config key is invalid: %s", key)
		}
	}

	return segments, nil
}

func ensureYAMLValueNode(node *yaml.Node, segments []string, traversed []string) (*yaml.Node, error) {
	if node == nil {
		return nil, fmt.Errorf("config tree is empty")
	}

	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			node.Content = []*yaml.Node{newMappingNode()}
		}

		return ensureYAMLValueNode(node.Content[0], segments, traversed)
	}

	if len(segments) == 0 {
		return node, nil
	}

	switch node.Kind {
	case yaml.MappingNode:
		segment := segments[0]
		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode := node.Content[index]
			valueNode := node.Content[index+1]
			if keyNode.Value != segment {
				continue
			}

			if len(segments) == 1 {
				return valueNode, nil
			}

			switch valueNode.Kind {
			case yaml.MappingNode, yaml.SequenceNode, yaml.DocumentNode:
				return ensureYAMLValueNode(valueNode, segments[1:], append(traversed, segment))
			default:
				return nil, fmt.Errorf("config key parent is not a mapping: %s", strings.Join(append(traversed, segment), "."))
			}
		}

		var valueNode *yaml.Node
		if len(segments) == 1 {
			valueNode = newScalarNode("")
		} else {
			valueNode = newMappingNode()
		}

		node.Content = append(node.Content, newKeyNode(segment), valueNode)
		if len(segments) == 1 {
			return valueNode, nil
		}

		return ensureYAMLValueNode(valueNode, segments[1:], append(traversed, segment))
	case yaml.SequenceNode:
		index, err := strconv.Atoi(segments[0])
		if err != nil || index < 0 || index >= len(node.Content) {
			return nil, fmt.Errorf("config key not found: %s", strings.Join(append(traversed, segments[0]), "."))
		}

		if len(segments) == 1 {
			return node.Content[index], nil
		}

		switch node.Content[index].Kind {
		case yaml.MappingNode, yaml.SequenceNode, yaml.DocumentNode:
			return ensureYAMLValueNode(node.Content[index], segments[1:], append(traversed, segments[0]))
		default:
			return nil, fmt.Errorf("config key parent is not a mapping: %s", strings.Join(append(traversed, segments[0]), "."))
		}
	default:
		return nil, fmt.Errorf("config key parent is not a mapping: %s", strings.Join(traversed, "."))
	}
}

func findYAMLNode(node *yaml.Node, segments []string) (*yaml.Node, bool) {
	if node == nil {
		return nil, false
	}

	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return nil, false
		}

		return findYAMLNode(node.Content[0], segments)
	}

	if len(segments) == 0 {
		return node, true
	}

	switch node.Kind {
	case yaml.MappingNode:
		segment := segments[0]
		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode := node.Content[index]
			valueNode := node.Content[index+1]
			if keyNode.Value != segment {
				continue
			}

			return findYAMLNode(valueNode, segments[1:])
		}
	case yaml.SequenceNode:
		index, err := strconv.Atoi(segments[0])
		if err != nil || index < 0 || index >= len(node.Content) {
			return nil, false
		}

		return findYAMLNode(node.Content[index], segments[1:])
	default:
		if len(segments) == 0 {
			return node, true
		}
	}

	return nil, false
}

func cloneYAMLNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}

	cloned := *node
	if len(node.Content) > 0 {
		cloned.Content = make([]*yaml.Node, len(node.Content))
		for index, child := range node.Content {
			cloned.Content[index] = cloneYAMLNode(child)
		}
	}

	return &cloned
}

func mergeYAMLNodes(base *yaml.Node, override *yaml.Node) *yaml.Node {
	if base == nil {
		return cloneYAMLNode(override)
	}
	if override == nil {
		return base
	}

	if base.Kind == yaml.DocumentNode {
		if len(base.Content) == 0 {
			base.Content = []*yaml.Node{newMappingNode()}
		}
		if override.Kind == yaml.DocumentNode {
			if len(override.Content) == 0 {
				return base
			}
			base.Content[0] = mergeYAMLNodes(base.Content[0], override.Content[0])
			return base
		}

		base.Content[0] = mergeYAMLNodes(base.Content[0], override)
		return base
	}

	if override.Kind == yaml.DocumentNode {
		if len(override.Content) == 0 {
			return base
		}

		return mergeYAMLNodes(base, override.Content[0])
	}

	if base.Kind == yaml.MappingNode && override.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(override.Content); index += 2 {
			overrideKeyNode := override.Content[index]
			overrideValueNode := override.Content[index+1]
			baseValueIndex, found := findMappingValueIndex(base, overrideKeyNode.Value)
			if !found {
				base.Content = append(base.Content, cloneYAMLNode(overrideKeyNode), cloneYAMLNode(overrideValueNode))
				continue
			}

			base.Content[baseValueIndex] = mergeYAMLNodes(base.Content[baseValueIndex], overrideValueNode)
		}

		return base
	}

	return cloneYAMLNode(override)
}

func findMappingValueIndex(node *yaml.Node, key string) (int, bool) {
	if node == nil || node.Kind != yaml.MappingNode {
		return 0, false
	}

	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return index + 1, true
		}
	}

	return 0, false
}

func deleteYAMLValueNode(node *yaml.Node, segments []string, traversed []string) (bool, bool, error) {
	if node == nil {
		return false, false, nil
	}

	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return false, false, nil
		}

		deleted, prune, err := deleteYAMLValueNode(node.Content[0], segments, traversed)
		if err != nil {
			return false, false, err
		}
		if deleted && prune {
			node.Content[0] = newMappingNode()
		}

		return deleted, false, nil
	}

	if len(segments) == 0 {
		return false, false, nil
	}

	switch node.Kind {
	case yaml.MappingNode:
		segment := segments[0]
		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode := node.Content[index]
			valueNode := node.Content[index+1]
			if keyNode.Value != segment {
				continue
			}

			if len(segments) == 1 {
				node.Content = append(node.Content[:index], node.Content[index+2:]...)
				return true, len(node.Content) == 0, nil
			}

			deleted, prune, err := deleteYAMLValueNode(valueNode, segments[1:], append(traversed, segment))
			if err != nil {
				return false, false, err
			}
			if !deleted {
				return false, false, nil
			}
			if prune {
				node.Content = append(node.Content[:index], node.Content[index+2:]...)
			}

			return true, len(node.Content) == 0, nil
		}
	case yaml.SequenceNode:
		index, err := strconv.Atoi(segments[0])
		if err != nil || index < 0 || index >= len(node.Content) {
			return false, false, nil
		}

		if len(segments) == 1 {
			node.Content = append(node.Content[:index], node.Content[index+1:]...)
			return true, len(node.Content) == 0, nil
		}

		deleted, prune, err := deleteYAMLValueNode(node.Content[index], segments[1:], append(traversed, segments[0]))
		if err != nil {
			return false, false, err
		}
		if !deleted {
			return false, false, nil
		}
		if prune {
			node.Content = append(node.Content[:index], node.Content[index+1:]...)
		}

		return true, len(node.Content) == 0, nil
	default:
		return false, false, fmt.Errorf("config key parent is not a mapping: %s", strings.Join(traversed, "."))
	}

	return false, false, nil
}

func newKeyNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
}

func newDocumentNode(content *yaml.Node) *yaml.Node {
	if content == nil {
		content = newMappingNode()
	}

	return &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{content},
	}
}

func newMappingNode() *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
}

func newScalarNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
}

func setScalarNodeValue(node *yaml.Node, value string) {
	node.Kind = yaml.ScalarNode
	node.Tag = "!!str"
	node.Style = 0
	node.Value = value
	node.Anchor = ""
	node.Alias = nil
	node.Content = nil
}
