package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"scripts/utils"
	"strings"
	"time"

	"github.com/cixtor/readability"
	"github.com/gocolly/colly"
	"golang.org/x/net/html"
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

	headers = append(headers, "id")
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

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func getRssArticles() []Article {
	articles := make([]Article, 0)

	paragraphs := make([]string, 0)

	c := colly.NewCollector(colly.MaxDepth(0))

	c.SetRequestTimeout(time.Hour)

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(2 * time.Second)
	})

	c.OnHTML(".ce-single-post-content-block", func(h *colly.HTMLElement) {
		for _, node := range h.DOM.Children().Nodes {
			if node.Data == "script" || node.Data == "figure" || node.Data == "div" || node.Data == "ins" {
				continue
			}
			for n := node.FirstChild; n != nil; n = n.NextSibling {
				if n.Type == html.NodeType(1) {
					paragraphs = append(paragraphs, n.Data)
				}
			}
		}
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

		article.Description = desc.TextContent
		article.Author = x.ChildText("dc:creator")

		reg_desc, _ := regexp.Compile("<img(.*)/>")
		reg_src, _ := regexp.Compile(`src="(.*).(jpg|png)"`)

		img := reg_desc.FindString(desc.Content)
		src := reg_src.FindString(img)

		src = strings.TrimLeft(src, "src=\"")
		src = strings.TrimRight(src, "\"")

		x.Request.Visit(article.Link)

		article.Image = src
		article.Content = strings.Join(paragraphs, "")

		article.Content = strings.Replace(article.Content, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		article.Description = strings.Replace(article.Description, "\n", "", -1)
		article.Description = strings.Replace(article.Description, "\t", "", -1)
		article.Description = strings.Replace(article.Description, "\"", "'", -1)

		paragraphs = paragraphs[:0]

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://coinedition.com/feed/")
	return articles
}

func main() {
	if err := os.Mkdir("../../output/coinedition.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/coinedition.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/coinedition.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/coinedition.com/%s.md", article.Title)

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
				"image": "thumbnail",
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
