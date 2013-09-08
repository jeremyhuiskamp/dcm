package main

import (
    "github.com/kamper/dcm/dcm"
    "github.com/kamper/dcm/dcmio"
    "flag"
    "fmt"
    "log"
    "os"
)

func main() {
    flag.Parse()
    file, err := os.Open(flag.Arg(0))
    if err != nil {
        log.Fatal(err)
    }

    p, err := dcmio.NewFileParser(file)
    if err != nil {
        log.Fatal(err)
    }

    for {
        tag, err := p.NextTag()
        if err != nil {
            log.Fatal(err)
        } else if tag == nil {
            break
        } else {
            fmt.Printf("%d:(%04X,%04X) %s #%d\n",
                tag.Offset,
                tag.Group,
                tag.Tag,
                dcm.VRName(tag.VR),
                tag.ValueLength)
        }
    }
}
