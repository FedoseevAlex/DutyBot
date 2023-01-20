package bot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"

	"github.com/FedoseevAlex/DutyBot/internal/config"
	"github.com/FedoseevAlex/DutyBot/internal/database/assignment"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/tasks"
	"github.com/FedoseevAlex/DutyBot/internal/utils"
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

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Stack().
			Msg("Failed to read body contents")
		return
	}

	var update tgbot.Update

	err = json.Unmarshal(bodyBytes, &update)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Stack().
			Msg("Unable to unmarshal json update from telegram")
		return
	}

	err = processUpdate(update)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Stack().
			Msg("Unable to process update from telegram")
	}
}

func StartBot() error {
	logger.InitLogger()
	err := initBot()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Stack().
			Msg("Unable to init bot")
		return err
	}

	if viper.GetBool("HookMode") {
		err = StartBotHook()
	} else {
		err = StartBotLongPoll()
	}

	if err != nil {
		logger.Log.Error().
			Err(err).
			Stack().
			Send()
		return err
	}

	return nil
}

func StartBotHook() error {
	http.HandleFunc("/"+bot.Token, handleRequests)

	err := http.ListenAndServeTLS(
		viper.GetString("ListenAddr"),
		viper.GetString("CertPath"),
		viper.GetString("KeyPath"),
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
	_, err := tasks.AddTask(viper.GetString("DutyAnnounceSchedule"), announce)
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
	_, err := tasks.AddTask(viper.GetString("FreeSlotsWarnSchedule"), checkFreeSlots)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Stack().
			Msg("Unable to schedule task")
	}
}

func initBot() error {
	if err := config.ReadConfig(); err != nil {
		logger.Log.Error().
			Err(err).
			Stack().
			Msg("Read config failed")
		return err
	}

	tasks.InitScheduler()
	initHandlers()

	_, err := assignment.InitAssignmentRepo(context.Background(), viper.GetString("DBConnectString"))
	if err != nil {
		logger.Log.Error().
			Stack().
			Err(err).
			Msg("failed go init assignment repo")
		return err
	}

	bot, err = tgbot.NewBotAPI(viper.GetString("BotToken"))
	if err != nil {
		logger.Log.Error().
			Stack().
			Err(err).
			Msg("failed to create bot")
		return err
	}
	scheduleAnnounceDutyTask(bot)
	scheduleFreeSlotsTask(bot)
	tasks.Start()
	logger.Log.Debug().Msg("Starting dutybot...")
	return nil
}
