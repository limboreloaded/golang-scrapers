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
	"github.com/flytam/filenamify"
	"github.com/gocolly/colly"
)

type Article struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Published   string `json:"published"`
	Image       string `json:"image"`
	Content     string `json:"content"`
	Author      string `json:"author"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "description")
	headers = append(headers, "published")
	headers = append(headers, "image")
	headers = append(headers, "content")
	headers = append(headers, "author")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Id)
	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Description)
	values = append(values, article.Published)
	values = append(values, article.Image)
	values = append(values, article.Content)
	values = append(values, article.Author)

	return values
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
func getRSSArticles() []Article {
	articles := make([]Article, 0)
	c := colly.NewCollector(colly.MaxDepth(0))

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("Proxy Address: %s\n", r.Request.ProxyURL)
		if r.StatusCode != 200 {
			log.Fatal("request failed with status")
		}
	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		article := Article{}
		article.Id = x.ChildText("guid")
		article.Title = x.ChildText("title")
		article.Link = x.ChildText("link")
		article.Description = x.ChildText("description")
		article.Published = x.ChildText("pubDate")
		article.Image = x.ChildAttr("enclosure", "url")
		article.Author = x.ChildText("dc:creator")

		r := readability.New()

		cont, _ := r.Parse(strings.NewReader(x.ChildText("content:encoded")), article.Link)

		article.Title = strings.Replace(article.Title, "🫨", "", -1)

		article.Content = strings.Replace(cont.TextContent, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		article.Description = strings.Replace(article.Description, "\n", "", -1)
		article.Description = strings.Replace(article.Description, "\t", "", -1)
		article.Description = strings.Replace(article.Description, "\"", "'", -1)

		fmt.Printf("Downloaded %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://metaversal.banklesshq.com/feed")

	return articles
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func main() {
	if err := os.Mkdir("../../output/metaversal.banklesshq.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	rssArticles := getRSSArticles()
	for _, article := range rssArticles {
		file, err := json.MarshalIndent(article, "", " ")

		filename, _ := filenamify.FilenamifyV2(article.Title)

		if err != nil {
			log.Fatal(err)
		}

		if err := os.Mkdir(fmt.Sprintf("../../output/metaversal.banklesshq.com/%s", filename), os.ModePerm); err != nil {
			if !strings.Contains(err.Error(), "file exists") {
				log.Fatal(err)
			}
		}

		filepathJson := fmt.Sprintf("../../output/metaversal.banklesshq.com/%s/article.json", filename)
		filepathCsv := fmt.Sprintf("../../output/metaversal.banklesshq.com/%s/article.csv", filename)
		filepathMd := fmt.Sprintf("../../output/metaversal.banklesshq.com/%s/article.md", filename)

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
