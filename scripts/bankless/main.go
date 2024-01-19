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
	Title   string `json:"title"`
	Link    string `json:"link"`
	Content string `json:"content"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "content")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Title)
	values = append(values, article.Link)
	values = append(values, article.Content)

	return values
}

func saveArticles() {
	c := colly.NewCollector(colly.MaxDepth(0))

	c.SetRequestTimeout(time.Minute * 3)

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "cf_chl_2=6975433d1525c15; cf_clearance=flZ4O9ZrCP_XcSXYLTpDH1Pe7w_O5PrL2Ny_.mJRziM-1701780867-0-1-37c16f54.773f9b89.56089283-160.0.0; PHPSESSID=2l7cldbq50rjve5ipkt4ppkvoi; isDark=1; _gcl_au=1.1.1783898777.1701780872; _ga=GA1.1.543392869.1701780872; signInPrompt=shown; _ga_DLG42PG8PY=GS1.1.1701780871.1.1.1701781830.60.0.0; userID2=4")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	})

	c.OnHTML("#filterContent", func(h *colly.HTMLElement) {
		for _, link := range h.ChildAttrs("a", "href") {
			article := Article{}

			article.Link = link

			c.OnHTML("title", func(h *colly.HTMLElement) {
				article.Title = h.Text
			})

			c.OnHTML(".contents", func(h *colly.HTMLElement) {
				article.Content = strings.Replace(h.Text, "\n", "", -1)
			})

			c.Visit(link)

			fmt.Printf("Downloaded: %s\n", article.Title)

			savePage(article)
		}
	})

	c.Visit("https://www.bankless.com/read/briefs")
}

func writeToCsv(csvFile os.File, article Article) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(article.GetHeaders())
	csvWriter.Write(article.GetValues())
	csvWriter.Flush()
}

func writeToMd(mdFile os.File, article Article) {
	var input = fmt.Sprintf(`---
	title: "%s"
	link: "%s"
	---
	%s`, article.Title, article.Link, article.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func savePage(article Article) {
	file, err := json.MarshalIndent(article, "", " ")

	if err != nil {
		log.Fatal(err)
	}

	if err := os.Mkdir(fmt.Sprintf("../../output/bankless.com/%s", article.Title), os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	filepathJson := fmt.Sprintf("../../output/bankless.com/%s/article.json", article.Title)
	filepathCsv := fmt.Sprintf("../../output/bankless.com/%s/article.csv", article.Title)
	filepathMd := fmt.Sprintf("../../output/bankless.com/%s/article.md", article.Title)

	f, err := os.OpenFile(filepathJson, os.O_RDONLY, 0666)

	if err != nil && f == nil {
		createdJson, json_err := os.OpenFile(filepathJson, os.O_WRONLY|os.O_CREATE, 0666)
		createdCsv, csv_err := os.OpenFile(filepathCsv, os.O_WRONLY|os.O_CREATE, 0666)
		createdMd, md_err := os.OpenFile(filepathMd, os.O_WRONLY|os.O_CREATE, 0666)

		if json_err != nil {
			log.Fatal(err)
		}

		createdJson.Write(file)

		if csv_err != nil {
			defer createdJson.Close()

		} else {
			writeToCsv(*createdCsv, article)
		}

		if md_err != nil {
			log.Fatal(err)
		}

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

func main() {
	if err := os.Mkdir("../../output/bankless.com", os.ModePerm); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Fatal(err)
		}
	}

	saveArticles()
}
