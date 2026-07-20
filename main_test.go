package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- SchemaGenerator Tests ---

func TestNewSchemaGenerator(t *testing.T) {
	g := NewSchemaGenerator()
	if g == nil {
		t.Fatal("expected non-nil generator")
	}
	if g.definitions == nil {
		t.Fatal("expected non-nil definitions")
	}
	if len(g.definitions) != 0 {
		t.Fatal("expected empty definitions")
	}
}

func TestNewSchemaGeneratorWithOptions(t *testing.T) {
	defs := map[string]*JSONSchema{"test": {Type: "object"}}
	g := NewSchemaGenerator(WithDefinitions(defs))
	if len(g.definitions) != 1 {
		t.Fatal("expected 1 definition")
	}
	if g.definitions["test"] == nil {
		t.Fatal("expected test definition")
	}
}

// --- SchemaFromType Tests ---

func TestSchemaFromType_String(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.SchemaFromType("string")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "string" {
		t.Errorf("expected type 'string', got '%s'", schema.Type)
	}
}

func TestSchemaFromType_Integer(t *testing.T) {
	g := NewSchemaGenerator()
	for _, typ := range []string{"int", "int64", "uint32"} {
		schema, err := g.SchemaFromType(typ)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", typ, err)
		}
		if schema.Type != "integer" {
			t.Errorf("%s: expected type 'integer', got '%s'", typ, schema.Type)
		}
	}
}

func TestSchemaFromType_Number(t *testing.T) {
	g := NewSchemaGenerator()
	for _, typ := range []string{"float32", "float64"} {
		schema, err := g.SchemaFromType(typ)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", typ, err)
		}
		if schema.Type != "number" {
			t.Errorf("%s: expected type 'number', got '%s'", typ, schema.Type)
		}
	}
}

func TestSchemaFromType_Bool(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.SchemaFromType("bool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "boolean" {
		t.Errorf("expected type 'boolean', got '%s'", schema.Type)
	}
}

func TestSchemaFromType_Array(t *testing.T) {
	g := NewSchemaGenerator()
	for _, typ := range []string{"[]string", "[]int", "[]bool"} {
		schema, err := g.SchemaFromType(typ)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", typ, err)
		}
		if schema.Type != "array" {
			t.Errorf("%s: expected type 'array', got '%s'", typ, schema.Type)
		}
		if schema.Items == nil {
			t.Fatal("expected Items to be set")
		}
	}
}

func TestSchemaFromType_Map(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.SchemaFromType("map[string]interface{}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", schema.Type)
	}
	if schema.AdditionalProperties == nil || !*schema.AdditionalProperties {
		t.Error("expected AdditionalProperties to be true")
	}
}

func TestSchemaFromType_Time(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.SchemaFromType("time.Time")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "string" {
		t.Errorf("expected type 'string', got '%s'", schema.Type)
	}
	if schema.Format != "date-time" {
		t.Errorf("expected format 'date-time', got '%s'", schema.Format)
	}
}

// --- GenerateFromValue Tests ---

func TestGenerateFromValue_Object(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(map[string]interface{}{"name": "test", "age": 25})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", schema.Type)
	}
	if schema.Properties == nil {
		t.Fatal("expected Properties to be set")
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("expected 'name' property")
	}
	if _, ok := schema.Properties["age"]; !ok {
		t.Error("expected 'age' property")
	}
}

func TestGenerateFromValue_EmptyObject(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", schema.Type)
	}
}

func TestGenerateFromValue_String(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "string" {
		t.Errorf("expected type 'string', got '%s'", schema.Type)
	}
	if schema.Default != "hello" {
		t.Errorf("expected default 'hello', got '%v'", schema.Default)
	}
}

func TestGenerateFromValue_Int(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "integer" {
		t.Errorf("expected type 'integer', got '%s'", schema.Type)
	}
	// Default is stored as interface{} — JSON number
	if schema.Default == nil {
		t.Error("expected non-nil default for int value")
	}
}

func TestGenerateFromValue_ZeroInt(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "integer" {
		t.Errorf("expected type 'integer', got '%s'", schema.Type)
	}
	if schema.Default != nil {
		t.Error("expected no default for zero value")
	}
}

func TestGenerateFromValue_Bool(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "boolean" {
		t.Errorf("expected type 'boolean', got '%s'", schema.Type)
	}
}

func TestGenerateFromValue_BoolFalse(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "boolean" {
		t.Errorf("expected type 'boolean', got '%s'", schema.Type)
	}
	if schema.Default != nil {
		t.Error("expected no default for false value")
	}
}

func TestGenerateFromValue_Float(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(3.14)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "number" {
		t.Errorf("expected type 'number', got '%s'", schema.Type)
	}
}

func TestGenerateFromValue_ArrayOfStrings(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue([]string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "array" {
		t.Errorf("expected type 'array', got '%s'", schema.Type)
	}
	if schema.Items == nil {
		t.Fatal("expected Items to be set")
	}
	if schema.Items.Type != "string" {
		t.Errorf("expected Items type 'string', got '%s'", schema.Items.Type)
	}
	if schema.MinItems == nil || *schema.MinItems != 3 {
		t.Errorf("expected MinItems 3, got %v", schema.MinItems)
	}
}

func TestGenerateFromValue_EmptyArray(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "array" {
		t.Errorf("expected type 'array', got '%s'", schema.Type)
	}
}

func TestGenerateFromValue_NestedObject(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateFromValue(map[string]interface{}{
		"user": map[string]interface{}{"name": "Alice", "active": true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Properties == nil {
		t.Fatal("expected Properties")
	}
	user, ok := schema.Properties["user"]
	if !ok {
		t.Fatal("expected 'user' property")
	}
	if user.Properties == nil {
		t.Fatal("expected user.Properties")
	}
	if _, ok := user.Properties["name"]; !ok {
		t.Error("expected user.name property")
	}
}

// --- GenerateSchema Tests (structs) ---

type TestUser struct {
	Name   string `json:"name" description:"User name" required:"true" pattern:"^[a-zA-Z]+$"`
	Age    int    `json:"age" minimum:"0" maximum:"150"`
	Active bool   `json:"active,omitempty"`
}

// Test struct from another "package" context using embedded approach
type TestPayload struct {
	Items []string `json:"items"`
	Count int      `json:"count"`
}

func TestGenerateSchema_InnerStructBehavior(t *testing.T) {
	// Structs from non-main packages get $ref behavior (named struct)
	g := NewSchemaGenerator()
	schema, err := g.GenerateSchema(TestUser{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Named struct: should have $ref in defs, not inline properties
	if schema.Defs == nil {
		t.Fatal("expected definitions for named struct")
	}
	if _, ok := schema.Defs["TestUser"]; !ok {
		t.Error("expected TestUser in definitions")
	}
}

func TestGenerateSchema_StructTags_Populated(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateSchema(TestPayload{Items: []string{"a", "b"}, Count: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TestPayload is named → goes to definitions. Check definition.
	if schema.Defs == nil {
		t.Fatal("expected definitions")
	}
	defSchema, ok := schema.Defs["TestPayload"]
	if !ok {
		t.Fatal("expected TestPayload in definitions")
	}
	if defSchema.Properties == nil {
		t.Fatal("expected Properties in TestPayload definition")
	}
	if _, ok := defSchema.Properties["items"]; !ok {
		t.Error("expected 'items' property in definition")
	}
	if _, ok := defSchema.Properties["count"]; !ok {
		t.Error("expected 'count' property in definition")
	}
}

// --- Named Struct with Definitions ---

type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
}

type Person struct {
	Name      string    `json:"name"`
	Addresses []Address `json:"addresses"`
}

func TestGenerateSchema_NamedStruct(t *testing.T) {
	g := NewSchemaGenerator()
	schema, err := g.GenerateSchema(Person{Name: "Bob", Addresses: []Address{{Street: "123 Main", City: "NYC"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Address is a named struct → should be in definitions
	if schema.Defs == nil {
		t.Fatal("expected definitions")
	}
	if _, ok := schema.Defs["Address"]; !ok {
		t.Error("expected Address definition")
	}
}

// --- RenderJSON Tests ---

func TestRenderJSON_Pretty(t *testing.T) {
	g := NewSchemaGenerator()
	schema := &JSONSchema{Type: "object", Title: "Test"}
	data, err := g.RenderJSON(schema, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "  ") {
		t.Error("expected pretty-printed output with indentation")
	}
	if !strings.Contains(string(data), "Test") {
		t.Error("expected title in output")
	}
}

func TestRenderJSON_Compact(t *testing.T) {
	g := NewSchemaGenerator()
	schema := &JSONSchema{Type: "string"}
	data, err := g.RenderJSON(schema, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), `"type":"string"`) {
		t.Error("expected compact output")
	}
}

// --- FromFile Tests ---

func TestGenerateSchemaFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.json")
	err := os.WriteFile(file, []byte(`{"name":"test","count":42}`), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	g := NewSchemaGenerator()
	schema, err := g.GenerateSchemaFromFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", schema.Type)
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("expected 'name' property")
	}
	if _, ok := schema.Properties["count"]; !ok {
		t.Error("expected 'count' property")
	}
}

func TestGenerateSchemaFromFile_NotFound(t *testing.T) {
	g := NewSchemaGenerator()
	_, err := g.GenerateSchemaFromFile("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestGenerateSchemaFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "invalid.json")
	os.WriteFile(file, []byte(`{not valid json}`), 0644)

	g := NewSchemaGenerator()
	_, err := g.GenerateSchemaFromFile(file)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- YAML Render Test ---

func TestRenderYAML(t *testing.T) {
	g := NewSchemaGenerator()
	schema := &JSONSchema{Type: "object", Properties: map[string]*JSONSchema{
		"name": {Type: "string", Default: "test"},
	}}
	data, err := g.RenderYAML(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "name") {
		t.Error("expected 'name' in YAML output")
	}
}

// --- boolPtr helper ---

func TestBoolPtr(t *testing.T) {
	p := boolPtr(true)
	if *p != true {
		t.Error("expected true")
	}
	p = boolPtr(false)
	if *p != false {
		t.Error("expected false")
	}
}

// --- JSON Round Trip ---

func TestJSONRoundTrip(t *testing.T) {
	g := NewSchemaGenerator()
	original := &JSONSchema{Type: "object", Title: "Test", Required: []string{"name"}}
	data, err := g.RenderJSON(original, true)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	var restored JSONSchema
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if restored.Type != "object" {
		t.Error("round trip: type mismatch")
	}
	if len(restored.Required) != 1 || restored.Required[0] != "name" {
		t.Error("round trip: required mismatch")
	}
	if restored.Title != "Test" {
		t.Error("round trip: title mismatch")
	}
}

// --- Edge Cases ---

func TestGenerateFromValue_TimeViaFromType(t *testing.T) {
	// time.Time is a named struct so GenerateFromValue wraps it in $ref;
	// test via SchemaFromType instead
	g := NewSchemaGenerator()
	schema, err := g.SchemaFromType("time.Time")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "string" {
		t.Errorf("expected type 'string', got '%s'", schema.Type)
	}
	if schema.Format != "date-time" {
		t.Errorf("expected format 'date-time', got '%s'", schema.Format)
	}
}

func TestGenerateSchema_NoErrorOnEmptyInput(t *testing.T) {
	g := NewSchemaGenerator()
	_, err := g.GenerateFromValue("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Struct Tag Handling (via schema inspection) ---

func TestStructWithEnumTag(t *testing.T) {
	type Status struct {
		Status string `json:"status" enum:"active,inactive,pending"`
	}
	g := NewSchemaGenerator()
	schema, err := g.GenerateSchema(Status{Status: "active"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Status is a named struct → in definitions
	if schema.Defs == nil {
		t.Fatal("expected definitions")
	}
	defSchema, ok := schema.Defs["Status"]
	if !ok {
		t.Fatal("expected Status in definitions")
	}
	if defSchema.Properties == nil {
		t.Fatal("expected properties in Status definition")
	}
	propSchema := defSchema.Properties["status"]
	if propSchema.Enum == nil || len(propSchema.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %v", propSchema.Enum)
	}
}

func TestStructWithFormatTag(t *testing.T) {
	type Email struct {
		Email string `json:"email" format:"email"`
	}
	g := NewSchemaGenerator()
	schema, err := g.GenerateSchema(Email{Email: "test@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Defs == nil {
		t.Fatal("expected definitions")
	}
	defSchema, ok := schema.Defs["Email"]
	if !ok {
		t.Fatal("expected Email in definitions")
	}
	propSchema := defSchema.Properties["email"]
	if propSchema.Format != "email" {
		t.Errorf("expected format 'email', got '%s'", propSchema.Format)
	}
}