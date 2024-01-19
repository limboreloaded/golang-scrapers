package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	Description string `json:"description"`
	Content     string `json:"content"`
	Author      string `json:"author"`
	Category    string `json:"category"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "guid")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "image")
	headers = append(headers, "published")
	headers = append(headers, "description")
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
	values = append(values, article.Description)
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
	var img string

	articles := make([]Article, 0)

	c := colly.NewCollector(colly.MaxDepth(0))

	url, _ := url.Parse("socks5://122.155.223.165:10203")

	c.SetProxyFunc(http.ProxyURL(url))

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "compass_uid=51f838e4-ee21-40d1-8d6c-4904faafea30; _ga=GA1.1.647314829.1700595461; ___nrbi=%7B%22firstVisit%22%3A1700595459%2C%22userId%22%3A%2251f838e4-ee21-40d1-8d6c-4904faafea30%22%2C%22userVars%22%3A%5B%5D%2C%22futurePreviousVisit%22%3A1701185199%2C%22timesVisited%22%3A2%7D; _ga_3WHFML52LK=GS1.1.1701199277.6.0.1701199277.0.0.0; cf_chl_2=a356405b4bb9849; cf_clearance=RSbjrON9b5TFAEe2DUjtnhDo4.aBw6HsrkhrnthvhdQ-1704471176-0-2-5032123a.dd5224ab.fc652da1-160.0.0")

		time.Sleep(2 * time.Second)
	})

	c.OnHTML(".attachment-large", func(h *colly.HTMLElement) {
		img = h.Attr("src")
	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		article := Article{}

		article.Id = x.ChildText("guid")
		article.Title = x.ChildText("title")
		article.Link = x.ChildText("link")
		article.Category = strings.Join(x.ChildTexts("category"), ", ")
		article.Published = x.ChildText("pubDate")

		r := readability.New()
		desc, err := r.Parse(strings.NewReader(x.ChildText("description")), article.Link)

		if err != nil {
			log.Fatal(err)
		}

		cont, err := r.Parse(strings.NewReader(x.ChildText("content:encoded")), article.Link)

		if err != nil {
			log.Fatal(err)
		}

		article.Content = cont.TextContent
		article.Description = desc.TextContent
		article.Author = x.ChildText("dc:creator")

		x.Request.Visit(article.Link)

		article.Image = img
		article.Description = strings.Replace(article.Description, "\n", "", -1)
		article.Description = strings.Replace(article.Description, "\t", "", -1)
		article.Description = strings.Replace(article.Description, "\"", "'", -1)

		article.Content = strings.Replace(article.Content, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://finbold.com/feed")
	return articles
}

func main() {
	if err := os.Mkdir("../../output/finbold.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/finbold.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/finbold.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/finbold.com/%s.md", article.Title)

		f, err := os.OpenFile(filepathJson, os.O_RDONLY, 0666)

		if err != nil && f == nil {
			createdJson, json_err := os.OpenFile(filepathJson, os.O_WRONLY|os.O_CREATE, 0666)
			createdCsv, csv_err := os.OpenFile(filepathCsv, os.O_WRONLY|os.O_CREATE, 0666)
			createdMd, md_err := os.OpenFile(filepathMd, os.O_WRONLY|os.O_CREATE, 0666)

			createdJson.Write(file)

			writeToCsv(*createdCsv, article)

			writeToMd(*createdMd, article)

			m := map[string]string{}

			utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

			if json_err != nil {
				log.Fatal(err)
			}

			if csv_err != nil {
				log.Fatal(err)
			}

			if md_err != nil {
				log.Fatal(err)
			}

			createdJson.Close()
			createdCsv.Close()
			createdMd.Close()
		} else {
			f.Close()
		}
	}
}
