package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// verbatim logic from lesson-06 cmd/qctl/main.go:53-61
func jsonOptNew(raw string) any {
	var v any
	if raw != "" {
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.UseNumber()
		decoder.Decode(&v)
	}
	return v
}

// verbatim logic from lesson-05 cmd/qctl/main.go (json.Unmarshal, err ignored)
func jsonOptOld(raw string) any {
	var v any
	if raw != "" {
		json.Unmarshal([]byte(raw), &v)
	}
	return v
}

// verbatim logic client.go lesson-06
func clientNew(b []byte) error {
	var out map[string]any
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	return d.Decode(&out)
}
func clientOld(b []byte) error {
	var out map[string]any
	return json.Unmarshal(b, &out)
}

func main() {
	inputs := []string{
		`{"big":9007199254740993}`,
		"{\"big\":9007199254740993}xyz",
		`{"big":9007199254740993} {"big":1}`,
		`{"big":}`,
		`123abc`,
		"{\"big\":9007199254740993}   \n",
		``,
	}
	for _, in := range inputs {
		o, _ := json.Marshal(jsonOptOld(in))
		n, _ := json.Marshal(jsonOptNew(in))
		fmt.Printf("qctl-input %q\n  old(Unmarshal)=%s\n  new(Decode)  =%s\n", in, o, n)
	}
	resps := []string{"{\"id\":1}\n{\"id\":2}", "", "notjson", "{\"id\":1}  ", "{\"id\":9007199254740993}"}
	for _, r := range resps {
		fmt.Printf("client %q oldErr=%v newErr=%v\n", r, clientOld([]byte(r)), clientNew([]byte(r)))
	}
	// worker numeric consumers worker.go:82-83,89-90
	for _, payload := range []map[string]any{
		{}, {"seconds": json.Number("0.3")}, {"seconds": "0.3"}, {"seconds": json.Number("1e999")}, {"n": json.Number("0")},
	} {
		number, ok := payload["seconds"].(json.Number)
		seconds, err := number.Float64()
		fmt.Printf("worker payload=%v assertOk=%v number=%q float=%v err=%v\n", payload, ok, string(number), seconds, err)
	}
	// worker http kind worker.go:111 unchanged Unmarshal
	b := []byte(`{"big":9007199254740993,"fraction":0.12345678901234567890123456789}`)
	var v any
	json.Unmarshal(b, &v)
	out, _ := json.Marshal(v)
	fmt.Printf("worker-http Unmarshal roundtrip: %s -> %s\n", b, out)
	b2 := []byte("{\"big\":9007199254740993}trailing")
	var v2 any
	err2 := json.Unmarshal(b2, &v2)
	fmt.Printf("worker-http trailing Unmarshal err=%v (fallback to string)\n", err2)
}
