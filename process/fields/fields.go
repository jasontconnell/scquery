package fields

import (
	"errors"

	"github.com/jasontconnell/sitecore/data"
)

type FieldProcessor interface {
	Process(m data.ItemMap, s string) (FieldProcessResult, error)
}

type FieldProcessResult struct {
	Value string
	Raw   string
}

var fieldProcessorMap map[string]FieldProcessor

func init() {
	fieldProcessorMap = make(map[string]FieldProcessor)
	fieldProcessorMap["General Link"] = generalLinkProcessor{}
	fieldProcessorMap["Single-Line Text"] = textProcessor{}
	fieldProcessorMap["Multi-Line Text"] = textProcessor{}
}

func Process(m data.ItemMap, tm data.TemplateMap, fv data.FieldValueNode) (FieldProcessResult, error) {
	item := m[fv.GetItemId()]
	template := tm[item.GetTemplateId()]

	f := template.GetField(fv.GetFieldId())

	fp := FieldProcessResult{}
	if f == nil {
		return fp, errors.New("couldn't find field for id " + fv.GetFieldId().String() + " " + fv.GetName())
	}

	processor, ok := fieldProcessorMap[f.GetType()]
	if !ok {
		return fp, errors.New("couldn't find processor for field " + fv.GetFieldId().String() + " " + fv.GetName() + " " + f.GetType())
	}
	return processor.Process(m, fv.GetValue())
}
