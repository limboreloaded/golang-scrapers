package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"scripts/utils"
	"strings"
	"time"

	"github.com/cixtor/readability"
	"github.com/gocolly/colly"
)

type SkipTillReader struct {
	rdr   *bufio.Reader
	delim []byte
	found bool
}

func NewSkipTillReader(reader io.Reader, delim []byte) *SkipTillReader {
	return &SkipTillReader{
		rdr:   bufio.NewReader(reader),
		delim: delim,
		found: false,
	}
}

func (str *SkipTillReader) Read(p []byte) (n int, err error) {
	if str.found {
		return str.rdr.Read(p)
	} else {
	outer:
		for {
			for i := range str.delim {
				var c byte
				c, err = str.rdr.ReadByte()
				if err != nil {
					n = 0
					return
				}
				if str.delim[i] != c {
					continue outer
				}
			}
			str.found = true
			str.rdr = bufio.NewReader(io.MultiReader(bytes.NewReader(str.delim), str.rdr))
			return str.Read(p)
		}
	}
}

type ReadTillReader struct {
	rdr   *bufio.Reader
	delim []byte
	found bool
}

func NewReadTillReader(reader io.Reader, delim []byte) *ReadTillReader {
	return &ReadTillReader{
		rdr:   bufio.NewReader(reader),
		delim: delim,
		found: false,
	}
}

func (rtr *ReadTillReader) Read(p []byte) (n int, err error) {
	if rtr.found {
		return 0, io.EOF
	} else {
	outer:
		for n < len(p) {
			for i := range rtr.delim {
				var c byte
				c, err = rtr.rdr.ReadByte()
				if err != nil && n > 0 {
					err = nil
					return
				} else if err != nil {
					return
				}
				p[n] = c
				n++
				if rtr.delim[i] != c {
					continue outer
				}
			}
			rtr.found = true
			break
		}
		if n == 0 {
			err = io.EOF
		}
		return
	}
}

type Article struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Link        string `json:"link"`
	Image       string `json:"image"`
	Published   string `json:"published"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

func (article Article) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "author")
	headers = append(headers, "link")
	headers = append(headers, "image")
	headers = append(headers, "published")
	headers = append(headers, "description")
	headers = append(headers, "content")

	return headers
}

func (article Article) GetValues() []string {
	values := make([]string, 0)

	values = append(values, article.Id)
	values = append(values, article.Title)
	values = append(values, article.Author)
	values = append(values, article.Link)
	values = append(values, article.Image)
	values = append(values, article.Published)
	values = append(values, article.Description)
	values = append(values, article.Content)

	return values
}

func extractStringFromLimit(content io.Reader, startString string, endString string) string {
	str := NewSkipTillReader(content, []byte(startString))
	rtr := NewReadTillReader(str, []byte(endString))
	res, err := ioutil.ReadAll(rtr)

	if err != nil {
		log.Fatal(err)
	}

	return string(res)
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

	c := colly.NewCollector(colly.MaxDepth(0))

	c.SetRequestTimeout(time.Minute * 3)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "_ga=GA1.1.1887371714.1700588643; _ga_F1PPFEMMP9=GS1.1.1700588652.1.1.1700588851.43.0.0; _ga_JZ46NG4KVG=deleted; Fbktio=%402tZYxi7%5BOQNBS; lPkcfoVHeTCtUQm=3yHpa9sU_EoGRu; VUJrfnGOHKZqDMa=d5YHVbw9kCSP; sgjuGraMD=FrODBLc; ppwp_wp_session=3fcb47476e418f1e2a97f5663dae26ea%7C%7C1701173611%7C%7C1701173251; cf_clearance=0x6Gh2JwrH0bGDYftYmdUEu23msPO6bPVwNTTFkemvA-1701171814-0-1-58d7ab6f.85b5a3f6.fd6f48f-160.2.1701171814; _ga_JZ46NG4KVG=GS1.1.1701171814.5.0.1701171818.0.0.0")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Fatal(err)
	})

	c.OnXML("rss/channel/item", func(x *colly.XMLElement) {
		article := Article{}

		article.Id = x.ChildText("guid")
		article.Title = x.ChildText("title")
		article.Link = x.ChildText("link")
		article.Published = x.ChildText("pubDate")
		article.Description = x.ChildText("description")
		article.Author = x.ChildText("dc:creator")

		r := readability.New()

		cont, _ := r.Parse(strings.NewReader(x.ChildText("content:encoded")), article.Link)
		article.Content = cont.TextContent

		reader := strings.NewReader(x.ChildText("content:encoded"))
		div := extractStringFromLimit(reader, "<div>", "</div>")

		reader = strings.NewReader(div)
		img := extractStringFromLimit(reader, "<img", ">")

		reader = strings.NewReader(img)
		src := extractStringFromLimit(reader, "https", "\"")

		article.Image = strings.TrimSuffix(src, "\"")

		article.Content = strings.Replace(article.Content, "\n", "", -1)
		article.Content = strings.Replace(article.Content, "\t", "", -1)
		article.Content = strings.Replace(article.Content, "\"", "'", -1)

		article.Description = strings.Replace(article.Description, "\n", "", -1)
		article.Description = strings.Replace(article.Description, "\t", "", -1)
		article.Description = strings.Replace(article.Description, "\"", "'", -1)

		fmt.Printf("Downloaded: %s\n", article.Title)

		articles = append(articles, article)
	})

	c.Visit("https://en.cryptonomist.ch/feed/")

	return articles
}

func main() {
	if err := os.Mkdir("../../output/en.cryptonomist.ch", os.ModePerm); err != nil {
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

		filepathJson := fmt.Sprintf("../../output/en.cryptonomist.ch/%s.json", article.Title)
		filepathCsv := fmt.Sprintf("../../output/en.cryptonomist.ch/%s.csv", article.Title)
		filepathMd := fmt.Sprintf("../../output/en.cryptonomist.ch/%s.md", article.Title)

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
