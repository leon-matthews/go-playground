package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	marshalBasic()
	marshalStructs()
	unmarshalBasic()
	unmarshalStructs()
	streaming()
}

func marshalBasic() {
	b, _ := json.Marshal(true)
	i, _ := json.Marshal(42)
	e, _ := json.Marshal(2.71828)
	fmt.Printf("%s %s %s\n", b, i, e)

	// Time
	t, _ := json.Marshal(time.Now())
	fmt.Printf("%s\n", t)

	// Slice
	s := []string{"apple", "peach", "pear"}
	b, _ = json.Marshal(s)
	fmt.Println(string(b))

	// Map
	m := map[string]int{"apple": 5, "lettuce": 7}
	b, _ = json.Marshal(m)
	fmt.Println(string(b))
}

type response1 struct {
	Page   int
	Fruits []string
}

type response2 struct {
	Page   int      `json:"page"`
	Fruits []string `json:"fruits"`
}

func marshalStructs() {
	r1 := &response1{1, []string{"apple", "peach", "pear"}}
	r2 := &response2{2, []string{"banana", "orange", "lemon"}}
	b1, _ := json.Marshal(r1)
	b2, _ := json.Marshal(r2)
	fmt.Println(string(b1))
	fmt.Println(string(b2))
}

func unmarshalBasic() {
	// Unmarshall JSON object into Go map
	b := []byte(`{"num":6.13,"strs":["apple","banana"]}`)
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		panic(err)
	}

	// Type assertion needed
	num, ok := data["num"].(float64)
	if !ok {
		panic("num is not a number")
	}
	fmt.Println(num)

	// Nested data requires a series of conversions.
	strs := data["strs"].([]any)
	s := strs[0].(string)
	fmt.Println(s)
}

func unmarshalStructs() {
	str := `{"page": 1, "fruits": ["apple", "peach"]}`
	r := response2{}
	json.Unmarshal([]byte(str), &r)
	fmt.Println(r)
}

// streaming without needing string/bytes as intermediate values
// Takes any [io.Reader] or [io.Writer] for the decoder and encoder, respectively.
func streaming() {
	// Stream out directly to stdout
	enc := json.NewEncoder(os.Stdout)
	d := map[string]int{"apple": 5, "lettuce": 7}
	enc.Encode(d)

	// Stream in from a [strings.Reader]
	str := `{"page": 1, "fruits": ["apple", "peach"]}`
	reader := strings.NewReader(str)
	dec := json.NewDecoder(reader)
	r := response2{}
	dec.Decode(&r)
	fmt.Println(r)
}
