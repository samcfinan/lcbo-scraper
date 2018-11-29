package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

func main() {
	fmt.Println("test")
	mainCollector := colly.NewCollector(
		colly.AllowedDomains("lcbo.com", "www.lcbo.com"),
		colly.CacheDir("./lcbo_cache"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36"),
		colly.Async(true),
	)

	// Supply random user agent to reduce chance of getting flagged.
	extensions.RandomUserAgent(mainCollector)
	// Provide valid referrer info
	extensions.Referrer(mainCollector)

	productCollector := mainCollector.Clone()

	// Log all pages MainCollector visits
	mainCollector.OnRequest(func(r *colly.Request) {
		log.Println("MainCollector visiting:", r.URL.String())
	})

	// Log all pages ProductCollector visits
	productCollector.OnRequest(func(r *colly.Request) {
		// log.Println("Visiting:", r.URL.String())
	})

	mainCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {

		// Ignore irrelevant links. Should only visit catalog pages.
		url := e.Attr("href")
		if strings.HasPrefix(url, "/lcbo/product") {

			// Only send product urls to productCollector
			productCollector.Visit(e.Request.AbsoluteURL(url))
		} else if strings.HasPrefix(url, "/lcbo/catalog") {
			e.Request.Visit(url)
		} else {
			return
		}

	})

	// Parsing product data below
	productCollector.OnResponse(parseProduct)

	// Run Crawler
	mainCollector.Visit("https://lcbo.com")
	mainCollector.Wait()
	productCollector.Wait()
}

func parseProduct(r *colly.Response) {
	if len(r.Body) == 0 {
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(r.Body)))
	if err != nil {
		log.Panic(err)
	}
	name := doc.Find("#prodName").Text()

	// Parse string as float
	volumeString := strings.Split(doc.Find(".product-volume").Text(), " ")
	volume, err := strconv.Atoi(volumeString[0])
	if err != nil {
		fmt.Println(err)
	}
	format := volumeString[len(volumeString)-1]

	// Parse string as float, strip $
	priceString := strings.Replace(doc.Find(".price-value").Text(), "$", "", -1)
	price, _ := strconv.ParseFloat(priceString, 64)
	if err != nil {
		fmt.Println(err)
	}

	details := doc.Find("dl")
	properties := details.ContentsFiltered("dd")

	alcoholString := strings.Replace(properties.Eq(1).Text(), "%", "", -1)
	alcohol, err := strconv.ParseFloat(alcoholString, 64)
	if err != nil {
		fmt.Println(err)
	}
	origin := properties.Eq(2).Text()
	producer := properties.Eq(3).Text()
	sugar := properties.Eq(4).Text()
	sweetness := properties.Eq(5).Text()
	style := properties.Eq(6).Text()
	varietal := properties.Eq(7).Text()

	product := Product{
		Name:      name,
		Volume:    volume,
		Price:     price,
		Alcohol:   alcohol,
		Format:    format,
		Producer:  producer,
		Sweetness: sweetness,
		Sugar:     sugar,
		Style:     style,
		Varietal:  varietal,
		Origin:    origin,
	}

	storeProduct(&product)
	// fmt.Printf("%s %d %f %s %s\n", product.Name, product.Volume, product.Price, product.Varietal, product.Origin)
	// fmt.Printf("%s %s %s\n", name, volumeFormat, price)
}

func storeProduct(p *Product) {
	fmt.Printf("%s %d %f %s %s\n", p.Name, p.Volume, p.Price, p.Varietal, p.Origin)
}

// Product is an LCBO product
type Product struct {
	Name      string
	Volume    int
	Price     float64
	Alcohol   float64
	Format    string
	Producer  string
	Sweetness string
	Sugar     string
	Style     string
	Varietal  string
	Origin    string
	Link      string
}
