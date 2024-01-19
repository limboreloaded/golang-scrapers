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

	"github.com/cixtor/readability"
	"github.com/gocolly/colly"
)

type Article struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Image       string `json:"image"`
	Published   string `json:"published"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Category    string `json:"category"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "image")
	headers = append(headers, "published")
	headers = append(headers, "description")
	headers = append(headers, "author")
	headers = append(headers, "category")
	headers = append(headers, "content")
	headers = append(headers, "post-id")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Id)
	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Image)
	values = append(values, article.Published)
	values = append(values, article.Description)
	values = append(values, article.Author)
	values = append(values, article.Category)
	values = append(values, article.Content)

	return values
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func writeToMd(mdFile os.File, article Article) {
	var input = fmt.Sprintf(`---
	id: "%s"
	title: "%s"
	link: "%s"
	published: "%s"
	description: "%s"
	category: "%s"
	author: "%s"
	image: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Description, article.Category, article.Author, article.Image, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getRssArticles() []Article {
	articles := make([]Article, 0)

	c := colly.NewCollector()

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(2 * time.Second)
	})

	img := ""
	paragraphs := make([]string, 0)

	c.OnHTML(".thumbnail-container", func(h *colly.HTMLElement) {
		img = h.ChildAttr("img", "src")
	})

	c.OnHTML(".content-inner", func(h *colly.HTMLElement) {
		h.ForEach("p", func(i int, h *colly.HTMLElement) {
			paragraphs = append(paragraphs, h.Text)
		})
	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		r := readability.New()

		link := x.ChildText("link")
		title := x.ChildText("title")
		published := x.ChildText("pubDate")
		author := x.ChildText("dc:creator")
		category := strings.Join(x.ChildTexts("category"), ", ")
		pid := x.ChildText("post-id")
		description := x.ChildText("description")

		description = strings.Replace(description, "\n", "", -1)
		description = strings.Replace(description, "\t", "", -1)
		description = strings.Replace(description, "\"", "'", -1)

		d, err := r.Parse(strings.NewReader(description), link)

		if err != nil {
			log.Fatal(err)
		}

		x.Request.Visit(link)

		article := Article{Link: link, Description: d.TextContent, Title: title, Published: published, Image: img, Id: pid, Author: author, Category: category, Content: strings.Join(paragraphs[:len(paragraphs)-1], "")}

		article.Content = strings.Replace(article.Content, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		paragraphs = paragraphs[:0]

		fmt.Printf("Downloaded: %s", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://dailyhodl.com/feed/")
	return articles
}

func main() {
	if err := os.Mkdir("../../output/dailyhodl.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	rssArticles := getRssArticles()
	for _, article := range rssArticles {
		file, err := json.MarshalIndent(article, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		filepathJson := fmt.Sprintf("../../output/dailyhodl.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/dailyhodl.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/dailyhodl.com/%s.md", article.Title)

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
