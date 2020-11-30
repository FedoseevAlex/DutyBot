package bot

import (
	"dutybot/internal/calendar"
	db "dutybot/internal/database"
	"dutybot/internal/utils"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	MarkdownParseMode   = "MarkdownV2"
	FreeslotsThreshold  = 10
	DefaultFreslotWeeks = 1
	NoParseMode         = ""
)

var handlers map[string]func(*tgbot.BotAPI, *tgbot.Message)

func initHandlers() {
	handlers = make(map[string]func(*tgbot.BotAPI, *tgbot.Message))
	handlers["help"] = help
	handlers["assign"] = assign
	handlers["show"] = show
	handlers["operator"] = operator
	handlers["freeslots"] = freeSlots
}

func processCommands(bot *tgbot.BotAPI, command *tgbot.Message) {
	handler, ok := handlers[command.Command()]
	if !ok {
		msg := tgbot.NewMessage(command.Chat.ID, "Unknown command. Try /help")
		_, err := bot.Send(msg)
		if err != nil {
			log.Print(err)
		}
		return
	}
	handler(bot, command)
}

func help(bot *tgbot.BotAPI, msg *tgbot.Message) {
	helpString := `Usage:
/help - look at this message again
/operator - tag current duty
/show [weeks (default=2)] - show duty schedule for some weeks ahead
/assign [date] - assign yourself for duty. Date should be in format DD-MM-YYYY
/freeslots [weeks default=1] - show free duty slots

Found a bug? Want some features?
Feel free to make an issue:
https://github.com/FedoseevAlex/DutyBot/issues
`
	answer := tgbot.NewMessage(msg.Chat.ID, helpString)
	_, err := bot.Send(answer)
	if err != nil {
		log.Print(err)
	}
}

func operator(bot *tgbot.BotAPI, msg *tgbot.Message) {
	as, err := db.GetTodaysAssignment(msg.Chat.ID)
	if err != nil {
		log.Print(err)
		reply := tgbot.NewMessage(msg.Chat.ID, "Couldn't fetch today's duty.")
		_, err := bot.Send(reply)
		if err != nil {
			log.Print(err)
		}
		return
	}

	operator := fmt.Sprintf("@%s", as.Operator.UserName)
	reply := tgbot.NewMessage(msg.Chat.ID, operator)
	_, err = bot.Send(reply)
	if err != nil {
		log.Print(err)
	}
}

// This function assumes that date
// is ordered like DD MM YYYY
func parseTime(probablyTime string) (t time.Time, err error) {
	r := regexp.MustCompile("([0-9]{1,2}).*?([0-9]{1,2}).*?([0-9]{4})")
	if !r.MatchString(probablyTime) {
		err = fmt.Errorf("'%s' is not look like DD MM YYYY", probablyTime)
		return
	}
	parts := r.FindAllStringSubmatch(probablyTime, 1)[0]
	date := fmt.Sprintf("%02s", parts[1])
	month := fmt.Sprintf("%02s", parts[2])
	year := parts[3]
	t, err = time.Parse(utils.DateFormat, fmt.Sprintf("%s-%s-%s", year, month, date))
	return
}

func sendMessage(bot *tgbot.BotAPI, chatID int64, message string, parseMode string) {
	msg := tgbot.NewMessage(
		chatID,
		message,
	)

	if parseMode != NoParseMode {
		msg.ParseMode = parseMode
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Print(err)
	}
}

func assign(bot *tgbot.BotAPI, msg *tgbot.Message) {
	dutydate, err := parseTime(msg.CommandArguments())
	if err != nil {
		log.Print(err)
		sendMessage(bot, msg.Chat.ID, "Something wrong with date", NoParseMode)
		return
	}

	if calendar.IsHoliday(dutydate) {
		answer := fmt.Sprintf(
			"`%s` is a holiday. No duty on holidays",
			dutydate.Format(utils.DateFormat),
		)
		sendMessage(bot, msg.Chat.ID, answer, MarkdownParseMode)
		return
	}

	if utils.GetToday().After(dutydate) {
		sendMessage(bot, msg.Chat.ID, "Assignment is possible only for a future date", NoParseMode)
		return
	}

	as, err := db.GetAssignmentByDate(msg.Chat.ID, dutydate)
	if err != nil {
		log.Print(err)
	}
	if as != nil {
		sendMessage(
			bot,
			msg.Chat.ID,
			fmt.Sprintf("This day already taken by `%s`", as.Operator.UserName),
			NoParseMode,
		)
		return
	}

	op := &db.Operator{
		UserName:  msg.From.UserName,
		FirstName: msg.From.FirstName,
		LastName:  msg.From.LastName,
	}
	err = op.GetByUserName()
	if err != nil {
		err = op.Insert()
		if err != nil {
			log.Print(err)
			return
		}
	}
	a := &db.Assignment{ChatID: msg.Chat.ID, DutyDate: dutydate, Operator: op}
	log.Printf("New assignment: %+v", a)
	err = a.Insert()
	if err != nil {
		log.Print(err)
		return
	}
	show(bot, msg)
}

func checkWeeks(weekArgument string) (int, error) {
	var weeks int

	if weekArgument == "" {
		return DefaultFreslotWeeks, nil
	}

	weeks, err := strconv.Atoi(weekArgument)
	if err != nil {
		return 0, fmt.Errorf("seems that %s is not a number", weekArgument)
	}

	if weeks > FreeslotsThreshold {
		return 0, fmt.Errorf("in the grim darkness of the far future there is only war")
	}

	if weeks <= 0 {
		return 0, fmt.Errorf("some serious QA here")
	}
	return weeks, nil
}

func freeSlots(bot *tgbot.BotAPI, msg *tgbot.Message) {
	weeks, err := checkWeeks(msg.CommandArguments())
	if err != nil {
		sendMessage(
			bot,
			msg.Chat.ID,
			err.Error(),
			NoParseMode,
		)
		return
	}

	slots, err := db.GetFreeSlots(weeks, msg.Chat.ID)
	if err != nil {
		log.Print(err)
	}

	freeSlots := utils.NewPrettyTable()
	for _, slot := range slots {
		freeSlots.AddRow([]string{slot.Format(utils.HumanDateFormat)})
		if err != nil {
			log.Print(err)
			return
		}
	}
	table, err := freeSlots.String()
	if err != nil {
		log.Print(err)
		return
	}

	sendMessage(bot, msg.Chat.ID, fmt.Sprintf("```\n%s\n```", table), MarkdownParseMode)
}

func show(bot *tgbot.BotAPI, msg *tgbot.Message) {
	weeks, err := checkWeeks(msg.CommandArguments())
	if err != nil {
		sendMessage(
			bot,
			msg.Chat.ID,
			err.Error(),
			NoParseMode,
		)
		return
	}

	assignments, err := db.GetAssignmentSchedule(weeks, msg.Chat.ID)
	if err != nil {
		sendMessage(bot, msg.Chat.ID, "Couldn't get assignments.", NoParseMode)
		return
	}

	schedule := utils.NewPrettyTable()

	for _, ass := range assignments {
		dutyDate := ass.DutyDate.Format(utils.HumanDateFormat)
		schedule.AddRow([]string{ass.Operator.UserName, dutyDate})
	}
	table, err := schedule.String()
	if err != nil {
		sendMessage(
			bot,
			msg.Chat.ID,
			fmt.Sprintf("Tabulation error: %s", err.Error()),
			NoParseMode,
		)
		return
	}

	sendMessage(bot, msg.Chat.ID, fmt.Sprintf("```\n%s\n```", table), MarkdownParseMode)
}
