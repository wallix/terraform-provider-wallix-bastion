package bastion

import "encoding/json"

// Exported for unit tests in package bastion_test.

func ConfigOptionsToValues(options []interface{}) map[string]interface{} {
	return configOptionsToValues(options)
}

func FilterConfigValues(current, managed map[string]interface{}) map[string]interface{} {
	return filterConfigValues(current, managed)
}

func AddConfigValuesPath(values map[string]interface{}, path []string) {
	addConfigValuesPath(values, path)
}

func SuppressEquivalentJSON(oldValue, newValue string) bool {
	return suppressEquivalentJSON("", oldValue, newValue, nil)
}

// MarshalConfigValuesForPut returns the PUT body of values, given the options returned by GET with metainfo.
//
//nolint:wrapcheck
func MarshalConfigValuesForPut(values map[string]interface{}, options []interface{}) ([]byte, error) {
	if err := prepareConfigValues(values, options, ""); err != nil {
		return nil, err
	}

	return json.Marshal(values)
}
