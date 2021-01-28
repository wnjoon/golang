package main

import (
	"os"
	"strings"

	scrapper "../jobscrapper"
	"github.com/labstack/echo"
)

// You should remove /v4 to get a 'go echo'
// go get github.com/labstack/echo

const fileName string = "../jobscrapper/jobs.csv"

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))

}

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(fileName, "jobs.csv")
}
