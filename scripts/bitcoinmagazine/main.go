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

func writeToMd(mdFile os.File, article Article) {
	var input = fmt.Sprintf(`---
	id: "%s"
	title: "%s"
	link: "%s"
	published: "%s"
	author: "%s"
	---
	%s
	`, article.Id, article.Title, article.Link, article.Published, article.Author, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
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
		guid := x.ChildText("guid")
		title := x.ChildText("title")
		link := x.ChildText("link")
		published := x.ChildText("pubDate")
		author := x.ChildText("dc:creator")

		r := readability.New()
		description := x.ChildText("description")
		cont := x.ChildText("content:encoded")

		content, _ := r.Parse(strings.NewReader(cont), link)

		title = strings.Replace(title, "\"", "'", -1)

		textcont := strings.Replace(content.TextContent, "\n", "", -1)
		textcont = strings.Replace(textcont, "\t", "", -1)
		textcont = strings.Replace(textcont, "\"", "'", -1)

		category := strings.Join(x.ChildTexts("category"), ", ")
		media_thumbnail := x.ChildAttr("media:content", "url")

		article := Article{Id: guid, Title: title, Link: link, Published: published, Author: author, Description: description, Image: media_thumbnail, Category: category, Content: textcont}

		articles = append(articles, article)

		fmt.Printf("Downloaded: %s\n", article.Title)
	})

	c.Visit("https://bitcoinmagazine.com/.rss/full/")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/bitcoinmagazine.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/bitcoinmagazine.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/bitcoinmagazine.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/bitcoinmagazine.com/%s.md", article.Title)

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

			m := map[string]string{
				"category": "categories",
			}

			utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

			createdJson.Close()
			createdCsv.Close()
			createdMd.Close()
		} else {
			f.Close()
		}
	}
}
