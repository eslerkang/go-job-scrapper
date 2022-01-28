package scrapper

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	ccsv "github.com/tsak/concurrent-csv-writer"
)

type extractedJob struct {
	id string
	location string
	title string
	salary string
	desc string
}

const jobViewURL string = "https://kr.indeed.com/viewjob?jk="

func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/jobs?q="+term+"&limit=50"
	totalPages := getPages(baseURL)
	jobs := []extractedJob{}

	c := make(chan []extractedJob)

	for i:=0; i<totalPages; i++ {
		go getPage(baseURL, i, c)
	}
	for i:=0; i<totalPages; i++ {
		extractedJobs := <- c
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)

	fmt.Println("Done,", len(jobs))
}

func writeJobs(jobs []extractedJob) {
	file, err := ccsv.NewCsvWriter("jobs.csv")
	checkErr(err)

	c := make(chan bool)

	defer file.Close()

	headers := []string{"Link", "Title", "Location", "Salary", "Description"}

	wErr := file.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		go writeJobLine(file, job, c)
	}

	for i:=0; i<len(jobs); i++ {
		<- c
	}
}

func writeJobLine(file *ccsv.CsvWriter, job extractedJob, c chan bool) {
	jobSlice := []string{jobViewURL+job.id, job.title, job.location, job.salary, job.desc}
	wErr := file.Write(jobSlice)
	checkErr(wErr)
	c <- true
}

func getPage(url string, page int, mainC chan<- []extractedJob) {
	jobs := []extractedJob{}

	c := make(chan extractedJob)

	pageURL := url + "&start=" + strconv.Itoa(page*50) + "&limit=50"
	fmt.Println("Requesting: ", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)
	defer res.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})
	for i:=0; i<searchCards.Length(); i++ {
		jobs = append(jobs, <- c)
	}	
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".jobTitle>span").Text())
	location := CleanString(card.Find(".companyLocation").Text())
	salary := CleanString(card.Find(".salary-snippet>span").Text())
	desc := CleanString(card.Find(".job-snippet").Text())
	c <- extractedJob{
		id: id,
		title: title,
		location: location,
		salary: salary,
		desc: desc,
	}
}

func getPages(url string) int {
	pages := 0
	
	res, err := http.Get(url)
	
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

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}