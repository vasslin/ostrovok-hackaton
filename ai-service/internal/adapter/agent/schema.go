package agent

import (
	"reflect"
	"strings"
)

func outputSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Pointer {
		return outputSchema(t.Elem())
	}
	switch t.Kind() {
	case reflect.Struct:
		properties := map[string]any{}
		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if !field.IsExported() || name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			if t.Name() == "PlanningConfig" && (name == "request" || name == "resolved") {
				continue
			}
			schema := outputSchema(field.Type)
			switch name {
			case "currency":
				schema["pattern"] = "^[A-Z]{3}$"
			case "departure_date", "return_date", "start_date", "end_date":
				schema["pattern"] = "^[0-9]{4}-[0-9]{2}-[0-9]{2}$"
			case "adults":
				schema["minimum"] = 1
				schema["maximum"] = 20
			case "score":
				schema["minimum"] = 0
				schema["maximum"] = 100
			case "budget", "amount", "max_price_per_night":
				if schema["type"] == "number" {
					schema["exclusiveMinimum"] = 0
				}
			case "status":
				schema["enum"] = []string{"advice", "questions", "generate"}
			case "theme":
				schema["enum"] = []string{"culture", "nature", "relax", "active"}
			case "source":
				schema["enum"] = []string{"ai_generated_idea"}
			case "version":
				schema["enum"] = []string{"1.0"}
			case "scope":
				schema["enum"] = []string{"total", "per_person", "per_night"}
			case "country_code":
				schema["pattern"] = "^[A-Z]{2}$"
			case "input_type":
				schema["enum"] = []string{"text", "date", "number"}
			case "advice":
				schema["maxItems"] = 3
			case "questions":
				schema["maxItems"] = 2
			}
			if field.Type.Kind() == reflect.Pointer && name != "config" && name != "score" {
				schema = map[string]any{"anyOf": []any{schema, map[string]string{"type": "null"}}}
			}
			if t.Name() == "planningCompletion" && name == "advice" {
				schema["maxItems"] = 0
			}
			properties[name] = schema
			optionalPlanning := false
			if t.Name() == "planningCompletion" && (name == "summary" || name == "warnings" || name == "assumptions") {
				optionalPlanning = true
			}
			if !optionalPlanning {
				required = append(required, name)
			}
		}
		return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": outputSchema(t.Elem())}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	default:
		return map[string]any{}
	}
}
