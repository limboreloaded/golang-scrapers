package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"scripts/utils"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"golang.org/x/net/html"
)

type Article struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Image     string `json:"image"`
	Published string `json:"published"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	Category  string `json:"category"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "image")
	headers = append(headers, "published")
	headers = append(headers, "content")
	headers = append(headers, "author")
	headers = append(headers, "category")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Id)
	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Image)
	values = append(values, article.Published)
	values = append(values, article.Content)
	values = append(values, article.Author)
	values = append(values, article.Category)

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
	image: "%s"
	published: "%s"
	author: "%s"
	category: "%s"
	---
	%s`, article.Id, article.Title, article.Link, article.Image, article.Published, article.Author, article.Category, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getRssArticles() []Article {
	articles := make([]Article, 0)

	paragraphs := make([]string, 0)

	var image string

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
		category := strings.Join(x.ChildTexts("category"), ", ")

		x.Request.Visit(link)

		content := strings.Join(paragraphs, "")

		content = strings.Replace(content, "\n", "", -1)

		article := Article{Title: title, Link: link, Id: guid, Published: published, Author: author, Category: category, Content: content, Image: image}

		paragraphs = paragraphs[:0]

		fmt.Printf("Downloading: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.OnHTML(".single-post-main-middle", func(h *colly.HTMLElement) {
		for _, node := range h.DOM.Children().Nodes {
			if node.Data == "div" || node.Data == "iframe" {
				continue
			}
			var b bytes.Buffer
			err := html.Render(&b, node)

			if err != nil {
				log.Fatal(err)
			}

			p := strings.NewReader(b.String())
			doc, _ := goquery.NewDocumentFromReader(p)
			paragraphs = append(paragraphs, doc.Text())
		}
	})

	c.OnHTML(".single-post-image", func(h *colly.HTMLElement) {
		time.Sleep(time.Second * 2)
		src := h.ChildAttr("img", "src")
		image = src
	})

	c.Visit("https://investing-api-eng.ambcrypto.com/feed/merge_category/")
	return articles
}

func main() {
	if err := os.Mkdir("../../output/ambcrypto.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/ambcrypto.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/ambcrypto.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/ambcrypto.com/%s.md", article.Title)

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
				"category":  "categories",
				"published": "created_at",
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
