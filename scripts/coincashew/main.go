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

var visited []string = make([]string, 0)

type PageData struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Published   string `json:"published"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

func isHrefRelevant(link string) bool {
	return !strings.Contains(link, "https://") && !strings.Contains(link, "#")
}

func (page PageData) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "published")
	headers = append(headers, "description")
	headers = append(headers, "content")

	return headers
}

func (page PageData) GetValues() []string {
	values := make([]string, 0)

	values = append(values, page.Title)
	values = append(values, page.Link)
	values = append(values, page.Published)
	values = append(values, page.Description)
	values = append(values, page.Content)

	return values
}

func writeToCsv(csvFile os.File, page PageData) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(page.GetHeaders())
	csvWriter.Write(page.GetValues())
	csvWriter.Flush()
}

func writeToMd(mdFile os.File, page PageData) {
	var input = fmt.Sprintf(`---
	title: "%s"
	link: "%s"
	published: "%s"
	description: "%s"
	---
	%s`,
		page.Title, page.Link, page.Published, page.Description, page.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func scrapePage(c *colly.Collector, href string) {
	time.Sleep(time.Second * 1)
	var page PageData
	sublinks := make([]string, 0)

	href = strings.Replace(href, "https://www.coincashew.com", "", -1)

	if strings.Contains(href, "http://") {
		return
	}

	if !isHrefRelevant(href) {
		return
	}

	page.Link = "https://www.coincashew.com" + href

	for i := 0; i < len(visited); i++ {
		if visited[i] == page.Link {
			return
		}
	}

	c.OnHTML("a", func(h *colly.HTMLElement) {
		var href string = h.Attr("href")

		href = strings.Replace(href, "https://www.coincashew.com", "", -1)

		if strings.Contains(href, "http://") {
			return
		}

		if !isHrefRelevant(href) {
			return
		}

		link := "https://www.coincashew.com" + href

		if link != page.Link {
			sublinks = append(sublinks, href)
		}
	})

	c.OnHTML("title", func(h *colly.HTMLElement) {
		page.Title, _ = filenamify.FilenamifyV2(h.Text)
	})

	c.OnHTML(".r-1pa6394", func(h *colly.HTMLElement) {
		page.Published = h.ChildAttr("span", "aria-label")
	})

	c.OnHTML(".css-1qaijid", func(h *colly.HTMLElement) {
		if strings.Contains(h.Attr("aria-label"), "UTC") {
			page.Published = h.Attr("aria-label")
		}
	})

	c.OnHTML(".contentParagraph_-qmPj", func(h *colly.HTMLElement) {
		page.Description = h.Text
	})

	c.OnHTML("main", func(h *colly.HTMLElement) {
		r := readability.New()

		cont, _ := r.Parse(strings.NewReader(h.Text), page.Link)

		page.Content = strings.Replace(cont.TextContent, "Comment on page", "", -1)
		page.Content = strings.Replace(page.Content, "\n", "", -1)
		page.Content = strings.Replace(page.Content, "\t", "", -1)
		page.Content = strings.Replace(page.Content, "\"", "'", -1)
	})

	visited = append(visited, page.Link)

	c.Visit(page.Link)

	fmt.Printf("Downloaded: %s\n", page.Title)

	if page.Title == "" {
		return
	}

	if err := os.Mkdir(fmt.Sprintf("../../output/coincashew.com/%s", page.Title), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		} else {
			return
		}
	}

	savePage(page)

	if len(sublinks) > 0 {
		for _, href := range sublinks {
			scrapePage(c, href)
		}
	}
}

func getMainPageLinks(c *colly.Collector) []string {
	links := make([]string, 0)

	c.OnHTML("a", func(h *colly.HTMLElement) {
		if isHrefRelevant(h.Attr("href")) {
			links = append(links, h.Attr("href"))
		}
	})

	c.Visit("https://www.coincashew.com/")

	return links
}

func savePage(page PageData) {
	file, err := json.MarshalIndent(page, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	filepathJson := fmt.Sprintf("../../output/coincashew.com/%s/article.json", page.Title)
	filepathCsv := fmt.Sprintf("../../output/coincashew.com/%s/article.csv", page.Title)
	filepathMd := fmt.Sprintf("../../output/coincashew.com/%s/article.md", page.Title)

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

		writeToCsv(*createdCsv, page)

		writeToMd(*createdMd, page)

		m := map[string]string{}

		utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

		createdJson.Close()
		createdCsv.Close()
		createdMd.Close()
	} else {
		f.Close()
	}
}

func main() {
	if err := os.Mkdir("../../output/coincashew.com", os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		if !strings.Contains(err.Error(), "Not Found") {
			log.Fatal(err)
		}
	})

	links := getMainPageLinks(c)

	for _, link := range links {
		scrapePage(c, link)
	}
}
