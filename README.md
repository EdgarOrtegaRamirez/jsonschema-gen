# JSON Schema Gen

Generate JSON Schema from JSON data, files, or Go type names. Automate API contract generation and validate data structures.

## Features

- **Generate from stdin** — Pipe JSON and get back a full JSON Schema
- **Generate from file** — Pass a JSON file path to introspect its schema
- **Generate from type** — Get schema for primitive Go types (string, int, bool, arrays, maps, time.Time)
- **Struct tag support** — Parse json, description, required, pattern, enum, format, minimum, maximum, minLength, maxLength tags
- **Definition reuse** — $defs support for recursive and complex structures
- **YAML output** — Convert to YAML for OpenAPI integration

## Usage

### Generate from stdin

```bash
echo '{"name":"test","age":25}' | jsonschema-gen generate
```

**Output:**
```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "default": "test"
    },
    "age": {
      "type": "integer",
      "default": 25
    }
  },
  "$defs": {}
}
```

### Generate from file

```bash
jsonschema-gen from-file ./data.json
```

### Generate from type

```bash
# Get schema for a string
jsonschema-gen from-type string

# Get schema for an array of strings
jsonschema-gen from-type "[]string"

# Get schema for a map
jsonschema-gen from-type "map[string]interface{}"
```

### Write to file

```bash
cat data.json | jsonschema-gen generate --output schema.json
jsonschema-gen from-file ./data.json --output schema.yaml
```

## Struct Tags (Go)

When generating from Go structs, these struct tags are respected:

| Tag | Purpose | Example |
|-----|---------|---------|
| `json` | Field name | `json:"user_name,omitempty"` |
| `description` | Schema description | `description:"User's display name"` |
| `required` | Make field required | `required:"true"` |
| `pattern` | Regex pattern | `pattern:"^[a-zA-Z]+$"` |
| `format` | Format hint | `format:"email"` |
| `enum` | Allowed values | `enum:"active,inactive,pending"` |
| `example` | Example value | `example:"john@example.com"` |
| `minimum` | Min value | `minimum:"0"` |
| `maximum` | Max value | `maximum:"100"` |
| `minLength` | Min string length | `minLength:"1"` |
| `maxLength` | Max string length | `maxLength:"255"` |

## Output Formats

### JSON (default)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {}
}
```

### YAML

For YAML output, pipe through `yq` or use a YAML-capable viewer:

```bash
jsonschema-gen generate | yq -P
```

## Integration

### OpenAPI/Swagger

```bash
# Generate schema for API response
cat api-response.json | jsonschema-gen generate > response-schema.json
```

### CI/CD Validation

```bash
# Validate JSON data against generated schema
echo '{"name":"test"}' | jsonschema-gen generate | jsonschema -d -f draft-07 data.json
```

## Installation

### Go

```bash
go install github.com/EdgarOrtegaRamirez/jsonschema-gen@latest
```

### Manual

```bash
git clone https://github.com/EdgarOrtegaRamirez/jsonschema-gen.git
cd jsonschema-gen
go build -o jsonschema-gen .
sudo mv jsonschema-gen /usr/local/bin/
```

## License

MIT © Edgar Ortega Ramirez
