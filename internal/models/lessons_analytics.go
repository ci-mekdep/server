package models

import (
	"sort"
	"strings"
	"time"
)

type ParentAnalyticsWeekly struct {
	SubjectRating *SubjectRatingByPeriodGrade `json:"subject_rating"`
	// TODO: deprecate below
	SubjectPercentByAreas []SubjectPercentByArea `json:"subject_percents"`
	GradeStreak           int                    `json:"grade_strike"`
}

type ParentAnalyticsSummary struct {
	GradeStreak   int                         `json:"grade_strike"`
	SubjectRating *SubjectRatingByPeriodGrade `json:"subject_rating"`
	SubjectGrades []SubjectGrade              `json:"subject_grades"`
}

type SubjectGrade struct {
	SubjectId   string    `json:"subject_id"`
	SubjectName string    `json:"subject_name"`
	Date        time.Time `json:"date"`
	GradeValue  string    `json:"grade_value"`
}
type SubjectPercentByArea struct {
	Title       string           `json:"title"`
	Point       int              `json:"point"`
	PointPrev   int              `json:"point_prev"`
	Percent     int              `json:"percent"`
	PercentPrev int              `json:"percent_prev"`
	BySubject   []SubjectPercent `json:"by_subject"`
}

type SubjectPercent struct {
	SubjectId       string `json:"subject_id"`
	SubjectFullName string `json:"name"`
	LessonsCount    int    `json:"lessons_count"`
	GradesCount     int    `json:"grades_count"`
	GradesValues1   *[]int `json:"-"`
	GradesValues2   *[]int `json:"-"`
	Point           int    `json:"point"`
	PointPrev       int    `json:"point_prev"`
	Percent         int    `json:"percent"`
	PercentPrev     int    `json:"percent_prev"`
}

type SubjectPeriodGradeFinished struct {
	Subject       Subject `json:"subject"`
	ClassroomName *string `json:"classroom_name"`
	StudentsCount int     `json:"students_count"`
	FinishedCount int     `json:"finished_count"`
}

type SubjectRating struct {
	StudentId     string              `json:"student_id"`
	SubjectId     string              `json:"subject_id"`
	SubjectName   string              `json:"name"`
	LessonsCount  int                 `json:"lessons_count"`
	GradesCount   int                 `json:"grades_count"`
	GradesValues1 *[]int              `json:"-"`
	GradesValues2 *[]int              `json:"-"`
	Point         int                 `json:"point"`
	PointPrev     int                 `json:"point_prev"`
	Rating        int                 `json:"rating"`
	RatingPrev    int                 `json:"rating_prev"`
	PeriodGrade   PeriodGradeResponse `json:"period_grade"`
}

type SubjectRatingByPeriodGrade struct {
	BySubject  []SubjectRating `json:"by_subject"`
	Rating     int             `json:"rating"`
	RatingPrev int             `json:"rating_prev"`
}

type SubjectLessonGrades struct {
	LessonDate      time.Time `json:"lesson_date"`
	LessonHours     int       `json:"lesson_hours"`
	GradesValues1   *[]int    `json:"-"`
	GradesValues2   *[]int    `json:"-"`
	GradesGoodCount int       `json:"-"`
	GradesCount     int       `json:"-"`
	GradesSum       int       `json:"-"`
}

func (sg *SubjectLessonGrades) SetOtherKeys() {
	if sg.GradesValues1 != nil {
		for _, v := range *sg.GradesValues1 {
			if v == 5 {
				sg.GradesGoodCount++
			}
			sg.GradesCount++
			sg.GradesSum += v
		}
	}
	if sg.GradesValues2 != nil {
		for _, v := range *sg.GradesValues2 {
			if v == 5 {
				sg.GradesGoodCount++
			}
			sg.GradesCount++
			sg.GradesSum += v
		}
	}
}

func (sp *SubjectRating) CalcPoint() {
	gradeWeight := map[int]int{
		5: 10,
		4: 5,
		3: 1,
		2: 0,
	}
	point := 0
	if sp.GradesValues1 == nil {
		sp.GradesValues1 = new([]int)
	}
	if sp.GradesValues2 == nil {
		sp.GradesValues2 = new([]int)
	}
	for _, v := range *sp.GradesValues1 {
		point += gradeWeight[v]
	}
	for _, v := range *sp.GradesValues2 {
		point += gradeWeight[v]
	}

	if sp.LessonsCount > 0 && sp.GradesCount*100/sp.LessonsCount <= 50 {
		point = point / 2
	}
	sp.Point = point
}

func (sp *SubjectPercent) CalcPoint() {
	gradeSum := 0
	percent := 0
	lessons := sp.LessonsCount

	pointsByGrades := map[int]int{
		5: 7,
		4: 3,
		3: 1,
		2: -1,
	}
	if sp.GradesValues1 == nil {
		sp.GradesValues1 = new([]int)
	}
	if sp.GradesValues2 == nil {
		sp.GradesValues2 = new([]int)
	}
	for _, v := range *sp.GradesValues1 {
		gradeSum += pointsByGrades[v]
	}
	for _, v := range *sp.GradesValues2 {
		gradeSum += pointsByGrades[v]
	}

	if gradeSum > 0 && lessons > 0 {
		percent = gradeSum * 100 / (lessons * pointsByGrades[5])
	}

	sp.Point = gradeSum
	sp.Percent = percent
}

func (sp *SubjectRatingByPeriodGrade) CalcRating() {
	if sp.BySubject != nil {
		sum := 0
		count := 0
		prevSum := 0
		prevCount := 0
		for _, v := range sp.BySubject {
			if v.Rating == 0 {
				continue
			}
			sum += v.Rating
			count++
			prevSum += v.RatingPrev
			prevCount++
		}
		if count > 0 {
			sp.Rating = sum / count
		}
		if prevCount > 0 {
			sp.RatingPrev = prevSum / prevCount
		}
	}
}

func (sp *SubjectPercentByArea) CalcPoint() {
	if sp.BySubject != nil {
		sumsOrdered := []int{}
		sumsOrderedPrev := []int{}

		for _, v := range sp.BySubject {
			sumsOrdered = append(sumsOrdered, v.Percent)
			sumsOrderedPrev = append(sumsOrderedPrev, v.PercentPrev)
		}

		sort.Slice(sumsOrdered, func(i, j int) bool {
			return sumsOrdered[i] > sumsOrdered[j]
		})
		sort.Slice(sumsOrderedPrev, func(i, j int) bool {
			return sumsOrderedPrev[i] > sumsOrderedPrev[j]
		})

		sum := 0
		count := 0
		sumPrev := 0
		countPrev := 0

		for k, v := range sumsOrdered {
			// get most 4 subjects
			if k > 3 {
				break
			}
			sum += v
			count++
		}
		for k, v := range sumsOrderedPrev {
			// get most 4 subjects
			if k > 3 {
				break
			}
			sumPrev += v
			countPrev++
		}

		if count > 0 {
			sp.Point = sum
			sp.Percent = sum / count
		}
		if countPrev > 0 {
			sp.PointPrev = sumPrev
			sp.PercentPrev = sumPrev / countPrev
		}
	}
}

func CalcRatingStudents(sp *[]SubjectRating) {
	index := map[string]int{}
	lastPoint := map[string]int{}
	sort.Slice(*sp, func(i, j int) bool {
		return (*sp)[i].SubjectName > (*sp)[j].SubjectName || (*sp)[i].Point > (*sp)[j].Point
	})
	for k, v := range *sp {
		if _, ok := index[v.SubjectName]; !ok {
			index[v.SubjectName] = 1
		}
		if v.Point > 0 {
			(*sp)[k].Rating = index[v.SubjectName]
		} else {
			(*sp)[k].Rating = 0
		}
		if lastPoint[v.SubjectName] != v.Point {
			index[v.SubjectName]++
		}
		lastPoint[v.SubjectName] = v.Point
	}
}

func CalcPercentByArea(sp *[]SubjectPercentByArea) {
	// sum := 0
	// count := 0
	// sumPrev := 0
	// countPrev := 0
	// for _, v := range *sp {
	// 	sum += v.Point
	// 	sumPrev += v.PointPrev
	// 	count++
	// 	countPrev++
	// }
	// for k, _ := range *sp {
	// 	if sum > 0 {
	// 		(*sp)[k].Percent = sum / count
	// 	}
	// 	if sumPrev > 0 {
	// 		(*sp)[k].PercentPrev = sumPrev / countPrev
	// 	}
	// }
}

func SubjectIsNoGrade(s Subject) bool {
	n := strings.ToLower(*s.Name)
	c := ""
	if s.Classroom != nil && s.Classroom.Name != nil {
		c = *s.Classroom.Name
	}
	if len(c) < 2 {
		c = "  "
	}
	if len(n) < 2 {
		n = "  "
	}
	if string(c[0]) == "1" && !strings.ContainsAny(string(c[1]), "1234567890") {
		return true
	}
	if string(n[0]) == "s" && (string(n[1]) == "y" || string(n[1]) == ".") {
		return true
	}
	return false
}
