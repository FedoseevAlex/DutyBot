package bot

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/FedoseevAlex/DutyBot/internal/calendar"
	"github.com/FedoseevAlex/DutyBot/internal/database/assignment"
	"github.com/FedoseevAlex/DutyBot/internal/logger"
	"github.com/FedoseevAlex/DutyBot/internal/utils"
)

const (
	MarkdownParseMode    = "MarkdownV2"
	FreeslotsThreshold   = 10
	DefaultFreeSlotWeeks = 2
	DefaultShowWeeks     = 2
	NoParseMode          = ""
	heHeProbability      = 0.1
)

type Command struct {
	Action     string
	Operator   string
	ChatID     int64
	Arguments  string
	KeyboardID int
}

type CommandResult struct {
	Error   error
	Message string
}

var handlers map[string]func(Command) error

func initHandlers() {
	handlers = map[string](func(Command) error){
		"help":      help,
		"assign":    assignAndPrint,
		"show":      show,
		"operator":  operator,
		"freeslots": freeSlots,
		"reset":     resetAssign,
		"buttons":   showButtons,
		"video":     reactToVideo,
	}
}

func showButtons(command Command) error {
	var err error

	from := utils.GetToday()
	if command.Arguments != "" {
		from, err = parseTime(command.Arguments)
		if err != nil {
			return err
		}
	}
	from = utils.GetStartOfWeek(from)

	schedule, err := assignment.AssignmentRepo.GetSchedule(
		context.Background(),
		from,
		from.Add(utils.WeekDuration),
		command.ChatID,
		true,
	)
	if err != nil {
		sendMessage(
			command.ChatID,
			fmt.Sprintf("Error getting assignments: %s", err),
			NoParseMode,
		)
		return err
	}

	sendKeyboard(command.ChatID, schedule)
	return nil
}

func processCallback(command Command) error {
	switch command.Action {
	case "assign":
		err := assign(command)
		if err != nil {
			return err
		}
		assignDate, _ := parseTime(command.Arguments)
		refreshKeyboard(command.ChatID, command.KeyboardID, assignDate)

	case "showWeek":
		from, err := parseTime(command.Arguments)
		if err != nil {
			return err
		}
		changeWeekOnKeyboard(command.ChatID, command.KeyboardID, from)

	case "reset":
		err := resetAssign(command)
		if err != nil {
			return err
		}
		date, _ := parseTime(command.Arguments)
		refreshKeyboard(command.ChatID, command.KeyboardID, date)
	}
	return nil
}

func processCommands(command Command) error {
	handler, ok := handlers[command.Action]
	if !ok {
		answer := tgbot.NewMessage(command.ChatID, "Unknown command. Try /help")
		_, err := bot.Send(answer)
		if err != nil {
			logger.Log.Error().Err(err).Send()
		}
		return errors.New("Unknown command")
	}
	err := handler(command)
	return err
}

func help(command Command) error {
	helpString := `Usage:
/help - look at this message again
/operator - tag current duty
/show [weeks (default=2)] - show duty schedule for some weeks ahead
/assign date - assign yourself for duty. Date should be in format DD-MM-YYYY
/reset [date default=Today] - clear specified date from assignments
/freeslots [weeks default=1] - show free duty slots
/buttons - show buttons for assignment

Found a bug? Want some features?
Feel free to make an issue:
https://github.com/FedoseevAlex/DutyBot/issues
`
	answer := tgbot.NewMessage(command.ChatID, helpString)
	_, err := bot.Send(answer)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}
	return nil
}

func reactToVideo(command Command) error {
	if rand.Float32() > heHeProbability {
		return nil
	}

	reply := tgbot.NewMessage(command.ChatID, "Hehehehehe")
	_, err := bot.Send(reply)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}
	return nil
}

func operator(command Command) error {
	as, err := assignment.AssignmentRepo.GetAssignmentByDate(
		context.Background(),
		utils.GetToday(),
		command.ChatID)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		reply := tgbot.NewMessage(command.ChatID, "Couldn't fetch today's duty.")
		_, err := bot.Send(reply)
		if err != nil {
			logger.Log.Error().Err(err).Send()
		}
		return err
	}
	if as.Operator == "" {
		reply := tgbot.NewMessage(command.ChatID, "No one is assigned for today")
		_, err := bot.Send(reply)
		if err != nil {
			logger.Log.Error().Err(err).Send()
		}
		return nil
	}

	operator := fmt.Sprintf("@%s", as.Operator)
	reply := tgbot.NewMessage(command.ChatID, operator)
	_, err = bot.Send(reply)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}
	return nil
}

// This function assumes that date
// is ordered like DD MM YYYY
func parseTime(probablyTime string) (time.Time, error) {
	r := regexp.MustCompile("([0-9]{1,2}).*?([0-9]{1,2}).*?([0-9]{4})")
	if !r.MatchString(probablyTime) {
		return time.Time{}, fmt.Errorf("'%s' is not look like DD MM YYYY", probablyTime)
	}
	parts := r.FindAllStringSubmatch(probablyTime, 1)[0]
	date := fmt.Sprintf("%02s", parts[1])
	month := fmt.Sprintf("%02s", parts[2])
	year := parts[3]
	t, err := time.Parse(utils.DateFormat, fmt.Sprintf("%s-%s-%s", year, month, date))
	return t, err
}

func sendMessage(chatID int64, message string, parseMode string) {
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

func checkDate(possibleDate string) (time.Time, error) {
	dutydate, err := parseTime(possibleDate)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return time.Time{}, err
	}

	if calendar.IsHoliday(dutydate) {
		answer := fmt.Errorf(
			"'%s' is a holiday. No duty on holidays",
			dutydate.Format(utils.DateFormat),
		)
		return time.Time{}, answer
	}

	if utils.GetToday().After(dutydate) {
		return time.Time{}, fmt.Errorf("assignment is possible only for a future date")
	}

	return dutydate, nil
}

func assignAndPrint(command Command) error {
	err := assign(command)
	if err != nil {
		return err
	}
	assignmentDate, _ := parseTime(command.Arguments)
	_, assignmentWeek := assignmentDate.ISOWeek()
	_, currentWeek := utils.GetToday().ISOWeek()
	weeks := assignmentWeek - currentWeek
	if weeks < DefaultShowWeeks {
		weeks = DefaultShowWeeks
	}

	assignments, err := getAssignmentsTable(command.ChatID, weeks)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		sendMessage(command.ChatID, err.Error(), NoParseMode)
		return err
	}
	sendMessage(
		command.ChatID,
		fmt.Sprintf("```\n%s\n```", assignments),
		MarkdownParseMode,
	)
	return nil
}

func assign(command Command) error {
	dutydate, err := checkDate(command.Arguments)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		sendMessage(command.ChatID, err.Error(), NoParseMode)
		return err
	}

	as, err := assignment.AssignmentRepo.GetAssignmentByDate(
		context.Background(),
		dutydate,
		command.ChatID)
	if err != nil {
		logger.Log.
			Error().
			Str("operation", "assign").
			Stack().
			Err(err).
			Send()
		return err
	}
	if as.Operator != "" {
		sendMessage(
			command.ChatID,
			fmt.Sprintf(
				"`%s` is taken by `%s` try `/reset %s`",
				as.At.Format(utils.AssignDateFormat),
				as.Operator,
				as.At.Format(utils.AssignDateFormat),
			),
			MarkdownParseMode,
		)
		return nil
	}

	a := assignment.Assignment{
		ChatID:    command.ChatID,
		At:        dutydate,
		Operator:  command.Operator,
		ID:        uuid.New(),
		CreatedAt: utils.GetToday(),
	}
	logger.Log.Printf("new assignment: %+v", a)
	err = assignment.AssignmentRepo.AddAssignment(
		context.Background(),
		a,
	)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return err
	}

	return nil
}

func resetAssign(command Command) error {
	var err error
	dutydate := utils.GetToday()

	if command.Arguments != "" {
		dutydate, err = checkDate(command.Arguments)
		if err != nil {
			logger.Log.Error().Err(err).Send()
			sendMessage(command.ChatID, err.Error(), NoParseMode)
			return err
		}
	}

	as, err := assignment.AssignmentRepo.GetAssignmentByDate(
		context.Background(),
		dutydate,
		command.ChatID)
	if err != nil {
		sendMessage(
			command.ChatID,
			fmt.Sprintf(
				"Error getting assignments for %s",
				dutydate.Format(utils.AssignDateFormat),
			),
			NoParseMode,
		)
		return err
	}
	if as.Operator == "" {
		return nil
	}

	err = assignment.AssignmentRepo.DeleteAssignment(context.Background(), as.ID)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		sendMessage(
			command.ChatID,
			"failed to reset assignments",
			NoParseMode,
		)
		return err
	}

	sendMessage(
		command.ChatID,
		fmt.Sprintf(
			"@%s is unassigned from %s",
			as.Operator,
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

func freeSlots(command Command) error {
	weeks, err := checkWeeks(command.Arguments)
	if err != nil {
		sendMessage(
			command.ChatID,
			err.Error(),
			NoParseMode,
		)
		return err
	}

	table, err := getFreeSlotsTable(command.ChatID, weeks)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return err
	}

	sendMessage(command.ChatID, table, NoParseMode)
	return nil
}

func getFreeSlotsTable(chatID int64, weeks int) (string, error) {
	slots, err := assignment.AssignmentRepo.GetFreeSlots(
		context.Background(),
		utils.GetToday().Add(utils.WeekDuration*time.Duration(weeks)),
		chatID)
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

func show(command Command) error {
	weeks, err := checkWeeks(command.Arguments)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		sendMessage(
			command.ChatID,
			err.Error(),
			NoParseMode,
		)
		return err
	}

	table, err := getAssignmentsTable(command.ChatID, weeks)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		sendMessage(
			command.ChatID,
			fmt.Sprintf("Tabulation error: %s", err.Error()),
			NoParseMode,
		)
		return err
	}

	if table == "" {
		table = "Nothing to show"
	}
	sendMessage(command.ChatID, fmt.Sprintf("```\n%s\n```", table), MarkdownParseMode)
	return nil
}

func getAssignmentsTable(chatID int64, weeks int) (string, error) {
	assignments, err := assignment.AssignmentRepo.GetAssignmentSchedule(
		context.Background(),
		utils.GetToday().Add(utils.WeekDuration*time.Duration(weeks)),
		chatID)
	if err != nil {
		logger.Log.Error().Stack().Err(err).Send()
		return "", fmt.Errorf("couldn't get assignments")
	}

	schedule := utils.NewPrettyTable()

	for _, ass := range assignments {
		dutyDate := ass.At.Format(utils.HumanDateFormat)
		schedule.AddRow([]string{ass.Operator, dutyDate})
	}
	table, err := schedule.String()
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return "", err
	}
	return table, nil
}
