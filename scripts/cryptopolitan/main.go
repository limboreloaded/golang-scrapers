package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"scripts/utils"
	"strings"

	// "github.com/cixtor/readability"

	"github.com/cixtor/readability"
	"github.com/gocolly/colly"
)

type Article struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Published   string `json:"published"`
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
	values = append(values, article.Image)

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
	image: "%s"
	---
	%s`,
		article.Id, article.Title, article.Link, article.Published, article.Description, article.Author, article.Image, article.Content)

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
		// r.Headers.Set("authority", "www.cryptopolitan.com")
		// r.Headers.Set("path", "/feed/")
		// r.Headers.Set("method", "GET")
		// r.Headers.Set("scheme", "https")
		// r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		// r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		// r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		// r.Headers.Set("Cache-Control", "max-age=0")
		// r.Headers.Set("Cookie", `cf_clearance=gqSUPJDGrwOD0D_tQA.8cR9GW9V_6L53DXIUY07aqDM-1700595957-0-1-58d7ab6f.c65b6469.e76c6d63-0.2.1700595957; holler-content-views={"374682":1,"374683":1}; _pk_id.27389.6bde=9dd498cde3fc04d0.1700595957.1.1700595957.1700595957.; _ga=GA1.1.566194511.1700595958; _ga_NHCRVZ3G4K=GS1.1.1700595958.1.0.1700596007.11.0.0; _ga_MLN81B438R=GS1.1.1700595958.1.0.1700596007.0.0.0`)
		// r.Headers.Set("Dnt", "1")
		// r.Headers.Set("If-Modified-Since", "Fri, 01 Dec 2023 14:38:29 GMT")
		// r.Headers.Set("If-None-Match", `W/"dbfc97f035f8fbac7ece63a9ab6c05d6"`)
		// r.Headers.Set("Sec-Ch-Ua", `"Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"`)
		// r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		// r.Headers.Set("Set-Ch-Ua-Platform", `"macOS"`)
		// r.Headers.Set("Sec-Fetch-Dest", "document")
		// r.Headers.Set("Sec-Fetch-Mode", "navigate")
		// r.Headers.Set("Sec-Fetch-Site", "none")
		// r.Headers.Set("Sec-Fetch-User", "?1")
		// r.Headers.Set("Sec-Gpc", "1")
		// r.Headers.Set("Upgrade-Insecure-Requests", "1")
		// r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		article := Article{}

		article.Id = x.ChildText("guid")
		article.Title = x.ChildText("title")
		article.Link = x.ChildText("link")
		article.Published = x.ChildText("pubDate")
		article.Author = x.ChildText("dc:creator")
		article.Image = x.ChildAttr("media:content", "url")

		r := readability.New()

		desc, _ := r.Parse(strings.NewReader(x.ChildText("description")), article.Link)
		cont, _ := r.Parse(strings.NewReader(x.ChildText("content:encoded")), article.Link)

		article.Description = desc.TextContent
		article.Content = cont.TextContent

		article.Content = strings.Replace(article.Content, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		article.Description = strings.Replace(article.Description, "\n", "", -1)
		article.Description = strings.Replace(article.Description, "\t", "", -1)
		article.Description = strings.Replace(article.Description, "\"", "'", -1)

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://www.cryptopolitan.com/feed/")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/cryptopolitan.com", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/cryptopolitan.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/cryptopolitan.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/cryptopolitan.com/%s.md", article.Title)

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

			m := map[string]string{}

			utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

			createdJson.Write(file)

			writeToCsv(*createdCsv, article)

			writeToMd(*createdMd, article)

			createdJson.Close()
			createdCsv.Close()
			createdMd.Close()
		} else {
			f.Close()
		}
	}
}
