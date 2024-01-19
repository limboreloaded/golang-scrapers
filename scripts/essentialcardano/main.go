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

type Page struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Link    string `json:"link"`
	Image   string `json:"image"`
}

func (page Page) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "content")
	headers = append(headers, "image")

	return headers
}

func (page Page) GetValues() []string {
	values := make([]string, 0)

	values = append(values, page.Title)
	values = append(values, page.Link)
	values = append(values, page.Content)
	values = append(values, page.Image)

	return values
}

func writeToCsv(csvFile os.File, page Page) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(page.GetHeaders())
	csvWriter.Write(page.GetValues())
	csvWriter.Flush()
}

func writeToMd(mdFile os.File, page Page) {
	var input = fmt.Sprintf(`---
	title: "%s"
	link: "%s"
	image: "%s"
	---
	%s`,
		page.Title, page.Link, page.Image, page.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func scrapePage(c *colly.Collector, href string) {
	time.Sleep(time.Second * 1)
	var page Page

	page.Link = href

	c.OnHTML("h1", func(h *colly.HTMLElement) {
		if h.Attr("class") == "Heading-sc-1xljlq0-0 SCkwr" {
			page.Title = h.Text
		}
	})

	c.OnHTML("article", func(h *colly.HTMLElement) {
		r := readability.New()

		title := h.ChildText("h1")

		cont, _ := r.Parse(strings.NewReader(h.Text), page.Link)
		page.Content = strings.Replace(cont.TextContent, title, "", -1)
		page.Content = strings.Replace(page.Content, "\n", "", -1)
		page.Content = strings.Replace(page.Content, "\t", "", -1)
		page.Content = strings.Replace(page.Content, "\"", "'", -1)
	})

	c.OnHTML("span > img", func(h *colly.HTMLElement) {
		page.Image = h.Attr("srcset")
	})

	c.Visit(page.Link)

	fmt.Printf("Downloaded: %s\n", page.Title)

	savePage(page)
}

func savePage(page Page) {
	title, _ := filenamify.FilenamifyV2(page.Title)

	if err := os.Mkdir(fmt.Sprintf("../../output/essentialcardano.io/%s", title), os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		} else {
			print("Found duplicate: " + title + "\n")
			return
		}
	}

	file, err := json.MarshalIndent(page, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	filepathJson := fmt.Sprintf("../../output/essentialcardano.io/%s/article.json", title)
	filepathCsv := fmt.Sprintf("../../output/essentialcardano.io/%s/article.csv", title)
	filepathMd := fmt.Sprintf("../../output/essentialcardano.io/%s/article.md", title)

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
	if err := os.Mkdir("../../output/essentialcardano.io", os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	all := make([]string, 0)
	links := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	// c.OnRequest(func(r *colly.Request) {
	// 	time.Sleep(time.Second * 4)
	// })

	c.OnHTML("a", func(h *colly.HTMLElement) {
		if strings.Contains(h.Attr("href"), "article") {
			all = append(all, "https://essentialcardano.io"+h.Attr("href"))
		}
	})

	c.Visit("https://essentialcardano.io/")

	for _, link := range all {
		if len(links) == 0 {
			links = append(links, link)
		} else {
			var found bool = false

			for i := 0; i < len(links); i++ {
				if links[i] == link {
					found = true
				}
			}

			if !found {
				links = append(links, link)
			}
		}
	}

	for _, link := range links {
		// print(link + "\n")
		scrapePage(c, link)
	}
}
