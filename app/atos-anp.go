package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"mongo"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func validateDateParams(start_date string, end_date string) {
	if start_date == "" {
		log.Fatalf("É necessário setar a flag -data_inicio")
	}
	if end_date == "" {
		log.Fatalf("É necessário setar a flag -end_date")
	}
	re_date := regexp.MustCompile(`\d{2}-\d{2}-\d{4}`)
	start_date_validate := re_date.FindAll([]byte(start_date), -1)
	if start_date_validate == nil {
		log.Fatalf("O parâmetro -data_inicio deve ser do fomato dd-mm-YYYY")
	}
	end_date_validate := re_date.FindAll([]byte(end_date), -1)
	if end_date_validate == nil {
		log.Fatalf("O parâmetro -data_fim deve ser do fomato dd-mm-YYYY")
	}
}

func main() {
	var BASE_URL string = "https://atosoficiais.com.br"
	var start_date string
	var end_date string
	flag.StringVar(&start_date, "data_inicio", "", "Data de Inicio no formato dd-mm-YYYY")
	flag.StringVar(&end_date, "data_fim", "", "Data Final no formato dd-m-YYYY")
	flag.Parse()
	validateDateParams(start_date, end_date)
	var START_URL string = fmt.Sprintf("https://atosoficiais.com.br/anp?q=&status_consolidacao=0&date_start=%v&date_end=%v", start_date, end_date)
	client, err := mongo.GetMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	initialRequest := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.80 Safari/537.36"),
	)
	initialRequest.AllowURLRevisit = false

	initialRequest.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay:       5 * time.Second,
	})
	paginator := initialRequest.Clone()
	parserLink := initialRequest.Clone()
	apiRequest := initialRequest.Clone()
	initialRequest.OnRequest(func(r *colly.Request) {
		fmt.Println("visitando", r.URL)
	})
	paginator.OnRequest(func(r *colly.Request) {
		fmt.Println("visitando", r.URL)
	})
	initialRequest.OnHTML("h4", func(e *colly.HTMLElement) {
		if e.Attr("class") == "small-title text-green" {
			re := regexp.MustCompile("[0-9]+")
			numberOfResults := re.FindAllString(e.Text, -1)
			if numberOfResults != nil {
				numberOfResults, _ := strconv.ParseFloat(numberOfResults[0], 32)
				numberOfPages := math.Ceil(numberOfResults / 10)
				messageNumPages := fmt.Sprintf("Há ao todo %v páginas a serem visitadas", numberOfPages)
				message := fmt.Sprintf("Foram encontrados %v valores a serem coletados", numberOfResults)
				numberOfPagesString := fmt.Sprintf("%v", numberOfPages)
				fmt.Println(message)
				fmt.Println(messageNumPages)
				pageNumber := e.Request.Ctx.Get("Pagina")
				if pageNumber == "" && numberOfResults > 10 {
					pageNumber := "2"
					e.Request.Ctx.Put("pageNumber", pageNumber)
					e.Request.Ctx.Put("numberOfResults", numberOfResults)
					e.Request.Ctx.Put("numberOfPages", numberOfPagesString)
					middle_url := fmt.Sprintf("/anp?q=&status_consolidacao=0&date_start=%v&date_end=%v&page=", start_date, end_date)
					url := BASE_URL + middle_url + pageNumber
					paginator.Request("GET", url, nil, e.Request.Ctx, nil)
				}
			} else {
				fmt.Println("Não foram encontrados resultados nessa pesquisa")
			}
		}
	})

	//parse links
	initialRequest.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "/anp/") {
			fmt.Println("Primeira página: ", link)
			url := BASE_URL + link
			parserLink.Visit(url)
		}

	})
	//parse links
	paginator.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "/anp/") {
			url := BASE_URL + link
			parserLink.Visit(url)
		}

	})
	//paginate over pages
	paginator.OnScraped(func(c *colly.Response) {
		GetpageNumber := c.Request.Ctx.Get("pageNumber")
		GetnumberOfPages := c.Request.Ctx.Get("numberOfPages")
		if GetpageNumber == "" {
			fmt.Println("Não foi possível identificar a página.")
		}
		var numberOfPages, _ = strconv.ParseFloat(GetnumberOfPages, 64)
		var pageNumber, _ = strconv.ParseFloat(GetpageNumber, 64)
		if pageNumber < numberOfPages {
			pageNumber++
			nextPage := fmt.Sprintf("%v", pageNumber)
			middle_url := fmt.Sprintf("/anp?q=&status_consolidacao=0&date_start=%v&date_end=%v&page=", start_date, end_date)
			url := BASE_URL + middle_url + nextPage
			c.Request.Ctx.Put("pageNumber", nextPage)
			c.Request.Ctx.Put("numberOfPages", GetnumberOfPages)
			paginator.Request("GET", url, nil, c.Request.Ctx, nil)
		}
	})
	//parserLink
	parserLink.OnHTML("button", func(e *colly.HTMLElement) {
		if e.Attr("class") == "btn btn-default btn-lg content-block-header-box btn-vinculados" {
			data_id := e.Attr("data-id")
			e.Request.Ctx.Put("id", data_id)
			source := e.DOM.ParentsUntil("~").Find("article")
			if html, err := source.Html(); err != nil {
				log.Fatal(err)
			} else {
				e.Request.Ctx.Put("html", html)
				sourceText := source.Text()
				e.Request.Ctx.Put("text", sourceText)
			}
			fetchApi := fmt.Sprintf("https://api.leismunicipais.com.br/atosoficiais/leis/%v/atos-vinculados", data_id)
			apiRequest.Request("GET", fetchApi, nil, e.Request.Ctx, nil)
		}
	})
	//Api
	apiRequest.OnResponse(func(e *colly.Response) {
		var dat mongo.AtosVinculados
		data_id := e.Ctx.Get("id")
		html := e.Ctx.Get("html")
		text := e.Ctx.Get("text")
		err := json.Unmarshal(e.Body, &dat)
		if err != nil {
			log.Fatal(err)
		} else {
			dat.Id = data_id
			dat.Html = html
			dat.Text = text
			mongo.CreateDocument(client, dat)

		}

	})
	initialRequest.Visit(START_URL)
}
