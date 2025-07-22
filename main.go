package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"

	"github.com/knadh/koanf/v2"
)

var config = koanf.New(".")
var envPrefix = "CRUSH_"

func main() {
	configure()

	bot, err := tgbotapi.NewBotAPI(config.String("token"))
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if strings.HasPrefix(update.Message.Chat.Title, config.String("chat.title")) && update.Message.Chat.ID == config.Int64("chat.id") {
			if !forgotten(update.Message.Text) {
				continue
			}

			log.Printf("%+v\n", update.Message.Chat)

			msg := tgbotapi.NewSticker(update.Message.Chat.ID, tgbotapi.FileID(config.String("alt.file.id")))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}

func forgotten(message string) bool {
	keywords := config.Strings("alt.keywords")
	if len(keywords) == 0 {
		keywords = strings.Split(config.String("alt.keywords"), " ")
	}

	for _, word := range strings.Split(message, " ") {
		for _, prefix := range keywords {
			if strings.HasPrefix(word, prefix) {
				return true
			}
		}
	}

	return false
}

func configure() {
	keyTransformer := func(key string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(key, envPrefix)), "_", ".")
	}

	envOpt := env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			// Transform the key.
			k = keyTransformer(k)

			// Transform the value into slices, if they contain spaces.
			// Eg: MYVAR_TAGS="foo bar baz" -> tags: ["foo", "bar", "baz"]
			// This is to demonstrate that string values can be transformed to any type
			// where necessary.
			if strings.Contains(v, " ") {
				return k, strings.Split(v, " ")
			}

			return k, v
		},
	}

	if err := config.Load(file.Provider(".env"), dotenv.ParserEnv(envPrefix, ".", keyTransformer)); err != nil {
		log.Fatalln(err)
	}

	if err := config.Load(env.Provider(".", envOpt), nil); err != nil {
		log.Fatalln(err)
	}
}
