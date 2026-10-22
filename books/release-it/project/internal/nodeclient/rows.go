package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"litefleet/internal/wire"
)

// Rows is a node's answer to a query, read as it arrives.
type Rows struct {
	node    string
	resp    *http.Response
	dec     *json.Decoder
	Columns []string
}

// openQuery sends the query and reads the answer up to its first row: the
// fields before "rows" must be exactly the ones the gateway knows.
func (c *Client) openQuery(ctx context.Context, rt wire.Route, body []byte) (*Rows, error) {
	resp, err := c.send(ctx, rt, "query", body)
	if err != nil {
		return nil, err
	}
	r := &Rows{node: rt.Node, resp: resp, dec: json.NewDecoder(resp.Body)}
	r.dec.UseNumber()
	if err := r.preamble(); err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("node %s: bad rows: %w", rt.Node, err)
	}
	return r, nil
}

func (r *Rows) preamble() error {
	if err := expect(r.dec, json.Delim('{')); err != nil {
		return err
	}
	for {
		key, err := r.key()
		if err != nil {
			return err
		}
		switch key {
		case "columns":
			if err := r.dec.Decode(&r.Columns); err != nil {
				return err
			}
		case "rows":
			return expect(r.dec, json.Delim('['))
		default:
			return fmt.Errorf("json: unknown field %q", key)
		}
	}
}

func (r *Rows) key() (string, error) {
	t, err := r.dec.Token()
	if err != nil {
		return "", err
	}
	k, ok := t.(string)
	if !ok {
		return "", fmt.Errorf("json: expected a field name, got %v", t)
	}
	return k, nil
}

func expect(dec *json.Decoder, want json.Delim) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if t != want {
		return fmt.Errorf("json: expected %v, got %v", want, t)
	}
	return nil
}

// WriteTo copies the answer to w row by row, as
// {"columns":[...],"rows":[...],"truncated":bool}, and closes it. An error
// after the first byte means the answer is incomplete.
func (r *Rows) WriteTo(w io.Writer) (int64, error) {
	defer r.Close()
	cw := &counter{w: w}
	cols, _ := json.Marshal(r.Columns)
	fmt.Fprintf(cw, `{"columns":%s,"rows":[`, cols)
	for first := true; r.dec.More(); first = false {
		var row json.RawMessage
		if err := r.dec.Decode(&row); err != nil {
			return cw.n, fmt.Errorf("node %s: rows cut short: %w", r.node, err)
		}
		if !first {
			cw.Write([]byte{','})
		}
		cw.Write(bytes.TrimSpace(row))
	}
	if err := expect(r.dec, json.Delim(']')); err != nil {
		return cw.n, fmt.Errorf("node %s: rows cut short: %w", r.node, err)
	}
	truncated := false
	for r.dec.More() {
		key, err := r.key()
		if err != nil {
			return cw.n, err
		}
		if key != "truncated" {
			return cw.n, fmt.Errorf("node %s: bad rows: json: unknown field %q", r.node, key)
		}
		if err := r.dec.Decode(&truncated); err != nil {
			return cw.n, err
		}
	}
	fmt.Fprintf(cw, `],"truncated":%t}`, truncated)
	return cw.n, cw.err
}

// Close releases the node's connection.
func (r *Rows) Close() { r.resp.Body.Close() }

type counter struct {
	w   io.Writer
	n   int64
	err error
}

func (c *counter) Write(p []byte) (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	n, err := c.w.Write(p)
	c.n += int64(n)
	c.err = err
	return n, err
}
