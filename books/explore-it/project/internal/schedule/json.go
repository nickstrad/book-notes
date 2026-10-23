package schedule

import "encoding/json"

func JSON(v any) string { b, _ := json.Marshal(v); return string(b) }
