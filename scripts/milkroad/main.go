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
	Id        string `json:"id"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Published string `json:"published"`
	Image     string `json:"image"`
	Content   string `json:"content"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "published")
	headers = append(headers, "image")
	headers = append(headers, "content")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Id)
	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Published)
	values = append(values, article.Image)
	values = append(values, article.Content)

	return values
}

func writeToMd(mdFile os.File, article Article) {
	var input = fmt.Sprintf(`---
	id: "%s"
	title: "%s"
	link: "%s"
	published: "%s"
	image: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Image, article.Content)

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

	c.OnHTML(".thumb", func(h *colly.HTMLElement) {
		articles = append(articles, h.ChildAttr("a", "href"))
	})

	c.Visit("https://milkroad.com/daily/")

	return articles
}

func getArticle(c *colly.Collector, link string) Article {
	article := Article{Link: link}

	c.OnHTML("article", func(h *colly.HTMLElement) {
		article.Id = h.Attr("id")
	})

	c.OnHTML(".beehiiv__body", func(h *colly.HTMLElement) {
		article.Image = h.ChildAttr("img", "data-src")

		r := readability.New()

		cont, _ := r.Parse(strings.NewReader(h.Text), link)
		article.Content = cont.TextContent
		article.Content = strings.Replace(article.Content, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)
	})

	c.OnHTML(".single-date", func(h *colly.HTMLElement) {
		article.Published = h.Text
	})

	c.OnHTML(".entry-title", func(h *colly.HTMLElement) {
		article.Title = h.Text
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
	if err := os.Mkdir("../../output/milkroad.com", os.ModePerm); err != nil {
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

		if err := os.Mkdir(fmt.Sprintf("../../output/milkroad.com/%s", article.Title), os.ModePerm); err != nil {
			if !strings.Contains(err.Error(), "file exists") {
				log.Fatal(err)
			}
		}

		filepathJson := fmt.Sprintf("../../output/milkroad.com/%s/article.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/milkroad.com/%s/article.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/milkroad.com/%s/article.md", article.Title)

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
