package bot

import (
	"bytes"
	"dutybot/internal/calendar"
	db "dutybot/internal/database"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"text/tabwriter"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

var handlers map[string]func(*tgbot.BotAPI, *tgbot.Message)

func initHandlers() {
	handlers = make(map[string]func(*tgbot.BotAPI, *tgbot.Message))
	handlers["help"] = help
	handlers["assign"] = assign
	handlers["show"] = show
	handlers["operator"] = operator
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
/assign [date] - assign yourself for duty. Date should be in format DD-MM-YYYY`
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

/// This function assumes that date
/// is ordered like DD MM YYYY
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
	t, err = time.Parse("02 01 2006", fmt.Sprintf("%s %s %s", date, month, year))
	return
}

func sendMessage(bot *tgbot.BotAPI, chatID int64, message string) {
	msg := tgbot.NewMessage(
		chatID,
		message,
	)
	msg.ParseMode = "Markdown"

	_, err := bot.Send(msg)
	if err != nil {
		log.Print(err)
	}
}

func Today() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(
		y,
		m,
		d,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

func assign(bot *tgbot.BotAPI, msg *tgbot.Message) {
	dutydate, err := parseTime(msg.CommandArguments())
	if err != nil {
		log.Print(err)
		sendMessage(
			bot,
			msg.Chat.ID,
			fmt.Sprintf("Something wrong with date: %s", err),
		)
		return
	}

	if Today().After(dutydate) {
		sendMessage(bot, msg.Chat.ID, "Assignment is possible only for a future date")
		return
	}

	as, err := db.GetAssignmentByDate(msg.Chat.ID, dutydate)
	if err != nil {
		log.Print(err)
	}
	if as != nil {
		log.Printf("%+v %+v", as, as.Operator)
		sendMessage(
			bot,
			msg.Chat.ID,
			fmt.Sprintf("This day already taken by `%s`", as.Operator.UserName),
		)
		return
	}

	if calendar.IsHoliday(dutydate) {
		answer := fmt.Sprintf(
			"%s is a holiday. No duty on holidays",
			dutydate.Format("02-01-2006"),
		)
		sendMessage(bot, msg.Chat.ID, answer)
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
	a := &db.Assignment{ChatID: msg.Chat.ID, DutyDate: dutydate.Unix(), Operator: op}
	log.Printf("New assignment: %+v", a)
	err = a.Insert()
	if err != nil {
		log.Print(err)
		return
	}
	show(bot, msg)
}

func show(bot *tgbot.BotAPI, msg *tgbot.Message) {
	weeks, err := strconv.Atoi(msg.CommandArguments())
	if err != nil {
		weeks = 2
	}

	assignments, err := db.GetAssignmentSchedule(weeks, msg.Chat.ID)
	if err != nil {
		sendMessage(bot, msg.Chat.ID, "Couldn't get assignments.")
		return
	}

	buf := make([]byte, 0)
	b := bytes.NewBuffer(buf)
	t := tabwriter.NewWriter(
		b,
		0,
		4,
		2,
		' ',
		tabwriter.TabIndent,
	)
	for _, ass := range assignments {
		dutyDate := time.Unix(ass.DutyDate, 0).Format("Mon Jan 02 2006")

		_, err = fmt.Fprintf(t, "`%s\t%s`\n", ass.Operator.UserName, dutyDate)
		if err != nil {
			log.Print(err)
			return
		}
	}
	err = t.Flush()
	if err != nil {
		log.Print(err)
		return
	}

	reply := tgbot.NewMessage(msg.Chat.ID, b.String())
	reply.ParseMode = "Markdown"
	_, err = bot.Send(reply)
	if err != nil {
		log.Print(err)
	}
}
