package bot

import (
	"dutybot/internal/logger"
	"fmt"

	db "dutybot/internal/database"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

func announceDutyTask(bot *tgbot.BotAPI) {
	msgFormat := "@%s is on duty today"
	logger.Log.Debug().Msg("Start duty announcing")
	ass, err := db.GetAllTodaysOperators()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("announceDutyTask job failed to get operators")
		return
	}

	for _, as := range ass {
		logger.Log.Debug().Msgf("Sending %+v\n", as)
		sendMessage(
			bot,
			as.ChatID,
			fmt.Sprintf(msgFormat, as.Operator.UserName),
			NoParseMode,
		)
	}
}

func warnAboutFreeSlots(bot *tgbot.BotAPI) {
	logger.Log.Debug().Msg("Start freeslots announcing")

	chats, err := db.GetAllChats()
	if err != nil {
		logger.Log.Error().
			Err(err).
			Msg("warnAboutFreeSlots job failed to get all chat IDs")
		return
	}

	for _, chatID := range chats {
		outputSlots, err := getFreeSlotsTable(chatID, DefaultFreeSlotWeeks)
		if err != nil {
			logger.Log.Error().
				Err(err).
				Msg("warnAboutFreeSlots job failed to tabulate free slots")
			return
		}

		if outputSlots == "" {
			continue
		}

		sendMessage(
			bot,
			chatID,
			fmt.Sprintf("Free slots still available!\n%s\n", outputSlots),
			NoParseMode,
		)
	}
}
