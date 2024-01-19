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

	"github.com/gocolly/colly"
)

type WikiPage struct {
	Title      string `json:"title"`
	Link       string `json:"link"`
	Categories string `json:"categories"`
	Tags       string `json:"tags"`
	IPFS       string `json:"ipfs"`
	TxHash     string `json:"tx_hash"`
	Published  string `json:"published"`
	Author     string `json:"author"`
	Content    string `json:"content"`
}

func (page WikiPage) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "categories")
	headers = append(headers, "tags")
	headers = append(headers, "ipfs")
	headers = append(headers, "tx_hash")
	headers = append(headers, "published")
	headers = append(headers, "author")
	headers = append(headers, "content")

	return headers
}

func (page WikiPage) GetValues() []string {
	values := make([]string, 0)

	values = append(values, page.Title)
	values = append(values, page.Link)
	values = append(values, page.Categories)
	values = append(values, page.Tags)
	values = append(values, page.IPFS)
	values = append(values, page.TxHash)
	values = append(values, page.Published)
	values = append(values, page.Author)
	values = append(values, page.Content)

	return values
}

func writeToCsv(csvFile os.File, page WikiPage) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(page.GetHeaders())
	csvWriter.Write(page.GetValues())
	csvWriter.Flush()
}

func writeToMd(mdFile os.File, page WikiPage) {
	var input = fmt.Sprintf(`---
	title: "%s"
	link: "%s"
	categories: "%s"
	tags: "%s"
	ipfs: "%s"
	tx_hash: "%s"
	published: "%s"
	author: "%s"
	---
	%s`,
		page.Title, page.Link, page.Categories, page.Tags, page.IPFS, page.TxHash, page.Published, page.Author, page.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func scrapePage(c *colly.Collector, href string) {
	time.Sleep(time.Second * 1)
	var date string = ""
	var months [12]string = [12]string{"January", "February", "March", "April", "May", "June", "July", "September", "October", "November", "December"}

	var page WikiPage

	page.Link = href

	c.OnHTML("meta", func(h *colly.HTMLElement) {
		if h.Attr("property") == "og:title" {
			page.Title = h.Attr("content")
		}
	})

	c.OnHTML(".css-0", func(h *colly.HTMLElement) {
		if strings.Contains(h.Text, "Categories") {
			page.Categories = h.ChildText("a")
		}

		if strings.Contains(h.ChildText("p"), "TX Hash") {
			page.TxHash = h.ChildAttr("a", "href")
		}
	})

	c.OnHTML(".css-qwu7cn", func(h *colly.HTMLElement) {
		page.Categories = h.ChildText("a")
	})

	c.OnHTML(".css-cq61k1", func(h *colly.HTMLElement) {
		page.Tags = h.ChildText("span")
	})

	c.OnHTML(".css-1mjbz9z", func(h *colly.HTMLElement) {
		page.IPFS = h.ChildAttr("a", "href")
	})

	c.OnHTML(".css-kc6ope", func(h *colly.HTMLElement) {
		link := h.ChildAttr("a", "href")
		page.TxHash = link
	})

	c.OnHTML("a", func(h *colly.HTMLElement) {
		href := h.Attr("href")

		if strings.Contains(href, "/account") {
			page.Author = "https://iq.wiki" + href
		}
	})

	c.OnHTML(".markdown_markdownBody__i1xqq", func(h *colly.HTMLElement) {
		pattern := regexp.MustCompile(`.css-(.*?)}`)

		page.Content = pattern.ReplaceAllString(h.ChildText("p"), "")
		page.Content = strings.Replace(page.Content, "\n", "", -1)
		page.Content = strings.Replace(page.Content, "\t", "", -1)
		page.Content = strings.Replace(page.Content, "\"", "'", -1)
	})

	c.OnHTML("td > p", func(h *colly.HTMLElement) {
		for _, month := range months {
			if strings.Contains(h.Text, month) && !strings.Contains(h.Text, "https://") {
				if date == "" && !strings.Contains(h.Text, "Created") && !strings.Contains(h.Text, "Created By") {
					date = strings.Replace(h.Text, ",", "", -1)
				}
			}
			page.Published = date
		}
	})

	c.Visit(page.Link)

	if page.Title == "" {
		return
	}

	if err := os.Mkdir(fmt.Sprintf("../../output/iqwiki.com/%s", page.Title), os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		} else {
			print("Found duplicate: " + page.Title + "\n")
			return

		}
	}

	savePage(page)

	print("Downloaded: " + page.Title + "\n")

}

func savePage(page WikiPage) {
	file, err := json.MarshalIndent(page, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	filepathJson := fmt.Sprintf("../../output/iqwiki.com/%s/article.json", page.Title)
	filepathCsv := fmt.Sprintf("../../output/iqwiki.com/%s/article.csv", page.Title)
	filepathMd := fmt.Sprintf("../../output/iqwiki.com/%s/article.md", page.Title)

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
	if err := os.Mkdir("../../output/iqwiki.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	c := colly.NewCollector()

	wikis := make([]string, 0)

	c.OnHTML("a", func(h *colly.HTMLElement) {
		if strings.Contains(h.Attr("href"), "https://") {
			return
		}

		if strings.Contains(h.Attr("href"), "/events") {
			return
		}

		if strings.Contains(h.Attr("href"), "/wiki") {
			wikis = append(wikis, "https://iq.wiki"+h.Attr("href"))
		}
	})

	c.Visit("https://iq.wiki/")

	for _, link := range wikis {
		scrapePage(c, link)
	}
}
