package bot

import (
	"dutybot/internal/config"
	db "dutybot/internal/database"
	"dutybot/internal/tasks"
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

func scheduleAnnounceDutyTask(bot *tgbot.BotAPI) {
	// Use this wrapper as NewTask accepts functions
	// with signature func () error
	announce := func() {
		announceDutyTask(bot)
	}
	_, err := tasks.AddTask(config.Cfg.DutyAnnounceSchedule, announce)
	if err != nil {
		log.Fatal("Unable to schedule task: ", err)
	}
}

func scheduleFreeSlotsTask(bot *tgbot.BotAPI) {
	checkFreeSlots := func() {
		warnAboutFreeSlots(bot)
	}
	_, err := tasks.AddTask(config.Cfg.FreeSlotsWarnSchedule, checkFreeSlots)
	if err != nil {
		log.Fatal("Unable to schedule task: ", err)
	}
}

func initBot() (bot *tgbot.BotAPI) {
	var err error
	tasks.InitScheduler()

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
	scheduleAnnounceDutyTask(bot)
	scheduleFreeSlotsTask(bot)
	tasks.Start()
	return
}
