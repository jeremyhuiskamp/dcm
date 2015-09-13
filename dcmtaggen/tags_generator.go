// Reads the dicom spec and produces constants for the tags, as well as the
// standard data dictionary
package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	tagRegex = regexp.MustCompile(`^\(([[:alnum:]]{4}),([[:alnum:]]{4})\)$`)
	multiplicityRegex = regexp.MustCompile(`([0-9])(-([0-9n]))?`)
)

// XML structs:

type Cell struct {
	Value   string `xml:",chardata"`
	EmValue string `xml:"emphasis"`
}

func (c Cell) GetValue() string {
	if strings.TrimSpace(c.Value) != "" {
		return c.Value
	}
	
	return c.EmValue
}

type Row struct {
	Cells []Cell `xml:"td>para"`
}

func (r Row) GetValues() []string {
	s := make([]string, len(r.Cells))
	for _, cell := range r.Cells {
		s = append(s, cell.GetValue())
	}
	return s
}

type Table struct {
	Rows []Row `xml:"tbody>tr"`
}

type Section struct {
	Tables []Table `xml:"table"`
}

type Chapter struct {
	Label    string    `xml:"label,attr"`
	Tables   []Table   `xml:"table"`
	Sections []Section `xml:"section"`
}

func (c Chapter) GetTables() []Table {
	tables := make([]Table, 10)
	
	if c.Tables != nil {
		tables = append(tables, c.Tables...)
	}
	
	if c.Sections != nil {
		for _, section := range c.Sections {
			if section.Tables != nil {
				tables = append(tables, section.Tables...)
			}
		}
	}
	
	return tables
}

type Book struct {
	Chapters []Chapter `xml:"chapter"`
}

// output prep:
type Element struct {
    Tag     string
    Keyword string
    VR 	    string
    VM      string
    Retired bool
    Desc    string
}

// Get the tag value for the element with "x" nibbles replaced with the given
// hex character.  Will presumably produce either a high or low boundary on
// the range that this tag can support.
func (e Element) GetTagBoundary(replacement string) *uint32 {
	matches := tagRegex.FindStringSubmatch(e.Tag)
	if (matches == nil) {
		return nil
	}
	
	hex := matches[1] + matches[2]
	hex = strings.Replace(hex, "x", replacement, -1)
	
	value, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return nil
	}
	
	value32 := uint32(value)
	return &value32
}

// Get the lowest tag the element can have, which will normally be the only
// one.
func (e Element) GetTagLowValue() *uint32 {
	return e.GetTagBoundary("0")
}

// Get the higest tag the element can have, which will normally be the same as
// the lowest, except for retired elements that are represented by a range of
// tags.
func (e Element) GetTagHighValue() *uint32 {
	return e.GetTagBoundary("F")
}

func (e Element) GetKeyword() string {
	return regexp.MustCompile(`[^[:alnum:]]`).ReplaceAllString(e.Keyword, "")
}

// Unimplemented, not sure if useful
func (e Element) GetLowMultiplicity() *uint32 {
	return nil
}

// Unimplemented, not sure if useful
func (e Element) GetHighMultiplicity() *uint32 {
	return nil
}

func (e Element) GetVR() string {
	tag := e.GetTagLowValue()
	
	if tag == nil || *tag == 0xFFFEE000 || *tag == 0xFFFEE00D || *tag == 0xFFFEE0DD {
		return "UN"
	}

	// TODO: parse better, validate known possible vrs
	if len(e.VR) > 1 {
		return e.VR[:2]
	}

	return "UN"
}

func (e Element) String() string {
	base := e.GetKeyword()
	
	if e.GetTagLowValue() != nil {
		base = base + fmt.Sprintf(" (%08X)", *e.GetTagLowValue())
	}
	
	if e.GetTagHighValue() != nil && e.GetTagHighValue() != e.GetTagLowValue() {
		base = base + fmt.Sprintf("..(%08X)", *e.GetTagHighValue())
	}
	
	return base  
}

func NewElement(row Row) Element {
	return Element{
		Tag:     row.Cells[0].GetValue(),
		Keyword: row.Cells[2].GetValue(),
		VR:      row.Cells[3].GetValue(),
		VM:      row.Cells[4].GetValue(),
		Retired: row.Cells[4].GetValue() == "RET",
		Desc:    row.Cells[1].GetValue(),
	}
}

type elements map[uint32]Element

// TODO: flags to trigger only specific go files data dictionary generation
// TODO: flag to automatically fetch latest spec
func main() {
	elements := make(elements)
	parseStandard("part06.xml", elements, "6", "7", "8")
	parseStandard("part07.xml", elements, "E")
	
	writeTags(elements)
	writeDictionary(elements)
}

func parseStandard(filename string, elements elements, chapters ... string) {
	stream, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	
	decoder := xml.NewDecoder(stream)
	var book Book
	err = decoder.Decode(&book)
	
	if err != nil {
		log.Fatal(err)
	}
	
	tagChapters := NewSet(chapters...)
	
	for _, chapter := range book.Chapters {
		if !tagChapters.Contains(chapter.Label) {
			continue
		}
		
		for _, table := range chapter.GetTables() {
			for _, row := range table.Rows {
				element := NewElement(row)
				
				if element.GetTagLowValue() != nil {
					elements[*element.GetTagLowValue()] = element
				} else {
					log.Printf("WARNING: no tag value for %s\n", element)
				}
			}
		}
	}
}

func forEach(elements elements, f func (element Element)) {
	keys := make(tagSlice, 0, len(elements))
	for key := range elements {
		keys = append(keys, key)
	}
	sort.Sort(keys)
	
	for _, tag := range keys {
		f(elements[tag])
	}
}

func writeTags(elements elements) {
	tags_go, err := os.Create("tags.go")
	if err != nil {
		log.Fatal(err)
	}
	defer tags_go.Close()
	
	fmt.Fprintln(tags_go, "package dcm")
	fmt.Fprintln(tags_go, "")
	fmt.Fprintln(tags_go, "// auto-generated, do not edit")
	fmt.Fprintln(tags_go, "")
	fmt.Fprintln(tags_go, "const (")
	
	maxLen := 0
	forEach(elements, func (element Element) {
		newLen := len(element.GetKeyword())
		if newLen > maxLen {
			maxLen = newLen 
		}
	})
	
	maxLenStr := fmt.Sprintf("%d", maxLen)
	forEach(elements, func (element Element) {
		// TODO: add comments with other attributes
		if element.GetKeyword() != "" {
			fmt.Fprintf(tags_go, "\t%-" + maxLenStr + "s = Tag(0x%08X)\n",
				element.GetKeyword(), *element.GetTagLowValue())
		}
	})
	
	fmt.Fprintln(tags_go, ")")
}

func writeDictionary(elements elements) {
	stddict_go, err := os.Create("stddict.go")
	if err != nil {
		log.Fatal(err)
	}
	defer stddict_go.Close()
	
	fmt.Fprintf(stddict_go, stddict_header)
	
	forEach(elements, func (element Element) {
		// TODO: support for multi-tag elements
		// - either define a single value and map it multiple times
		// - or enhance datadict.go to search for these somehow...
		fmt.Fprintf(stddict_go, "\t\t" + elementSpecPattern + "\n",
			*element.GetTagLowValue(),
			*element.GetTagLowValue(),
			*element.GetTagHighValue(),
			element.GetVR(),
			element.Retired,
			element.Desc,
			element.GetKeyword(),
		)
	})
	
	fmt.Fprintf(stddict_go, "\t})\n")
}

const stddict_header = `package dcm

// auto-generated, do not edit

var stddict = NewDataDictionary("",
	map[Tag]ElementSpec{
`

const elementSpecPattern = `Tag(0x%08X): ElementSpec{tag: Tag(0x%08X), ` +
	`maxValue: Tag(0x%08X), vr: %s, retired: %t, desc: "%s", keyword: "%s"},`

// dumb util for set membership check
type Set map[string]*struct{}
func NewSet(items ...string) Set {
	set := make(Set)
	for _, item := range items {
		set[item] = nil
	}
	return set
}
func (set Set) Contains(item string) bool {
	_, ok := set[item]
	return ok
}

// for sorting of uint32s
type tagSlice []uint32
func (p tagSlice) Len() int           { return len(p) }
func (p tagSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p tagSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
