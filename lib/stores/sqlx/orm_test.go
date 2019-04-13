package sqlx

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	scanFn        func(v ...interface{}) error
	mockedScanner struct {
		Cols     []string
		Count    int
		ScanFunc scanFn
	}
)

func (m *mockedScanner) Columns() ([]string, error) {
	return m.Cols, nil
}

func (m *mockedScanner) Err() error {
	return nil
}

func (m *mockedScanner) Next() bool {
	m.Count--
	return m.Count >= 0
}

func (m *mockedScanner) Scan(v ...interface{}) error {
	return m.ScanFunc(v...)
}

func TestUnmarshalRowBool(t *testing.T) {
	const expect = true
	var value bool
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*bool); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowInt(t *testing.T) {
	const expect = int(2)
	var value int
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowInt8(t *testing.T) {
	const expect = int8(3)
	var value int8
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int8); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowInt16(t *testing.T) {
	const expect = int16(4)
	var value int16
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int16); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowInt32(t *testing.T) {
	const expect = int32(5)
	var value int32
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int32); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowInt64(t *testing.T) {
	const expect = int64(6)
	var value int64
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int64); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowUint(t *testing.T) {
	const expect = uint(2)
	var value uint
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowUint8(t *testing.T) {
	const expect = uint8(3)
	var value uint8
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint8); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowUint16(t *testing.T) {
	const expect = uint16(4)
	var value uint16
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint16); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowUint32(t *testing.T) {
	const expect = uint32(5)
	var value uint32
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint32); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowUint64(t *testing.T) {
	const expect = uint64(6)
	var value uint64
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint64); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowFloat32(t *testing.T) {
	const expect = float32(7)
	var value float32
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*float32); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowFloat64(t *testing.T) {
	const expect = float64(8)
	var value float64
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*float64); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowString(t *testing.T) {
	const expect = "hello"
	var value string
	doTestUnmarshalRowPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*string); ok {
			*i = expect
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowStruct(t *testing.T) {
	var value = new(struct {
		Name string
		Age  int
	})
	m := mockedScanner{
		Cols:  []string{"name", "age"},
		Count: 1,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString("liao")
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(5)
			return nil
		},
	}
	assert.Nil(t, UnmarshalRow(value, &m))
	assert.Equal(t, "liao", value.Name)
	assert.Equal(t, 5, value.Age)
}

func TestUnmarshalRowStructWithTags(t *testing.T) {
	var value = new(struct {
		Age  int    `db:"age"`
		Name string `db:"name"`
	})
	m := mockedScanner{
		Cols:  []string{"name", "age"},
		Count: 1,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString("liao")
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(5)
			return nil
		},
	}
	assert.Nil(t, UnmarshalRow(value, &m))
	assert.Equal(t, "liao", value.Name)
	assert.Equal(t, 5, value.Age)
}

func TestUnmarshalRowsBool(t *testing.T) {
	var expect = []bool{true, false}
	var value []bool
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*bool); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt(t *testing.T) {
	var expect = []int{2, 3}
	var value []int
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt8(t *testing.T) {
	var expect = []int8{2, 3}
	var value []int8
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int8); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt16(t *testing.T) {
	var expect = []int16{2, 3}
	var value []int16
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int16); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt32(t *testing.T) {
	var expect = []int32{2, 3}
	var value []int32
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int32); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt64(t *testing.T) {
	var expect = []int64{2, 3}
	var value []int64
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int64); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint(t *testing.T) {
	var expect = []uint{2, 3}
	var value []uint
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint8(t *testing.T) {
	var expect = []uint8{2, 3}
	var value []uint8
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint8); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint16(t *testing.T) {
	var expect = []uint16{2, 3}
	var value []uint16
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint16); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint32(t *testing.T) {
	var expect = []uint32{2, 3}
	var value []uint32
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint32); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint64(t *testing.T) {
	var expect = []uint64{2, 3}
	var value []uint64
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint64); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsFloat32(t *testing.T) {
	var expect = []float32{2, 3}
	var value []float32
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*float32); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsFloat64(t *testing.T) {
	var expect = []float64{2, 3}
	var value []float64
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*float64); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsString(t *testing.T) {
	var expect = []string{"hello", "world"}
	var value []string
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*string); ok {
			*i = expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsBoolPtr(t *testing.T) {
	yes := true
	no := false
	var expect = []*bool{&yes, &no}
	var value []*bool
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*bool); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsIntPtr(t *testing.T) {
	two := 2
	three := 3
	var expect = []*int{&two, &three}
	var value []*int
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt8Ptr(t *testing.T) {
	two := int8(2)
	three := int8(3)
	var expect = []*int8{&two, &three}
	var value []*int8
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int8); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt16Ptr(t *testing.T) {
	two := int16(2)
	three := int16(3)
	var expect = []*int16{&two, &three}
	var value []*int16
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int16); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt32Ptr(t *testing.T) {
	two := int32(2)
	three := int32(3)
	var expect = []*int32{&two, &three}
	var value []*int32
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int32); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsInt64Ptr(t *testing.T) {
	two := int64(2)
	three := int64(3)
	var expect = []*int64{&two, &three}
	var value []*int64
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*int64); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUintPtr(t *testing.T) {
	two := uint(2)
	three := uint(3)
	var expect = []*uint{&two, &three}
	var value []*uint
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint8Ptr(t *testing.T) {
	two := uint8(2)
	three := uint8(3)
	var expect = []*uint8{&two, &three}
	var value []*uint8
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint8); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint16Ptr(t *testing.T) {
	two := uint16(2)
	three := uint16(3)
	var expect = []*uint16{&two, &three}
	var value []*uint16
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint16); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint32Ptr(t *testing.T) {
	two := uint32(2)
	three := uint32(3)
	var expect = []*uint32{&two, &three}
	var value []*uint32
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint32); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsUint64Ptr(t *testing.T) {
	two := uint64(2)
	three := uint64(3)
	var expect = []*uint64{&two, &three}
	var value []*uint64
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*uint64); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsFloat32Ptr(t *testing.T) {
	two := float32(2)
	three := float32(3)
	var expect = []*float32{&two, &three}
	var value []*float32
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*float32); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsFloat64Ptr(t *testing.T) {
	two := float64(2)
	three := float64(3)
	var expect = []*float64{&two, &three}
	var value []*float64
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*float64); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsStringPtr(t *testing.T) {
	hello := "hello"
	world := "world"
	var expect = []*string{&hello, &world}
	var value []*string
	index := 0
	doTestUnmarshalRowsPrimitive(t, &value, func(v interface{}) {
		if i, ok := v.(*string); ok {
			*i = *expect[index]
			index++
		}
	}, func(actual interface{}) bool {
		return doCompare(expect, actual)
	})
}

func TestUnmarshalRowsStruct(t *testing.T) {
	var expect = []struct {
		Name string
		Age  int64
	}{
		{
			Name: "first",
			Age:  2,
		},
		{
			Name: "second",
			Age:  3,
		},
	}
	var value []struct {
		Name string
		Age  int64
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "age"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(expect[index].Age)
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))

	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.Age, value[i].Age)
	}
}

func TestUnmarshalRowsStructWithNullStringType(t *testing.T) {
	var expect = []struct {
		Name       string
		NullString sql.NullString
	}{
		{
			Name: "first",
			NullString: sql.NullString{
				String: "firstnullstring",
				Valid:  true,
			},
		},
		{
			Name: "second",
			NullString: sql.NullString{
				String: "",
				Valid:  false,
			},
		},
	}
	var value []struct {
		Name       string `db:"name"`
		NullString sql.NullString
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "value"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).Set(reflect.ValueOf(expect[index].NullString))
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))
	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.NullString.String, value[i].NullString.String)
		assert.Equal(t, each.NullString.Valid, value[i].NullString.Valid)
	}
}

func TestUnmarshalRowsStructWithTags(t *testing.T) {
	var expect = []struct {
		Name string
		Age  int64
	}{
		{
			Name: "first",
			Age:  2,
		},
		{
			Name: "second",
			Age:  3,
		},
	}
	var value []struct {
		Age  int64  `db:"age"`
		Name string `db:"name"`
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "age"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(expect[index].Age)
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))

	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.Age, value[i].Age)
	}
}

func TestUnmarshalRowsStructAndEmbeddedAnonymousStructWithTags(t *testing.T) {
	type Embed struct {
		Value int64 `db:"value"`
	}

	var expect = []struct {
		Name  string
		Age   int64
		Value int64
	}{
		{
			Name:  "first",
			Age:   2,
			Value: 3,
		},
		{
			Name:  "second",
			Age:   3,
			Value: 4,
		},
	}
	var value []struct {
		Name string `db:"name"`
		Age  int64  `db:"age"`
		Embed
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "age", "value"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(expect[index].Age)
			reflect.Indirect(reflect.ValueOf(v[2])).SetInt(expect[index].Value)
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))

	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.Age, value[i].Age)
		assert.Equal(t, each.Value, value[i].Value)
	}
}

func TestUnmarshalRowsStructAndEmbeddedStructPtrAnonymousWithTags(t *testing.T) {
	type Embed struct {
		Value int64 `db:"value"`
	}

	var expect = []struct {
		Name  string
		Age   int64
		Value int64
	}{
		{
			Name:  "first",
			Age:   2,
			Value: 3,
		},
		{
			Name:  "second",
			Age:   3,
			Value: 4,
		},
	}
	var value []struct {
		Name string `db:"name"`
		Age  int64  `db:"age"`
		*Embed
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "age", "value"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(expect[index].Age)
			reflect.Indirect(reflect.ValueOf(v[2])).SetInt(expect[index].Value)
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))

	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.Age, value[i].Age)
		assert.Equal(t, each.Value, value[i].Value)
	}
}

func TestUnmarshalRowsStructPtr(t *testing.T) {
	var expect = []*struct {
		Name string
		Age  int64
	}{
		{
			Name: "first",
			Age:  2,
		},
		{
			Name: "second",
			Age:  3,
		},
	}
	var value []*struct {
		Name string
		Age  int64
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "age"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(expect[index].Age)
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))

	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.Age, value[i].Age)
	}
}

func TestUnmarshalRowsStructWithTagsPtr(t *testing.T) {
	var expect = []*struct {
		Name string
		Age  int64
	}{
		{
			Name: "first",
			Age:  2,
		},
		{
			Name: "second",
			Age:  3,
		},
	}
	var value []*struct {
		Age  int64  `db:"age"`
		Name string `db:"name"`
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "age"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(expect[index].Age)
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))

	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.Age, value[i].Age)
	}
}

func TestUnmarshalRowsStructWithTagsPtrWithInnerPtr(t *testing.T) {
	var expect = []*struct {
		Name string
		Age  int64
	}{
		{
			Name: "first",
			Age:  2,
		},
		{
			Name: "second",
			Age:  3,
		},
	}
	var value []*struct {
		Age  *int64 `db:"age"`
		Name string `db:"name"`
	}
	index := 0
	m := mockedScanner{
		Cols:  []string{"name", "age"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			reflect.Indirect(reflect.ValueOf(v[0])).SetString(expect[index].Name)
			reflect.Indirect(reflect.ValueOf(v[1])).SetInt(expect[index].Age)
			index++
			return nil
		},
	}

	assert.Nil(t, UnmarshalRows(&value, &m))

	for i, each := range expect {
		assert.Equal(t, each.Name, value[i].Name)
		assert.Equal(t, each.Age, *value[i].Age)
	}
}

func doCompare(expect, actual interface{}) bool {
	rt := reflect.TypeOf(actual)
	rv := reflect.ValueOf(actual)
	if rt.Kind() == reflect.Ptr {
		rvi := reflect.Indirect(rv)
		if rvi.CanInterface() {
			return reflect.DeepEqual(expect, rvi.Interface())
		} else {
			return false
		}
	} else {
		return reflect.DeepEqual(expect, actual)
	}
}

func doTestUnmarshalRowPrimitive(t *testing.T, value interface{}, setter func(v interface{}),
	compare func(actual interface{}) bool) {
	m := mockedScanner{
		Cols:  []string{"any"},
		Count: 1,
		ScanFunc: func(v ...interface{}) error {
			for _, each := range v {
				setter(each)
			}

			return nil
		},
	}
	assert.Nil(t, UnmarshalRow(value, &m))
	assert.True(t, compare(value))
}

func doTestUnmarshalRowsPrimitive(t *testing.T, value interface{}, setter func(v interface{}),
	compare func(actual interface{}) bool) {
	m := mockedScanner{
		Cols:  []string{"any"},
		Count: 2,
		ScanFunc: func(v ...interface{}) error {
			for _, each := range v {
				setter(each)
			}

			return nil
		},
	}
	assert.Nil(t, UnmarshalRows(value, &m))
	assert.True(t, compare(value))
}
