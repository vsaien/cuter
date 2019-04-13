package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func JsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func JsonUnmarshalBytes(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := jsonUnmarshal(decoder, v); err != nil {
		return formatError(string(data), err)
	}

	return nil
}

func JsonUnmarshalString(str string, v interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(str))
	if err := jsonUnmarshal(decoder, v); err != nil {
		return formatError(str, err)
	}

	return nil
}

func JsonUnmarshalReader(reader io.Reader, v interface{}) error {
	var buf strings.Builder
	teeReader := io.TeeReader(reader, &buf)
	decoder := json.NewDecoder(teeReader)
	if err := jsonUnmarshal(decoder, v); err != nil {
		return formatError(buf.String(), err)
	}

	return nil
}

func jsonUnmarshal(decoder *json.Decoder, v interface{}) error {
	decoder.UseNumber()
	return decoder.Decode(v)
}

func formatError(v string, err error) error {
	return fmt.Errorf("string: `%s`, error: `%s`", v, err.Error())
}
