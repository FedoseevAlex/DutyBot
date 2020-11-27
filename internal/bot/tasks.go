package bot

import (
	db "dutybot/internal/database"
	"dutybot/internal/utils"
	"fmt"
	"log"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	// Default number of weeks to check for free slots
	freeSlotsWeeks int = 1
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
		fmt.Printf("Sending %+v\n", as)
		msg := tgbot.NewMessage(as.ChatID, fmt.Sprintf(msgFormat, as.Operator.UserName))
		msg.DisableNotification = true

		_, err := bot.Send(msg)
		if err != nil {
			log.Println(
				"announceDutyTask job failed to send message to telegram: ",
				err,
			)
			return
		}
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
		slots, err := db.GetFreeSlots(freeSlotsWeeks, chatID)
		if err != nil {
			log.Println(
				"warnAboutFreeSlots job failed to get free slots: ",
				err,
			)
			return
		}

		if len(slots) == 0 {
			continue
		}

		freeslots := utils.NewPrettyTable()
		for _, slot := range slots {
			freeslots.AddRow([]string{slot.Format(utils.HumanDateFormat)})
		}
		outputSlots, err := freeslots.String()
		if err != nil {
			log.Println(
				"warnAboutFreeSlots job failed to tabulate free slots: ",
				err,
			)
			return
		}

		msg := tgbot.NewMessage(
			chatID,
			fmt.Sprintf("Free slots still available! \n%s", outputSlots),
		)

		_, err = bot.Send(msg)
		if err != nil {
			log.Println(
				"warnAboutFreeSlots job failed to send message to telegram: ",
				err,
			)
			return
		}
	}
}
