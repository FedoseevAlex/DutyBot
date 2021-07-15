package bot

import (
	"dutybot/internal/config"
	"dutybot/internal/tasks"
	"dutybot/internal/utils"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"

	db "dutybot/internal/database"
)

var bot *tgbot.BotAPI

const logFilePermissions = 0o666

func handleRequests(_ http.ResponseWriter, req *http.Request) {
	defer utils.Close(req.Body)

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("Failed to read body contents: ", err)
		return
	}

	var update tgbot.Update
	var msg *tgbot.Message

	err = json.Unmarshal(bodyBytes, &update)
	if err != nil {
		log.Println("Unable to unmarshal json update from telegram: ", err)
		return
	}

	if update.EditedMessage != nil {
		msg = update.EditedMessage
	} else {
		msg = update.Message
	}

	if msg.IsCommand() {
		processCommands(bot, msg)
	}
}

func StartBotHook() {
	config.ReadConfig(config.DefaultConfigPath)

	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)

	logfile, err := os.OpenFile(config.Cfg.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC, logFilePermissions)
	if err != nil {
		log.Println("Unable to open a log file: ", config.Cfg.LogPath)
	} else {
		defer utils.Close(logfile)
		log.SetOutput(logfile)
	}

	err = initBot()
	if err != nil {
		log.Println("Unable to init bot: ", err)
		return
	}

	http.HandleFunc("/"+bot.Token, handleRequests)

	err = http.ListenAndServeTLS(
		config.Cfg.ListenAddr,
		config.Cfg.CertPath,
		config.Cfg.KeyPath,
		nil,
	)
	if err != nil {
		log.Println("Unable to start https server: ", err)
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

func initBot() error {
	tasks.InitScheduler()
	initHandlers()
	err := db.Init(config.Cfg.DBDriver, config.Cfg.DBConnectString)
	if err != nil {
		log.Print(err)
		return err
	}

	bot, err = tgbot.NewBotAPI(config.Cfg.BotToken)
	if err != nil {
		log.Print(err)
		return err
	}
	scheduleAnnounceDutyTask(bot)
	scheduleFreeSlotsTask(bot)
	tasks.Start()
	log.Println("Starting dutybot...")
	return nil
}
