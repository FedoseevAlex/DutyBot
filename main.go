package main

import (
	//db "dutybot/internal/database"
	cal "dutybot/internal/calendar"
	"time"
)

func main() {
	//db.CreateSchema()
	//var op *db.Operator = &db.Operator{UserName: "first Operator", FirstName: "name"}
	//op.Insert()
	//op.Delete()
	var date time.Time;
	date, err := time.Parse(cal.DateFormat, "10-09-2020")
	if err != nil {
		panic(err)
	}
	cal.CheckDay(date)
}
