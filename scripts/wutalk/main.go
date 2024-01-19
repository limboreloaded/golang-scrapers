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

	"github.com/flytam/filenamify"
	"github.com/gocolly/colly"
)

type Article struct {
	Title     string `json:"title"`
	Link      string `json:"link"`
	Image     string `json:"image"`
	Published string `json:"published"`
	Author    string `json:"author"`
	Content   string `json:"content"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "image")
	headers = append(headers, "published")
	headers = append(headers, "author")
	headers = append(headers, "content")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Image)
	values = append(values, article.Published)
	values = append(values, article.Author)
	values = append(values, article.Content)

	return values
}

func writeToMd(mdFile os.File, article Article) {
	var input = fmt.Sprintf(`---
	title: "%s"
	link: "%s"
	published: "%s"
	author: "%s"
	image: "%s"
	---
	%s`,
		article.Title, article.Link, article.Published, article.Author, article.Image, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getArticles(c *colly.Collector) []Article {
	articles := make([]Article, 0)

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second * 2)
	})

	c.OnHTML(".aListItem", func(h *colly.HTMLElement) {
		article := Article{}

		pict := h.ChildAttr("img", "src")
		title := strings.Replace(h.ChildText(".aListItemTit"), "&nbsp;", "", -1)
		link := strings.Replace(h.ChildAttr("a", "href"), "\u0026", "&", -1)

		published := h.ChildText(".aListItemTime")

		article.Author = ""
		article.Image = pict
		article.Title = title
		article.Published = published
		article.Link = link

		c.OnHTML("strong", func(h *colly.HTMLElement) {
			if strings.Contains(h.Text, "Author：") {
				article.Author = strings.Replace(h.Text, "Author：", "", -1)
			}

			if strings.Contains(h.Text, "Author:") {
				article.Author = strings.Replace(h.Text, "Author: ", "", -1)
			}
		})

		c.OnHTML(".path", func(h *colly.HTMLElement) {
			split := strings.Split(h.Text, ".")
			article.Published = strings.Replace(split[1], " ", "", -1)
			article.Published = strings.Replace(article.Published, "\t", "", -1)
			article.Published = strings.Replace(article.Published, "\n", "", -1)

			if article.Author == "" {
				article.Author = strings.Replace(split[0], " ", "", -1)
				article.Author = strings.Replace(article.Author, "\t", "", -1)
				article.Author = strings.Replace(article.Author, "\n", "", -1)
			}
		})

		c.OnHTML(".entry-content", func(h *colly.HTMLElement) {
			entry := h.Text

			article.Content = strings.Replace(entry, fmt.Sprintf("Author: %s", article.Author), "", -1)
			article.Content = strings.Replace(article.Content, fmt.Sprintf("Author：%s", article.Author), "", -1)

			article.Content = strings.Replace(article.Content, "\n", "", -1)
			article.Content = strings.Replace(article.Content, "\t", "", -1)
			article.Content = strings.Replace(article.Content, "\"", "'", -1)
		})

		c.Visit(article.Link)

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://www.wu-talk.com/html/en/")

	return articles
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func main() {
	if err := os.Mkdir("../../output/wutalk.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	articles := getArticles(c)

	for _, article := range articles {
		file, err := json.MarshalIndent(article, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		title, _ := filenamify.FilenamifyV2(article.Title)

		filepathJson := fmt.Sprintf("../../output/wutalk.com/%s.json", title)
		filepathCsv := fmt.Sprintf("../../output/wutalk.com/%s.csv", title)
		filepathMd := fmt.Sprintf("../../output/wutalk.com/%s.md", title)

		f, err := os.OpenFile(filepathJson, os.O_RDONLY, 0666)

		if err != nil && f == nil {
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

			writeToCsv(*createdCsv, article)

			writeToMd(*createdMd, article)

			m := map[string]string{}

			utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

			createdJson.Close()
			createdCsv.Close()
			createdMd.Close()
		} else {
			f.Close()
		}
	}
}
