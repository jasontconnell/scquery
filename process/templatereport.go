package process

import (
	"fmt"
	"html/template"
	"os"
	"sort"

	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

type TemplateReportModel struct {
	Templates []TemplateModel
}

type TemplateModel struct {
	Name     string
	ID       string
	Inherits []*TemplateModel
	Fields   []FieldModel
}

type FieldModel struct {
	Name string
	Type string
}

func ExecuteTemplateReport(connstr, path string) error {
	tlist, err := api.LoadTemplates(connstr)
	if err != nil {
		return fmt.Errorf("getting templates: %w", err)
	}

	tmap := api.GetTemplateMap(tlist)
	tmap = api.FilterTemplateMap(tmap, []string{path})
	templates := []TemplateModel{}

	mmap := make(map[uuid.UUID]TemplateModel)

	for _, tmpl := range tmap {
		tmodel := getTemplateModel(tmpl, mmap, 0)

		templates = append(templates, tmodel)
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	reportModel := TemplateReportModel{Templates: templates}
	t, err := template.New("templateReport.txt").ParseFiles("./tmpl/templateReport.txt")
	if err != nil {
		return fmt.Errorf("problem parsing template: %w", err)
	}
	return t.Execute(os.Stdout, reportModel)
}

func getTemplateModel(tmpl data.TemplateNode, mmap map[uuid.UUID]TemplateModel, count int) TemplateModel {
	model := TemplateModel{ID: idstring(tmpl.GetId()), Name: tmpl.GetName()}
	if count > 5 {
		return model
	}

	if existing, ok := mmap[tmpl.GetId()]; ok {
		return existing
	}

	for _, base := range tmpl.GetBaseTemplates() {
		var bmodel TemplateModel
		bmodel, ok := mmap[base.GetId()]
		if !ok {
			bmodel = getTemplateModel(base, mmap, count+1)
			mmap[base.GetId()] = bmodel
		}

		model.Inherits = append(model.Inherits, &bmodel)
	}

	model.Fields = getTemplateFields(tmpl)
	return model
}

func getTemplateFields(tmpl data.TemplateNode) []FieldModel {
	fields := []FieldModel{}

	for _, f := range tmpl.GetFields() {
		fm := FieldModel{Name: f.GetName(), Type: f.GetType()}
		fields = append(fields, fm)
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
	return fields
}
