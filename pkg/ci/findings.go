// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package ci

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

type Finding struct {
	ToolName  string `json:"toolName"`
	RuleID    string `json:"ruleID"`
	Level     string `json:"level"`
	FilePath  string `json:"filePath"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	StartCol  int    `json:"startCol"`
	EndCol    int    `json:"endCol"`
	Message   string `json:"message"`
}

type JSONToFindingsMappings struct {
	// Key containing the array of findings in which to search using jsonMappingInfo
	BaseArrayKey string
	ToolName     JSONMappingInfo
	RuleID       JSONMappingInfo
	// Severity level of the finding, valid values are `debug`, `notice`, `warning`, `error`
	// Based on GitHub workflow commands, see https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#setting-a-debug-message
	// Other common values are mapped automatically:
	// `info` -> `notice`
	// `low` -> `notice`
	// `medium` -> `warning`
	// `critical` -> `error`
	// `high` -> `error`
	Level     JSONMappingInfo
	FilePath  JSONMappingInfo
	StartLine JSONMappingInfo
	EndLine   JSONMappingInfo
	StartCol  JSONMappingInfo
	EndCol    JSONMappingInfo
	Message   JSONMappingInfo
}

type JSONMappingInfo struct {
	// JSON key to find, can use dot notation to find nested keys like `foo.bar.baz`. Keys whose value is a string array will be
	// converted to a string by joining the values with " - ".
	Key string
	// Value to use if the key is not found or is empty. Internally uses strconv.Atoi to convert the value to an int if mapping type is int
	DefaultValue string
	// Do not try to find key and use this value instead
	OverrideValue string
	// Transform found value using this regex
	ValueTransformerRegex string
	// Discard whole finding if this regex for current key does not match
	GlobalSelectorRegex string
	// Select if the regex does not match (kind-of global negative lookahead, not supported in Go)
	InvertGlobalSelector bool
	// Another jsonMappingInfo to use as a suffix
	// Can be used to compose a value from multiple sources, like <mapping1 result><mapping2 result><mapping3 result>
	// Use in conjunction with OverrideKey to set constant values
	// Only enabled for string values
	Suffix *JSONMappingInfo
}

type JSONInfos struct {
	// Type or linter output to parse
	// Default is handling json array of findings
	// "plain": Handle plain text output, where any output is considered a finding, with such finding
	// being populated with the OverrideKey values from the jsonMappingInfo
	// "stream": Handle JSON stream output, internally converted to simple JSON array
	// "none": Do not parse output, do not treat it as finding
	Type string
	// Whether to read the JSON from stderr instead of stdout
	ReadFromStderr bool
	Mappings       JSONToFindingsMappings
}

var (
	ErrCantConvert         = fmt.Errorf("error converting value to a valid type")
	ErrCantConvertToInt    = fmt.Errorf("error converting value to int")
	ErrCantConvertToString = fmt.Errorf("error converting value to string")
	ErrCantCompileRegex    = fmt.Errorf("error compiling regex")
	ErrUnsupportedType     = fmt.Errorf("unsupported type")
	ErrJSONNotArray        = fmt.Errorf("json is not an array")
	ErrJSONNotMap          = fmt.Errorf("json is not a map")
	ErrJSONNoSuchKey       = fmt.Errorf("json does not contain key")
)

func FindingsFromJSON(str string, jsonInfo JSONInfos) ([]Finding, error) {
	if str == "" {
		return nil, nil
	}

	switch jsonInfo.Type {
	case "stream":
		strAsArray := ""

		for _, line := range strings.Split(str, "\n") {
			if strings.TrimSpace(line) != "" {
				strAsArray += line + ","
			}
		}

		strAsArray = "[" + strings.TrimSuffix(strAsArray, ",") + "]"
		str = strAsArray
	case "object":
		str = "[" + str + "]"
	}

	var jsonm any

	err := json.Unmarshal([]byte(str), &jsonm)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %w", err)
	}

	if jsonInfo.Mappings.BaseArrayKey != "" {
		njsonm, ok := jsonm.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"json does not contain key %s: %w",
				jsonInfo.Mappings.BaseArrayKey,
				ErrJSONNoSuchKey,
			)
		}

		jsonm = njsonm[jsonInfo.Mappings.BaseArrayKey]
	}

	jsonArray, ok := jsonm.([]any)
	if !ok {
		return nil, ErrJSONNotArray
	}

	var findings []Finding

	for _, item := range jsonArray {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, ErrJSONNotArray
		}

		keep, finding, err := findingFromJSONObject(m, jsonInfo.Mappings)
		if err != nil {
			return nil, fmt.Errorf("error parsing json object: %w", err)
		}

		if keep {
			findings = append(findings, finding)
		}
	}

	return findings, nil
}

func findingFromJSONObject(
	jsonm map[string]any,
	mappings JSONToFindingsMappings,
) (bool, Finding, error) {
	var finding Finding

	mappingFields := []struct {
		mapping JSONMappingInfo
		field   any
	}{
		{mappings.ToolName, &finding.ToolName},
		{mappings.RuleID, &finding.RuleID},
		{mappings.Level, &finding.Level},
		{mappings.FilePath, &finding.FilePath},
		{mappings.StartLine, &finding.StartLine},
		{mappings.EndLine, &finding.EndLine},
		{mappings.StartCol, &finding.StartCol},
		{mappings.EndCol, &finding.EndCol},
		{mappings.Message, &finding.Message},
	}

	for _, mf := range mappingFields {
		shouldKeep, err := setValue(jsonm, mf.mapping, mf.field)
		if err != nil {
			return false, finding, fmt.Errorf("error setting value for %v: %w", mf.mapping.Key, err)
		}

		if !shouldKeep {
			return false, Finding{
				ToolName:  "",
				RuleID:    "",
				Level:     "",
				FilePath:  "",
				StartLine: 0,
				EndLine:   0,
				StartCol:  0,
				EndCol:    0,
				Message:   "",
			}, nil
		}
	}

	finding.Level = strings.ToLower(finding.Level)

	return true, finding, nil
}

func applyGlobalSelector(jsonm map[string]any, mapInfo JSONMappingInfo) (bool, error) {
	if jsonm[mapInfo.Key] == nil {
		return false, nil
	}

	if s, ok := jsonm[mapInfo.Key].(string); ok {
		if s == "" {
			return false, nil
		}

		res, err := globalSelectFromRegex(s, mapInfo.GlobalSelectorRegex)
		if err != nil {
			return false, fmt.Errorf("error applying global selector regex: %w", err)
		}

		if mapInfo.InvertGlobalSelector {
			res = !res
		}

		if !res {
			return false, nil
		}
	}

	return true, nil
}

func setValue(
	jsonm map[string]any,
	mapInfo JSONMappingInfo,
	field any,
) (bool, error) {
	jsonTargetKey := jsonm

	if strings.Contains(mapInfo.Key, ".") {
		keys := strings.Split(mapInfo.Key, ".")
		for _, key := range keys {
			if jsonTargetKey[key] == nil {
				break
			}

			jsonTargetKey = handleKeyType(jsonTargetKey, key)
			if jsonTargetKey == nil {
				return false, fmt.Errorf(
					"error converting %s to a valid type: %w",
					key,
					ErrCantConvert,
				)
			}
		}

		mapInfo.Key = keys[len(keys)-1]
	}

	if mapInfo.GlobalSelectorRegex != "" {
		shouldKeep, err := applyGlobalSelector(jsonTargetKey, mapInfo)
		if err != nil || !shouldKeep {
			return false, err
		}
	}

	switch v := field.(type) {
	case *string:
		return true, setStringValue(jsonTargetKey, jsonm, mapInfo, v)
	case *int:
		return true, setIntValue(jsonTargetKey, mapInfo, v)
	default:
		return false, fmt.Errorf("unsupported type %T: %w", field, ErrUnsupportedType)
	}
}

func handleKeyType(jsonm map[string]any, key string) map[string]any {
	switch value := jsonm[key].(type) {
	case map[string]any:
		return value
	case []any:
		if len(value) == 0 {
			return nil
		}

		m, ok := value[0].(map[string]any)
		if ok {
			return m
		}

		handleKeyType(m, key)

		return nil
	case string:
		return map[string]any{key: value}
	case int:
		return map[string]any{key: value}
	case float64:
		return map[string]any{key: int(value)}
	case bool:
		return map[string]any{key: value}
	default:
		return nil
	}
}

func globalSelectFromRegex(str, regex string) (bool, error) {
	exp, err := regexp.Compile(regex)
	if err != nil {
		return false, fmt.Errorf("error compiling regex: %w", err)
	}

	if exp == nil {
		return false, ErrCantCompileRegex
	}

	if str == "" {
		return false, nil
	}

	return exp.MatchString(str), nil
}

func getDefaultStringValue(i JSONMappingInfo, defaultValue string) string {
	if i.OverrideValue != "" {
		return i.OverrideValue
	}

	if i.Key == "" {
		return defaultValue
	}

	return defaultValue
}

func getDefaultIntValue(i JSONMappingInfo, defaultValue int) int {
	if i.Key == "" {
		return defaultValue
	}

	return defaultValue
}

func applyValueTransformerRegex(val, regex string) (string, error) {
	if regex == "" {
		return val, nil
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return "", fmt.Errorf("error compiling regex: %w", err)
	}

	m := r.FindStringSubmatch(val)
	if len(m) <= 1 {
		return val, nil
	}

	return m[1], nil
}

func computeValueWithoutOverride(
	jsonTargetKey map[string]any,
	mapInfo JSONMappingInfo,
	field *string,
) error {
	switch val := jsonTargetKey[mapInfo.Key].(type) {
	case nil:
		slog.Debug("key not found in json", slog.String("key", mapInfo.Key), slog.Any("json", jsonTargetKey))

		return nil
	case string:
		*field = val
	case int:
		*field = strconv.Itoa(val)
	case float64:
		*field = strconv.Itoa(int(val))
	case bool:
		if val {
			*field = "true"
		} else {
			*field = "false"
		}
	case []any:
		// Perform aggregation
		var values []string

		for _, v := range val {
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("error converting %s to string: %w", mapInfo.Key, ErrCantConvertToString)
			}

			values = append(values, str)
		}

		*field = strings.Join(values, " - ")
	case any:
		*field = fmt.Sprintf("%v", val)
	default:
		return fmt.Errorf("error converting %s to string: %w", mapInfo.Key, ErrCantConvertToString)
	}

	transformedVal, err := applyValueTransformerRegex(*field, mapInfo.ValueTransformerRegex)
	if err != nil {
		return err
	}

	*field = transformedVal

	return nil
}

func setStringValue(
	jsonTargetKey map[string]any,
	jsonm map[string]any,
	mapInfo JSONMappingInfo,
	field *string,
) error {
	*field = getDefaultStringValue(mapInfo, mapInfo.DefaultValue)

	if mapInfo.OverrideValue == "" {
		err := computeValueWithoutOverride(jsonTargetKey, mapInfo, field)
		if err != nil {
			return fmt.Errorf("error computing value without override: %w", err)
		}
	}

	if mapInfo.Suffix != nil {
		var suffix string

		_, err := setValue(jsonm, *mapInfo.Suffix, &suffix)
		if err != nil {
			return fmt.Errorf("error setting suffix: %w", err)
		}

		*field += suffix
	}

	return nil
}

func setIntValue(jsonm map[string]any, mapInfo JSONMappingInfo, field *int) error {
	var def int

	if mapInfo.DefaultValue != "" {
		num, err := strconv.Atoi(mapInfo.DefaultValue)
		if err != nil {
			return fmt.Errorf(
				"error converting default value %s to int: %w",
				mapInfo.DefaultValue,
				err,
			)
		}

		def = num
	}

	*field = getDefaultIntValue(mapInfo, def)

	if jsonm[mapInfo.Key] == nil {
		return nil
	}

	value := jsonm[mapInfo.Key]
	switch val := value.(type) {
	case int:
		*field = val
	case float64:
		*field = int(val)
	case string:
		r, err := regexp.Compile(mapInfo.ValueTransformerRegex)
		if err != nil {
			return fmt.Errorf("error compiling regex: %w", err)
		}

		matches := r.FindStringSubmatch(val)
		if len(matches) <= 1 {
			return nil
		}

		parsedValue, err := strconv.Atoi(matches[1])
		if err != nil {
			return fmt.Errorf("error converting %s to int: %w", mapInfo.Key, err)
		}

		*field = parsedValue
	default:
		return fmt.Errorf("cannot convert %s to int: %w", mapInfo.Key, ErrCantConvertToInt)
	}

	return nil
}
