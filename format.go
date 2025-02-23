package konfig

import (
	"encoding/json"
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v2"
)

type FormatHandler interface {
	Unmarshal(data []byte, v interface{}) error
	Marshal(v interface{}) ([]byte, error)
}

type JSONFormat struct{}

func (j *JSONFormat) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j *JSONFormat) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

type YAMLFormat struct{}

func (y *YAMLFormat) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func (y *YAMLFormat) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

type TOMLFormat struct{}

func (t *TOMLFormat) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

func (t *TOMLFormat) Marshal(v interface{}) ([]byte, error) {
	return toml.Marshal(v)
}

func createFormatHandler(format EncodingFormat) (FormatHandler, error) {
	switch format {
	case JSON:
		return &JSONFormat{}, nil
	case YAML:
		return &YAMLFormat{}, nil
	case TOML:
		return &TOMLFormat{}, nil
	default:
		return nil, fmt.Errorf("unknown format: %d", format)
	}
}
