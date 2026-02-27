package actions

import (
	"encoding/json"
)

// GenerateJSONSchema returns the JSON Schema for tenderly.yaml as a map.
func GenerateJSONSchema() map[string]interface{} {
	schema := obj(
		"$schema", "https://json-schema.org/draft/2020-12/schema",
		"title", "Tenderly Actions Configuration",
		"description", "Schema for tenderly.yaml Web3 Actions configuration",
		"type", "object",
		"properties", obj(
			"actions", obj(
				"type", "object",
				"description", "Map of project slug to project actions configuration",
				"additionalProperties", refDef("ProjectActions"),
			),
		),
		"additionalProperties", true,
		"$defs", buildDefs(),
	)
	return schema
}

// GenerateJSONSchemaString returns the JSON Schema as a pretty-printed JSON string.
func GenerateJSONSchemaString() (string, error) {
	schema := GenerateJSONSchema()
	bytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func buildDefs() map[string]interface{} {
	defs := map[string]interface{}{
		// Primitive types
		"StrField":       defStrField(),
		"NetworkField":   defNetworkField(),
		"AddressValue":   defAddressValue(),
		"AddressField":   defAddressField(),
		"SignatureValue": defSignatureValue(),
		"IntValue":       defIntValue(),
		"IntField":       defIntField(),
		"StatusField":    defStatusField(),
		"Hex64":          defHex64(),
		"AnyValue":       defAnyValue(),
		"MapValue":       defMapValue(),

		// Composite types
		"ContractValue":      defContractValue(),
		"AccountValue":       defAccountValue(),
		"ParameterCondValue": defParameterCondValue(),
		"FunctionValue":      defFunctionValue(),
		"FunctionField":      defFunctionField(),
		"EventEmittedValue":  defEventEmittedValue(),
		"EventEmittedField":  defEventEmittedField(),
		"LogEmittedValue":    defLogEmittedValue(),
		"LogEmittedField":    defLogEmittedField(),
		"StateChangedValue":  defStateChangedValue(),
		"StateChangedField":  defStateChangedField(),
		"EthBalanceValue":    defEthBalanceValue(),
		"EthBalanceField":    defEthBalanceField(),
		"TransactionFilter":  defTransactionFilter(),

		// Trigger types
		"PeriodicTrigger":    defPeriodicTrigger(),
		"WebhookTrigger":     defWebhookTrigger(),
		"BlockTrigger":       defBlockTrigger(),
		"TransactionTrigger": defTransactionTrigger(),
		"AlertTrigger":       defAlertTrigger(),

		// Top-level
		"TriggerUnparsed": defTriggerUnparsed(),
		"ActionSpec":      defActionSpec(),
		"ProjectActions":  defProjectActions(),
	}
	return defs
}

// --- Primitive type definitions ---

func defStrField() map[string]interface{} {
	return singleOrArray(obj("type", "string"))
}

func defNetworkField() map[string]interface{} {
	strOrInt := obj(
		"oneOf", arr(
			obj("type", "string"),
			obj("type", "integer"),
		),
	)
	return obj(
		"oneOf", arr(
			obj("type", "string"),
			obj("type", "integer"),
			obj(
				"type", "array",
				"items", strOrInt,
			),
		),
	)
}

func defAddressValue() map[string]interface{} {
	return obj(
		"type", "string",
		"pattern", AddressRegex,
	)
}

func defAddressField() map[string]interface{} {
	return singleOrArray(refDef("AddressValue"))
}

func defSignatureValue() map[string]interface{} {
	return obj(
		"oneOf", arr(
			obj("type", "string", "pattern", SigRegex),
			obj("type", "integer"),
		),
	)
}

func defIntValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"gte", obj("type", "integer"),
			"lte", obj("type", "integer"),
			"eq", obj("type", "integer"),
			"gt", obj("type", "integer"),
			"lt", obj("type", "integer"),
			"not", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defIntField() map[string]interface{} {
	return singleOrArray(refDef("IntValue"))
}

func defStatusField() map[string]interface{} {
	return singleOrArray(obj("type", "string"))
}

func defHex64() map[string]interface{} {
	return obj(
		"oneOf", arr(
			obj("type", "string", "pattern", "^0x"),
			obj("type", "integer"),
		),
	)
}

func defAnyValue() map[string]interface{} {
	return obj(
		"oneOf", arr(
			obj("type", "string"),
			refDef("IntValue"),
			refDef("MapValue"),
		),
	)
}

func defMapValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"key", obj("type", "string"),
			"value", refDef("AnyValue"),
		),
		"required", arr("key", "value"),
		"additionalProperties", false,
	)
}

// --- Composite type definitions ---

func defContractValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"address", refDef("AddressValue"),
			"invocation", obj(
				"type", "string",
				"enum", toInterfaceSlice(Invocations),
			),
		),
		"required", arr("address"),
		"additionalProperties", false,
	)
}

func defAccountValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"address", refDef("AddressValue"),
		),
		"required", arr("address"),
		"additionalProperties", false,
	)
}

func defParameterCondValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"name", obj("type", "string"),
			"string", obj("type", "string"),
			"int", refDef("IntValue"),
		),
		"required", arr("name"),
		"additionalProperties", false,
	)
}

func defFunctionValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"contract", refDef("ContractValue"),
			"signature", refDef("SignatureValue"),
			"name", obj("type", "string"),
			"parameter", refDef("MapValue"),
			"not", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defFunctionField() map[string]interface{} {
	return singleOrArray(refDef("FunctionValue"))
}

func defEventEmittedValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"contract", refDef("ContractValue"),
			"id", obj("type", "string"),
			"name", obj("type", "string"),
			"parameters", obj(
				"type", "array",
				"items", refDef("ParameterCondValue"),
			),
			"not", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defEventEmittedField() map[string]interface{} {
	return singleOrArray(refDef("EventEmittedValue"))
}

func defLogEmittedValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"startsWith", obj(
				"type", "array",
				"items", refDef("Hex64"),
				"minItems", 1,
			),
			"contract", refDef("ContractValue"),
			"matchAny", obj("type", "boolean"),
			"not", obj("type", "boolean"),
		),
		"required", arr("startsWith"),
		"additionalProperties", false,
	)
}

func defLogEmittedField() map[string]interface{} {
	return singleOrArray(refDef("LogEmittedValue"))
}

func defStateChangedValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"contract", refDef("ContractValue"),
			"key", obj("type", "string"),
			"field", obj("type", "string"),
			"value", refDef("AnyValue"),
			"previousValue", refDef("AnyValue"),
		),
		"additionalProperties", false,
	)
}

func defStateChangedField() map[string]interface{} {
	return singleOrArray(refDef("StateChangedValue"))
}

func defEthBalanceValue() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"value", refDef("IntValue"),
			"account", refDef("AccountValue"),
			"contract", refDef("ContractValue"),
		),
		"required", arr("value"),
		"additionalProperties", false,
	)
}

func defEthBalanceField() map[string]interface{} {
	return singleOrArray(refDef("EthBalanceValue"))
}

func defTransactionFilter() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"network", refDef("NetworkField"),
			"status", refDef("StatusField"),
			"from", refDef("AddressField"),
			"to", refDef("AddressField"),
			"value", refDef("IntField"),
			"gasLimit", refDef("IntField"),
			"gasUsed", refDef("IntField"),
			"fee", refDef("IntField"),
			"contract", refDef("ContractValue"),
			"function", refDef("FunctionField"),
			"eventEmitted", refDef("EventEmittedField"),
			"logEmitted", refDef("LogEmittedField"),
			"ethBalance", refDef("EthBalanceField"),
			"stateChanged", refDef("StateChangedField"),
		),
		"additionalProperties", false,
	)
}

// --- Trigger definitions ---

func defPeriodicTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"oneOf", arr(
			obj(
				"properties", obj(
					"interval", obj(
						"type", "string",
						"enum", toInterfaceSlice(Intervals),
					),
				),
				"required", arr("interval"),
			),
			obj(
				"properties", obj(
					"cron", obj("type", "string"),
				),
				"required", arr("cron"),
			),
		),
		"additionalProperties", false,
	)
}

func defWebhookTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"authenticated", obj("type", "boolean"),
		),
		"additionalProperties", false,
	)
}

func defBlockTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"network", refDef("NetworkField"),
			"blocks", obj(
				"type", "integer",
				"minimum", 1,
			),
		),
		"required", arr("network", "blocks"),
		"additionalProperties", false,
	)
}

func defTransactionTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"status", refDef("StatusField"),
			"filters", obj(
				"type", "array",
				"items", refDef("TransactionFilter"),
				"minItems", 1,
			),
		),
		"required", arr("status", "filters"),
		"additionalProperties", false,
	)
}

func defAlertTrigger() map[string]interface{} {
	return obj(
		"type", "object",
		"additionalProperties", true,
	)
}

// --- Top-level definitions ---

func defTriggerUnparsed() map[string]interface{} {
	triggerSchema := obj(
		"type", "object",
		"properties", obj(
			"type", obj(
				"type", "string",
				"enum", toInterfaceSlice(TriggerTypes),
			),
			"periodic", refDef("PeriodicTrigger"),
			"webhook", refDef("WebhookTrigger"),
			"block", refDef("BlockTrigger"),
			"transaction", refDef("TransactionTrigger"),
			"alert", refDef("AlertTrigger"),
		),
		"required", arr("type"),
		"allOf", arr(
			ifThenTrigger("periodic"),
			ifThenTrigger("webhook"),
			ifThenTrigger("block"),
			ifThenTrigger("transaction"),
		),
		"additionalProperties", false,
	)
	return triggerSchema
}

func defActionSpec() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"description", obj("type", "string"),
			"function", obj(
				"type", "string",
				"pattern", "^.+:.+$",
				"description", "Entry point in the format file:functionName",
			),
			"execution_type", obj(
				"type", "string",
				"enum", arr(ParallelExecutionType, SequentialExecutionType),
			),
			"trigger", refDef("TriggerUnparsed"),
		),
		"required", arr("function", "trigger", "execution_type"),
		"additionalProperties", false,
	)
}

func defProjectActions() map[string]interface{} {
	return obj(
		"type", "object",
		"properties", obj(
			"runtime", obj(
				"type", "string",
				"enum", toInterfaceSlice(SupportedRuntimes),
			),
			"sources", obj("type", "string"),
			"dependencies", obj("type", "string"),
			"specs", obj(
				"type", "object",
				"description", "Map of action name to action spec",
				"additionalProperties", refDef("ActionSpec"),
			),
		),
		"required", arr("runtime", "sources", "specs"),
		"additionalProperties", false,
	)
}

// --- Helpers ---

// singleOrArray produces a oneOf schema: either a single item or an array of items.
func singleOrArray(itemSchema map[string]interface{}) map[string]interface{} {
	return obj(
		"oneOf", arr(
			itemSchema,
			obj(
				"type", "array",
				"items", itemSchema,
			),
		),
	)
}

// ifThenTrigger creates an if/then block: if type == triggerName, then triggerName is required.
func ifThenTrigger(triggerName string) map[string]interface{} {
	return obj(
		"if", obj(
			"properties", obj(
				"type", obj("const", triggerName),
			),
		),
		"then", obj(
			"required", arr(triggerName),
		),
	)
}

// refDef creates a $ref to a definition in $defs.
func refDef(name string) map[string]interface{} {
	return obj("$ref", "#/$defs/"+name)
}

// obj builds a map from alternating key-value pairs.
func obj(kvs ...interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(kvs)/2)
	for i := 0; i < len(kvs)-1; i += 2 {
		m[kvs[i].(string)] = kvs[i+1]
	}
	return m
}

// arr builds a slice of interface{}.
func arr(items ...interface{}) []interface{} {
	return items
}

// toInterfaceSlice converts a []string to []interface{} for JSON marshaling.
func toInterfaceSlice(ss []string) []interface{} {
	result := make([]interface{}, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}
