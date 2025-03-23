package util

import (
	"fmt"
	"runtime"
)

func RecoverFromPanic() {
	if r := recover(); r != nil {
		fmt.Println("🔥 Panic Recovered:", r)
		buf := make([]uintptr, 10) // 最大10フレーム取得
		n := runtime.Callers(2, buf)
		frames := runtime.CallersFrames(buf[:n])
		fmt.Println("📌 Stack Trace:")
		for frame, more := frames.Next(); more; frame, more = frames.Next() {
			fmt.Printf("  - %s\n    %s:%d\n", frame.Function, frame.File, frame.Line)
		}
	}
}
