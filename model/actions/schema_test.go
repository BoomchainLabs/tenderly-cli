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
		"TransactionStatusField", "Hex64",
		"ContractValue", "AddressOnlyContractValue", "ParameterCondValue",
		"FunctionValue", "FunctionField",
		"EventEmittedValue", "EventEmittedField",
		"LogEmittedValue", "LogEmittedField",
		"TransactionFilter",
		"PeriodicTrigger", "WebhookTrigger", "BlockTrigger",
		"TransactionTrigger", "AlertTrigger",
		"TriggerUnparsed", "ActionSpec", "ProjectActions",
	}

	// Unsupported features should NOT be in schema
	removedDefs := []string{
		"AnyValue", "MapValue", "AccountValue",
		"StateChangedValue", "StateChangedField",
		"EthBalanceValue", "EthBalanceField",
	}
	for _, name := range removedDefs {
		_, exists := defs[name]
		assert.False(t, exists, "$defs should NOT contain %s (unsupported)", name)
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

	// Properties should be at top level (not inside oneOf) to work with additionalProperties
	props := periodic["properties"].(map[string]interface{})
	interval := props["interval"].(map[string]interface{})
	enumVals := interval["enum"].([]interface{})

	require.Len(t, enumVals, len(Intervals))
	for i, v := range enumVals {
		assert.Equal(t, Intervals[i], v)
	}

	// oneOf should only contain required constraints
	oneOf := periodic["oneOf"].([]interface{})
	require.Len(t, oneOf, 2)
	first := oneOf[0].(map[string]interface{})
	assert.Equal(t, []interface{}{"interval"}, first["required"])
	second := oneOf[1].(map[string]interface{})
	assert.Equal(t, []interface{}{"cron"}, second["required"])
}

func TestGenerateJSONSchema_PeriodicCronHasPattern(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	periodic := defs["PeriodicTrigger"].(map[string]interface{})
	props := periodic["properties"].(map[string]interface{})
	cronField := props["cron"].(map[string]interface{})
	assert.Equal(t, CronPattern, cronField["pattern"])
}

func TestGenerateJSONSchema_AddressPattern(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	addrValue := defs["AddressValue"].(map[string]interface{})
	assert.Equal(t, AddressRegexCI, addrValue["pattern"], "schema should use case-insensitive address regex")
}

func TestGenerateJSONSchema_FunctionValueRequiresContract(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	fn := defs["FunctionValue"].(map[string]interface{})
	required := fn["required"].([]interface{})
	assert.Contains(t, required, "contract")
}

func TestGenerateJSONSchema_FunctionValueOneOfSignatureOrName(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	fn := defs["FunctionValue"].(map[string]interface{})
	oneOf := fn["oneOf"].([]interface{})
	require.Len(t, oneOf, 2)
}

func TestGenerateJSONSchema_EventEmittedRequiresContract(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	ev := defs["EventEmittedValue"].(map[string]interface{})
	required := ev["required"].([]interface{})
	assert.Contains(t, required, "contract")
}

func TestGenerateJSONSchema_EventEmittedOneOfIdOrName(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	ev := defs["EventEmittedValue"].(map[string]interface{})
	oneOf := ev["oneOf"].([]interface{})
	require.Len(t, oneOf, 2)
}

func TestGenerateJSONSchema_StatusFieldEnum(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	sf := defs["StatusField"].(map[string]interface{})
	oneOf := sf["oneOf"].([]interface{})
	// First branch is the single value with enum
	single := oneOf[0].(map[string]interface{})
	enumVals := single["enum"].([]interface{})
	assert.Equal(t, []interface{}{"success", "fail"}, enumVals)
}

func TestGenerateJSONSchema_TransactionStatusFieldEnum(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	tsf := defs["TransactionStatusField"].(map[string]interface{})
	oneOf := tsf["oneOf"].([]interface{})
	single := oneOf[0].(map[string]interface{})
	enumVals := single["enum"].([]interface{})
	assert.Equal(t, []interface{}{"mined", "confirmed10"}, enumVals)
}

func TestGenerateJSONSchema_TransactionFilterRequiresNetwork(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	tf := defs["TransactionFilter"].(map[string]interface{})
	required := tf["required"].([]interface{})
	assert.Contains(t, required, "network")
}

func TestGenerateJSONSchema_TransactionFilterNoUnsupportedFields(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	tf := defs["TransactionFilter"].(map[string]interface{})
	props := tf["properties"].(map[string]interface{})
	_, hasEthBalance := props["ethBalance"]
	_, hasStateChanged := props["stateChanged"]
	assert.False(t, hasEthBalance, "ethBalance should not be in TransactionFilter (not yet supported)")
	assert.False(t, hasStateChanged, "stateChanged should not be in TransactionFilter (not yet supported)")
}

func TestGenerateJSONSchema_LogEmittedContractNoInvocation(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	aocv := defs["AddressOnlyContractValue"].(map[string]interface{})
	props := aocv["properties"].(map[string]interface{})
	_, hasInvocation := props["invocation"]
	assert.False(t, hasInvocation, "LogEmitted contract should not have invocation")
}

func TestGenerateJSONSchema_FunctionValueHasParameters(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	fn := defs["FunctionValue"].(map[string]interface{})
	props := fn["properties"].(map[string]interface{})
	params, hasParams := props["parameters"]
	assert.True(t, hasParams, "FunctionValue should have parameters field")
	paramsObj := params.(map[string]interface{})
	assert.Equal(t, "array", paramsObj["type"])
}

func TestGenerateJSONSchema_ParameterCondStringSupportsNotForm(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	pcv := defs["ParameterCondValue"].(map[string]interface{})
	props := pcv["properties"].(map[string]interface{})
	strField := props["string"].(map[string]interface{})
	oneOf := strField["oneOf"].([]interface{})
	require.Len(t, oneOf, 2, "string field should support plain string OR {exact, not} object")

	// First branch: plain string
	assert.Equal(t, "string", oneOf[0].(map[string]interface{})["type"])
	// Second branch: object with exact/not
	objBranch := oneOf[1].(map[string]interface{})
	assert.Equal(t, "object", objBranch["type"])
	objProps := objBranch["properties"].(map[string]interface{})
	_, hasExact := objProps["exact"]
	_, hasNot := objProps["not"]
	assert.True(t, hasExact)
	assert.True(t, hasNot)
}

func TestGenerateJSONSchema_TransactionFilterMinConstraint(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	tf := defs["TransactionFilter"].(map[string]interface{})
	anyOf := tf["anyOf"].([]interface{})
	require.Len(t, anyOf, 5, "anyOf should require at least one of from/to/function/eventEmitted/logEmitted")

	expectedFields := []string{"from", "to", "function", "eventEmitted", "logEmitted"}
	for i, entry := range anyOf {
		branch := entry.(map[string]interface{})
		required := branch["required"].([]interface{})
		assert.Equal(t, expectedFields[i], required[0])
	}
}

func TestGenerateJSONSchema_ExecutionTypeOptional(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	action := defs["ActionSpec"].(map[string]interface{})
	required := action["required"].([]interface{})
	assert.Contains(t, required, "function")
	assert.Contains(t, required, "trigger")
	assert.NotContains(t, required, "execution_type", "execution_type should be optional (defaults to sequential)")
}

func TestGenerateJSONSchema_ActionNamePattern(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	pa := defs["ProjectActions"].(map[string]interface{})
	props := pa["properties"].(map[string]interface{})
	specs := props["specs"].(map[string]interface{})
	pp := specs["patternProperties"].(map[string]interface{})
	_, hasPattern := pp[ActionNamePattern]
	assert.True(t, hasPattern, "specs should validate action names with patternProperties")
	assert.Equal(t, false, specs["additionalProperties"], "specs should reject invalid action names")
}

func TestGenerateJSONSchema_Hex64Pattern(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	hex64 := defs["Hex64"].(map[string]interface{})
	oneOf := hex64["oneOf"].([]interface{})
	strBranch := oneOf[0].(map[string]interface{})
	assert.Equal(t, "^0x[0-9a-fA-F]+$", strBranch["pattern"], "Hex64 should validate hex characters")
}

func TestGenerateJSONSchema_SignaturePatternCaseInsensitive(t *testing.T) {
	schema := GenerateJSONSchema()
	defs := schema["$defs"].(map[string]interface{})
	sig := defs["SignatureValue"].(map[string]interface{})
	oneOf := sig["oneOf"].([]interface{})
	strBranch := oneOf[0].(map[string]interface{})
	assert.Equal(t, SigRegexCI, strBranch["pattern"], "schema should use case-insensitive sig regex")
}
