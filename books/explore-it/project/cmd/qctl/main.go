package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"queuelet/internal/client"
	"queuelet/internal/worker"
	"strconv"
	"strings"
)

func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: qctl queue|enqueue|jobs|job|worker|schedule|stats")
	}
	c := client.Env()
	opts := map[string]string{}
	pos := []string{}
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			key := strings.TrimPrefix(args[i], "--")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
				opts[key] = args[i]
			} else {
				opts[key] = "true"
			}
		} else {
			pos = append(pos, args[i])
		}
	}
	p := func(i int) string {
		if i < len(pos) {
			return pos[i]
		}
		return ""
	}
	method, path := "GET", ""
	var data any
	jsonOpt := func(key string) any {
		var v any
		if raw := opts[key]; raw != "" {
			json.Unmarshal([]byte(raw), &v)
		}
		return v
	}
	integer := func(key string, fallback int) int {
		n, e := strconv.Atoi(opts[key])
		if e != nil {
			return fallback
		}
		return n
	}
	body := func() map[string]any {
		if v, ok := jsonOpt("json").(map[string]any); ok {
			return v
		}
		return map[string]any{}
	}
	switch p(0) {
	case "queue":
		path = "/v1/queues"
		switch p(1) {
		case "ls":
		case "create":
			method = "POST"
			v := body()
			v["name"] = p(2)
			if opts["lease-seconds"] != "" {
				v["lease_seconds"] = integer("lease-seconds", 30)
			}
			if opts["max-attempts"] != "" {
				v["max_attempts"] = integer("max-attempts", 3)
			}
			data = v
		case "get", "rm", "pause", "resume", "purge", "patch":
			path += "/" + url.PathEscape(p(2))
			switch p(1) {
			case "rm":
				method = "DELETE"
			case "patch":
				method = "PATCH"
				data = body()
			case "pause", "resume", "purge":
				method = "POST"
				path += "/" + p(1)
			}
		default:
			return fmt.Errorf("unknown queue verb")
		}
	case "enqueue":
		method = "POST"
		path = "/v1/queues/" + url.PathEscape(p(1)) + "/jobs"
		v := body()
		v["kind"] = opts["kind"]
		if v["kind"] == "" {
			v["kind"] = "echo"
		}
		if opts["payload"] != "" {
			v["payload"] = jsonOpt("payload")
		}
		v["priority"] = integer("priority", 5)
		if opts["delay"] != "" {
			v["delay_seconds"] = integer("delay", 0)
		}
		if opts["run-at"] != "" {
			v["run_at"] = opts["run-at"]
		}
		if opts["max-attempts"] != "" {
			v["max_attempts"] = integer("max-attempts", 3)
		}
		if opts["idempotency-key"] != "" {
			v["idempotency_key"] = opts["idempotency-key"]
		}
		data = v
	case "batch":
		method = "POST"
		path = "/v1/queues/" + url.PathEscape(p(1)) + "/jobs:batch"
		data = body()
	case "jobs":
		path = "/v1/queues/" + url.PathEscape(p(2)) + "/jobs"
		query := url.Values{}
		for _, key := range []string{"state", "kind", "limit", "cursor"} {
			if opts[key] != "" {
				query.Set(key, opts[key])
			}
		}
		path += "?" + query.Encode()
	case "job":
		path = "/v1/jobs/" + p(2)
		switch p(1) {
		case "get":
		case "attempts":
			path += "/attempts"
		case "rm":
			method = "DELETE"
		case "cancel", "retry", "ack", "nack", "heartbeat":
			method = "POST"
			path += "/" + p(1)
			v := body()
			if opts["lease-token"] != "" {
				v["lease_token"] = opts["lease-token"]
			}
			if opts["result"] != "" {
				v["result"] = jsonOpt("result")
			}
			if opts["error"] != "" {
				v["error"] = opts["error"]
			}
			if opts["extend-seconds"] != "" {
				v["extend_seconds"] = integer("extend-seconds", 30)
			}
			data = v
		default:
			return fmt.Errorf("unknown job verb")
		}
	case "worker":
		if p(1) == "run" {
			kinds := opts["kinds"]
			if kinds == "" {
				kinds = "echo,sleep,fail,crash,http"
			}
			return worker.Run(context.Background(), c, "qctl-worker", p(2), strings.Split(kinds, ","))
		}
		path = "/v1/workers"
		switch p(1) {
		case "create":
			method = "POST"
			kinds := opts["kinds"]
			if kinds == "" {
				kinds = "echo"
			}
			data = map[string]any{"name": p(2), "kinds": strings.Split(kinds, ",")}
		case "rm":
			method = "DELETE"
			path += "/" + p(2)
		default:
			return fmt.Errorf("unknown worker verb")
		}
	case "lease":
		method = "POST"
		path = "/v1/queues/" + p(1) + "/lease"
		data = map[string]any{"worker_id": integer("worker-id", 0), "max_jobs": integer("max-jobs", 1)}
	case "schedule":
		path = "/v1/schedules"
		switch p(1) {
		case "ls":
		case "create":
			method = "POST"
			data = body()
		case "get", "rm", "patch", "run":
			path += "/" + p(2)
			switch p(1) {
			case "rm":
				method = "DELETE"
			case "patch":
				method = "PATCH"
				data = body()
			case "run":
				method = "POST"
				path += "/run"
			}
		default:
			return fmt.Errorf("unknown schedule verb")
		}
	case "stats":
		path = "/v1/queues/" + p(1) + "/stats"
		if opts["namespace"] != "" {
			path = "/v1/namespaces/" + opts["namespace"] + "/stats"
		}
	default:
		return fmt.Errorf("unknown command %s", p(0))
	}
	b, status, e := c.Do(method, path, data)
	if e != nil {
		return e
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d: %s", status, b)
	}
	if len(b) > 0 {
		var v any
		if json.Unmarshal(b, &v) == nil {
			out, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Print(string(b))
		}
	}
	return nil
}
