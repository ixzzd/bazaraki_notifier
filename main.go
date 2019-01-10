package main

import (
  "fmt"
  "regexp"
  "bufio"
  "os"
  "reflect"
  "strconv"
  "io/ioutil"
  "github.com/PuerkitoBio/goquery"
  "github.com/Syfaro/telegram-bot-api"
  "time"
  "net/http"
  "sync"
)

var mutex sync.Mutex

func _check(err error) {
  if err != nil {
    panic(err)
  }
}

func createFile(path string) {
  // check if file exists
  var _, err = os.Stat(path)

  // create file if not exists
  if os.IsNotExist(err) {
    var file, err = os.Create(path)
    _check(err)

    defer file.Close()
  }

  fmt.Println("File Created Successfully", path)
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

func telegramBot() {
  bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
  _check(err)

  u := tgbotapi.NewUpdate(0)

  updates, err := bot.GetUpdatesChan(u)

  for update := range updates {
    if update.Message == nil {
      continue
    }

    data_folder := os.Getenv("DATA_FOLDER") + "/"

    // Make sure that message in text
    if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

      chat_folder := data_folder + strconv.FormatInt(update.Message.Chat.ID, 10)

      switch update.Message.Text {
      case "/start":

        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a Bazaraki notification bot!")
        bot.Send(msg)
        msg1 := tgbotapi.NewMessage(update.Message.Chat.ID, "Send me Bazaraki advertisements list URL sorted by newest to start receiving notifications.")
        bot.Send(msg1)
        msg2 := tgbotapi.NewMessage(update.Message.Chat.ID, "To stop receiving notifications send me /stop")
        bot.Send(msg2)

        if os.Getenv("NOTIFY_TO_CHAT") != "" {
          chat_id_int, err := strconv.ParseInt(os.Getenv("NOTIFY_TO_CHAT"), 10, 64)
          _check(err)

          msg := tgbotapi.NewMessage(chat_id_int, "New user: @" + update.Message.From.UserName)
          bot.Send(msg)
        }

        fmt.Println("Start chat with id:" + strconv.FormatInt(update.Message.Chat.ID, 10) + ". User: @" + update.Message.From.UserName)

      case "/stop":
        err := os.RemoveAll(chat_folder)
        _check(err)

        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You successfully stop following all advertisements")
        bot.Send(msg)

      default:

        url := update.Message.Text

        fmt.Println("request: " + url)

        // Request the HTML page.
        res, err := http.Get(url)

        if err != nil {
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "URL is wrong"))
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Correct URL sample: https://www.bazaraki.com/real-estate/houses-and-villas-rent/lemesos-district-limassol/?price_min=500&price_max=1000"))
          continue
        }

        defer res.Body.Close()
        if res.StatusCode != 200 {
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "URL is wrong"))
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Correct URL sample: https://www.bazaraki.com/real-estate/houses-and-villas-rent/lemesos-district-limassol/?price_min=500&price_max=1000"))
          continue
        }

        // Load the HTML document
        doc, err := goquery.NewDocumentFromReader(res.Body)
        if err != nil {
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "URL is wrong"))
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Correct URL sample: https://www.bazaraki.com/real-estate/houses-and-villas-rent/lemesos-district-limassol/?price_min=500&price_max=1000"))
          continue
        }

        if len(doc.Find(".list-announcement").Nodes) == 0 {
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "URL is wrong"))
          bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Correct URL sample: https://www.bazaraki.com/real-estate/houses-and-villas-rent/lemesos-district-limassol/?price_min=500&price_max=1000"))
          continue
        }

        // Create advertisement list
        os.MkdirAll(chat_folder, os.ModePerm);

        advertisements_path := chat_folder + "/advertisements"
        os.Remove(advertisements_path)
        createFile(advertisements_path)

        createFile(chat_folder + "/sended_links")

        // Write url to advertisements list
        lines, err := readLines(advertisements_path)
        _check(err)

        lines = append(lines, url)

        err = writeLines(lines, advertisements_path)
        _check(err)

        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Now you are following only this url: " + url)
        bot.Send(msg)
        check_updates(false)
      }
    } else {
      msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Send URL for subscribe")
      bot.Send(msg)

    }
  }
}

func check_updates(notify bool) {
  mutex.Lock()
  defer mutex.Unlock()

  bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
  _check(err)

  data_folder := os.Getenv("DATA_FOLDER") + "/"

  folders, err := ioutil.ReadDir(data_folder)
  _check(err)


  for _, folder := range folders {
    if folder.IsDir() {
      chat_id := folder.Name()

      advertisements, err := readLines(data_folder + chat_id + "/advertisements")
      _check(err)

      for _, url := range advertisements {
        sended_links_path := data_folder + chat_id + "/sended_links"

        lines, err := readLines(sended_links_path)
        _check(err)

        doc, err := goquery.NewDocument(url)
        doc.Find("a").Each(func(i int, s *goquery.Selection) {
          link, _ := s.Attr("href")
          isAdv, _ := regexp.MatchString(`/adv/\d{7}_.*/`, link)
          if isAdv {
            if ! Contains(lines, link) {
              lines = append(lines, link)

              if notify {
                advUrl := "https://www.bazaraki.com" + link
                chat_id_int, err := strconv.ParseInt(chat_id, 10, 64)
                _check(err)

                msg := tgbotapi.NewMessage(chat_id_int, advUrl)
                bot.Send(msg)
              }
            }
          }
        })

        err = writeLines(lines, sended_links_path)
        _check(err)
      }
    }
  }
}

func main() {
  go telegramBot()

  checking_interval := 300

  if os.Getenv("CHECKING_INTERVAL") != "" {
    parsed_int, err := strconv.Atoi(os.Getenv("CHECKING_INTERVAL"))
    _check(err)
    checking_interval = parsed_int
  }

  for {
    check_updates(true)
    time.Sleep(time.Second * time.Duration(checking_interval))
  }
}

