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
	headers = append(headers, "content")

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
	author: "%s"
	image: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Description, article.Author, article.Image, article.Content)

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

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		title := x.ChildText("title")
		link := x.ChildText("link")
		guid := x.ChildText("guid")
		published := x.ChildText("pubDate")
		author := x.ChildText("dc:creator")
		image := x.ChildAttr("media:thumbnail", "url")

		r := readability.New()

		desc, _ := r.Parse(strings.NewReader(x.ChildText("description")), link)
		cont, _ := r.Parse(strings.NewReader(x.ChildText("content:encoded")), link)

		content := cont.TextContent
		content = strings.Replace(content, "\n", "", -1)
		content = strings.Replace(content, "\t", "", -1)
		content = strings.Replace(content, "\"", "'", -1)

		description := desc.TextContent
		description = strings.Replace(description, "\n", "", -1)
		description = strings.Replace(description, "\t", "", -1)
		description = strings.Replace(description, "\"", "'", -1)

		article := Article{Title: title, Link: link, Id: guid, Published: published, Author: author, Description: description, Content: content, Image: image}

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://coinpaprika.com/news/feed/")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/coinpaprika.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/coinpaprika.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/coinpaprika.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/coinpaprika.com/%s.md", article.Title)

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
