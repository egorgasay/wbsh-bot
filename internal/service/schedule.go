package service

import (
	"bot/config"
	"bot/internal/storage"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
)

type Schedule struct {
	path          string
	sheetName     string
	maxPairPerDay int
	storage       *storage.Storage
	m             map[group]WorkWeek
}

func NewSchedule(c config.Config) (*Schedule, error) {

	return &Schedule{
		path:          c.PathToSchedule,
		sheetName:     c.SheetName,
		maxPairPerDay: c.MaxPairPerDay,
	}, nil
}

func (s *Schedule) Update() (err error) {
	f, err := excelize.OpenFile(s.path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	defer func() {
		// Close the spreadsheet.
		if errClose := f.Close(); err != nil {
			err = fmt.Errorf("close file: %w", errClose)
		}
	}()

	cols, err := f.GetCols(s.sheetName)
	if err != nil {
		return fmt.Errorf("get cols: %w", err)
	}

	s.m = colsToMap(cols, s.maxPairPerDay)

	return nil
}

func (s *Schedule) GetWeekByGroup(groupName string) (WorkWeek, error) {
	if s.m == nil {
		return nil, fmt.Errorf("no schedule")
	}

	if w, ok := s.m[group(groupName)]; ok {
		return w, nil
	}

	return nil, fmt.Errorf("group not found")
}

type kabAndPair struct {
	kab  string
	pair string
}

type group string

type groups map[group]kabAndPair
type WorkWeek []day
type day []kabAndPair

func (w WorkWeek) IsNext(i int) bool {
	return len(w)-1 > i
}

func (d day) IsNext(i int) bool {
	return len(d)-1 > i
}

type Pair struct {
	Teacher string
	Subject string
	Room    string
	Group   byte
}

const (
	no byte = iota
	first
	second
)

func (p Pair) formatKabAndPair(kap kabAndPair) (Pair, error) {
	rawPair := kap.pair
	rawPair = strings.Replace(rawPair, "Гр", "гр", -1)
	rawPair = strings.Replace(rawPair, "ГР", "гр", -1)
	rawPair = strings.Replace(rawPair, "гР", "гр", -1)

	if strings.Contains(rawPair, "гр. ") {
		split := strings.SplitAfter(rawPair, "гр. ")
		split = split[1:]

		switch len(split) {
		case 0:
			return p, fmt.Errorf("wrong pair")
		case 1:
			teacher, subject := teacherAndSubject(split[0])
			if strings.HasPrefix(subject, "1") {
				subject = strings.Replace(subject, "1 ", "", 1)
				res.WriteString(fmt.Sprintln("Предмет первой группы:", subject))
				res.WriteString(fmt.Sprintln("Первая группа преподаватель:", teacher))
				res.WriteString(fmt.Sprintln("Кабинет первой группы:", kap.kab))
			} else {
				subject = strings.Replace(subject, "2 ", "", 1)
				res.WriteString(fmt.Sprintln("Предмет второй группы:", subject))
				res.WriteString(fmt.Sprintln("Вторая группа преподаватель:", teacher))
				res.WriteString(fmt.Sprintln("Кабинет второй группы:", kap.kab))
			}
		case 2:
			g1 := strings.Trim(split[0], " \n1гр.")
			g2 := strings.Trim(split[1], " \n2гр.")

			teacher1, subject1 := teacherAndSubject(g1)
			teacher2, subject2 := teacherAndSubject(g2)
			kabs := strings.Split(kap.kab, "\n")

			res.WriteString(fmt.Sprintln("Предмет первой группы:", subject1))
			res.WriteString(fmt.Sprintln("Первая группа преподаватель:", teacher1))
			res.WriteString(fmt.Sprintln("Кабинет первой группы:", kabs[0]))

			res.WriteString(fmt.Sprintln("----------------------------------"))

			res.WriteString(fmt.Sprintln("Предмет второй группы:", subject2))
			res.WriteString(fmt.Sprintln("Вторая группа преподаватель:", teacher2))
			res.WriteString(fmt.Sprintln("Кабинет второй группы:", kabs[1]))
		}
		return res.String()
	}

	teacher, subject := teacherAndSubject(rawPair)
	res.WriteString(fmt.Sprintln("Предмет:", subject))
	res.WriteString(fmt.Sprintln("Преподаватель:", teacher))

	return res.String()
}

func teacherAndSubject(str string) (string, string) {
	const cutSet = " \n"

	str = strings.Replace(str, "\n", " ", -1)

	cleanArr := func(s []string) string {
		return strings.Trim(strings.Join(s, " "), cutSet)
	}

	split := strings.Split(str, " ")
	teacher := split[len(split)-2:]
	subject := split[:len(split)-2]
	return cleanArr(teacher), cleanArr(subject)
}

func colsToMap(cols [][]string, maxPairPerDay int) map[group]WorkWeek {
	mp := make(map[group]WorkWeek)
	dayPair := cols[0]
	timePair := cols[1]
	_, _ = dayPair, timePair

	cols = cols[2:]

	for colsIndx, col := range cols {
		gname := col[0]

		if gname == "" {
			continue
		}

		col = col[1:]

		week := make(WorkWeek, 5)
		j := maxPairPerDay
		i := 0
		for cellIndex, cell := range col {
			if cell == "" {
				j--
				continue
			}

			if j < 0 {
				j = maxPairPerDay
				i++
			}
			j--

			if i == 5 {
				break
			}

			week[i] = append(week[i], kabAndPair{
				pair: cell,
				kab:  cols[colsIndx+1][cellIndex+1],
			})
		}

		mp[group(gname)] = week
	}

	return mp
}
