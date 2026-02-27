package actions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJSONSchemaString_ValidJSON(t *testing.T) {
	schemaStr, err := GenerateJSONSchemaString()
	require.NoError(t, err)
	require.NotEmpty(t, schemaStr)

	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(schemaStr), &parsed)
	require.NoError(t, err, "schema must be valid JSON")
}

func TestGenerateJSONSchema_HasSchemaKey(t *testing.T) {
	schema := GenerateJSONSchema()
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema["$schema"])
}

func TestGenerateJSONSchema_HasAllDefs(t *testing.T) {
	schema := GenerateJSONSchema()
	defs, ok := schema["$defs"].(map[string]interface{})
	require.True(t, ok, "$defs must be a map")

	expectedDefs := []string{
		"StrField", "NetworkField", "AddressValue", "AddressField",
		"SignatureValue", "IntValue", "IntField", "StatusField",
		"Hex64", "AnyValue", "MapValue",
		"ContractValue", "AccountValue", "ParameterCondValue",
		"FunctionValue", "FunctionField",
		"EventEmittedValue", "EventEmittedField",
		"LogEmittedValue", "LogEmittedField",
		"StateChangedValue", "StateChangedField",
		"EthBalanceValue", "EthBalanceField",
		"TransactionFilter",
		"PeriodicTrigger", "WebhookTrigger", "BlockTrigger",
		"TransactionTrigger", "AlertTrigger",
		"TriggerUnparsed", "ActionSpec", "ProjectActions",
	}

	for _, name := range expectedDefs {
		_, exists := defs[name]
		assert.True(t, exists, "$defs should contain %s", name)
	}
}

func TestGenerateJSONSchema_RuntimeEnum(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	projectActions := defs["ProjectActions"].(map[string]interface{})
	props := projectActions["properties"].(map[string]interface{})
	runtime := props["runtime"].(map[string]interface{})
	enumVals := runtime["enum"].([]interface{})

	require.Len(t, enumVals, len(SupportedRuntimes))
	for i, v := range enumVals {
		assert.Equal(t, SupportedRuntimes[i], v)
	}
}

func TestGenerateJSONSchema_TriggerTypeEnum(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	trigger := defs["TriggerUnparsed"].(map[string]interface{})
	props := trigger["properties"].(map[string]interface{})
	typeField := props["type"].(map[string]interface{})
	enumVals := typeField["enum"].([]interface{})

	require.Len(t, enumVals, len(TriggerTypes))
	for i, v := range enumVals {
		assert.Equal(t, TriggerTypes[i], v)
	}
}

func TestGenerateJSONSchema_IntervalEnum(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	periodic := defs["PeriodicTrigger"].(map[string]interface{})
	oneOf := periodic["oneOf"].([]interface{})

	// First oneOf entry should have interval with enum
	intervalOption := oneOf[0].(map[string]interface{})
	props := intervalOption["properties"].(map[string]interface{})
	interval := props["interval"].(map[string]interface{})
	enumVals := interval["enum"].([]interface{})

	require.Len(t, enumVals, len(Intervals))
	for i, v := range enumVals {
		assert.Equal(t, Intervals[i], v)
	}
}

func TestGenerateJSONSchema_AddressPattern(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	addrValue := defs["AddressValue"].(map[string]interface{})
	assert.Equal(t, AddressRegex, addrValue["pattern"])
}
