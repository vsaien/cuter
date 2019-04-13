package messages

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validMessage = `{
        "HM": {
            "uid": "1234",
            "name": "kevin",
            "number": 5678,
            "nan": "not a number",
            "msg": {
                "t": "text"
            }
        }
    }`
)

var (
	invalidMessageSet = []string{
		`not a valid json`,
		`{"HM": "not a valid chaoxin protocol"}`,
		`{"HM": "not a valid chaoxin protocol", "DUMMY": "illegal node"}`,
		`[{"HM": "first"}, {"DUMMY": "second"}]`,
	}
)

func TestNewJsonMessageOnInvalidInput(t *testing.T) {
	for _, str := range invalidMessageSet {
		_, err := NewJsonMessage(str)
		assert.NotNil(t, err)
	}
}

func TestJsonMessageGetAttrInt(t *testing.T) {
	// given
	message, err := NewJsonMessage(validMessage)
	assert.Nil(t, err)

	// when
	value, ok := message.GetAttrInt("uid")

	// then
	assert.True(t, ok)
	assert.Equal(t, 1234, value)

	_, ok = message.GetAttrInt("nonexist")
	assert.False(t, ok)

	_, ok = message.GetAttrInt("number")
	assert.False(t, ok)

	_, ok = message.GetAttrInt("nan")
	assert.False(t, ok)
}

func TestJsonMessageGetAttrMap(t *testing.T) {
	// given
	message, err := NewJsonMessage(validMessage)
	assert.Nil(t, err)

	// when
	value, ok := message.GetAttrMap("msg")
	assert.True(t, ok)

	// then
	msgContent, ok := value["t"]
	assert.True(t, ok)
	assert.Equal(t, "text", msgContent)

	_, ok = message.GetAttrMap("nonexist")
	assert.False(t, ok)
}

func TestJsonMessageGetAttrString(t *testing.T) {
	// given
	message, err := NewJsonMessage(validMessage)
	assert.Nil(t, err)

	// when
	value, ok := message.GetAttrString("name")

	// then
	assert.True(t, ok)
	assert.Equal(t, "kevin", value)

	_, ok = message.GetAttrString("nonexist")
	assert.False(t, ok)
}

func TestJsonMessageGetContentNode(t *testing.T) {
	// given
	message, err := NewJsonMessage(validMessage)
	assert.Nil(t, err)

	// when
	content := message.Data

	// then
	verifyMapAttribute(t, content, "uid", "1234")
	verifyMapAttribute(t, content, "name", "kevin")
	verifyMapAttribute(t, content, "number", json.Number("5678"))
	verifyMapAttribute(t, content, "nan", "not a number")

	msgNode, ok := content["msg"]
	assert.True(t, ok)
	msgMap, ok := msgNode.(map[string]interface{})
	assert.True(t, ok)
	value, ok := msgMap["t"]
	assert.True(t, ok)
	assert.Equal(t, "text", value)
}

func TestJsonMessageGetTagName(t *testing.T) {
	message, err := NewJsonMessage(validMessage)
	assert.Nil(t, err)
	assert.Equal(t, "HM", message.TagName)
}

func TestJsonMessageMarshal(t *testing.T) {
	// given
	message, err := NewJsonMessage(validMessage)
	assert.Nil(t, err)

	// when
	str, err := message.Marshal()
	assert.Nil(t, err)
	newMessage, err := NewJsonMessage(str)
	assert.Nil(t, err)

	// then
	assert.True(t, compareAttrInt(message, newMessage, "uid"))
	assert.True(t, compareAttrString(message, newMessage, "name"))
}

func TestJsonMessageSetAttrValue(t *testing.T) {
	// given
	message, err := NewJsonMessage(validMessage)
	assert.Nil(t, err)

	// when
	message.SetAttr("ten", "10")

	// then
	actual, ok := message.GetAttrInt("ten")
	assert.True(t, ok)
	assert.Equal(t, 10, actual)
}

func TestJsonMessageFill(t *testing.T) {
	const s = `{"HM": {"id": 1453730145353127267, "name": "person"}}`
	jmsg, err := NewJsonMessage(s)
	assert.Nil(t, err)

	var v struct {
		Name string `key:"name"`
		Id   int64  `key:"id,string"`
	}

	assert.Nil(t, jmsg.Fill(&v))
	assert.Equal(t, "person", v.Name)
	assert.Equal(t, int64(1453730145353127267), v.Id)
}

func compareAttrInt(first, second *JsonMessage, field string) bool {
	firstValue, ok := first.GetAttrInt(field)
	if !ok {
		return false
	}

	secondValue, ok := second.GetAttrInt(field)
	if !ok {
		return false
	}

	return firstValue == secondValue
}

func compareAttrString(first, second *JsonMessage, field string) bool {
	firstValue, ok := first.GetAttrString(field)
	if !ok {
		return false
	}

	secondValue, ok := second.GetAttrString(field)
	if !ok {
		return false
	}

	return firstValue == secondValue
}

func verifyMapAttribute(t *testing.T, dict map[string]interface{}, field string, expect interface{}) {
	if value, ok := dict[field]; !ok {
		t.Fail()
	} else if value != expect {
		t.Fail()
	}
}
