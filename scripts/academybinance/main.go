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

type AppData struct {
	DynamicIds []string `json:"dynamicIds"`
	Ssr        bool     `json:"ssr"`
	State      AppState `json:"appState"`
}

type AppState struct {
	Loader DataLoader `json:"loader"`
}

type DataLoader struct {
	Data DataByRouteId `json:"dataByRouteId"`
}

type DataByRouteId struct {
	Id D174 `json:"d174"`
}

type D174 struct {
	Glossaries []Glossary `json:"glossaries"`
}

type Glossary struct {
	Title      string        `json:"title"`
	Slug       string        `json:"slug"`
	Excerpt    string        `json:"excerpt"`
	Visible    bool          `json:"visible"`
	Difficulty DifficultyObj `json:"difficulty"`
	Content    string        `json:"content"`
}

type DifficultyObj struct {
	Level int    `json:"level"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

func (glossary Glossary) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "slug")
	headers = append(headers, "excerpt")
	headers = append(headers, "difficulty")
	headers = append(headers, "content")

	return headers
}

func (glossary Glossary) GetValues() []string {
	values := make([]string, 0)

	values = append(values, glossary.Title)
	values = append(values, glossary.Slug)
	values = append(values, glossary.Excerpt)
	values = append(values, glossary.Difficulty.Label)
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
	glossary.Excerpt = strings.Replace(glossary.Excerpt, "\n", "", -1)

	glossary.Excerpt = strings.Replace(glossary.Excerpt, "\"", "", -1)

	var input = fmt.Sprintf(`---
	title: "%s"
	slug: "%s"
	difficulty: "%s"
	excerpt: "%s"
	---
	%s`, glossary.Title, glossary.Slug, glossary.Difficulty.Label, strings.Replace(glossary.Excerpt, "\n", "", -1), glossary.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getGlossaries(c *colly.Collector) {
	glossaries := make([]Glossary, 0)

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		if !strings.Contains(err.Error(), "Not Found") {
			log.Fatal(err)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second * 2)
	})

	c.OnHTML("#__APP_DATA", func(h *colly.HTMLElement) {
		data := AppData{}
		json.Unmarshal([]byte(h.Text), &data)

		for _, glossary := range data.State.Loader.Data.Id.Glossaries {

			var content = ""

			c.OnHTML(".css-g5rsps", func(h *colly.HTMLElement) {
				content = h.Text

				h.ForEach(".css-13l1dht", func(i int, h *colly.HTMLElement) {
					text := h.Text
					if strings.Contains(text, "Community Submission - Author:") {
						content = strings.Replace(content, text, "", 1)
					}
					if strings.Contains(text, "Learn more:") {
						content = strings.Replace(content, text, "", 1)
					}
					if strings.Contains(text, "Learn More:") {
						content = strings.Replace(content, text, "", 1)
					}
					if strings.Contains(text, "Also learn:") {
						content = strings.Replace(content, text, "", 1)
					}
					if strings.Contains(text, "Also Learn:") {
						content = strings.Replace(content, text, "", 1)
					}
				})

				nav := h.ChildText(".css-1j2hb68")
				title := h.ChildText("h1")
				difficulty := h.ChildText(".css-11w9015")
				content = strings.Replace(content, nav, "", 1)
				content = strings.Replace(content, title, "", 1)
				content = strings.Replace(content, difficulty, "", 1)
			})

			c.Visit(fmt.Sprintf("https://academy.binance.com/en/glossary/%s", glossary.Slug))

			glossary.Content = content

			fmt.Printf("Downloaded: %s\n", glossary.Title)

			glossaries = append(glossaries, glossary)

			file, err := json.MarshalIndent(glossary, "", " ")
			if err != nil {
				log.Fatal(err)
			}

			glossary.Title = strings.Replace(glossary.Title, "\n", "", -1)
			glossary.Slug = strings.Replace(glossary.Slug, "\n", "", -1)
			glossary.Excerpt = strings.Replace(glossary.Excerpt, "\n", "", -1)
			glossary.Difficulty.Label = strings.Replace(glossary.Difficulty.Label, "\n", "", -1)
			glossary.Content = strings.Replace(glossary.Content, "\n", "", -1)

			if err := os.Mkdir(fmt.Sprintf("../../output/academy.binance.com/%s", glossary.Title), os.ModePerm); err != nil {
				if !strings.Contains(err.Error(), "file exists") {
					log.Fatal(err)
				}
			}

			filepathJson := fmt.Sprintf("../../output/academy.binance.com/%s/article.json", glossary.Title)
			filepathCsv := fmt.Sprintf("../../output/academy.binance.com/%s/article.csv", glossary.Title)
			filepathMd := fmt.Sprintf("../../output/academy.binance.com/%s/article.md", glossary.Title)

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

				m := map[string]string{
					"title": "term",
				}

				utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

				createdJson.Close()
				createdMd.Close()
				createdCsv.Close()

			} else {
				f.Close()
			}
		}
	})

	c.Visit("https://academy.binance.com/en/glossary")
}

func main() {
	if err := os.Mkdir("../../output/academy.binance.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	getGlossaries(c)
}
