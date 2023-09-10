package service

import (
	"bot/config"
	"bot/internal/storage"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
)

type ScheduleService struct {
	path          string
	sheetName     string
	maxPairPerDay int
	storage       *storage.Storage
	schedule      map[group]WorkWeek
}

type group string

type (
	kabAndPair struct {
		kab  string
		pair string
	}
	day  []kabAndPair
	week []day
)

type (
	PairEntity []Pair
	WorkDay    []PairEntity
	WorkWeek   []WorkDay
)

func NewSchedule(c config.Config) (*ScheduleService, error) {

	return &ScheduleService{
		path:          c.PathToSchedule,
		sheetName:     c.SheetName,
		maxPairPerDay: c.MaxPairPerDay,
	}, nil
}

func (s *ScheduleService) Update() (err error) {
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

	m := colsToMap(cols, s.maxPairPerDay)

	var all = make(map[group]WorkWeek, len(m))
	for group, week := range m {
		all[group] = make(WorkWeek, len(week))
		for i, day := range week {
			all[group][i] = make(WorkDay, len(day))
			for j, pair := range day {
				all[group][i][j], err = newFromKabAndPair(pair)
				if err != nil {
					return fmt.Errorf("newFromKabAndPair: %w", err)
				}
			}
		}
	}

	s.schedule = all

	return nil
}

func (s *ScheduleService) GetWeekByGroup(groupName string) (WorkWeek, error) {
	if s.schedule == nil {
		return nil, fmt.Errorf("no schedule")
	}

	if w, ok := s.schedule[group(groupName)]; ok {
		return w, nil
	}

	return nil, fmt.Errorf("group not found")
}

func (s *ScheduleService) GetDayByGroup(groupName string, offset int) (WorkDay, error) {
	if s.schedule == nil {
		return nil, fmt.Errorf("no schedule")
	}

	w, err := s.GetWeekByGroup(groupName)
	if err != nil {
		return nil, err
	}

	if !w.IsNext(offset) {
		return nil, fmt.Errorf("offset out of range")
	}

	return w[offset], nil
}

func (w week) IsNext(i int) bool {
	return len(w)-1 > i
}

func (w WorkWeek) IsNext(i int) bool {
	return len(w) > i
}

func (d day) IsNext(i int) bool {
	return len(d)-1 > i
}

type Pair struct {
	Teacher string
	Subject string
	Room    string
	Group   int
}

const (
	no = iota
	first
	second
)

func newFromKabAndPair(kap kabAndPair) ([]Pair, error) {
	rawPair := kap.pair
	rawPair = strings.Replace(rawPair, "Гр", "гр", -1)
	rawPair = strings.Replace(rawPair, "ГР", "гр", -1)
	rawPair = strings.Replace(rawPair, "гР", "гр", -1)

	if strings.Contains(rawPair, "гр. ") {
		split := strings.SplitAfter(rawPair, "гр. ")
		split = split[1:]

		switch len(split) {
		case 0:
			return nil, fmt.Errorf("wrong pair")
		case 1:
			var pair Pair
			teacher, subject := teacherAndSubject(split[0])
			if strings.HasPrefix(subject, "1") {
				subject = strings.Replace(subject, "1 ", "", 1)
				pair.Group = first
			} else {
				subject = strings.Replace(subject, "2 ", "", 1)
				pair.Group = second
			}
			pair.Teacher = teacher
			pair.Subject = subject

			return []Pair{pair}, nil
		case 2:
			g1 := strings.Trim(split[0], " \n1гр.")
			g2 := strings.Trim(split[1], " \n2гр.")

			g2Split := strings.Split(g2, " ")

			var teacher1, subject1, teacher2, subject2 string

			if len(g2Split) < 3 {
				teacher2 = g2
				g1Split := strings.Split(g1, " ")
				teacher1 = strings.Join(g1Split[0:1], " ")
				subject1 = strings.Join(g1Split[2:], " ")
				subject2 = subject1
			} else {
				teacher1, subject1 = teacherAndSubject(g1)
				teacher2, subject2 = teacherAndSubject(g2)
			}

			kabs := strings.Split(kap.kab, "\n")

			pair1 := Pair{
				Teacher: teacher1,
				Subject: subject1,
				Room:    kabs[0],
				Group:   first,
			}

			pair2 := Pair{
				Teacher: teacher2,
				Subject: subject2,
				Room:    kabs[1],
				Group:   second,
			}

			return []Pair{pair1, pair2}, nil
		}
	}

	teacher, subject := teacherAndSubject(rawPair)
	pair := Pair{Teacher: teacher, Subject: subject, Room: kap.kab, Group: no}

	return []Pair{pair}, nil
}

func teacherAndSubject(str string) (string, string) {
	const cutSet = " \n"

	str = strings.Replace(str, "\n", " ", -1)

	cleanArr := func(s []string) string {
		return strings.Trim(strings.Join(s, " "), cutSet)
	}

	split := strings.Split(str, " ")

	if len(split) < 3 {
		return "", ""
	}

	teacher := split[len(split)-2:]
	subject := split[:len(split)-2]
	return cleanArr(teacher), cleanArr(subject)
}

func colsToMap(cols [][]string, maxPairPerDay int) map[group]week {
	mp := make(map[group]week)
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

		week := make(week, 5)
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
