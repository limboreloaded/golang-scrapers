package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"scripts/utils"
	"strings"

	"github.com/flytam/filenamify"
	"github.com/gocolly/colly"
)

type Glossary struct {
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	Content    string `json:"content"`
}

func (glossary Glossary) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "difficulty")
	headers = append(headers, "content")

	return headers
}

func (glossary Glossary) GetValues() []string {
	values := make([]string, 0)

	values = append(values, glossary.Title)
	values = append(values, glossary.Difficulty)
	values = append(values, glossary.Content)

	return values
}

func writeToMd(mdFile os.File, glossary Glossary) {
	var input = fmt.Sprintf(`---
	title: "%s"
	difficulty: "%s"
	---
	%s`,
		glossary.Title, glossary.Difficulty, glossary.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func writeToCsv(csvFile os.File, glossary Glossary) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(glossary.GetHeaders())
	csvWriter.Write(glossary.GetValues())
	csvWriter.Flush()
}

func getGlossary(c *colly.Collector, link string) Glossary {
	glossary := Glossary{}

	c.OnHTML(".bjHQTa", func(h *colly.HTMLElement) {
		glossary.Title = h.Text
	})

	c.OnHTML(".Label__StyledLabel-sc-1t4rrpc-0", func(h *colly.HTMLElement) {
		glossary.Difficulty = h.Text
	})

	c.OnHTML(".dCEoLC", func(h *colly.HTMLElement) {
		glossary.Content = h.Text
		glossary.Content = strings.Replace(glossary.Content, "\n", "", -1)
		glossary.Content = strings.Replace(glossary.Content, "\t", "", -1)
		glossary.Content = strings.Replace(glossary.Content, "\"", "'", -1)
	})

	c.OnHTML(".jzoMnb", func(h *colly.HTMLElement) {
		glossary.Content = strings.Replace(glossary.Content, h.Text, "", 1)
	})

	c.Visit(link)

	glossary.Content = strings.Replace(glossary.Content, glossary.Title, "", 1)
	glossary.Content = strings.Replace(glossary.Content, glossary.Difficulty, "", 1)

	if glossary.Title == "" {
		return glossary
	}

	fmt.Printf("Downloaded: %s\n", glossary.Title)

	return glossary
}

func main() {
	if err := os.Mkdir("../../output/coinmarketcap.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}
	c := colly.NewCollector()

	c.OnHTML("a", func(h *colly.HTMLElement) {
		if strings.Contains(h.Attr("href"), "/academy/glossary") {
			glossary := getGlossary(c, "https://coinmarketcap.com"+h.Attr("href"))

			if glossary.Title == "" {
				return
			}

			file, err := json.MarshalIndent(glossary, "", " ")
			if err != nil {
				log.Fatal(err)
			}

			title, _ := filenamify.FilenamifyV2(glossary.Title)

			if err := os.Mkdir(fmt.Sprintf("../../output/coinmarketcap.com/%s", title), os.ModePerm); err != nil {
				if !strings.Contains(err.Error(), "file exists") {
					log.Fatal(err)
				}
			}

			filepathJson := fmt.Sprintf("../../output/coinmarketcap.com/%s/article.json", title)
			filepathCsv := fmt.Sprintf("../../output/coinmarketcap.com/%s/article.csv", title)
			filepathMd := fmt.Sprintf("../../output/coinmarketcap.com/%s/article.md", title)

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
	})

	c.Visit("https://coinmarketcap.com/academy/glossary")
}
