package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

// NextDate вычисляет следующую дату для задачи по правилу повторения
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %w", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	parts := strings.Split(repeat, " ")

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("не указан интервал в днях")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("недопустимый интервал дней")
		}
		for {
			date = date.AddDate(0, 0, days)
			if date.After(now) {
				break
			}
		}
		return date.Format(DateFormat), nil

	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if date.After(now) {
				break
			}
		}
		return date.Format(DateFormat), nil

	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("не указаны дни недели")
		}

		var weekdays [8]bool
		for _, p := range strings.Split(parts[1], ",") {
			d, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || d < 1 || d > 7 {
				return "", fmt.Errorf("недопустимый день недели: %s", p)
			}
			weekdays[d] = true
		}

		if date.Before(now) {
			date = now
		}

		for i := 1; i <= 7; i++ {
			next := date.AddDate(0, 0, i)
			wd := int(next.Weekday())
			if wd == 0 {
				wd = 7
			}
			if weekdays[wd] {
				return next.Format(DateFormat), nil
			}
		}
		return "", fmt.Errorf("не найден подходящий день недели")

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("неверный формат правила m")
		}

		var days []int
		for _, p := range strings.Split(parts[1], ",") {
			d, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || d < -2 || d == 0 || d > 31 {
				return "", fmt.Errorf("недопустимый день месяца: %s", p)
			}
			days = append(days, d)
		}

		// если указаны конкретные месяцы, запоминаем их
		var months [13]bool
		allMonths := true
		if len(parts) == 3 {
			allMonths = false
			for _, p := range strings.Split(parts[2], ",") {
				m, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil || m < 1 || m > 12 {
					return "", fmt.Errorf("недопустимый месяц: %s", p)
				}
				months[m] = true
			}
		}

		if date.Before(now) {
			date = now
		}

		// перебираем дни вперёд, ищем подходящий
		for i := 1; i <= 366*2; i++ {
			next := date.AddDate(0, 0, i)

			if !allMonths && !months[int(next.Month())] {
				continue
			}

			d := next.Day()
			lastDay := time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			for _, day := range days {
				if day == -1 && d == lastDay {
					return next.Format(DateFormat), nil
				}
				if day == -2 && d == lastDay-1 {
					return next.Format(DateFormat), nil
				}
				if day > 0 && d == day {
					return next.Format(DateFormat), nil
				}
			}
		}
		return "", fmt.Errorf("не найдена подходящая дата")

	default:
		return "", fmt.Errorf("неподдерживаемый формат правила: %s", parts[0])
	}
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, "неверный формат параметра now", http.StatusBadRequest)
			return
		}
	}

	result, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(result))
}
