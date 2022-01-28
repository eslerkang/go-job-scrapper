package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id string
	location string
	title string
	salary string
	desc string
}

var baseURL string = "https://kr.indeed.com/jobs?q=python&limit=50"

func main() {
	totalPages := getPages()
	jobs := []extractedJob{}

	for i:=0; i<totalPages; i++ {
		jobs = append(jobs, getPage(i)...)
	}
	fmt.Println(jobs)
}

func getPage(page int) []extractedJob {
	jobs := []extractedJob{}

	pageURL := baseURL + "&start=" + strconv.Itoa(page*50) + "&limit=50"
	fmt.Println("Requesting: ", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)
	defer res.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")

	searchCards.Each(func(i int, card *goquery.Selection) {
		jobs = append(jobs, extractJob(card))
	})

	return jobs
}

func extractJob(card *goquery.Selection) extractedJob {
	id, _ := card.Attr("data-jk")
	title := cleanString(card.Find(".jobTitle>span").Text())
	location := cleanString(card.Find(".companyLocation").Text())
	salary := cleanString(card.Find(".salary-snippet>span").Text())
	desc := cleanString(card.Find(".job-snippet").Text())
	return extractedJob{
		id: id,
		title: title,
		location: location,
		salary: salary,
		desc: desc,
	}
}

func getPages() int {
	pages := 0
	
	res, err := http.Get(baseURL)
	
	checkErr(err)
	checkCode(res)
	
	defer res.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatal("Request Failed")
	}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}