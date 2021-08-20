package observiqexporter

// Type preEncodedJSON aliases []byte, represents JSON that has already been encoded
// Does not support unmarshalling
type preEncodedJSON []byte

// The marshaled JSON is just the []byte that this type aliases.
func (p preEncodedJSON) MarshalJSON() ([]byte, error) {
	return p, nil
}
