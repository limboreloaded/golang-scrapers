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
	Id        string `json:"id"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Published string `json:"published"`
	Author    string `json:"author"`
	Content   string `json:"content"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
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
	---
	%s`, article.Id, article.Title, article.Link, article.Published, article.Author, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getArticleLinks(c *colly.Collector) []string {
	articles := make([]string, 0)

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second * 2)
	})

	c.OnHTML(".sc-bztbEt dkJaxr", func(h *colly.HTMLElement) {
		articles = append(articles, fmt.Sprintf("https://benzinga.com%s", h.Attr("href")))
	})

	c.OnHTML(".news-block-headline", func(h *colly.HTMLElement) {
		articles = append(articles, fmt.Sprintf("https://benzinga.com%s", h.Attr("href")))
	})

	c.Visit("https://benzinga.com")

	return articles
}

func getArticle(c *colly.Collector, link string) Article {
	var title string = ""
	var author string = ""
	var date string = ""
	var content string = ""

	article := Article{Id: link, Link: link}

	c.OnHTML("#article-body", func(h *colly.HTMLElement) {
		if content == "" {
			content = h.ChildText(".core-block")
			content = strings.Replace(content, "\n", "", -1)
			content = strings.Replace(content, "\t", "", -1)
			content = strings.Replace(content, "\"", "'", -1)
			content = strings.Replace(content, "\\", "", -1)
			article.Content = content
		}
	})

	c.OnHTML(".date", func(h *colly.HTMLElement) {
		if date == "" {
			date = h.Text
			article.Published = h.Text
		}
	})

	c.OnHTML(".author-name", func(h *colly.HTMLElement) {
		if author == "" {
			author = h.Text
			article.Author = h.Text
		}
	})

	c.OnHTML(".layout-title", func(h *colly.HTMLElement) {
		if title == "" {
			title = h.Text
			article.Title = h.Text
		}
	})

	c.Visit(link)

	if article.Title == "" {
		return Article{}
	}

	fmt.Printf("Saving: %s\n", article.Title)

	return article
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func main() {
	if err := os.Mkdir("../../output/benzinga.com", os.ModePerm); err != nil {
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
			continue
		}

		filepathJson := fmt.Sprintf("../../output/benzinga.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/benzinga.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/benzinga.com/%s.md", article.Title)

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
				"title": "name",
			}

			utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

			createdJson.Close()
			createdCsv.Close()
			createdMd.Close()
		} else {
			f.Close()
		}
	}

	// for _, article := range rssArticles {
	// 	file, err := json.MarshalIndent(article, "", " ")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	filepathJson := fmt.Sprintf("../../output/web3isdoinggreat.com/%s.json", article.Title)
	// 	filepathCsv := fmt.Sprintf("../../output/web3isdoinggreat.com/%s.csv", article.Title)
	// 	filepathMd := fmt.Sprintf("../../output/web3isdoinggreat.com/%s.md", article.Title)

	// 	f, err := os.OpenFile(filepathJson, os.O_RDONLY, 0666)

	// 	if err != nil && f == nil {
	// 		createdJson, json_err := os.OpenFile(filepathJson, os.O_WRONLY|os.O_CREATE, 0666)
	// 		createdCsv, csv_err := os.OpenFile(filepathCsv, os.O_WRONLY|os.O_CREATE, 0666)
	// 		createdMd, md_err := os.OpenFile(filepathMd, os.O_WRONLY|os.O_CREATE, 0666)

	// 		createdJson.Write(file)

	// 		writeToCsv(*createdCsv, article)

	// 		createdMd.Write(file)

	// 		if json_err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		if csv_err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		if md_err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		 createdJson.Close()
	// 		 createdCsv.Close()
	// 		 createdMd.Close()
	// 	} else {
	// 		 f.Close()
	// 	}
	// }
}
