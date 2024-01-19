package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"scripts/utils"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Glossary struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (glossary Glossary) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "content")

	return headers
}

func (glossary Glossary) GetValues() []string {
	values := make([]string, 0)

	values = append(values, glossary.Title)
	values = append(values, glossary.Content)

	return values
}

func writeToCsv(csvFile os.File, glossary Glossary) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(glossary.GetHeaders())
	csvWriter.Write(glossary.GetValues())
	csvWriter.Flush()
}

func writeToMd(mdFile os.File, glossary Glossary) {
	var input = fmt.Sprintf(`---
	title: "%s"
	---
	%s`,
		glossary.Title, glossary.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getGlossaries(c *colly.Collector) []Glossary {
	glossaries := make([]Glossary, 0)

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnHTML("script", func(h *colly.HTMLElement) {
		glossary := Glossary{}

		glossary.Title = h.ChildText("h5")
		glossary.Content = h.ChildText("p")

		print(glossary.Title)

		glossaries = append(glossaries, glossary)
	})

	c.OnHTML("html", func(h *colly.HTMLElement) {
		print(h.Text)
	})

	c.Visit("https://www.coindesk.com/learn/glossary/")

	return glossaries
}

func main() {
	if err := os.Mkdir("../../output/coindeskglossary", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	glosaries := getGlossaries(c)

	for _, glossary := range glosaries {
		file, err := json.MarshalIndent(glossary, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		if err := os.Mkdir(fmt.Sprintf("../../output/coindeskglossary/%s", glossary.Title), os.ModePerm); err != nil {
			if !strings.Contains(err.Error(), "file exists") {
				log.Fatal(err)
			}
		}

		filepathJson := fmt.Sprintf("../../output/coindeskglossary/%s/article.json", glossary.Title)
		filepathCsv := fmt.Sprintf("../../output/coindeskglossary/%s/article.csv", glossary.Title)
		filepathMd := fmt.Sprintf("../../output/coindeskglossary/%s/article.md", glossary.Title)

		f, err := os.OpenFile(filepathJson, os.O_RDONLY, 0666)

		if err != nil && f == nil {

			f.Close()

			createdJson, json_err := os.OpenFile(filepathJson, os.O_WRONLY|os.O_CREATE, 0666)
			createdCsv, csv_err := os.OpenFile(filepathCsv, os.O_WRONLY|os.O_CREATE, 0666)
			createdMd, md_err := os.OpenFile(filepathMd, os.O_WRONLY|os.O_CREATE, 0666)

			if json_err != nil {
				log.Fatal(err)
			}

			if csv_err != nil {
				log.Fatal(err)
			}

			if md_err != nil {
				log.Fatal(err)
			}

			createdJson.Write(file)

			writeToCsv(*createdCsv, glossary)

			writeToMd(*createdMd, glossary)

			m := map[string]string{}

			utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

			createdJson.Close()
			createdMd.Close()
			createdCsv.Close()

		} else {
			f.Close()
		}
	}
}
