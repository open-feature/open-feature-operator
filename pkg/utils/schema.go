package utils

/*
This is a temporary reference of the Open Feature schema
https://github.com/open-feature/playground/blob/main/schemas/flag.schema.json
*/
func GetSchema() string {
	return `
	{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id": "https://openfeature.dev/flag.schema.json",
		"title": "OpenFeature Feature Flags",
		"type": "object",
		"patternProperties": {
		  "^[A-Za-z]+$": {
			"description": "The flag key that uniquely represents the flag.",
			"type": "object",
			"properties": {
			  "name": {
				"type": "string"
			  },
			  "description": {
				"type": "string"
			  },
			  "returnType": {
				"type": "string",
				"enum": ["boolean", "string", "number", "object"],
				"default": "boolean"
			  },
			  "variants": {
				"type": "object",
				"patternProperties": {
				  "^[A-Za-z]+$": {
					"properties": {
					  "value": {
						"type": ["string", "number", "boolean", "object"]
					  }
					}
				  },
				  "additionalProperties": false
				},
				"minProperties": 2,
				"default": { "enabled": true, "disabled": false }
			  },
			  "defaultVariant": {
				"type": "string",
				"default": "enabled"
			  },
			  "state": {
				"type": "string",
				"enum": ["enabled", "disabled"],
				"default": "enabled"
			  },
			  "rules": {
				"type": "array",
				"items": {
				  "$ref": "#/$defs/rule"
				},
				"default": []
			  }
			},
			"required": ["state"],
			"additionalProperties": false
		  }
		},
		"additionalProperties": false,
	  
		"$defs": {
		  "rule": {
			"type": "object",
			"description": "A rule that ",
			"properties": {
			  "action": {
				"description": "The action that should be taken if at least one condition evaluates to true.",
				"type": "object",
				"properties": {
				  "variant": {
					"type": "string",
					"description": "The variant that should be return if one of the conditions evaluates to true."
				  }
				},
				"required": ["variant"],
				"additionalProperties": false
			  },
			  "conditions": {
				"type": "array",
				"description": "The conditions that should that be evaluated.",
				"items": {
				  "type": "object",
				  "properties": {
					"context": {
					  "type": "string",
					  "description": "The context key that should be evaluated in this condition"
					},
					"op": {
					  "type": "string",
					  "description": "The operation that should be performed",
					  "enum": ["equals", "starts_with", "ends_with"]
					},
					"value": {
					  "type": "string",
					  "description": "The value that should be evaluated"
					}
				  },
				  "required": ["context", "op", "value"],
				  "additionalProperties": false
				}
			  }
			},
			"required": ["action", "conditions"],
			"additionalProperties": false
		  }
		}
	  }
	`
}
