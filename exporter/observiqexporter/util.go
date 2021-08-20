package observiqexporter

import (
	"strings"

	"go.opentelemetry.io/collector/model/pdata"
)

const dotReplacement = "_"

// flattenStringMap "flattens" the input map into a map of string -> string, de-dotting keys and merging sub-maps
//   in order to create the flattened map.
func flattenStringMap(m map[string]interface{}) map[string]string {
	mOut := make(map[string]string)
	flattenStringMapInternal("", mOut, m)
	return mOut
}

// flattenStringMapInternal recursively flattens mIn into mOut, appending keyPrefix to all keys in mIn.
func flattenStringMapInternal(keyPrefix string, mOut map[string]string, mIn map[string]interface{}) {
	if keyPrefix != "" {
		keyPrefix += dotReplacement
	}
	for k, v := range mIn {
		switch vTyped := v.(type) {
		case string:
			mOut[keyPrefix+k] = vTyped
		case map[string]interface{}:
			flattenStringMapInternal(keyPrefix+k, mOut, vTyped)
		}
	}
}

/*
	Transform AttributeMap to native Go map, skipping keys with nil values, and replacing dots in keys with _
*/
func attributeMapToBaseType(m pdata.AttributeMap) map[string]interface{} {
	mapOut := make(map[string]interface{}, m.Len())
	m.Range(func(k string, v pdata.AttributeValue) bool {
		val := attributeValueToBaseType(v)
		if val != nil {
			dedotedKey := strings.ReplaceAll(k, ".", dotReplacement)
			mapOut[dedotedKey] = val
		}
		return true
	})
	return mapOut
}

/*
	attrib is the attribute value to convert to it's native Go type - skips nils in arrays/maps
*/
func attributeValueToBaseType(attrib pdata.AttributeValue) interface{} {
	switch attrib.Type() {
	case pdata.AttributeValueTypeString:
		return attrib.StringVal()
	case pdata.AttributeValueTypeBool:
		return attrib.BoolVal()
	case pdata.AttributeValueTypeInt:
		return attrib.IntVal()
	case pdata.AttributeValueTypeDouble:
		return attrib.DoubleVal()
	case pdata.AttributeValueTypeMap:
		attribMap := attrib.MapVal()
		return attributeMapToBaseType(attribMap)
	case pdata.AttributeValueTypeArray:
		arrayVal := attrib.ArrayVal()
		slice := make([]interface{}, 0, arrayVal.Len())
		for i := 0; i < arrayVal.Len(); i++ {
			val := attributeValueToBaseType(arrayVal.At(i))
			if val != nil {
				slice = append(slice, val)
			}
		}
		return slice
	case pdata.AttributeValueTypeNull:
		return nil
	}
	return nil
}
