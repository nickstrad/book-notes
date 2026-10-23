package main

import (
	"fmt"
	"queuelet/internal/schedule"
	"time"
)

func main() {
	after := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	n, _ := schedule.Next("0 9 * * *", "America/New_York", -14400, after)
	loc, _ := time.LoadLocation("America/New_York")
	off, _ := schedule.Offset("America/New_York", after)
	fmt.Println("stored-offset next:", n.UTC(), "=", n.In(loc).Format("15:04 MST"), "| offset now would be", off)
}
