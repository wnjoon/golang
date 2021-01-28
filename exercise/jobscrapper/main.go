package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var baseURL string = "https://kr.indeed.com/jobs?q=python"
var jobPerPage int = 10

type jobInfo struct {
	id       string
	location string
	title    string
	salary   string
	summary  string
}

func main() {

	defer elapsedTime("Scrapper Test", "start")()

	var jobs []jobInfo

	c := make(chan []jobInfo)

	totalPages := getPageNumber()

	for i := 0; i < totalPages; i++ {
		// combine all slices into one slice
		go getPageInfo(i, c)
		// extractJobs...: if ... -> [x]+[x] => [xx] / else [x] + [x] -> [[x][x]] (... = content appending)
	}

	for i := 0; i < totalPages; i++ {
		jobsPerPage := <-c
		jobs = append(jobs, jobsPerPage...)
	}

	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

/*
 * getPageNumber
 * - Get number of pages about searched job
 */
func getPageNumber() int {

	pageNumber := 0

	res, err := http.Get(baseURL)
	checkError(err)
	checkResponseCode(res)

	defer res.Body.Close()

	// goquery: give response body to goquery
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	// get every item searched
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pageNumber = s.Find("a").Length() // Show me how many links in pagination
	})
	return pageNumber
}

/*
 * getPageInfo
 * - Get information about jobs per page
 */
func getPageInfo(page int, mainC chan<- []jobInfo) {

	var jobs []jobInfo
	c := make(chan jobInfo)

	pageURL := baseURL + "&start=" + strconv.Itoa(page*jobPerPage)
	fmt.Println("Requesting", pageURL)

	res, err := http.Get(pageURL)

	checkError(err)
	checkResponseCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	// get all cards
	jobSearchCard := doc.Find(".jobsearch-SerpJobCard")
	// s -> each card
	jobSearchCard.Each(func(i int, card *goquery.Selection) {
		go extractJobCard(card, c)
	})

	for i := 0; i < jobSearchCard.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

/*
 * writeJobs
 * -  Write searched information to csv
 */
func writeJobs(jobs []jobInfo) {
	// csv library (make excel)
	file, err := os.Create("jobs.csv")
	checkError(err)

	c := make(chan []string)

	w := csv.NewWriter(file)

	// flush : data injection
	defer w.Flush()

	headers := []string{"Link", "Title", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkError(wErr)

	// for loop every job
	for _, job := range jobs {
		go writeJob(job, c)
	}

	for i := 0; i < len(jobs); i++ {
		jobSlice := <-c
		jwErr := w.Write(jobSlice)
		checkError(jwErr)
	}
}

/*
 * writeJob
 * - Homework to adjust goroutine to writejobs for distributing workload using seperate execution
 */
func writeJob(job jobInfo, c chan []string) {
	applyLink := "https://kr.indeed.com/viewjob?jk="
	jobSlice := []string{applyLink + job.id, job.title, job.location, job.salary, job.summary}
	c <- jobSlice
}

/*
 * extractJobCard
 * - Extract information from job card
 * - Set variables in jobInfo structure
 */
func extractJobCard(card *goquery.Selection, c chan<- jobInfo) {
	id, _ := card.Attr("data-jk") // can not initialize variable with two values
	title := cleanString(card.Find(".title>a").Text())
	location := cleanString(card.Find(".sjcl").Text())
	salary := cleanString(card.Find(".salaryText").Text())
	summary := cleanString(card.Find(".summary").Text())

	c <- jobInfo{id: id, title: title, location: location, salary: salary, summary: summary}
}

/*
 * checkError
 * - Check error from http.Get
 */
func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

/*
 * checkResponseCode
 * - Check response code from http.Get
 */
func checkResponseCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status", res.StatusCode)
	}
}

/*
 * cleanString
 * - Clean blanks from result and get pure string
 * - strings.TrimSpace: spaces will be deleted by TrimSpace(front to end)
 * - strings.Fields: seperate all words using space -> remove all tiny spaces and just put 'pure' words into array
 * - strings.Join: "hello       9    f" -> "hello 9 f" (when using seperator " ")
 */
func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

/*
 * elapsedTime
 * - To get a time
 */
func elapsedTime(tag string, msg string) func() {
	if msg != "" {
		log.Printf("[%s] %s", tag, msg)
	}

	start := time.Now()
	return func() { log.Printf("[%s] Elipsed Time: %s", tag, time.Since(start)) }
}
