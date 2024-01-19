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
	Id            string `json:"id"`
	Title         string `json:"title"`
	Link          string `json:"link"`
	Published     string `json:"published"`
	Description   string `json:"description"`
	Content       string `json:"content"`
	Author        string `json:"author"`
	SlashComments string `json:"slash_comments"`
	CommentRss    string `json:"comment_rss"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "guid")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "published")
	headers = append(headers, "description")
	headers = append(headers, "content")
	headers = append(headers, "author")
	headers = append(headers, "slash_comments")
	headers = append(headers, "comment_rss")

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
	values = append(values, article.SlashComments)
	values = append(values, article.CommentRss)

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
	slash_comments: "%s"
	comment_rss: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Description, article.Author, article.SlashComments, article.CommentRss, article.Content)

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
		article := Article{}

		article.Id = x.ChildText("guid")
		article.Title = x.ChildText("title")
		article.Link = x.ChildText("link")
		article.Published = x.ChildText("pubDate")
		article.Author = x.ChildText("dc:creator")
		article.SlashComments = x.ChildText("slash:comments")
		article.CommentRss = x.ChildText("wfw:commentRss")

		r := readability.New()

		desc, _ := r.Parse(strings.NewReader(x.ChildText("description")), article.Link)
		cont, _ := r.Parse(strings.NewReader(x.ChildText("content:encoded")), article.Link)

		article.Description = strings.Replace(desc.TextContent, "\n", "", -1)
		article.Description = strings.Replace(article.Description, "\t", "", -1)
		article.Description = strings.Replace(article.Description, "\"", "'", -1)

		article.Content = strings.Replace(cont.TextContent, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		x.Request.Visit(article.Link)

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://coinjournal.net/feed/")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/coinjournal.net", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/coinjournal.net/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/coinjournal.net/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/coinjournal.net/%s.md", article.Title)

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
