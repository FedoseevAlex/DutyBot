package bot

import (
	db "dutybot/internal/database"
	"fmt"
	"log"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

func announceDutyTask(bot *tgbot.BotAPI) {
	msgFormat := "@%s is on duty today"
	log.Println("Start duty announcing")
	ass, err := db.GetAllTodaysOperators()
	if err != nil {
		log.Println(
			"announceDutyTask job failed to get operators: ",
			err,
		)
		return
	}

	for _, as := range ass {
		log.Printf("Sending %+v\n", as)
		sendMessage(
			bot,
			as.ChatID,
			fmt.Sprintf(msgFormat, as.Operator.UserName),
			NoParseMode,
		)
	}
}

func warnAboutFreeSlots(bot *tgbot.BotAPI) {
	log.Println("Start freeslots announcing")

	chats, err := db.GetAllChats()
	if err != nil {
		log.Println(
			"warnAboutFreeSlots job failed to get all chat IDs: ",
			err,
		)
		return
	}

	for _, chatID := range chats {
		outputSlots, err := getFreeSlotsTable(chatID, DefaultFreeSlotWeeks)
		if err != nil {
			log.Println(
				"warnAboutFreeSlots job failed to tabulate free slots: ",
				err,
			)
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
