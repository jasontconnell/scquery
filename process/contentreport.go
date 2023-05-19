package process

import (
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"

	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

type ContentReportModel struct {
	Contents []ContentModel
}

type ContentModel struct {
	ID         string
	Name       string
	Path       string
	UrlPath    string
	Url        string
	HasUrl     bool
	Template   TemplateModel
	Renderings []RenderingModel
}

type RenderingModel struct {
	ID           string
	Name         string
	DataSourceID string
}

func ExecuteContentReport(connstr, path, homePath, baseUrl string) error {
	items, err := api.LoadItems(connstr)
	if err != nil {
		return fmt.Errorf("loading items: %w", err)
	}

	templates, err := api.LoadTemplates(connstr)
	if err != nil {
		return fmt.Errorf("loading templtes: %w", err)
	}

	fieldValues, err := api.LoadFieldsParallel(connstr, 12)
	if err != nil {
		return fmt.Errorf("loading field values: %w", err)
	}

	tmap := api.GetTemplateMap(templates)
	_, imap := api.LoadItemMap(items)

	api.SetTemplates(imap, tmap)
	api.AssignFieldValues(imap, fieldValues)

	err = api.MapAllLayouts(imap, tmap, true)
	if err != nil {
		return fmt.Errorf("mapping layouts: %w", err)
	}

	imap = api.FilterItemMap(imap, []string{path})

	contents := []ContentModel{}
	for _, item := range imap {
		contents = append(contents, getContentModel(item, homePath, baseUrl))
	}

	sort.Slice(contents, func(i, j int) bool {
		return contents[i].Path < contents[j].Path
	})

	reportModel := ContentReportModel{Contents: contents}

	t, err := template.New("contentReport.txt").ParseFiles("./tmpl/contentReport.txt")
	if err != nil {
		return fmt.Errorf("problem parsing template: %w", err)
	}
	return t.Execute(os.Stdout, reportModel)
}

func getContentModel(item data.ItemNode, homePath, baseUrl string) ContentModel {
	var tmodel TemplateModel
	if item.GetTemplate() != nil {
		tmodel = TemplateModel{Name: item.GetTemplate().GetName(), ID: idstring(item.GetTemplateId())}
	}

	hasUrl := false
	url := ""
	if strings.HasPrefix(item.GetPath(), homePath) {
		if len(item.GetRenderings()) > 0 {
			hasUrl = true
			url = baseUrl + getUrlPath(strings.TrimPrefix(item.GetPath(), homePath))
		}
	}

	cm := ContentModel{
		ID:         idstring(item.GetId()),
		Name:       item.GetName(),
		Path:       item.GetPath(),
		UrlPath:    getUrlPath(item.GetPath()),
		HasUrl:     hasUrl,
		Url:        url,
		Template:   tmodel,
		Renderings: getRenderings(item),
	}

	return cm
}

func getRenderings(item data.ItemNode) []RenderingModel {
	list := []RenderingModel{}
	for _, d := range item.GetRenderings() {
		for _, r := range d.RenderingInstances {
			if r.Rendering.Item == nil {
				continue
			}
			rmodel := RenderingModel{
				ID:           idstring(r.Rendering.Item.GetId()),
				Name:         r.Rendering.Item.GetName(),
				DataSourceID: idstring(r.DataSourceId),
			}

			list = append(list, rmodel)
		}
	}
	return list
}
