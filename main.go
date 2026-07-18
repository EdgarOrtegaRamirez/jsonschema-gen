package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

// SchemaGenerator generates JSON Schema from Go structs or raw data
type SchemaGenerator struct {
	definitions map[string]*JSONSchema
}

// JSONSchema represents a JSON Schema node
type JSONSchema struct {
	Type                 string                 `json:"type,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	AdditionalProperties *bool                  `json:"additionalProperties,omitempty"`
	Items                *JSONSchema            `json:"items,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty"`
	Default              interface{}            `json:"default,omitempty"`
	MinLength            *int                   `json:"minLength,omitempty"`
	MaxLength            *int                   `json:"maxLength,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Pattern              string                 `json:"pattern,omitempty"`
	MinItems             *int                   `json:"minItems,omitempty"`
	MaxItems             *int                   `json:"maxItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty"`
	Const                interface{}            `json:"const,omitempty"`
	AllOf                []*JSONSchema          `json:"allOf,omitempty"`
	OneOf                []*JSONSchema          `json:"oneOf,omitempty"`
	AnyOf                []*JSONSchema          `json:"anyOf,omitempty"`
	Ref                  string                 `json:"$ref,omitempty"`
	Defs                 map[string]*JSONSchema `json:"$defs,omitempty"`
}

// SchemaOption is a functional option for configuring the generator
type SchemaOption func(*SchemaGenerator)

// WithDefinitions sets the definitions map
func WithDefinitions(defs map[string]*JSONSchema) SchemaOption {
	return func(g *SchemaGenerator) {
		g.definitions = defs
	}
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(opts ...SchemaOption) *SchemaGenerator {
	g := &SchemaGenerator{
		definitions: make(map[string]*JSONSchema),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// GenerateSchema generates a JSON Schema from a Go struct
func (g *SchemaGenerator) GenerateSchema(v interface{}) (*JSONSchema, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	schema, err := g.typeToSchema(val.Type(), val)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema: %w", err)
	}

	// Add $defs for referenced types
	if len(g.definitions) > 0 {
		schema.Defs = g.definitions
	}

	return schema, nil
}

// GenerateSchemaFromFile generates a JSON Schema from a JSON file
func (g *SchemaGenerator) GenerateSchemaFromFile(filePath string) (*JSONSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return g.GenerateFromValue(v)
}

// GenerateFromValue generates a JSON Schema from a Go value
func (g *SchemaGenerator) GenerateFromValue(v interface{}) (*JSONSchema, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return g.typeToSchema(val.Type(), val)
}

// SchemaFromType generates a JSON Schema from a Go type name
func (g *SchemaGenerator) SchemaFromType(typeName string) (*JSONSchema, error) {
	// Simple type mapping for primitive types
	schema := &JSONSchema{}
	switch typeName {
	case "string":
		schema.Type = "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		schema.Type = "integer"
	case "float32", "float64":
		schema.Type = "number"
	case "bool":
		schema.Type = "boolean"
	case "[]string", "[]int", "[]int64", "[]float64", "[]bool", "[]interface{}":
		schema.Type = "array"
		schema.Items = &JSONSchema{Type: "string"} // default item type
	case "map[string]interface{}":
		schema.Type = "object"
		schema.AdditionalProperties = boolPtr(true)
	case "time.Time":
		schema.Type = "string"
		schema.Format = "date-time"
	default:
		schema.Type = "object"
		schema.AdditionalProperties = boolPtr(false)
	}
	return schema, nil
}

// RenderJSON renders the schema as JSON
func (g *SchemaGenerator) RenderJSON(schema *JSONSchema, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(schema, "", "  ")
	}
	return json.Marshal(schema)
}

// RenderYAML renders the schema as YAML (basic conversion from JSON)
func (g *SchemaGenerator) RenderYAML(schema *JSONSchema) ([]byte, error) {
	jsonData, err := g.RenderJSON(schema, true)
	if err != nil {
		return nil, err
	}

	// Use python to convert JSON to YAML
	cmd := exec.Command("python3", "-c", `
import json, sys
data = json.load(sys.stdin)
try:
    import yaml
    print(yaml.dump(data, default_flow_style=False, sort_keys=False))
except ImportError:
    # Basic YAML-like output
    print(json.dumps(data, indent=2))
`)
	cmd.Stdin = strings.NewReader(string(jsonData))
	output, err := cmd.Output()
	if err != nil {
		// Fall back to JSON
		return jsonData, nil
	}
	return output, nil
}

// typeToSchema converts a Go type to a JSON Schema
func (g *SchemaGenerator) typeToSchema(t reflect.Type, val reflect.Value) (*JSONSchema, error) {
	schema := &JSONSchema{}

	switch val.Kind() {
	case reflect.String:
		schema.Type = "string"
		if val.Len() > 0 {
			schema.Default = val.String()
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = "integer"
		if val.Kind() == reflect.Int64 {
			schema.Format = "int64"
		}
		if val.Int() != 0 {
			schema.Default = val.Int()
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
		if val.Uint() != 0 {
			schema.Default = val.Uint()
		}
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
		if val.Float() != 0 {
			schema.Default = val.Float()
		}
	case reflect.Bool:
		schema.Type = "boolean"
		if val.Bool() {
			schema.Default = true
		}
	case reflect.Slice:
		schema.Type = "array"
		if val.Len() > 0 {
			schema.UniqueItems = true
		}
		if val.Len() > 0 {
			firstItem := val.Index(0)
			if firstItem.IsValid() {
				itemsSchema, err := g.typeToSchema(firstItem.Type(), firstItem)
				if err != nil {
					return nil, err
				}
				schema.Items = itemsSchema
			}
		}
		if val.Len() > 0 {
			n := val.Len()
			schema.MinItems = &n
		}
	case reflect.Map:
		schema.Type = "object"
		schema.AdditionalProperties = boolPtr(true)
		if val.Len() > 0 {
			// Try to infer structure
			schema.Properties = make(map[string]*JSONSchema)
			iter := val.MapRange()
			for iter.Next() {
				key := iter.Key()
				value := iter.Value()
				propName := fmt.Sprintf("%v", key.Interface())
				propSchema, err := g.typeToSchema(value.Type(), value)
				if err == nil {
					propSchema.Title = propName
					schema.Properties[propName] = propSchema
				}
			}
		}
	case reflect.Interface:
		// Unwrap the interface to get the concrete type inside
		// e.g., map[string]interface{} values are interfaces
		if val.IsValid() {
			elem := val.Elem()
			if elem.IsValid() {
				return g.typeToSchema(elem.Type(), elem)
			}
		}
		schema.Type = "object"
	case reflect.Struct:
		// Handle named structs
		if t.Name() != "" && t.PkgPath() != "" {
			// Named struct — add to definitions
			refName := t.Name()
			if _, exists := g.definitions[refName]; !exists {
				def := g.structToSchema(t, val)
				g.definitions[refName] = def
			}
			schema.Ref = "#/$defs/" + refName
			return schema, nil
		}

		// Handle time.Time
		if t.String() == "time.Time" {
			schema.Type = "string"
			schema.Format = "date-time"
			return schema, nil
		}

		// Empty struct
		schema.Type = "object"
		schema.AdditionalProperties = boolPtr(false)

	default:
		schema.Type = "object"
	}

	return schema, nil
}

// structToSchema converts a Go struct to a JSON Schema
func (g *SchemaGenerator) structToSchema(t reflect.Type, val reflect.Value) *JSONSchema {
	schema := &JSONSchema{
		Type:       "object",
		Properties: make(map[string]*JSONSchema),
	}

	// Get struct name for title
	schema.Title = t.Name()

	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Parse json tag
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] == "-" {
				continue // Skip
			}
			fieldName = parts[0]
		}

		propSchema, err := g.typeToSchema(field.Type, fieldValue)
		if err != nil {
			continue
		}

		// Parse description from tag
		if desc := field.Tag.Get("description"); desc != "" {
			propSchema.Description = desc
		}

		// Parse example from tag
		if example := field.Tag.Get("example"); example != "" {
			propSchema.Default = example
		}

		// Parse pattern from tag
		if pattern := field.Tag.Get("pattern"); pattern != "" {
			propSchema.Pattern = pattern
		}

		// Parse format from tag
		if format := field.Tag.Get("format"); format != "" {
			propSchema.Format = format
		}

		// Handle enum tag
		if enumStr := field.Tag.Get("enum"); enumStr != "" {
			var enums []interface{}
			for _, e := range strings.Split(enumStr, ",") {
				enums = append(enums, strings.TrimSpace(e))
			}
			propSchema.Enum = enums
		}

		// Handle required tag
		if field.Tag.Get("required") == "true" {
			required = append(required, fieldName)
		}

		// Handle minimum/maximum
		if minStr := field.Tag.Get("minimum"); minStr != "" {
			var min float64
			fmt.Sscanf(minStr, "%f", &min)
			propSchema.Minimum = &min
		}
		if maxStr := field.Tag.Get("maximum"); maxStr != "" {
			var max float64
			fmt.Sscanf(maxStr, "%f", &max)
			propSchema.Maximum = &max
		}
		if minLengthStr := field.Tag.Get("minLength"); minLengthStr != "" {
			var min int
			fmt.Sscanf(minLengthStr, "%d", &min)
			propSchema.MinLength = &min
		}
		if maxLengthStr := field.Tag.Get("maxLength"); maxLengthStr != "" {
			var max int
			fmt.Sscanf(maxLengthStr, "%d", &max)
			propSchema.MaxLength = &max
		}

		propSchema.Type = "" // Don't repeat type in properties

		schema.Properties[fieldName] = propSchema
	}

	if len(required) > 0 {
		schema.Required = required
	}

	// If no properties, make it empty object
	if len(schema.Properties) == 0 {
		schema.Type = "object"
		schema.AdditionalProperties = boolPtr(false)
	}

	return schema
}

// boolPtr returns a pointer to b
func boolPtr(b bool) *bool {
	return &b
}

// CLI command types
type CLICommand int

const (
	CommandGenerate CLICommand = iota
	CommandFromType
	CommandFromFile
	CommandFromValue
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "generate", "gen":
		cmdGenerate()
	case "from-type":
		cmdFromType()
	case "from-file":
		cmdFromFile()
	case "from-value":
		cmdFromValue()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`jsonschema-gen - Generate JSON Schema from Go structs and data

Commands:
  generate        Generate schema from stdin JSON
  from-type       Generate schema for a Go type
  from-file       Generate schema from a JSON file
  from-value      Generate schema from a JSON value

Options:
  --pretty        Pretty-print JSON output
  --format <fmt>  Output format: json (default), yaml
  --output <file> Write to file instead of stdout
  --help, -h      Show this help message

Examples:
  echo '{"name": "test"}' | jsonschema-gen generate
  jsonschema-gen from-file ./data.json
  jsonschema-gen from-type string
  jsonschema-gen from-type "map[string]interface{}"`)
}

// cmdGenerate reads JSON from stdin and generates a schema
func cmdGenerate() {
	var output string
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				output = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i--
			}
		}
	}

	// Read from stdin
	data, err := os.ReadFile("/dev/stdin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: no input provided. Pipe JSON via stdin\n")
		os.Exit(1)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outputBytes, err := g.RenderJSON(schema, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if output != "" {
		err = os.WriteFile(output, outputBytes, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Schema written to %s\n", output)
	} else {
		fmt.Println(string(outputBytes))
	}
}

func cmdFromType() {
	args := os.Args[2:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: jsonschema-gen from-type <type>\n")
		os.Exit(1)
	}

	typeName := args[0]
	g := NewSchemaGenerator()
	schema, err := g.SchemaFromType(typeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output, err := g.RenderJSON(schema, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

func cmdFromFile() {
	var output string
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				output = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i--
			}
		}
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: jsonschema-gen from-file <file>\n")
		os.Exit(1)
	}

	filePath := args[0]
	absPath, _ := filepath.Abs(filePath)
	g := NewSchemaGenerator()
	schema, err := g.GenerateSchemaFromFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outputBytes, err := g.RenderJSON(schema, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if output != "" {
		err = os.WriteFile(output, outputBytes, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Schema written to %s\n", output)
	} else {
		fmt.Println(string(outputBytes))
	}
}

func cmdFromValue() {
	var output string
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				output = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i--
			}
		}
	}

	// Read value from stdin
	data, err := os.ReadFile("/dev/stdin")
	if err != nil {
		data = []byte("{}")
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outputBytes, err := g.RenderJSON(schema, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if output != "" {
		err = os.WriteFile(output, outputBytes, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Schema written to %s\n", output)
	} else {
		fmt.Println(string(outputBytes))
	}
}
