package opampextension

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

// bestEffortMarshal attempts to perform a "best effort" version of yaml.Marshal.
// It will attempt to marshal the config, and marshal elements that could not
// be marshalled as a string "<cannot unmarshal>".
func bestEffortMarshal(m map[string]any) ([]byte, error) {
	marshalableMap := makeStringMapBestEffortMarshalable(m)
	return yaml.Marshal(marshalableMap)
}

// bestEffortMarshalable implements yaml.Marshaller.
// It first checks if a value is marshalable, and if it is not, returns "<cannot marshal>" as its value.
type bestEffortMarshalable struct {
	val any
}

func (m bestEffortMarshalable) MarshalYAML() (val any, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			val = "<cannot marshal>"
		}
	}()

	_, err = yaml.Marshal(m.val)
	if err != nil {
		return "<cannot marshal>", nil
	}

	return m.val, nil
}

// makeStringMapBestEffortMarshalable makes the map swallow any yaml marshal errors
func makeStringMapBestEffortMarshalable(m map[string]any) map[string]any {
	for k := range m {
		m[k] = toMarshalableValue(m[k])
	}

	return m
}

// makeStringMapBestEffortMarshalable makes the value swallow any yaml marshal errors
func toMarshalableValue(v any) any {
	switch vTyped := v.(type) {
	// for map[string]any, we can dive in deeper and attempt
	// to wrap any terminal elements (we know map[string]any marshals
	// if all values marshal)
	case map[string]any:
		return makeStringMapBestEffortMarshalable(vTyped)
	// These are types we know will be marshalable to YAML
	case []int, []int8, []int16, []int32, []int64,
		int, int8, int16, int32, int64,
		*int, *int8, *int16, *int32, *int64,
		[]uint, []uint8, []uint16, []uint32, []uint64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr,
		[]float32, []float64,
		float32, float64,
		*float32, *float64,
		bool,
		*bool,
		string,
		*string,
		nil:
		return vTyped
	}

	// For slices of unknown types, we iterate over the elements
	// (a slice is marshalable if all its elements are marshalable)
	if reflect.TypeOf(v).Kind() == reflect.Slice {
		sliceVal := reflect.ValueOf(v)
		returnSlice := make([]any, 0, sliceVal.Len())
		for i := 0; i < sliceVal.Len(); i++ {
			curVal := sliceVal.Index(i).Interface()
			returnSlice = append(returnSlice, toMarshalableValue(curVal))
		}

		return returnSlice
	}

	return bestEffortMarshalable{val: v}
}
