package main

func newBool(val bool) *bool {
    b := val
    return &b
}

func newUint32(val uint32) *uint32 {
    i := val
    return &i
}