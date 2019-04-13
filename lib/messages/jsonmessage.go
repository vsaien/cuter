package messages

import (
	"encoding/json"
	"strconv"

	"github.com/vsaien/cuter/lib/codec"
	"github.com/vsaien/cuter/lib/mapping"
)

// Not thread safe
type JsonMessage struct {
	Raw      string
	TagName  string
	Data     map[string]interface{}
	Modified bool
}

func NewJsonMessage(s string) (*JsonMessage, error) {
	var f interface{}
	if err := codec.JsonUnmarshalString(s, &f); err != nil {
		return nil, err
	}

	m, ok := f.(map[string]interface{})
	if !ok {
		return nil, JsonMessageError{"Invalid json message format"}
	}
	if len(m) != 1 {
		return nil, JsonMessageError{"Should be only one tag name"}
	}

	var message *JsonMessage
	for k, v := range m {
		data, ok := v.(map[string]interface{})
		if !ok {
			return nil, JsonMessageError{"Message content should be a json node"}
		}

		message = &JsonMessage{
			Raw:     s,
			TagName: k,
			Data:    data,
		}
	}

	return message, nil
}

func NewJsonMessageFromMap(tag string, attributes map[string]interface{}) (*JsonMessage, error) {
	bs, err := codec.JsonMarshal(map[string]map[string]interface{}{
		tag: attributes,
	})
	if err != nil {
		return nil, err
	}

	return &JsonMessage{
		Raw:     string(bs),
		TagName: tag,
		Data:    attributes,
	}, nil
}

func (message *JsonMessage) Fill(v interface{}) error {
	return mapping.UnmarshalKey(message.Data, v)
}

func (message *JsonMessage) GetAttrInt(key string) (int, bool) {
	value, ok := message.GetAttrString(key)
	if !ok {
		return 0, false
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return result, true
}

func (message *JsonMessage) GetAttrInt64(key string) (int64, bool) {
	value, ok := message.GetAttrString(key)
	if !ok {
		return 0, false
	}
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return result, true
}

func (message *JsonMessage) GetAttrMap(key string) (map[string]interface{}, bool) {
	value, ok := message.getAttr(key)
	if !ok {
		return nil, false
	}
	result, ok := value.(map[string]interface{})
	return result, ok
}

func (message *JsonMessage) GetAttrString(key string) (string, bool) {
	value, ok := message.getAttr(key)
	if !ok {
		return "", false
	}
	result, ok := value.(string)
	return result, ok
}

func (message *JsonMessage) GetAttrUint(key string) (uint, bool) {
	result, ok := message.GetAttrUint32(key)
	return uint(result), ok
}

func (message *JsonMessage) GetAttrUint32(key string) (uint32, bool) {
	value, ok := message.GetAttrString(key)
	if !ok {
		return 0, false
	}
	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(result), true
}

func (message *JsonMessage) GetAttrUint64(key string) (uint64, bool) {
	value, ok := message.GetAttrString(key)
	if !ok {
		return 0, false
	}
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return result, true
}

func (message *JsonMessage) GetInt64(key string) (int64, bool) {
	value, ok := message.getAttr(key)
	if !ok {
		return 0, false
	}

	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}

	result, err := number.Int64()
	if err != nil {
		return 0, false
	}

	return result, true
}

func (message *JsonMessage) Marshal() (string, error) {
	if message.Modified {
		payload, err := json.Marshal(message.Data)
		if err != nil {
			return "", err
		}

		bytes, err := json.Marshal(map[string]json.RawMessage{
			message.TagName: payload,
		})
		if err != nil {
			return "", err
		}

		return string(bytes), nil
	} else {
		return message.Raw, nil
	}
}

func (message *JsonMessage) DelAttr(key string) {
	if _, ok := message.Data[key]; ok {
		message.Modified = true
		delete(message.Data, key)
	}
}

func (message *JsonMessage) SetAttr(key string, value interface{}) {
	message.Modified = true
	switch v := value.(type) {
	case []byte:
		message.Data[key] = json.RawMessage(v)
	default:
		message.Data[key] = value
	}
}

func (message *JsonMessage) getAttr(key string) (interface{}, bool) {
	value, ok := message.Data[key]
	return value, ok
}

type JsonMessageError struct {
	Reason string
}

func (e JsonMessageError) Error() string {
	return e.Reason
}
