package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// AddTagToFile adds a tag to the frontmatter tags array in a markdown file
// If the file doesn't have frontmatter or tags, it won't modify the file
// Returns true if the tag was added, false if it already existed or couldn't be added
func AddTagToFile(filePath string, tag string) (bool, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse frontmatter
	frontmatterEnd, frontmatter, err := extractFrontmatter(content)
	if err != nil || frontmatterEnd == 0 {
		// No frontmatter or couldn't parse - don't modify
		return false, nil
	}

	// Parse YAML frontmatter
	var fm map[string]interface{}
	if err := yaml.Unmarshal(frontmatter, &fm); err != nil {
		return false, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Check if tags exist
	tagsRaw, hasTagsField := fm["tags"]
	if !hasTagsField {
		// No tags field - don't add it
		return false, nil
	}

	// Convert tags to string slice
	var tags []string
	switch v := tagsRaw.(type) {
	case []interface{}:
		for _, tag := range v {
			if strTag, ok := tag.(string); ok {
				tags = append(tags, strTag)
			}
		}
	case []string:
		tags = v
	default:
		// Unknown tags format - don't modify
		return false, nil
	}

	// Check if tag already exists
	for _, existingTag := range tags {
		if existingTag == tag {
			return false, nil // Tag already exists
		}
	}

	// Add the tag
	tags = append(tags, tag)
	fm["tags"] = tags

	// Serialize back to YAML with inline array style for tags
	newFrontmatter, err := marshalFrontmatterWithFlowTags(fm)
	if err != nil {
		return false, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Reconstruct the file
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(newFrontmatter)
	buf.WriteString("---\n")
	buf.Write(content[frontmatterEnd:])

	// Write back to file
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return false, fmt.Errorf("failed to write file: %w", err)
	}

	return true, nil
}

// extractFrontmatter extracts the YAML frontmatter from markdown content
// Returns the end position of frontmatter and the frontmatter bytes
func extractFrontmatter(content []byte) (int, []byte, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))

	// Check first line is "---"
	if !scanner.Scan() {
		return 0, nil, fmt.Errorf("empty file")
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		return 0, nil, fmt.Errorf("no frontmatter found")
	}

	startPos := len(scanner.Text()) + 1 // +1 for newline

	// Collect frontmatter until closing "---"
	var frontmatter bytes.Buffer
	endPos := startPos

	for scanner.Scan() {
		line := scanner.Text()
		endPos += len(line) + 1 // +1 for newline

		if strings.TrimSpace(line) == "---" {
			// Found closing delimiter
			return endPos, frontmatter.Bytes(), nil
		}

		frontmatter.WriteString(line)
		frontmatter.WriteByte('\n')
	}

	return 0, nil, fmt.Errorf("frontmatter not closed")
}

// marshalFrontmatterWithFlowTags marshals frontmatter with inline array style for tags
func marshalFrontmatterWithFlowTags(fm map[string]interface{}) ([]byte, error) {
	// Create a YAML node
	var node yaml.Node
	if err := node.Encode(fm); err != nil {
		return nil, err
	}

	// Find and modify the tags field to use flow style
	if err := setFlowStyleForTags(&node); err != nil {
		return nil, err
	}

	// Marshal with the modified node
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&node); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// setFlowStyleForTags recursively finds the tags field and sets it to flow style
func setFlowStyleForTags(node *yaml.Node) error {
	if node.Kind != yaml.DocumentNode && node.Kind != yaml.MappingNode {
		return nil
	}

	// For document nodes, recurse into content
	if node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			if err := setFlowStyleForTags(child); err != nil {
				return err
			}
		}
		return nil
	}

	// For mapping nodes, look for "tags" key
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		// Check if this is the "tags" key
		if keyNode.Value == "tags" && valueNode.Kind == yaml.SequenceNode {
			// Set the sequence to flow style (inline array)
			valueNode.Style = yaml.FlowStyle

			// Set all string elements to use double quotes
			for _, item := range valueNode.Content {
				if item.Kind == yaml.ScalarNode && item.Tag == "!!str" {
					item.Style = yaml.DoubleQuotedStyle
				}
			}
		}
	}

	return nil
}
