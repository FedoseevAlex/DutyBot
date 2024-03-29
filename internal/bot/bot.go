package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"

	"github.com/FedoseevAlex/DutyBot/internal/config"
	"github.com/FedoseevAlex/DutyBot/internal/database/assignment"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/tasks"
	"github.com/FedoseevAlex/DutyBot/internal/utils"
)

var bot *tgbot.BotAPI

func processUpdate(update tgbot.Update) error {
	var command Command
	switch {
	case update.Message != nil:
		if !update.Message.IsCommand() {
			// Avoid non command messages (e.g. reply)
			return nil
		}
		command = Command{
			Action:    update.Message.Command(),
			Arguments: update.Message.CommandArguments(),
			Operator:  update.SentFrom().UserName,
			ChatID:    update.FromChat().ID,
		}
	case update.EditedMessage != nil:
		command = Command{
			Action:    update.EditedMessage.Command(),
			Arguments: update.EditedMessage.CommandArguments(),
			Operator:  update.SentFrom().UserName,
			ChatID:    update.FromChat().ID,
		}
	case update.CallbackQuery != nil:
		action, arguments, _ := strings.Cut(update.CallbackData(), " ")
		command = Command{
			Action:     action,
			Arguments:  arguments,
			Operator:   update.SentFrom().UserName,
			ChatID:     update.FromChat().ID,
			KeyboardID: update.CallbackQuery.Message.MessageID,
		}
		return processCallback(command)
	default:
		logger.Log.Info().Str("update", fmt.Sprintf("%+v", update)).Send()
		return nil
	}

	err := processCommands(command)
	if err != nil {
		return err
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
			Bool("hook_mode", viper.GetBool("HookMode")).
			Stack().
			Err(err).
			Send()
		return err
	}

	return nil
}

func StartBotHook() error {
	botURL := "/" + bot.Token
	http.HandleFunc(botURL, handleRequests)

	webhookConfig, err := tgbot.NewWebhookWithCert(
		viper.GetString("ExternalAddress")+botURL,
		tgbot.FilePath(viper.GetString("CertPath")),
	)
	if err != nil {
		return fmt.Errorf("create webhook msg: %w", err)
	}

	_, err = bot.Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("start boot hook: set webhook: %w", err)
	}
	logger.Log.Info().Msg("Webhook was set!")

	err = http.ListenAndServeTLS(
		viper.GetString("ListenAddress"),
		viper.GetString("CertPath"),
		viper.GetString("KeyPath"),
		nil,
	)
	if err != nil {
		return fmt.Errorf("start boot hook: listen and serve tls: %w", err)
	}
	logger.Log.Debug().Msg("Server shutdown")
	return nil
}

func StartBotLongPoll() error {
	updateConfig := tgbot.UpdateConfig{}
	updateConfig.Timeout = 5

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		err := processUpdate(update)
		if err != nil {
			logger.Log.Error().
				Stack().
				Err(err).
				Msg("Unable to process update")
		}
	}

	return nil
}

func scheduleAnnounceDutyTask() {
	announce := func() {
		announceDutyTask()
	}
	_, err := tasks.AddTask(viper.GetString("DutyAnnounceSchedule"), announce)
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Unable to schedule task")
	}
}

func scheduleFreeSlotsTask() {
	checkFreeSlots := func() {
		warnAboutFreeSlots()
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
	scheduleAnnounceDutyTask()
	scheduleFreeSlotsTask()
	tasks.Start()
	logger.Log.Debug().Msg("Starting dutybot...")
	return nil
}
