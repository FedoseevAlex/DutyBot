package bot

import (
	"dutybot/internal/config"
	db "dutybot/internal/database"
	"dutybot/internal/tasks"
	"fmt"
	"log"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

func StartBot() {
	bot := initBot()

	u := tgbot.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}
	initHandlers()

	for update := range updates {
		var msg *tgbot.Message

		if update.EditedMessage != nil {
			msg = update.EditedMessage
		} else {
			msg = update.Message
		}

		if msg.IsCommand() {
			go processCommands(bot, msg)
		}
	}
}

func announceADuty(bot *tgbot.BotAPI) error {
	msgFormat := "@%s is on duty today"
	ass, err := db.GetAllTodaysOperators()
	if err != nil {
		log.Fatal(err)
	}
	for _, as := range ass {
		fmt.Printf("Sending %+v\n", as)
		msg := tgbot.NewMessage(as.ChatID, fmt.Sprintf(msgFormat, as.Operator.UserName))
		msg.DisableNotification = true

		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func startAnnouncing(bot *tgbot.BotAPI) {
	// Use this wrapper as NewTask accepts functions
	// with signature func () error
	announce := func() error {
		if err := announceADuty(bot); err != nil {
			return err
		}
		return nil
	}
	tasks.NewTask(announce, config.Cfg.DutyShift, config.Cfg.DutyStartAt).Start()
}

func initBot() (bot *tgbot.BotAPI) {
	var err error

	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
	config.ReadConfig()

	err = db.Init(config.Cfg.DBDriver, config.Cfg.DBConnectString)
	if err != nil {
		log.Fatal(err)
	}

	bot, err = tgbot.NewBotAPI(config.Cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}
	startAnnouncing(bot)
	return
}
