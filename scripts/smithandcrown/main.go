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

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second * 2)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnHTML(".gl-block__links", func(h *colly.HTMLElement) {
		var glossary Glossary

		c.OnHTML(".glossary-title__wrap", func(h *colly.HTMLElement) {
			glossary.Title = h.Text
		})

		c.OnHTML(".description-col", func(h *colly.HTMLElement) {
			glossary.Content = h.Text
			glossary.Content = strings.Replace(glossary.Content, "\n", "", -1)
			glossary.Content = strings.Replace(glossary.Content, "\t", "", -1)
			glossary.Content = strings.Replace(glossary.Content, "\"", "'", -1)
		})

		c.Visit(fmt.Sprintf("https://www.smithandcrown.com/%s", h.ChildAttr("a", "href")))

		fmt.Printf("Downloaded: %s\n", glossary.Title)

		glossaries = append(glossaries, glossary)
	})

	c.Visit("https://www.smithandcrown.com/glossary")

	return glossaries
}

func main() {
	if err := os.Mkdir("../../output/smithandcrown", os.ModePerm); err != nil {
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

		if err := os.Mkdir(fmt.Sprintf("../../output/smithandcrown/%s", glossary.Title), os.ModePerm); err != nil {
			if !strings.Contains(err.Error(), "file exists") {
				log.Fatal(err)
			}
		}

		filepathJson := fmt.Sprintf("../../output/smithandcrown/%s/article.json", glossary.Title)
		filepathCsv := fmt.Sprintf("../../output/smithandcrown/%s/article.csv", glossary.Title)
		filepathMd := fmt.Sprintf("../../output/smithandcrown/%s/article.md", glossary.Title)

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
