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

	"github.com/flytam/filenamify"
	"github.com/gocolly/colly"
)

type Article struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Published string `json:"published"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	Image     string `json:"image"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "published")
	headers = append(headers, "author")
	headers = append(headers, "content")
	headers = append(headers, "image")

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
	image: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Author, article.Image, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getArticles(c *colly.Collector) []Article {
	articles := make([]Article, 0)
	links := make([]string, 0)

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second * 2)
	})

	c.OnHTML(".news-feed", func(h *colly.HTMLElement) {
		local := h.ChildAttrs("a", "href")

		for _, link := range local {
			if strings.Contains(link, "https://cryptoslate.com") {
				links = append(links, link)
			}
		}
	})

	c.OnHTML("article", func(h *colly.HTMLElement) {
		for _, link := range links {
			if link == h.ChildAttr("a", "href") {
				article := getArticle(c, link)

				article.Id = h.Attr("id")

				first := strings.Split(h.ChildAttr("img", "data-src"), "src=")

				if len(first) == 1 {
					continue
				}

				second := strings.Split(first[1], "&")

				article.Image = second[0]

				articles = append(articles, article)
			}
		}
	})

	c.Visit("https://cryptoslate.com/alpha-home/")

	return articles
}

func getArticle(c *colly.Collector, link string) Article {
	var content string = ""
	var title string = ""

	article := Article{Link: link}

	c.OnHTML(".post-title", func(h *colly.HTMLElement) {
		if title == "" {
			title = h.Text
			article.Title = h.Text
		}
	})

	c.OnHTML(".title", func(h *colly.HTMLElement) {
		if title == "" {
			title = h.ChildText("h1")
			article.Title = title
		}
	})

	c.OnHTML("#csjump-introduction", func(h *colly.HTMLElement) {
		if content == "" {
			content = h.Text
			article.Content = content
			article.Content = strings.Replace(content, "\n", "", -1)
			article.Content = strings.Replace(content, "\t", "", -1)
			article.Content = strings.Replace(content, "\"", "'", -1)
		}
	})

	c.OnHTML("article > p", func(h *colly.HTMLElement) {
		if content == "" {
			content = h.Text
			article.Content = content
			article.Content = strings.Replace(content, "\n", "", -1)
			article.Content = strings.Replace(content, "\t", "", -1)
			article.Content = strings.Replace(content, "\"", "'", -1)
		}
	})

	c.OnHTML(".cs-alpha-paywall-teaser", func(h *colly.HTMLElement) {
		h.ForEach("p", func(i int, h *colly.HTMLElement) {
			if i == 0 && content == "" {
				content = h.Text
				article.Content = content
				article.Content = strings.Replace(content, "\n", "", -1)
				article.Content = strings.Replace(content, "\t", "", -1)
				article.Content = strings.Replace(content, "\"", "'", -1)
			}
		})
	})

	c.OnHTML(".author-info", func(h *colly.HTMLElement) {
		article.Author = h.ChildText("a")

		time := h.ChildText(".time")

		article.Published = strings.Replace(h.Text, time, "", -1)
		article.Published = strings.Replace(article.Published, article.Author, "", -1)

	})

	c.OnHTML(".name", func(h *colly.HTMLElement) {
		article.Author = h.ChildText("a")
	})

	c.OnHTML(".post-date", func(h *colly.HTMLElement) {
		time := h.ChildText(".time")

		article.Published = strings.Replace(h.Text, time, "", -1)
		article.Published = strings.Replace(article.Published, "Updated:", "", -1)
	})

	c.OnHTML(".date", func(h *colly.HTMLElement) {
		article.Published = strings.Replace(h.Text, "Updated:", "", -1)
	})

	c.OnHTML(".cover > img", func(h *colly.HTMLElement) {
		print(h.Attr("src-set"))
	})

	c.Visit(link)

	article.Published = strings.Replace(article.Published, "•", "", -1)
	article.Published = strings.Replace(article.Published, "/", "", -1)

	return article
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func main() {
	if err := os.Mkdir("../../output/cryptoslate.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	articles := getArticles(c)

	for _, article := range articles {
		if article.Title == "" {
			continue
		}

		fmt.Printf("Downloaded: %s\n", article.Title)

		file, err := json.MarshalIndent(article, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		filename, _ := filenamify.FilenamifyV2(article.Title)

		filepathJson := fmt.Sprintf("../../output/cryptoslate.com/%s.json", filename)
		filepathCsv := fmt.Sprintf("../../output/cryptoslate.com/%s.csv", filename)
		filepathMd := fmt.Sprintf("../../output/cryptoslate.com/%s.md", filename)

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
