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
	Id        string `json:"id"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Image     string `json:"image"`
	Published string `json:"published"`
	Author    string `json:"author"`
	Content   string `json:"content"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "image")
	headers = append(headers, "published")
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
	values = append(values, article.Author)
	values = append(values, article.Content)

	return values
}

func writeToMd(mdFile os.File, article Article) {
	var input = fmt.Sprintf(`---
	id: "%s"
	title: "%s"
	link: "%s"
	published: "%s"
	author: "%s"
	image: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Author, article.Image, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getArticleLinks(c *colly.Collector) []string {
	articles := make([]string, 0)

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		if strings.Contains(err.Error(), "Not Acceptable") {
			print("\n")
		}
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second * 2)
	})

	c.OnHTML("article", func(h *colly.HTMLElement) {
		link := h.ChildAttr("a", "href")

		if strings.Contains(link, "https://research.despread.io/") {
			articles = append(articles, h.ChildAttr("a", "href"))
		}
	})

	c.Visit("https://research.despread.io/")

	return articles
}

func getArticle(c *colly.Collector, link string) Article {
	article := Article{Id: link, Link: link}

	c.OnHTML(".post-hero__figure", func(h *colly.HTMLElement) {
		foundimg := strings.Split(h.ChildAttr("img", "data-srcset"), ",")
		article.Image = strings.Replace(fmt.Sprintf("https://research.despread.io%s", foundimg[len(foundimg)-2]), " 1000w", "", 1)
	})

	c.OnHTML(".post-hero__title", func(h *colly.HTMLElement) {
		title, _ := filenamify.FilenamifyV2(h.Text)
		article.Title = title
	})

	c.OnHTML(".post-info__author", func(h *colly.HTMLElement) {
		article.Author = h.Text
	})

	c.OnHTML(".post-info__dr", func(h *colly.HTMLElement) {
		article.Published = h.ChildAttr("time", "datetime")
	})

	c.OnHTML("article", func(h *colly.HTMLElement) {
		r := readability.New()

		content, _ := r.Parse(strings.NewReader(h.Text), link)
		article.Content = content.TextContent
		article.Content = strings.Replace(article.Content, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)
	})

	c.Visit(link)

	fmt.Printf("Downloaded: %s\n", article.Title)

	return article
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func main() {
	if err := os.Mkdir("../../output/research.despread.io", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	links := getArticleLinks(c)

	for _, link := range links {
		article := getArticle(c, link)

		file, err := json.MarshalIndent(article, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		if article.Title == "" {
			return
		}

		filepathJson := fmt.Sprintf("../../output/research.despread.io/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/research.despread.io/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/research.despread.io/%s.md", article.Title)

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
