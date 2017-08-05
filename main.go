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

func selectAndSend(tag string, chatID int) {
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
		sendMessage(chatID, publishDate.String() + " " + title + " " + description, "")	
		count++
	}
	if count == 0 {
		sendMessage(chatID, "Вакансий нет", "")
	}
}

func main() {
	SetWebhook()
	port := os.Getenv("PORT")
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

		db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Println(err)
		}
		count := 0
		switch update.Message.Text {
			case "Все вакансии":
				sendMessage(update.Message.Chat.Id, "Все вакансии", "")
			case "Программисты":
				k := string(replyMarkup([][]string{{"Все"}, {"C➕➕"}, {"Python"}, {"Golang"}}))
				sendMessage(update.Message.Chat.Id, "Программисты", k)
				//sendMessage(update.Message.Chat.Id, "Доступные команды: 1. 📰\\news - последние новости города и области\n2. 🎉\\events - события города")
				//log.Println(message)
			case "Все":
				count = 0
				rows, err := db.Query("SELECT publish_date, title, description FROM Jobs WHERE section = 'programmers'")
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
					sendMessage(update.Message.Chat.Id, publishDate.String() + " " + title + " " + description, string(replyMarkup([][]string{{"Все (ещё)"}, {"Назад"}})))
					count++
					if count == 4 {
						break
					}
				}
				if count == 0 {
					sendMessage(update.Message.Chat.Id, "Вакансий нет", "")
				}
			case "Все (ещё)":
				rows, err := db.Query("SELECT publish_date, title, description FROM Jobs WHERE section = 'programmers'")
				if err != nil {
					log.Println(err)
				}
				i := 0
				count := count + 4
				for rows.Next() {
					if i < (count - 4) {
						i++
						continue
					}
					var publishDate time.Time
					var title, description string
					err = rows.Scan(&publishDate, &title, &description)
					if err != nil {
						log.Println(err)
					}
					sendMessage(update.Message.Chat.Id, publishDate.String() + " " + title + " " + description, string(replyMarkup([][]string{{"Все (ещё 5)"}, {"Назад"}})))
					i++
					if i == count {
						break
					}
				}
			case "Назад":
				sendMessage(update.Message.Chat.Id, "Программисты", string(replyMarkup([][]string{{"Все"}, {"C➕➕"}, {"Python"}, {"Golang"}})))
			case "C➕➕":
				selectAndSend("c++", update.Message.Chat.Id)
			case "Python":
				selectAndSend("python", update.Message.Chat.Id)
			case "Golang":
				selectAndSend("golang", update.Message.Chat.Id)
			case "Дизайнеры":
				sendMessage(update.Message.Chat.Id, "Дизайнеры", "")
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
