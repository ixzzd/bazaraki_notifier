package main

import (
  "fmt"
  "regexp"
  "bufio"
  "os"
  "github.com/PuerkitoBio/goquery"
  "github.com/Syfaro/telegram-bot-api"
)

func _check(err error) {
  if err != nil {
    panic(err)
  }
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
  file, err := os.Create(path)
  if err != nil {
    return err
  }
  defer file.Close()

  w := bufio.NewWriter(file)
  for _, line := range lines {
    fmt.Fprintln(w, line)
  }
  return w.Flush()
}

func Contains(a []string, x string) bool {
  for _, n := range a {
    if x == n {
      return true
    }
  }
  return false
}

func main() {
  url := "https://www.bazaraki.com/real-estate/houses-and-villas-rent/?price_max=1000&city_districts=5730&city_districts=5737&city_districts=5049&city_districts=5682&city_districts=5683&city_districts=5684&city_districts=5687&city_districts=5689&city_districts=5690&city_districts=5691&city_districts=5692&city_districts=5693&city_districts=5695"

  fmt.Println("request: " + url)

  doc, err := goquery.NewDocument(url)
  _check(err)

  lines, err := readLines("data")
  _check(err)

  bot, err := tgbotapi.NewBotAPI("BOT_API_KEY")
  _check(err)


  doc.Find("a").Each(func(i int, s *goquery.Selection) {
    link, _ := s.Attr("href")
    isAdv, _ := regexp.MatchString(`/adv/\d{7}_.*/`, link)
    if isAdv {
      if ! Contains(lines, link) {
        lines = append(lines, link)

        advUrl := "https://www.bazaraki.com" + link
        msg_eg := tgbotapi.NewMessage(CHAT_ID, advUrl)
        bot.Send(msg_eg)
      }
    }
  })

 err = writeLines(lines, "data")
  _check(err)
}
