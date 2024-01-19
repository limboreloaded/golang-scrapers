package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"scripts/utils"
	"strconv"
	"strings"

	"github.com/cixtor/readability"
	"github.com/gocolly/colly"
)

type ArticleRes struct {
	Status bool      `json:"status"`
	Msg    string    `json:"msg"`
	Data   []Article `json:"data"`
}

type Article struct {
	Title       string `json:"title"`
	AltTitle    string `json:"alt_title"`
	Link        string `json:"url"`
	LikesCounts int    `json:"likes_counts"`
	Published   string `json:"time_elapse1"`
	Image       string `json:"src"`
	Author      string `json:"author"`
	Description string `json:"short_desc"`
	AuthorUrl   string `json:"author_url"`
	Content     string `json:"content"`
}

type JsonArticle struct {
	Title       string `json:"title"`
	AltTitle    string `json:"alt_title"`
	Link        string `json:"link"`
	LikesCounts int    `json:"likes_counts"`
	Published   string `json:"published"`
	Image       string `json:"image"`
	Author      string `json:"author"`
	Description string `json:"description"`
	AuthorUrl   string `json:"author_url"`
	Content     string `json:"content"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "alt_title")
	headers = append(headers, "url")
	headers = append(headers, "likes_counts")
	headers = append(headers, "published")
	headers = append(headers, "image")
	headers = append(headers, "author")
	headers = append(headers, "description")
	headers = append(headers, "author_url")
	headers = append(headers, "content")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Title)
	values = append(values, article.AltTitle)
	values = append(values, article.Link)
	values = append(values, string(rune(article.LikesCounts)))
	values = append(values, article.Published)
	values = append(values, article.Image)
	values = append(values, article.Author)
	values = append(values, article.Description)
	values = append(values, article.AuthorUrl)
	values = append(values, article.Content)

	return values
}

func writeToMd(mdFile os.File, article Article) {
	likes_counts := strconv.Itoa(article.LikesCounts)

	var input = fmt.Sprintf(`---
	title: "%s"
	alt_title: "%s"
	link: "%s"
	likes_counts: "%s"
	published: "%s"
	image: "%s"
	author: "%s"
	description: "%s"
	author_url: "%s"
	---
	%s`,
		article.Title, article.AltTitle, article.Link, likes_counts, article.Published, article.Image, article.Author, article.Description, article.AuthorUrl, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func getArticles(c *colly.Collector) []Article {
	returned := make([]Article, 0)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://www.coingabbar.com/en/getNews", nil)
	req.Header.Set("X-Csrf-Token", "DVWMgswR9n7p30oNvfrvfD7MRognjZB538qCWVVE")
	req.Header.Set("Cookie", "_ga=GA1.1.1817975449.1700588488; SELECTED_CURRENCY=INR; SELECTED_CURRENCY_SYMBOL=â‚¹; SELECTED_CURRENCY_PRICE=83.133071; AWSALB=fx56N993xN96vg2MuZHGJhqwr89kD8MPl8iwadjLiHP1sW+FeLDXpgeSGTaeje3Ggq+gD8Ok5JUxtuJhVRHs7kEVJy3fmySiPGQwgp8lTWHbO2pQw9va++AOe3w+; AWSALBCORS=fx56N993xN96vg2MuZHGJhqwr89kD8MPl8iwadjLiHP1sW+FeLDXpgeSGTaeje3Ggq+gD8Ok5JUxtuJhVRHs7kEVJy3fmySiPGQwgp8lTWHbO2pQw9va++AOe3w+; XSRF-TOKEN=eyJpdiI6Ikh1bVFiSU53Q2N0YTYrUWlQbTV2OEE9PSIsInZhbHVlIjoiWkpadFhCR2tudXAvbkp6OE5qQzVIQjFoMmpTT2hFZVVKVlAvazZnaEF2bkFyOVJWZ2Z5UDZQZHJyR0FweE1DZGdUUERMeW5OdEF1a05PbmJTVTNFSUg5OWVuQ3I5MGw2eHo2bkQvU1lJU0h4Z0hIWkhzcXpkc241RnU5Tm5mZ3IiLCJtYWMiOiI4OWY5MzYxOWI3MzQ1N2JhOTU3ZDA3NTk5Y2JjMmYyNzAzMjc3NTcyYzUyODcyNGMwN2VkYTFjNGM3YTdiYWY2IiwidGFnIjoiIn0%3D; coin_gabbar_session=eyJpdiI6ImZrbk1lZFRITW80RThVZnA0M2o3VlE9PSIsInZhbHVlIjoiVURZeTJWdUdvS1RaUzBScTFjNW9zNkd4QWNVTThSbmdkOXdDV1p1aHNNaUNHQ2NtU0YvTjEzTXdIQzJ5MGRVWHZiQnBGMThBR0xzUzBydC8xNkVQUDVyQVlLQ1JYb0xHNjJjT05tKzVCYTB2ekV5Zi9SRHhEaklabHlMY2ZFMkkiLCJtYWMiOiJiNTI0ZWEyMWI0Yjg3Y2FmNzc2ZjlhZjRmM2U1Njg4NWQ3Y2JjZWEyYzhlMjk5YTliMGM0ZDI1Nzk0NDI1NmIzIiwidGFnIjoiIn0%3D; _ga_46B68L0GCY=GS1.1.1704807204.15.1.1704807204.60.0.0")

	res, _ := client.Do(req)

	body, _ := ioutil.ReadAll(res.Body)
	var translated ArticleRes

	json.Unmarshal(body, &translated)

	for _, article := range translated.Data {
		c.OnHTML(".news", func(h *colly.HTMLElement) {
			r := readability.New()

			content, _ := r.Parse(strings.NewReader(h.Text), article.Link)
			article.Content = content.TextContent

			article.Content = strings.Replace(article.Content, "\n", "", -1)
			article.Content = strings.Replace(article.Content, "\t", "", -1)
			article.Content = strings.Replace(article.Content, "\"", "'", -1)

			article.Description = strings.Replace(article.Description, "\n", "", -1)
			article.Description = strings.Replace(article.Description, "\t", "", -1)
			article.Description = strings.Replace(article.Description, "\"", "'", -1)

			returned = append(returned, article)
		})

		c.Visit(article.Link)

		fmt.Printf("Downloaded: %s\n", article.Title)
	}

	return returned
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func main() {
	if err := os.Mkdir("../../output/coingabbar.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	articles := getArticles(c)

	for _, article := range articles {

		json_article := JsonArticle{Title: article.Title, AltTitle: article.AltTitle, Link: article.Link, LikesCounts: article.LikesCounts, Published: article.Published, Description: article.Description, Image: article.Image, Content: article.Content, AuthorUrl: article.AuthorUrl, Author: article.Author}

		file, err := json.MarshalIndent(json_article, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		filepathJson := fmt.Sprintf("../../output/coingabbar.com/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/coingabbar.com/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/coingabbar.com/%s.md", article.Title)

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
