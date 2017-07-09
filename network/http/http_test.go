package http

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJSON(t *testing.T) {
	input := `{"test":"data","validation":"process"}`
	expected := map[string]interface{}{
		"test":       "data",
		"validation": "process",
	}

	var actual map[string]interface{}
	err := UnmarshalJSON(bytes.NewReader([]byte(input)), &actual)
	if err != nil {
		fmt.Printf("decoding err: %v\n", err)
	}

	assert.EqualValues(t, expected, actual)
}
