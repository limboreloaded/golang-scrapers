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
	"strings"
	"time"

	"github.com/cixtor/readability"
	"github.com/flytam/filenamify"
)

var links [10]string = [10]string{"https://www.pointer.gg/tutorials/Setting-Up-Your-Wallet/6e23e8e8-6760-45fa-8fb3-330400ac03ac", "https://www.pointer.gg/tutorials/solid-solidity/a7ec3eff-fc59-481d-bcb2-1224b3e9c0f7", "https://www.pointer.gg/tutorials/thirdweb-nft-lootbox/feda78d2-d35c-4b77-a1d3-182aa16070d9", "https://www.pointer.gg/tutorials/create-a-web3-forum-with-polygon/1cb8f005-08f4-48a2-9d82-cd963e16f7f1", "https://www.pointer.gg/tutorials/solana-pay-irl-payments/944eba7e-82c6-4527-b55c-5411cdf63b23", "https://www.pointer.gg/tutorials/build-a-dex-with-stacks/56abb3a4-05c1-4608-b096-f82189e9f759", "https://www.pointer.gg/tutorials/solana-anchor/6f9afc45-309d-4f3e-abfc-aa40dd09fd0a", "https://www.pointer.gg/tutorials/polygon-amm/c725cd07-447b-46b3-b544-81516b172c5f", "https://www.pointer.gg/tutorials/polygon-amm2/65775edd-eb14-4f44-8b0b-38778c51fc5d", "https://www.pointer.gg/tutorials/solana-nft-collection/0c289046-8e7a-42fe-9d38-9f7ef52c0d7d"}

type InitData struct {
	Id                   string     `json:"id"`
	TutorialPublications []PageData `json:"tutorial_publications"`
	Slug                 string     `json:"slug"`
}

type PageData struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Link       string `json:"link"`
	Content    string `json:"content"`
	Published  string `json:"created_at"`
	TutorialId string `json:"tutorial_id"`
}

func (page PageData) GetHeaders() []string {
	headers := make([]string, 0)

	headers = append(headers, "id")
	headers = append(headers, "title")
	headers = append(headers, "link")
	headers = append(headers, "content")
	headers = append(headers, "published")
	headers = append(headers, "tutorial_id")

	return headers
}

func (page PageData) GetValues() []string {
	values := make([]string, 0)

	values = append(values, page.Id)
	values = append(values, page.Title)
	values = append(values, page.Link)
	values = append(values, page.Content)
	values = append(values, page.Published)
	values = append(values, page.TutorialId)

	return values
}

func writeToMd(mdFile os.File, page PageData) {
	var input = fmt.Sprintf(`---
	id: "%s"
	title: "%s"
	link: "%s"
	published: "%s"
	tutorial_id: "%s"
	---
	%s`,
		page.Id, page.Title, page.Link, page.Published, page.TutorialId, page.Content)

	input = strings.Replace(input, "\t", "", -1)

	mdFile.Write([]byte(strings.Replace(input, "\"", "", -1)))
}

func writeToCsv(csvFile os.File, page PageData) {
	csvWriter := csv.NewWriter(&csvFile)
	csvWriter.Write(page.GetHeaders())
	csvWriter.Write(page.GetValues())
	csvWriter.Flush()
}

func savePage(page PageData) {
	title, _ := filenamify.FilenamifyV2(page.Title + "-" + page.TutorialId)

	if err := os.Mkdir(fmt.Sprintf("../../output/pointer.gg/%s", title), os.ModePerm); err != nil {
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

	filepathJson := fmt.Sprintf("../../output/pointer.gg/%s/article.json", title)
	filepathCsv := fmt.Sprintf("../../output/pointer.gg/%s/article.csv", title)
	filepathMd := fmt.Sprintf("../../output/pointer.gg/%s/article.md", title)

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

		m := map[string]string{
			"category":  "categories",
			"published": "created_at",
		}

		utils.ChangeKeysByPath(m, createdJson, createdMd, createdCsv)

		createdJson.Close()
		createdCsv.Close()
		createdMd.Close()
	} else {
		f.Close()
	}
}

func scrapePage(link string) {
	time.Sleep(time.Second * 1)
	var page PageData

	page.Link = link

	split := strings.Split(link, "/")

	slug := split[4]

	client := &http.Client{}

	url := "https://gcyrkvfwpvlwjrrklxnh.supabase.co/rest/v1/tutorials?select=*%2Ctutorial_drafts%21tutorial_drafts_tutorial_id_fkey%28*%29%2Ctutorial_publications%28*%29&slug=eq." + slug

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiIsImlhdCI6MTY0MDY0NzQ1MywiZXhwIjoxOTU2MjIzNDUzfQ.aMlksrJ6pR0aqG-xb9hl_usDQw83Wj1pSBWoLefMEp0")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYW5vbiIsImlhdCI6MTY0MDY0NzQ1MywiZXhwIjoxOTU2MjIzNDUzfQ.aMlksrJ6pR0aqG-xb9hl_usDQw83Wj1pSBWoLefMEp0")

	res, _ := client.Do(req)

	res_body, _ := ioutil.ReadAll(res.Body)

	var init []InitData

	json.Unmarshal(res_body, &init)

	for _, publication := range init[0].TutorialPublications {
		r := readability.New()

		cont, _ := r.Parse(strings.NewReader(publication.Content), fmt.Sprintf("https://www.pointer.gg/tutorials/%s/%s", slug, publication.Id))

		publication.Link = fmt.Sprintf("https://www.pointer.gg/tutorials/%s/%s", slug, publication.Id)
		publication.Content = cont.TextContent
		publication.Content = strings.Replace(publication.Content, "\n", "", -1)
		publication.Content = strings.Replace(publication.Content, "\t", "", -1)
		publication.Content = strings.Replace(publication.Content, "\"", "'", -1)

		fmt.Printf("Downloaded: %s\n", publication.Title)

		savePage(publication)
	}
}

func main() {
	if err := os.Mkdir("../../output/pointer.gg", os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	for _, link := range links {
		scrapePage(link)
	}
}
