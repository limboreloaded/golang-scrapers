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

type Article struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Published   string `json:"published"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Author      string `json:"author"`
	Image       string `json:"image"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "published")
	headers = append(headers, "description")
	headers = append(headers, "content")
	headers = append(headers, "author")
	headers = append(headers, "category")
	headers = append(headers, "image")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Id)
	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Published)
	values = append(values, article.Description)
	values = append(values, article.Content)
	values = append(values, article.Author)
	values = append(values, article.Category)
	values = append(values, article.Image)

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
	category: "%s"
	description: "%s"
	author: "%s"
	image: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Category, article.Description, article.Author, article.Image, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getRssArticles() []Article {
	var content string

	articles := make([]Article, 0)

	c := colly.NewCollector()

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(2 * time.Second)
	})

	c.OnHTML(".at-body", func(h *colly.HTMLElement) {
		content = h.Text
	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		title := x.ChildText("title")
		link := x.ChildText("link")
		guid := x.ChildText("guid")
		published := x.ChildText("pubDate")
		author := x.ChildText("dc:creator")
		description := x.ChildText("description")
		category := strings.Join(x.ChildTexts("category"), ", ")
		media_content := x.ChildAttr("media:content", "url")

		x.Request.Visit(link)

		media_content = strings.Replace(media_content, "\n", "", -1)
		media_content = strings.Replace(media_content, "\t", "", -1)
		media_content = strings.Replace(media_content, "\"", "'", -1)

		description = strings.Replace(description, "\n", "", -1)
		description = strings.Replace(description, "\t", "", -1)
		description = strings.Replace(description, "\"", "'", -1)

		article := Article{Title: title, Link: link, Id: guid, Published: published, Author: author, Description: description, Image: media_content, Category: category, Content: content}

		fmt.Printf("Dowloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://www.coindesk.com/arc/outboundfeeds/rss/")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/coindesk.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/coindesk.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/coindesk.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/coindesk.com/%s.md", article.Title)

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
