package test

import "time"

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int32) *int32 {
	return &i
}

func Int32SlicePtr(i []int32) *[]int32 {
	return &i
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func BoolPtr(b bool) *bool {
	return &b
}

func Float64Ptr(f float64) *float64 {
	return &f
}
