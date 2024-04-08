package modules

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type Result struct {
	BuildServer  string `json:"build_server"`
	DownloadLink string `json:"download_link"`
	ImageLink    string `json:"image_link"`
	ImageCode    string `json:"image_code"`
	Info         string `json:"info"`
}

func Ephoto360(urlEphoto string, texts ...string) (interface{}, error) {

	pattern := `^https:\/\/en\.ephoto360\.com\/[\w-]+-\d+\.html$`

	validUrlEphoto, err := regexp.MatchString(pattern, urlEphoto)
	if err != nil {
		return nil, err
	}

	if !validUrlEphoto {
		return nil, errors.New("invalid ephoto url")
	}

	res, err := http.Get(urlEphoto)
	if err != nil {
		return nil, errors.New("failed to get information of website")
	}
	res.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	cookies := res.Cookies()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	findToken := doc.Find("#token")
	tokenValue, ok := findToken.Attr("value")
	if !ok {
		log.Fatal("Cannot get token value")
	}

	findBuildServer := doc.Find("#build_server")
	buildServerValue, ok := findBuildServer.Attr("value")
	if !ok {
		log.Fatal("Cannot get build_server value")
	}

	findBuildServerId := doc.Find("#build_server_id")
	buildServerIdValue, ok := findBuildServerId.Attr("value")
	if !ok {
		log.Fatal("Cannot get build_server_id value")
	}

	var radioSelects []string
	findRadioSelect := doc.Find("input[name='radio0[radio]']")
	findRadioSelect.Each(func(i int, s *goquery.Selection) {
		radioSelectValue, ok := s.Attr("value")
		if !ok {
			fmt.Println("radio0[radio] not available")
		}
		// fmt.Println(radioSelect)
		radioSelects = append(radioSelects, radioSelectValue)
	})

	var body bytes.Buffer
	bodyWriter := multipart.NewWriter(&body)

	// shuffle slice elements
	if len(radioSelects) > 0 {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(radioSelects))))
		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}
		bodyWriter.WriteField("radio0[radio]", radioSelects[randomIndex.Int64()])
	}

	for i, text := range texts {
		bodyWriter.WriteField("text["+strconv.Itoa(i)+"]", text)
	}
	bodyWriter.WriteField("submit", "GO")
	bodyWriter.WriteField("token", tokenValue)
	bodyWriter.WriteField("build_server", buildServerValue)
	bodyWriter.WriteField("build_server_id", buildServerIdValue)

	// close form data
	bodyWriter.Close()

	request, err := http.NewRequest("POST", urlEphoto, &body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// add cookie to header
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	request.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	request.Header.Set("Referer", "https://en.ephoto360.com/naruto-shippuden-logo-style-text-effect-online-808.html")
	request.Header.Set("Origin", "https://en.ephoto360.com")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer response.Body.Close()

	scrape, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	findValueInput := scrape.Find("#form_value_input")
	valueInput, ok := findValueInput.Attr("value")
	if !ok {
		return nil, errors.New("cannot get form value input")
	}

	// fmt.Println(valueInput)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(valueInput), &data)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, errors.New("Error parsing JSON:" + err.Error())
	}
	var radioValue string
	if radio0, ok := data["radio0"].(map[string]interface{}); ok {
		if radio, ok := radio0["radio"].(string); ok {
			// fmt.Println("Radio:", radio)
			radioValue = radio
		}
	}

	var bodyCreateImg bytes.Buffer
	form := multipart.NewWriter(&bodyCreateImg)
	for i, text := range texts {
		form.WriteField("text["+strconv.Itoa(i)+"]", text)
	}
	form.WriteField("id", data["id"].(string))
	form.WriteField("token", data["token"].(string))
	form.WriteField("radio0[radio]", radioValue)
	form.WriteField("build_server", data["build_server"].(string))
	form.WriteField("build_server_id", data["build_server_id"].(string))

	// close form data
	form.Close()

	createImg, err := http.NewRequest("POST", "https://en.ephoto360.com/effect/create-image", &bodyCreateImg)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	for _, cookie := range cookies {
		createImg.AddCookie(cookie)
	}

	createImg.Header.Set("Content-Type", form.FormDataContentType())
	createImg.Header.Set("Referer", "https://en.ephoto360.com/naruto-shippuden-logo-style-text-effect-online-808.html")
	createImg.Header.Set("Origin", "https://en.ephoto360.com")

	// Mengirim request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(createImg)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, errors.New("error when creating image")
	}
	defer resp.Body.Close()

	readBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	var dataFormInput map[string]interface{}
	err = json.Unmarshal(readBody, &dataFormInput)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	result := Result{
		BuildServer:  data["build_server"].(string),
		ImageLink:    dataFormInput["image"].(string),
		ImageCode:    dataFormInput["image_code"].(string),
		DownloadLink: data["build_server"].(string) + dataFormInput["image"].(string),
		Info:         dataFormInput["info"].(string),
	}

	return result, nil
}
