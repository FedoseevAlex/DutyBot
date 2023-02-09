package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/FedoseevAlex/DutyBot/internal/database/assignment"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/utils"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var chatKeyboards = map[int64]int{}

func makeCalendarButtons(schedule []assignment.Assignment) tgbot.InlineKeyboardMarkup {
	if len(schedule) == 0 {
		return tgbot.InlineKeyboardMarkup{}
	}
	keyboard := make([][]tgbot.InlineKeyboardButton, 0)

	for _, assignment := range schedule {
		buttons := make([]tgbot.InlineKeyboardButton, 0, 2)
		buttons = append(buttons, tgbot.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s", assignment.At.Format("02 Jan, Mon"), assignment.Operator),
			fmt.Sprintf("assign %s", assignment.At.Format(utils.AssignDateFormat))),
		)
		if assignment.Operator != "" {
			buttons = append(buttons, tgbot.NewInlineKeyboardButtonData(
				"reset",
				fmt.Sprintf("reset %s", assignment.At.Format(utils.AssignDateFormat))),
			)
		}
		keyboard = append(keyboard, tgbot.NewInlineKeyboardRow(buttons...))
	}

	scheduleStart := schedule[0].At
	manageRow := tgbot.NewInlineKeyboardRow(
		tgbot.NewInlineKeyboardButtonData(
			"<",
			fmt.Sprintf("showWeek %s", scheduleStart.Add(-utils.WeekDuration).Format(utils.AssignDateFormat)),
		),
		tgbot.NewInlineKeyboardButtonData(
			">",
			fmt.Sprintf("showWeek %s", scheduleStart.Add(utils.WeekDuration).Format(utils.AssignDateFormat)),
		),
	)
	keyboard = append(keyboard, manageRow)

	return tgbot.NewInlineKeyboardMarkup(keyboard...)
}

func changeWeekOnKeyboard(chatID int64, keyboardID int, from time.Time) {
	from = utils.GetStartOfWeek(from)
	schedule, err := assignment.AssignmentRepo.GetSchedule(
		context.Background(),
		from,
		from.Add(utils.WeekDuration),
		chatID)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
	}
	keyboard := makeCalendarButtons(schedule)
	edit := tgbot.NewEditMessageReplyMarkup(chatID, keyboardID, keyboard)

	_, err = bot.Send(edit)
	if err != nil {
		logger.Log.Warn().Stack().Err(err).Send()
	}
}

func refreshKeyboard(chatID int64, keyboardID int, from time.Time) {
	from = utils.GetStartOfWeek(from)
	schedule, err := assignment.AssignmentRepo.GetSchedule(
		context.Background(),
		from,
		from.Add(utils.WeekDuration),
		chatID)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
	}
	keyboard := makeCalendarButtons(schedule)
	edit := tgbot.NewEditMessageReplyMarkup(chatID, keyboardID, keyboard)

	_, err = bot.Send(edit)
	if err != nil {
		logger.Log.Warn().Stack().Err(err).Send()
	}
}

func sendKeyboard(chatID int64, schedule []assignment.Assignment) {
	keyboard := makeCalendarButtons(schedule)

	answer := tgbot.NewMessage(chatID, "It's time to choose")
	answer.ReplyMarkup = keyboard

	response, err := bot.Send(answer)
	if err != nil {
		logger.Log.Error().Err(err).Send()
	}
	logger.Log.Debug().Str("response", fmt.Sprintf("%+v", response)).Send()

	removeOldKeyboard(chatID)
	chatKeyboards[chatID] = response.MessageID
}

func removeOldKeyboard(chatID int64) {
	keyboardID, ok := chatKeyboards[chatID]
	if !ok {
		return
	}
	rm := tgbot.NewDeleteMessage(chatID, keyboardID)
	_, err := bot.Request(rm)
	if err != nil {
		logger.Log.Warn().Stack().Err(err).Send()
	}
}
