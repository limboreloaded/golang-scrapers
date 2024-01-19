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
	headers = append(headers, "description")
	headers = append(headers, "content")
	headers = append(headers, "author")
	headers = append(headers, "category")
	headers = append(headers, "image")

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
	values = append(values, article.Category)
	values = append(values, article.Image)

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

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func getRssArticles() []Article {
	var content string

	articles := make([]Article, 0)

	c := colly.NewCollector()

	c.SetRequestTimeout(time.Minute)

	url, _ := url.Parse("socks5://138.68.16.30:13469")

	c.SetProxyFunc(http.ProxyURL(url))

	c.OnError(func(r *colly.Response, err error) {
		print(r.Request.ProxyURL)
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		print(r.ProxyURL)

		r.Headers.Set("Host", "cointelegraph.com")
		r.Headers.Set("Cookie", "__cf_bm=VrXQZRTTnKvQD6QEojGU35.hK9zfZGxk3nfG6YR.TBM-1704467128-1-AciVfyYhPGK+RNU9AT8aRixFFWx/iVebcbWMZVowl8Zq/HING3gcjyZ+kRUUk8WebHTLzWmbfesFIvi1jvLG4Ow=")
		r.Headers.Set("Connection", "Close")
		r.Headers.Set("User-Agent", "Paw/3.2.4 (Macintosh; OS X/12.6.3) GCDHTTPRequest")
	})

	c.OnHTML(".post-content", func(h *colly.HTMLElement) {
		content = h.Text
		content = strings.Replace(content, "\n", "", -1)
		content = strings.Replace(content, "\t", "", -1)
		content = strings.Replace(content, "\"", "'", -1)
	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		guid := x.ChildText("guid")
		title := x.ChildText("title")
		link := x.ChildText("link")
		published := x.ChildText("pubDate")
		author := x.ChildText("dc:creator")

		r := readability.New()
		desc := x.ChildText("description")

		description, _ := r.Parse(strings.NewReader(desc), link)

		desc = strings.Replace(description.TextContent, "\n", "", -1)
		desc = strings.Replace(desc, "\t", "", -1)
		desc = strings.Replace(desc, "\"", "'", -1)

		category := strings.Join(x.ChildTexts("category"), ", ")
		media_content := x.ChildAttr("media:content", "url")

		c.Visit(link)

		article := Article{Id: guid, Title: title, Link: link, Published: published, Author: author, Description: desc, Image: media_content, Category: category, Content: content}

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://cointelegraph.com/rss")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/cointelegraph.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/cointelegraph.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/cointelegraph.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/cointelegraph.com/%s.md", article.Title)

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

			writeToMd(*createdMd, article)

			writeToCsv(*createdCsv, article)

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
