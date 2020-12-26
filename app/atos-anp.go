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

type Error struct {
	message string
	isError bool
}

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

func buildStartURL() string {
	var start_date string
	var end_date string
	//type to code
	var mappingAtos = make(map[string]string)
	mappingAtos["Ata"] = "106"
	mappingAtos["Autorização"] = "151"
	mappingAtos["Despacho"] = "152"
	mappingAtos["IN_Financeira_Administrativa"] = "202"
	mappingAtos["IN_Gestão_Interna"] = "203"
	mappingAtos["IN_Gestão_Técnica"] = "204"
	mappingAtos["IN_Recursos_Humanos"] = "200"
	mappingAtos["IN_Segurança_Operacional"] = "201"
	mappingAtos["Instrução_Normativa"] = "9"
	mappingAtos["Portaria_ANP"] = "157"
	mappingAtos["Portaria_Conjunta"] = "153"
	mappingAtos["Portaria_Técnica"] = "183"
	mappingAtos["Resolução"] = "24"
	mappingAtos["Resolução_Conjunta"] = "156"
	mappingAtos["Resolução_de_Diretoria_RD"] = "163"
	//date params
	flag.StringVar(&start_date, "data_inicio", "", "Data de Inicio no formato dd-mm-YYYY")
	flag.StringVar(&end_date, "data_fim", "", "Data Final no formato dd-m-YYYY")
	//normative acts types
	collAta := flag.Bool("ata", false, "Para coletar do tipo Ata.")
	collAutorização := flag.Bool("autorização", false, "Para coletar do tipo Autorização.")
	collDespacho := flag.Bool("despacho", false, "Para coletar do tipo Despacho.")
	collInFinAdm := flag.Bool("in_fin_adm", false, "Para coletar do tipo IN Financeira Administrativa.")
	collInGesInterna := flag.Bool("in_ges_interna", false, "Para coletar do tipo IN Gestão Interna")
	collInGesTecnica := flag.Bool("in_ges_tec", false, "Para coletar do tipo IN Gestão Técnica")
	collInRecHumanos := flag.Bool("in_rec_humanos", false, "Para coletar do tipo IN Recursos Humanos.")
	collInSegOp := flag.Bool("in_seg_op", false, "Para coletar do tipo IN Segurança Operacional.")
	collInstrNorm := flag.Bool("instr_norm", false, "Para coletar do tipo Instrução Normativa.")
	collPortAnp := flag.Bool("port_anp", false, "Para coletar do tipo Portaria ANP.")
	collPortConj := flag.Bool("port_conj", false, "Para coletar do tipo Portaria Conjunta.")
	collPortTecnica := flag.Bool("port_tecnica", false, "Para coletar do tipo Portaria Técnica.")
	collRes := flag.Bool("resolução", false, "Para coletar do tipo Resolução.")
	collResConj := flag.Bool("res_conjunta", false, "Para coletar do tipo Resolução Conjunta.")
	collResRD := flag.Bool("res_diretoria", false, "Para coletar do tipo Resolução de Diretoria RD.")
	collAll := flag.Bool("all", false, "Para coletar todos os tipos.")
	flag.Parse()
	validateDateParams(start_date, end_date)
	var TypeBool = make(map[string]bool)
	TypeBool["all"] = *collAll
	TypeBool["Ata"] = *collAta
	TypeBool["Autorização"] = *collAutorização
	TypeBool["Despacho"] = *collDespacho
	TypeBool["IN_Financeira_Administrativa"] = *collInFinAdm
	TypeBool["IN_Gestão_Interna"] = *collInGesInterna
	TypeBool["IN_Gestão_Técnica"] = *collInGesTecnica
	TypeBool["IN_Recursos_Humanos"] = *collInRecHumanos
	TypeBool["IN_Segurança_Operacional"] = *collInSegOp
	TypeBool["Instrução_Normativa"] = *collInstrNorm
	TypeBool["Portaria_ANP"] = *collPortAnp
	TypeBool["Portaria_Conjunta"] = *collPortConj
	TypeBool["Portaria_Técnica"] = *collPortTecnica
	TypeBool["Resolução"] = *collRes
	TypeBool["Resolução_Conjunta"] = *collResConj
	TypeBool["Resolução_de_Diretoria_RD"] = *collResRD
	//build a query string
	if TypeBool["all"] == true {
		var START_URL = fmt.Sprintf("https://atosoficiais.com.br/anp?q=&status_consolidacao=0&date_start=%v&date_end=%v", start_date, end_date)
		return START_URL
	} else {
		queryString := ""
		for key, value := range TypeBool {
			if key == "all" {
				continue
			} else {
				if value == true {
					queryString = queryString + "&types=" + mappingAtos[key]
				}
			}

		}
		if queryString != "" {
			START_URL := fmt.Sprintf("https://atosoficiais.com.br/anp?q=%v&date_start=%v&date_end=%v", queryString, start_date, end_date)
			return START_URL
		} else {
			// se nenhum parâmetro de tipo de normativo for setado, a query será realizada como -all
			var START_URL = fmt.Sprintf("https://atosoficiais.com.br/anp?q=&status_consolidacao=0&date_start=%v&date_end=%v", start_date, end_date)
			return START_URL
		}

	}
}

func buildPaginationURL(url string, pageNumber string, first_pagination bool) (string, Error) {
	if first_pagination {
		paginationURL := url + "&page=" + pageNumber
		return paginationURL, Error{isError: false}
	} else if strings.Contains(url, "page=") {
		getIndex := strings.LastIndex(url, "page=")
		if getIndex != -1 {
			paginationURL := url[:getIndex] + "&page=" + pageNumber
			return paginationURL, Error{isError: false}
		}
	}
	return "", Error{message: "Não foi possível identificar a URL de paginação", isError: true}
}

func main() {
	var BASE_URL string = "https://atosoficiais.com.br"
	START_URL := buildStartURL()
	client, err := mongo.GetMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	initialRequest := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.80 Safari/537.36"),
	)
	initialRequest.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay:       2 * time.Second,
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
					url, err := buildPaginationURL(e.Request.URL.String(), pageNumber, true)
					if err.isError == true {
						log.Fatal(err.message)
					}
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
			url, err := buildPaginationURL(c.Request.URL.String(), nextPage, false)
			if err.isError == true {
				log.Fatal(err.message)
			}
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
		if e.StatusCode == 200 {
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
				err = mongo.CreateDocument(client, dat)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			messageError := fmt.Sprintf("Request para %v com status code: %v", e.Request.URL.String(), e.StatusCode)
			fmt.Println(messageError)
		}

	})
	initialRequest.Visit(START_URL)
}
