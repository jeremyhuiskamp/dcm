package main

import (
    //"bytes"
    //"encoding/binary"
    "flag"
    "fmt"
    //"io"
)

func main() {
    flag.Parse()
    fmt.Printf("Got arguments: %s\n", flag.Args())
}
