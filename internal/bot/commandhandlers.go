package bot

import (
	"dutybot/internal/calendar"
	"dutybot/internal/logger"
	"dutybot/internal/utils"
	"fmt"
	"regexp"
	"strconv"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"

	db "dutybot/internal/database"
)

const (
	MarkdownParseMode    = "MarkdownV2"
	FreeslotsThreshold   = 10
	DefaultFreeSlotWeeks = 2
	DefaultShowWeeks     = 2
	NoParseMode          = ""
)

var handlers map[string]func(*tgbot.BotAPI, *tgbot.Message) error

func initHandlers() {
	handlers = make(map[string]func(*tgbot.BotAPI, *tgbot.Message) error)
	handlers["help"] = help
	handlers["assign"] = assign
	handlers["show"] = show
	handlers["operator"] = operator
	handlers["freeslots"] = freeSlots
	handlers["reset"] = resetAssign
}

func processCommands(bot *tgbot.BotAPI, msg *tgbot.Message) error {
	handler, ok := handlers[msg.Command()]
	if !ok {
		answer := tgbot.NewMessage(msg.Chat.ID, "Unknown command. Try /help")
		_, err := bot.Send(answer)
		if err != nil {
			logger.Log.Error().Err(err).Send()
		}
		return errors.New("Unknown command")
	}
	err := handler(bot, msg)
	return err
}

func help(bot *tgbot.BotAPI, msg *tgbot.Message) error {
	helpString := `Usage:
/help - look at this message again
/operator - tag current duty
/show [weeks (default=2)] - show duty schedule for some weeks ahead
/assign date - assign yourself for duty. Date should be in format DD-MM-YYYY
/reset [date default=Today] - clear specified date from assignments
/freeslots [weeks default=1] - show free duty slots

Found a bug? Want some features?
Feel free to make an issue:
https://github.com/FedoseevAlex/DutyBot/issues
`
	answer := tgbot.NewMessage(msg.Chat.ID, helpString)
	_, err := bot.Send(answer)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}
	return nil
}

func operator(bot *tgbot.BotAPI, msg *tgbot.Message) error {
	as, err := db.GetTodaysAssignment(msg.Chat.ID)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		reply := tgbot.NewMessage(msg.Chat.ID, "Couldn't fetch today's duty.")
		_, err := bot.Send(reply)
		if err != nil {
			logger.Log.Error().Err(err).Send()
		}
		return err
	}

	operator := fmt.Sprintf("@%s", as.Operator.UserName)
	reply := tgbot.NewMessage(msg.Chat.ID, operator)
	_, err = bot.Send(reply)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}
	return nil
}

// This function assumes that date
// is ordered like DD MM YYYY
func parseTime(probablyTime string) (*time.Time, error) {
	r := regexp.MustCompile("([0-9]{1,2}).*?([0-9]{1,2}).*?([0-9]{4})")
	if !r.MatchString(probablyTime) {
		return nil, fmt.Errorf("'%s' is not look like DD MM YYYY", probablyTime)
	}
	parts := r.FindAllStringSubmatch(probablyTime, 1)[0]
	date := fmt.Sprintf("%02s", parts[1])
	month := fmt.Sprintf("%02s", parts[2])
	year := parts[3]
	t, err := time.Parse(utils.DateFormat, fmt.Sprintf("%s-%s-%s", year, month, date))
	return &t, err
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
		logger.Log.Error().Err(err).Send()
	}
}

func checkDate(possibleDate string) (*time.Time, error) {
	dutydate, err := parseTime(possibleDate)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return nil, err
	}

	if calendar.IsHoliday(dutydate) {
		answer := fmt.Errorf(
			"'%s' is a holiday. No duty on holidays",
			dutydate.Format(utils.DateFormat),
		)
		return nil, answer
	}

	if utils.GetToday().After(*dutydate) {
		return nil, fmt.Errorf("assignment is possible only for a future date")
	}

	return dutydate, nil
}

func assign(bot *tgbot.BotAPI, msg *tgbot.Message) error {
	dutydate, err := checkDate(msg.CommandArguments())
	if err != nil {
		logger.Log.Error().Err(err).Send()
		sendMessage(bot, msg.Chat.ID, err.Error(), NoParseMode)
		return err
	}

	as, err := db.GetAssignmentByDate(msg.Chat.ID, dutydate)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}
	if as != nil {
		sendMessage(
			bot,
			msg.Chat.ID,
			fmt.Sprintf("This day already taken by `%s`", as.Operator.UserName),
			NoParseMode,
		)
		return nil
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
			logger.Log.Error().Err(err).Send()
			return err
		}
	}

	a := &db.Assignment{ChatID: msg.Chat.ID, DutyDate: *dutydate, Operator: op}
	logger.Log.Printf("new assignment: %+v", a)
	err = a.Insert()
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}

	assignments, err := getAssignmentsTable(msg.Chat.ID, DefaultShowWeeks)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		sendMessage(bot, msg.Chat.ID, err.Error(), NoParseMode)
		return err
	}
	sendMessage(
		bot,
		msg.Chat.ID,
		fmt.Sprintf("```\n%s\n```", assignments),
		MarkdownParseMode,
	)
	return nil
}

func resetAssign(bot *tgbot.BotAPI, msg *tgbot.Message) error {
	var err error
	dutydate := utils.GetToday()

	if msg.CommandArguments() != "" {
		dutydate, err = checkDate(msg.CommandArguments())
		if err != nil {
			logger.Log.Error().Err(err).Send()
			sendMessage(bot, msg.Chat.ID, err.Error(), NoParseMode)
			return err
		}
	}

	as, err := db.GetAssignmentByDate(msg.Chat.ID, dutydate)
	if err != nil {
		sendMessage(
			bot,
			msg.Chat.ID,
			fmt.Sprintf(
				"no assignments for %s",
				dutydate.Format(utils.AssignDateFormat),
			),
			NoParseMode,
		)
		return err
	}

	err = as.Delete()
	if err != nil {
		logger.Log.Error().Err(err).Send()
		sendMessage(
			bot,
			msg.Chat.ID,
			"failed to reset assignments",
			NoParseMode,
		)
		return err
	}

	sendMessage(
		bot,
		msg.Chat.ID,
		fmt.Sprintf(
			"@%s is unassigned from %s",
			as.Operator.UserName,
			dutydate.Format(utils.AssignDateFormat),
		),
		NoParseMode,
	)
	return nil
}

func checkWeeks(weekArgument string) (int, error) {
	var weeks int

	if weekArgument == "" {
		return DefaultFreeSlotWeeks, nil
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

func freeSlots(bot *tgbot.BotAPI, msg *tgbot.Message) error {
	weeks, err := checkWeeks(msg.CommandArguments())
	if err != nil {
		sendMessage(
			bot,
			msg.Chat.ID,
			err.Error(),
			NoParseMode,
		)
		return err
	}

	table, err := getFreeSlotsTable(msg.Chat.ID, weeks)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}

	sendMessage(bot, msg.Chat.ID, table, NoParseMode)
	return nil
}

func getFreeSlotsTable(chatID int64, weeks int) (string, error) {
	slots, err := db.GetFreeSlots(weeks, chatID)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return "", err
	}

	if len(slots) == 0 {
		return "", nil
	}

	freeSlots := utils.NewPrettyTable()
	for _, slot := range slots {
		freeSlots.AddRow([]string{"/assign", slot.Format(utils.AssignDateFormat)})
	}
	table, err := freeSlots.String()
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return "", err
	}
	return table, nil
}

func show(bot *tgbot.BotAPI, msg *tgbot.Message) error {
	weeks, err := checkWeeks(msg.CommandArguments())
	if err != nil {
		logger.Log.Error().Err(err).Send()
		sendMessage(
			bot,
			msg.Chat.ID,
			err.Error(),
			NoParseMode,
		)
		return err
	}

	table, err := getAssignmentsTable(msg.Chat.ID, weeks)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		sendMessage(
			bot,
			msg.Chat.ID,
			fmt.Sprintf("Tabulation error: %s", err.Error()),
			NoParseMode,
		)
		return err
	}

	if table == "" {
		table = "Nothing to show"
	}
	sendMessage(bot, msg.Chat.ID, fmt.Sprintf("```\n%s\n```", table), MarkdownParseMode)
	return nil
}

func getAssignmentsTable(chatID int64, weeks int) (string, error) {
	assignments, err := db.GetAssignmentSchedule(weeks, chatID)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return "", fmt.Errorf("couldn't get assignments")
	}

	schedule := utils.NewPrettyTable()

	for _, ass := range assignments {
		dutyDate := ass.DutyDate.Format(utils.HumanDateFormat)
		schedule.AddRow([]string{ass.Operator.UserName, dutyDate})
	}
	table, err := schedule.String()
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return "", err
	}
	return table, nil
}
