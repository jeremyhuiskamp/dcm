package main

import (
    "github.com/kamper/dcm/dcm"
    "github.com/kamper/dcm/dcmio"
    "flag"
    "fmt"
    "log"
    "io"
    "io/ioutil"
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
        if err == io.EOF {
            fmt.Println("EOF")
            break
        } else if err != nil {
            log.Fatal(err)
        } else {
            written, err := io.Copy(ioutil.Discard, tag.Value)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Printf("(%04X,%04X) VR=%s, VL=%d\n",
                tag.Group, tag.Tag, dcm.VRName(tag.VR), written)
        }
    }
}
