// Package markdown provides utilities for parsing and manipulating markdown documents.
// It uses the Goldmark library for parsing and supports YAML frontmatter via goldmark-meta.
package markdown

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Document represents a parsed markdown document
type Document struct {
	// FilePath is the path to the markdown file
	FilePath string

	// Content is the raw markdown content
	Content []byte

	// Metadata contains the YAML frontmatter
	Metadata map[string]any

	// AST is the parsed markdown abstract syntax tree
	AST ast.Node

	// Source is the source text reference for AST navigation
	Source []byte
}

// Parser handles markdown parsing
type Parser struct {
	md goldmark.Markdown
}

// NewParser creates a new markdown parser
func NewParser() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return &Parser{
		md: md,
	}
}

// ParseFile parses a markdown file and returns a Document
func (p *Parser) ParseFile(filePath string) (*Document, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(filePath, content)
}

// Parse parses markdown content and returns a Document
func (p *Parser) Parse(filePath string, content []byte) (*Document, error) {
	doc := &Document{
		FilePath: filePath,
		Content:  content,
		Source:   content,
	}

	// Create parser context
	ctx := parser.NewContext()

	// Parse the markdown
	doc.AST = p.md.Parser().Parse(text.NewReader(content), parser.WithContext(ctx))

	// Extract metadata (frontmatter)
	metaData := meta.Get(ctx)
	if metaData != nil {
		doc.Metadata = metaData
	} else {
		doc.Metadata = make(map[string]any)
	}

	return doc, nil
}

// WalkAST walks the AST and calls the visitor function for each node
func (doc *Document) WalkAST(visitor func(node ast.Node, entering bool) ast.WalkStatus) {
	_ = ast.Walk(doc.AST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return visitor(n, entering), nil
	})
}

// GetMetadata returns a metadata value by key
func (doc *Document) GetMetadata(key string) (any, bool) {
	val, ok := doc.Metadata[key]
	return val, ok
}

// GetMetadataString returns a metadata value as a string
func (doc *Document) GetMetadataString(key string) (string, bool) {
	val, ok := doc.GetMetadata(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetMetadataStringSlice returns a metadata value as a string slice
func (doc *Document) GetMetadataStringSlice(key string) ([]string, bool) {
	val, ok := doc.GetMetadata(key)
	if !ok {
		return nil, false
	}

	// Handle []any from YAML parser
	if slice, ok := val.([]any); ok {
		result := make([]string, 0, len(slice))
		for _, item := range slice {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result, true
	}

	// Handle []string directly
	if strSlice, ok := val.([]string); ok {
		return strSlice, true
	}

	return nil, false
}

// GetNodeText extracts text content from a node
func (doc *Document) GetNodeText(node ast.Node) string {
	var buf bytes.Buffer

	// Walk the node and collect text
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if text, ok := n.(*ast.Text); ok {
				buf.Write(text.Segment.Value(doc.Source))
			}
		}
		return ast.WalkContinue, nil
	})

	return buf.String()
}

// GetHeadings returns all headings in the document
func (doc *Document) GetHeadings() []Heading {
	var headings []Heading

	doc.WalkAST(func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if heading, ok := node.(*ast.Heading); ok {
			headings = append(headings, Heading{
				Level: heading.Level,
				Text:  doc.GetNodeText(heading),
				Node:  heading,
			})
		}

		return ast.WalkContinue
	})

	return headings
}

// Heading represents a markdown heading
type Heading struct {
	Level int
	Text  string
	Node  ast.Node
}
