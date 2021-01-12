package fields

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/data"
)

type generalLink struct {
	XmlName     xml.Name `xml:"link"`
	LinkType    string   `xml:"linktype,attr"`
	Url         string   `xml:"url,attr"`
	Target      string   `xml:"target,attr"`
	Text        string   `xml:"text,attr"`
	Id          string   `xml:"id,attr"`
	Title       string   `xml:"title,attr"`
	QueryString string   `xml:"querystring,attr"`
}

type generalLinkProcessor struct {
}

func (g generalLinkProcessor) Process(m data.ItemMap, s string) (FieldProcessResult, error) {
	b := bytes.NewBufferString(s)
	dec := xml.NewDecoder(b)

	link := generalLink{}
	err := dec.Decode(&link)
	if err != nil {
		return FieldProcessResult{}, fmt.Errorf("parsing xml %s", s)
	}

	val := link.Url
	switch link.LinkType {
	case "internal":
		uid, err := uuid.Parse(link.Id)
		if err != nil {
			i, ok := m[uid]
			if ok {
				val = i.GetPath()
			}
		}
	}

	return FieldProcessResult{Value: val, Raw: s}, nil
}
