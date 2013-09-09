package main

import (
    "github.com/kamper/dcm/dcm"
    "github.com/kamper/dcm/dcmio"
    "github.com/kamper/dcm/dcmtag"
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

    nest := 0
    for {
        tag, err := p.NextTag()

        if err != nil {
            log.Fatal(err)

        } else if tag == nil {
            break

        } else {
            fmt.Printf("%d:%s(%04X,%04X)%s #%d %s\n",
                tag.Offset,
                indent(nest),
                tag.Group,
                tag.Tag,
                vrToString(tag.VR),
                tag.ValueLength,
                desc(tag))

            // hmm, how can we compare identity and not values?
            if tag.VR != nil && *tag.VR == dcm.SQ {
                nest++
            } else if tag.Group == 0xFFFE && tag.Tag == 0xE0DD {
                nest--
            }
        }
    }
}

func indent(nest int) (s string) {
    for ; nest > 0; nest-- {
        s += ">"
    }
    return
}

func vrToString(vr *dcm.VR) (s string) {
    if vr != nil {
        return " " + vr.Name
    }
    return
}

func desc(tag *dcmio.Tag) string {
    t := dcmtag.GetGroupTag(tag.Group, tag.Tag)
    if t != nil {
        return t.Desc
    }

    return "??"
}
