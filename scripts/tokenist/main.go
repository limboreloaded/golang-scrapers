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

	// "github.com/cixtor/readability"

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
	headers = append(headers, "category")
	headers = append(headers, "description")
	headers = append(headers, "content")
	headers = append(headers, "author")
	headers = append(headers, "image")

	return headers
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

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Id)
	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Published)
	values = append(values, article.Category)
	values = append(values, article.Description)
	values = append(values, article.Content)
	values = append(values, article.Author)
	values = append(values, article.Image)

	return values
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func getRssArticles() []Article {
	var image string

	articles := make([]Article, 0)

	c := colly.NewCollector()

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(2 * time.Second)
	})

	c.OnHTML(".section-post__img-img", func(h *colly.HTMLElement) {
		image = h.ChildAttr("img", "data-lazy-src")
	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		article := Article{}

		article.Id = x.ChildText("guid")
		article.Title = x.ChildText("title")
		article.Link = x.ChildText("link")
		article.Category = strings.Join(x.ChildTexts("category"), ", ")
		article.Published = x.ChildText("pubDate")
		article.Author = x.ChildText("dc:creator")

		r := readability.New()

		desc, _ := r.Parse(strings.NewReader(x.ChildText("description")), article.Link)
		cont, _ := r.Parse(strings.NewReader(x.ChildText("content:encoded")), article.Link)

		article.Description = strings.Replace(desc.TextContent, "\n", "", -1)
		article.Description = strings.Replace(article.Description, "\t", "'", -1)
		article.Description = strings.Replace(article.Description, "\"", "'", -1)

		article.Content = strings.Replace(cont.TextContent, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		x.Request.Visit(article.Link)

		article.Image = image

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://tokenist.com/feed/")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/tokenist.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/tokenist.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/tokenist.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/tokenist.com/%s.md", article.Title)

		f, err := os.OpenFile(filepathJson, os.O_RDONLY, 0666)

		if err != nil && f == nil {
			createdJson, json_err := os.OpenFile(filepathJson, os.O_WRONLY|os.O_CREATE, 0666)
			createdCsv, csv_err := os.OpenFile(filepathCsv, os.O_WRONLY|os.O_CREATE, 0666)
			createdMd, md_err := os.OpenFile(filepathMd, os.O_WRONLY|os.O_CREATE, 0666)

			createdJson.Write(file)

			writeToCsv(*createdCsv, article)

			writeToMd(*createdMd, article)

			if json_err != nil {
				log.Fatal(err)
			}

			if csv_err != nil {
				log.Fatal(err)
			}

			if md_err != nil {
				log.Fatal(err)
			}

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
