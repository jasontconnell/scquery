package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jasontconnell/scquery/conf"
	"github.com/jasontconnell/scquery/process"
	"github.com/jasontconnell/sitecore/data"
)

func main() {
	c := flag.String("c", "config.json", "config filename with connectionString")
	tid := flag.String("template", "", "template id to search")
	report := flag.String("report", "", "run template report")
	path := flag.String("path", "", "path to search or include")
	flds := flag.String("fields", "", "csv of fields and values, like Field=Value to search")
	output := flag.String("output", "", "csv of fields to output per line")
	order := flag.String("order", "", "order by field")
	language := flag.String("lang", "en", "language")

	homePath := flag.String("home", "/sitecore/content/Home", "home root path")
	baseUrl := flag.String("baseUrl", "", "the root url")
	flag.Parse()

	if *tid == "" && *path == "" && *report != "" {
		flag.PrintDefaults()
		return
	}

	if *flds != "" {
		log.Println("parsing", *flds)
	}
	queryFields := process.GetFieldsQuery(strings.Split(*flds, ","))

	start := time.Now()
	cfg := conf.LoadConfig(*c)

	if *report == "" {
		var templateId uuid.UUID

		if *tid != "" {
			templateId = uuid.Must(uuid.Parse(*tid))
		}

		resultFields := []process.FieldResult{}
		if output != nil {
			sp := strings.Split(*output, ",")
			for _, s := range sp {
				resultFields = append(resultFields, process.FieldResult{Name: strings.Trim(s, " ")})
			}
		}

		log.Println("running query", templateId, *path, queryFields, resultFields, *language)
		res, err := process.RunQuery(cfg.ConnectionString, templateId, *path, queryFields, resultFields, data.GetLanguage(*language))
		if err != nil {
			log.Fatal(err)
		}

		if *order != "" && len(res) > 0 {
			orderIndex := -1
			fst := res[0]

			for i, f := range fst.Fields {
				if f.Name == *order {
					orderIndex = i
					break
				}
			}

			sort.Slice(res, func(i, j int) bool {
				iv, jv := res[i].Id.String(), res[j].Id.String()
				if orderIndex != -1 && orderIndex < len(res[i].Fields) {
					iv, jv = res[i].Fields[orderIndex].Value, res[j].Fields[orderIndex].Value
				}
				return iv < jv
			})
		}

		for _, r := range res {
			line := ""
			for _, f := range r.Fields {
				line += f.Value + ","
			}
			fmt.Println(strings.TrimSuffix(line, ","))
		}
	} else {
		if *report == "template" {
			err := process.ExecuteTemplateReport(cfg.ConnectionString, *path)
			if err != nil {
				log.Fatal(err)
			}
		} else if *report == "content" {
			err := process.ExecuteContentReport(cfg.ConnectionString, *path, *homePath, *baseUrl)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Println("finished", time.Since(start))
}
