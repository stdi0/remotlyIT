package main

import (
	_ "github.com/lib/pq"
	"encoding/json"
	"html/template"
	"database/sql"
	"io/ioutil"
	"strings"
	"time"
	"log"
	"net/http"
	"strconv"
	"os"
	//"fmt"
)


type Update struct {
		UpdateId int `json:"update_id"`
		Message struct {
			MessageId int `json:"message_id"`
			From      struct {
				Id           int    `json:"id"`
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				Id        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				ChatType  string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	}

type Message struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageId int `json:"message_id"`
			From      struct {
				Id           int    `json:"id"`
				FirstName    string `json:"first_name"`
				Username	 string `json:"username"`
			} `json:"from"`
			Chat struct {
				Id        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				ChatType  string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"result"`
	} 

type ReplyKeyboardMarkup struct {
	Keyboard [][]string `json:"keyboard"`
	ResizeKeyboard bool `json:"resize_keyboard"`
	OneTimeKeyboard bool `json:"one_time_keyboard"`
}

const (
	Token  string = "446256177:AAEA4xrX-nFy3qF5ynm_AqH-tQNtfCsI3OM"
	UrlApiTelegram string = "https://api.telegram.org/bot"
	WebHookUrl string = "https://glacial-island-37216.herokuapp.com/"
)

func SetWebhook() {
	path := UrlApiTelegram + Token
	query := "/setWebhook?url=" + WebHookUrl
	c := http.Client{}
	resp, err := c.Get(path + query)
	if err != nil {
		log.Println(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	resp.Body.Close()
}

func sendMessage(chatID int, text string, replyMarkup string) Message {
	path := UrlApiTelegram + Token
	query := "/sendMessage?chat_id="
	c := http.Client{}
	resp, err := c.Get(path + query + strconv.Itoa(chatID) + "&text=" + text + "&reply_markup=" + replyMarkup)
	if err != nil {
		log.Println(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	message := Message{}
	err = json.Unmarshal(body, &message)
	if err != nil {
		log.Println(err)
	}
	resp.Body.Close()
	return message
} 

func replyMarkup(keyboard [][]string) []byte {
	replyMarkup := ReplyKeyboardMarkup{
		Keyboard: keyboard, 
		ResizeKeyboard: true, 
		OneTimeKeyboard: true,
	}
	j, _ := json.Marshal(replyMarkup)
	return j
}

func tagSend(tag string, chatID int, text string) int {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
	}
	rows, err := db.Query("SELECT job_id FROM Tags WHERE tag = '" + tag + "'")
	if err != nil {
		log.Println(err)
	}
	count := 0
	for rows.Next() {
		var jobID int
		err = rows.Scan(&jobID)
		if err != nil {
			log.Println(err)
		}
		var publishDate time.Time
		var title, description string
		err := db.QueryRow("SELECT publish_date, title, description FROM Jobs WHERE id = '" + strconv.Itoa(jobID) + "'").Scan(&publishDate, &title, &description)
		if err != nil {
			log.Println(err)
		}
		sendMessage(chatID, publishDate.Format("2006-01-02") + " " + title + "%0A" + description, string(replyMarkup([][]string{{text}, {"Назад"}})))	
		count++
	}
	if count == 0 {
		sendMessage(chatID, "Вакансий нет", string(replyMarkup([][]string{{text}, {"Назад"}})))
	}
	return count
}

func tagCountSend(tag string, chatID int, count int, text string) int {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
	}
	rows, err := db.Query("SELECT job_id FROM Tags WHERE tag = '" + tag + "'")
	if err != nil {
		log.Println(err)
	}
	i := 0
	count = count + 4
	foo := false
	for rows.Next() {
		if i < (count - 4) {
			i++
			continue
		}
		foo = true
		var jobID int
		err = rows.Scan(&jobID)
		if err != nil {
			log.Println(err)
		}
		var publishDate time.Time
		var title, description string
		err := db.QueryRow("SELECT publish_date, title, description FROM Jobs WHERE id = '" + strconv.Itoa(jobID) + "'").Scan(&publishDate, &title, &description)
		if err != nil {
			log.Println(err)
		}
		sendMessage(chatID, publishDate.Format("2006-01-02") + " " + title + "%0A" + description, string(replyMarkup([][]string{{text}, {"Назад"}})))	
		i++
		if i == count {
			break
		}
	}
	if foo == false {
		sendMessage(chatID, "Вакансий больше нет :)", string(replyMarkup([][]string{{"Назад"}})))
	}
	return count
}

func sectionSend(section string, chatID int, text string) int {
	count := 0
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
	}
	rows, err := db.Query("SELECT publish_date, title, description FROM Jobs WHERE section = '" + section + "'")
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var publishDate time.Time
		var title, description string
		err = rows.Scan(&publishDate, &title, &description)
		if err != nil {
			log.Println(err)
		}
		sendMessage(chatID, publishDate.Format("2006-01-02") + " " + title + "%0A" + description, string(replyMarkup([][]string{{text}, {"Назад"}})))
		count++
		if count == 4 {
			break
		}
	}
	if count == 0 {
		sendMessage(chatID, "Вакансий нет", string(replyMarkup([][]string{{text}, {"Назад"}})))
	}
	return count
}

func sectionCountSend(section string, chatID int, count int, text string) int {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
	}
	rows, err := db.Query("SELECT publish_date, title, description FROM Jobs WHERE section = '" + section + "'")
	if err != nil {
		log.Println(err)
	}
	i := 0
	count = count + 4
	foo := false
	for rows.Next() {
		if i < (count - 4) {
			i++
			continue
		}
		foo = true
		var publishDate time.Time
		var title, description string
		err = rows.Scan(&publishDate, &title, &description)
		if err != nil {
			log.Println(err)
		}
		sendMessage(chatID, publishDate.Format("2006-01-02") + " " + title + "%0A" + description, string(replyMarkup([][]string{{text}, {"Назад"}})))
		i++
		if i == count {
			break
		}
	}
	if foo == false {
		sendMessage(chatID, "Вакансий больше нет :)", string(replyMarkup([][]string{{"Назад"}})))
	}
	return count
}

func main() {
	SetWebhook()
	port := os.Getenv("PORT")
	var count int
	var pointer string
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, _ := ioutil.ReadAll(r.Body)
		log.Println("Body: ", string(body))
		update := Update{}
		err := json.Unmarshal(body, &update)
		if err != nil {
			log.Println(err)
		}
		log.Println("Update: ", update)
		switch update.Message.Text {
			case "Все вакансии":
				pointer = "Все вакансии"
				count = sectionSend("programmers' OR section = 'designers", update.Message.Chat.Id, "Все вакансии (ещё)")
			case "Все вакансии (ещё)":
				count = sectionCountSend("programmers' OR section = 'designers", update.Message.Chat.Id, count, "Все вакансии (ещё)")
			case "Программисты":
				pointer = "Все вакансии"
				//k := string(replyMarkup([][]string{{"Все"}, {"Java", "Python"}, {"PHP", "C#"}, {"JavaScript", "C/C➕➕"}, {"Golang", "Ruby"}, {"Назад"}}))
				sendMessage(update.Message.Chat.Id, "Вакансии для программистов", string(replyMarkup([][]string{{"Все"}, {"Java", "Python"}, {"Назад"}})))
				//sendMessage(update.Message.Chat.Id, "Доступные команды: 1. 📰\\news - последние новости города и области\n2. 🎉\\events - события города")
				//log.Println(message)
			case "Все":
				pointer = "Программисты"
				count = sectionSend("programmers", update.Message.Chat.Id, "Все (ещё)")
			case "Все (ещё)":
				pointer = "Программисты"
				count = sectionCountSend("programmers", update.Message.Chat.Id, count, "Все (ещё)")
			case "Java": 
				pointer = "Программисты"
				count = tagSend("c++", update.Message.Chat.Id, "Java (ещё)")	
			case "PHP": 
				pointer = "Программисты"
				count = tagSend("c++", update.Message.Chat.Id, "PHP (ещё)")	
			case "JavaScript": 
				pointer = "Программисты"
				count = tagSend("c++", update.Message.Chat.Id, "JavaScript (ещё)")	
			case "Ruby": 
				pointer = "Программисты"
				count = tagSend("c++", update.Message.Chat.Id, "Ruby (ещё)")	
			case "C/C➕➕": 
				pointer = "Программисты"
				count = tagSend("c++", update.Message.Chat.Id, "C➕➕ (ещё)")
			case "C/C➕➕ (ещё)":
				pointer = "Программисты"
				count = tagCountSend("c++", update.Message.Chat.Id, count, "C➕➕ (ещё)")
			case "Python":
				pointer = "Программисты"
				count = tagSend("python", update.Message.Chat.Id, "Python (ещё)")
			case "Python (ещё)":
				pointer = "Программисты"
				count = tagCountSend("c++", update.Message.Chat.Id, count, "Python (ещё)")
			case "Golang":
				pointer = "Программисты"
				count = tagSend("golang", update.Message.Chat.Id, "Golang (ещё)")
			case "Golang (ещё)":
				pointer = "Программисты"
				count = tagCountSend("c++", update.Message.Chat.Id, count, "Golang (ещё)")
			case "Назад":
				if pointer == "Все вакансии" {
					sendMessage(update.Message.Chat.Id, "Главное меню", string(replyMarkup([][]string{{"Все вакансии"}, {"Программисты"}, {"Дизайнеры"}})))
				} else if pointer == "Программисты" {
					sendMessage(update.Message.Chat.Id, "Вакансии для программистов", string(replyMarkup([][]string{{"Все"}, {"Java", "Python"}, {"PHP", "C#"}, {"JavaScript", "C/C➕➕"}, {"Golang", "Ruby"}, {"Назад"}})))
					pointer = "Все вакансии"
				}
			case "Дизайнеры":
				pointer = "Все вакансии"
				count = sectionSend("designers", update.Message.Chat.Id, "Все (ещё)")
			default:
				sendMessage(update.Message.Chat.Id, "Это сообщение отобразится при отправке /start", string(replyMarkup([][]string{{"Все вакансии"}, {"Программисты"}, {"Дизайнеры"}})))
				//log.Println(message)
		}
		/*for _, v := range update.Result {
			fmt.Println(v.Message.Text)
			message := sendMessage(v.Message.Chat.Id, v.Message.Text, 0)
			if message.Result.Text != v.Message.Text {
				log.Println(message)
			}
		}*/
		r.Body.Close()
	}
	})
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
			if err != nil {
				log.Println(err)
			}
			var lastID int
			if err = db.QueryRow("INSERT INTO Jobs (publish_date, title, description, section) VALUES ($1, $2, $3, $4) RETURNING id", time.Now(), r.Form["title"][0], r.Form["description"][0], r.Form["section"][0]).Scan(&lastID); err != nil {
					log.Println(err)
			}
			s := strings.Split(r.Form["tags"][0], ",")
			for _, v := range s {
				if _, err = db.Exec("INSERT INTO Tags (job_id, tag) VALUES ($1, $2)", lastID, strings.ToLower(strings.TrimSpace(v))); err != nil {
						log.Println(err)
				}
			}
		}
		t, _ := template.ParseFiles("post.html")
		t.Execute(w, nil)
	})
	http.ListenAndServe(":"+port, nil)
	return
}
