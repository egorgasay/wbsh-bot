package service

import (
	"bot/config"
	"bot/internal/constant"
	"bot/internal/storage"
	"fmt"
	"github.com/xuri/excelize/v2"
	"sort"
	"strings"
)

type ScheduleService struct {
	paths         []string
	maxPairPerDay int
	storage       *storage.Storage
	schedule      map[group]WorkWeek // todo: add lock
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

func toDay(i int) string {
	switch i {
	case 0:
		return "Понедельник"
	case 1:
		return "Вторник"
	case 2:
		return "Среда"
	case 3:
		return "Четверг"
	case 4:
		return "Пятница"
	case 5:
		return "Суббота"
	case 6:
		return "Воскресенье"
	}

	return ""
}

func findGroup(pe PairEntity, groupID int) (Pair, error) {
	//if groupID == -1 && len(pe) == 2 { // TODO:
	//	return service.Pair{
	//		Teacher: "",
	//		Subject: "",
	//		Room:    "",
	//		Group:   0,
	//	}, nil
	//
	//}
	//

	if pe == nil {
		return Pair{}, constant.ErrNoPair
	}

	for _, p := range pe {
		if p.Group == groupID || p.Group == 0 {
			return p, nil
		}
	}
	return Pair{}, constant.ErrGroupNotFound
}

func DayToString(day WorkDay, needNew bool, offset int, subGroup int) string {
	var sb strings.Builder
	if len(day) > 0 {
		if needNew {
			sb.WriteString(fmt.Sprintf("Твое ближайшее расписание: \n\n"))
		} else {
			sb.WriteString(fmt.Sprintf("День: %s\n\n", toDay(offset)))
		}
	}

	if len(day) == 0 {
		sb.WriteString("Нет пар на этот день")
		return sb.String()
	}

	for i, pairE := range day {
		actualPair, err := findGroup(pairE, subGroup)
		if err != nil {
			switch err {
			case constant.ErrGroupNotFound:
				sb.WriteString(
					fmt.Sprintf(
						"№%d\nПара у другой группы\n\n", i+1,
					),
				)
			case constant.ErrNoPair:
				sb.WriteString(
					fmt.Sprintf(
						"№%d\nПара не найдена, проверьте на сайте на всякий случай)\n\n", i+1,
					),
				)
			}
		} else {
			sb.WriteString(
				fmt.Sprintf(
					"№%d\nПредмет: %s\nКабинет: %s\nПреподаватель: %s\n\n",
					i+1, actualPair.Subject, actualPair.Room, actualPair.Teacher,
				),
			)
		}
	}

	return sb.String()
}

func NewSchedule(c config.Config) (*ScheduleService, error) {

	return &ScheduleService{
		paths:         c.Files,
		maxPairPerDay: c.MaxPairPerDay,
	}, nil
}

func (s *ScheduleService) Update() (err error) {
	var allCols [][]string

	for _, path := range s.paths {
		err = func() error {
			f, err := excelize.OpenFile(path)
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}

			defer func() {
				// Close the spreadsheet.
				if errClose := f.Close(); err != nil {
					err = fmt.Errorf("close file: %w", errClose)
				}
			}()

			for _, sheet := range f.GetSheetMap() {
				cols, err := f.GetCols(sheet)
				if err != nil {
					return fmt.Errorf("get cols: %w", err)
				}

				allCols = append(allCols, cols...)
			}

			return nil
		}()

		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
	}

	m := colsToMap(allCols, 5)
	delete(m, "День\nнеде")
	delete(m, "Время")

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

func (s *ScheduleService) GetDayGroupNames() []string {
	var names []string
	for g := range s.schedule {
		names = append(names, string(g))
	}

	return names
}

func (s *ScheduleService) VerifyGroup(g string) bool {
	if s.schedule == nil {
		return false
	}

	if _, ok := s.schedule[group(g)]; ok {
		return true
	}

	return false
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
	all = iota - 1
	no
	first
	second
)

func newFromKabAndPair(kap kabAndPair) ([]Pair, error) {
	rawPair := kap.pair

	if rawPair == "Нет" {
		return nil, nil
	}

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
			pair.Room = kap.kab

			return []Pair{pair}, nil
		case 2:
			g1 := strings.Trim(split[0], " \n1гр.")
			g2 := strings.Trim(split[1], " \n2гр.")
			g1 = strings.Replace(g1, "\n", " ", -1)
			g2 = strings.Replace(g2, "\n", " ", -1)

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
			if len(kabs) != 2 {
				kabs = strings.Split(kap.kab, " ")
				if len(kabs) != 2 {
					kabs = []string{kap.kab, kap.kab}
				}
			}

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
		return "Нет информации", str
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

	var iCap = []int{1, 1, 1, 1, 1, 1, 1}

	var idx = 0
	for _, el := range dayPair[2:] {
		if el != "" {
			idx++
			continue
		}
		iCap[idx]++
	}

	cols = cols[2:]

	lengths := allLengths(cols)

	goodPairs := lengths[0]
	wrongPairs := -1

	goodKabs := 0
	if len(lengths) == 1 {
		goodKabs = lengths[0]
	} else {
		goodKabs = lengths[1]
	}
	wrongKabs := -1

	if len(lengths) == 4 {
		goodPairs = lengths[0]
		wrongPairs = lengths[1]
		goodKabs = lengths[2]
		wrongKabs = lengths[3]
	}

	if len(lengths) == 3 {
		goodPairs = lengths[0]
		wrongPairs = lengths[1]
		goodKabs = lengths[0]
		wrongKabs = lengths[1]
	}

	// because of col = col[1:]
	goodPairs--
	goodKabs--
	wrongPairs--
	wrongKabs--

	clearData := func(col []string) []string {
		var tempCols []string
		var first = col[0]

		col = col[1:]

		for i := 0; i < len(col); i++ {
			if i%2 == 0 {
				tempCols = append(tempCols, col[i])
			}
		}

		res := tempCols[:func() int {
			if len(col) < 31 {
				return len(tempCols)
			}
			return 31
		}()]

		return append([]string{first}, res...)
	}

	for colsIndx, col := range cols {
		if len(col) >= wrongKabs {
			cols[colsIndx] = clearData(col)
		}
	}

	for colsIndx, col := range cols {
		gname := col[0]

		if gname == "" || gname == "День\nнеде" || gname == "Время" {
			continue
		}

		col = col[1:]
		week := make(week, 5)

		var iCapCopy = make([]int, len(iCap))
		copy(iCapCopy, iCap)

		i := 0
		for cellIndex, cell := range col {
			if iCap[i] == 0 {
				i++
			}
			if i == 5 {
				break
			}

			iCap[i]--

			if cell == "" {
				continue
			}

			if colsIndx == 185 {
				print()
			}

			week[i] = append(week[i], kabAndPair{
				pair: cell,
				kab:  cols[colsIndx+1][cellIndex+1],
			})

			if cell == "ВПР" {
				week[i] = append(week[i], kabAndPair{
					pair: cell,
					kab:  cols[colsIndx+1][cellIndex+1],
				})
			}
		}
		copy(iCap, iCapCopy)

		mp[group(gname)] = week
	}

	return mp
}

func allLengths(cols [][]string) []int {
	var lengths = make(map[int]struct{})
	for _, col := range cols {
		lengths[len(col)] = struct{}{}
	}

	var res []int
	for l := range lengths {
		res = append(res, l)
	}

	sort.Ints(res)

	return res
}
