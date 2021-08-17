package bot

import (
	"dutybot/internal/config"
	"dutybot/internal/logger"
	"dutybot/internal/tasks"
	"dutybot/internal/utils"
	"encoding/json"
	"io/ioutil"
	"net/http"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"

	db "dutybot/internal/database"
)

var bot *tgbot.BotAPI

func processUpdate(update tgbot.Update) error {
	var msg *tgbot.Message
	if update.EditedMessage != nil {
		msg = update.EditedMessage
	} else {
		msg = update.Message
	}

	if msg.IsCommand() {
		err := processCommands(bot, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleRequests(_ http.ResponseWriter, req *http.Request) {
	defer utils.Close(req.Body)

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Failed to read body contents")
		return
	}

	var update tgbot.Update

	err = json.Unmarshal(bodyBytes, &update)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to unmarshal json update from telegram")
		return
	}

	err = processUpdate(update)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to process update from telegram")
	}
}

func StartBot() error {
	err := initBot()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to init bot")
		return err
	}

	if false {
		err = StartBotHook()
	} else {
		err = StartBotLongPoll()
	}

	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("")
		return err
	}

	return nil
}

func StartBotHook() error {
	http.HandleFunc("/"+bot.Token, handleRequests)

	err := http.ListenAndServeTLS(
		config.Cfg.ListenAddr,
		config.Cfg.CertPath,
		config.Cfg.KeyPath,
		nil,
	)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to start https server")
		return err
	}
	logger.Log.Debug().Msg("Server shutdown")
	return nil
}

func StartBotLongPoll() error {
	updateConfig := tgbot.UpdateConfig{}
	updateConfig.Timeout = 5

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to start long poll bot")
		return err
	}

	for update := range updates {
		err = processUpdate(update)
		if err != nil {
			logger.Log.Error().
				Stack().
				Err(err).
				Msg("Unable to process update")
		}
	}

	return nil
}

func scheduleAnnounceDutyTask(bot *tgbot.BotAPI) {
	// Use this wrapper as NewTask accepts functions
	// with signature func () error
	announce := func() {
		announceDutyTask(bot)
	}
	_, err := tasks.AddTask(config.Cfg.DutyAnnounceSchedule, announce)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to schedule task")
	}
}

func scheduleFreeSlotsTask(bot *tgbot.BotAPI) {
	checkFreeSlots := func() {
		warnAboutFreeSlots(bot)
	}
	_, err := tasks.AddTask(config.Cfg.FreeSlotsWarnSchedule, checkFreeSlots)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to schedule task")
	}
}

func initBot() error {
	config.ReadConfig()
	logger.InitLogger(config.Cfg.LogPath)
	tasks.InitScheduler()
	initHandlers()

	err := db.Init(config.Cfg.DBDriver, config.Cfg.DBConnectString)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return err
	}

	bot, err = tgbot.NewBotAPI(config.Cfg.BotToken)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return err
	}
	scheduleAnnounceDutyTask(bot)
	scheduleFreeSlotsTask(bot)
	tasks.Start()
	logger.Log.Debug().Msg("Starting dutybot...")
	return nil
}
