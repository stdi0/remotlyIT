package main

import (
	_ "github.com/lib/pq"
	"encoding/json"
	"html/template"
	"database/sql"
	"io/ioutil"
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

		switch update.Message.Text {
			case "Программисты":
				k := string(replyMarkup([][]string{{"Все"}, {"C➕➕"}, {"Python"}, {"Golang"}}))
				sendMessage(update.Message.Chat.Id, "Программисты", k)
				log.Println("JSON:", k)
				//sendMessage(update.Message.Chat.Id, "Доступные команды: 1. 📰\\news - последние новости города и области\n2. 🎉\\events - события города")
				//log.Println(message)
			case "Все":
				rows, err := db.Query("SELECT publish_date, title, description FROM Jobs WHERE section = 'programmers'")
				if err != nil {
					log.Println(err)
				}
				for rows.Next() {
					var publish_date time.Time
					var title, description string
					err = rows.Scan(&publish_date, &title, &description)
					if err != nil {
						log.Println(err)
					}
					sendMessage(update.Message.Chat.Id, publish_date.String() + " " + title + " " + description, "")
				}
			case "Дизайнеры":
				sendMessage(update.Message.Chat.Id, "Дизайнеры", "")
			case "Все вакансии":
				sendMessage(update.Message.Chat.Id, "Все вакансии", "")
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
			if _, err = db.Exec("INSERT INTO Tags (job_id, tag) VALUES ($1, $2)", lastID, r.Form["tags"][0]); err != nil {
					log.Println(err)
			}
		}
		t, _ := template.ParseFiles("post.html")
		t.Execute(w, nil)
	})
	http.ListenAndServe(":"+port, nil)
	return
}
