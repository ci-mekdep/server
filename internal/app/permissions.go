package app

import (
	"slices"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/mekdep/server/config"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
)

var rolePermissions = map[models.Role][]Permission{
	models.RoleAdmin: []Permission{
		PermAdminSchools,
		PermAdminClassrooms,
		PermAdminUsers,
		PermAdminSubjects,
		PermAdminSubjectsExams,
		PermAdminTimetables,
		PermAdminShifts,
		PermAdminPeriods,
		PermAdminTopics,
		PermAdminContactItems,
		PermAdminBooks,
		PermAdminSettings,
		PermAdminTeacherExcuses,
		PermAdminPayments,
		PermAdminSchoolTransfers,
		PermToolReportForms,
		PermToolNotifier,
		PermToolReports,
		PermToolReportsData,
		PermToolExport,
		PermJournal,
	},
	models.RoleOrganization: []Permission{
		PermAdminSchools,
		PermAdminClassrooms,
		PermAdminUsers,
		PermAdminSubjects,
		PermAdminSubjectsExams,
		PermAdminTimetables,
		PermAdminShifts,
		PermAdminTopics,
		PermAdminPeriods,
		PermAdminTeacherExcuses,
		PermAdminSchoolTransfers,
		PermToolNotifier,
		PermToolReportForms,
		PermToolReports,
		PermToolReportsData,
		PermToolExport,
		PermJournal,
	},
	models.RoleOperator: []Permission{
		PermAdminUsers,
		PermAdminContactItems,
		PermAdminTopics,
		PermAdminBooks,
	},
	models.RolePrincipal: []Permission{
		PermAdminSchools,
		PermAdminClassrooms,
		PermAdminUsers,
		PermAdminSubjects,
		PermAdminSubjectsExams,
		PermAdminTimetables,
		PermAdminShifts,
		PermAdminPeriods,
		PermAdminTeacherExcuses,
		PermAdminSchoolTransfers,
		PermToolReportForms,
		PermJournal,
		PermToolNotifier,
		PermToolReports,
		PermToolReportsData,
		PermToolExport,
	},
	models.RoleTeacher: []Permission{
		PermTeacher,
		PermJournal,
		PermToolNotifier,
		PermToolReportForms,
		PermAdminUsers,
		PermToolExport,
	},
	models.RoleParent: []Permission{
		PermChildren,
		PermDiary,
		PermAnalytics,
		PermPayments,
		PermUnlimited,
		PermPlus,
		PermAdminPeriods,
	},
	models.RoleStudent: []Permission{
		PermChildren,
		PermPayments,
		PermDiary,
		PermAnalytics,
		PermUnlimited,
		PermPlus,
		PermAdminPeriods,
	},
}
var rolePermissionsRead = map[models.Role][]Permission{
	models.RoleAdmin: []Permission{
		PermToolLogs,
		PermToolReportsData,
		PermToolImport,
		PermToolNotifier,
		PermToolExport,
		PermJournal,
	},
	models.RoleOrganization: []Permission{
		PermAdminSettings,
		PermToolLogs,
		PermToolImport,
		PermToolNotifier,
		PermToolExport,
		PermToolReportsData,
		PermJournal,
	},
	models.RoleOperator: []Permission{
		PermAdminClassrooms,
		PermAdminSubjects,
		PermAdminSubjectsExams,
		PermAdminTimetables,
		PermAdminShifts,
		PermAdminPeriods,
		PermAdminReports,
		PermAdminSettings,
		PermAdminTeacherExcuses,
		PermAdminUsers,
		PermAdminSchools,
		PermAdminPayments,
		PermToolReports,
		PermToolReportForms,
		PermToolReportsData,
		PermToolNotifier,
		PermJournal,
	},
	models.RolePrincipal: []Permission{
		PermAdminSettings,
		PermToolLogs,
		PermToolImport,
		PermToolNotifier,
		PermToolExport,
		PermJournal,
	},
	models.RoleTeacher: []Permission{
		PermAdminTeacherExcuses,
		PermAdminSchools,
		PermAdminPeriods,
		PermAdminSubjects,
		PermAdminSubjectsExams,
		PermToolReportForms,
		PermAdminClassrooms,
		PermToolExport,
	},
	models.RoleParent:  []Permission{},
	models.RoleStudent: []Permission{},
}

type Permission interface{}

var (
	PermAdminSchools         Permission = "admin_schools"
	PermAdminClassrooms      Permission = "admin_classrooms"
	PermAdminUsers           Permission = "admin_users"
	PermAdminSubjects        Permission = "admin_subjects"
	PermAdminSubjectsExams   Permission = "admin_subjects_exams"
	PermAdminTimetables      Permission = "admin_timetables"
	PermAdminShifts          Permission = "admin_shifts"
	PermAdminPeriods         Permission = "admin_periods"
	PermAdminTopics          Permission = "admin_topics"
	PermAdminContactItems    Permission = "admin_contact_items"
	PermAdminBooks           Permission = "admin_books"
	PermAdminReports         Permission = "admin_reports"
	PermAdminSettings        Permission = "admin_settings"
	PermAdminTeacherExcuses  Permission = "admin_teacher_excuses"
	PermAdminPayments        Permission = "admin_payments"
	PermAdminSchoolTransfers Permission = "admin_school_transfers"

	PermToolReports     Permission = "tool_reports"
	PermToolReportForms Permission = "tool_report_forms"
	PermToolReportsData Permission = "tool_reports_data"
	PermToolLogs        Permission = "tool_logs"
	PermToolImport      Permission = "tool_import"
	PermToolReset       Permission = "tool_reset"
	PermToolNotifier    Permission = "tool_notifier"
	PermToolExport      Permission = "tool_export"

	PermTopics Permission = "topics"

	PermJournal   Permission = "journal"
	PermDiary     Permission = "diary"
	PermChildren  Permission = "children"
	PermAnalytics Permission = "analytics"
	PermPayments  Permission = "payments"
	PermUnlimited Permission = "unlimited"
	PermPlus      Permission = "plus"

	PermTeacher Permission = "teacher"
	PermUser    Permission = "user"
)

func (a App) userAction(ses *utils.Session, p Permission, f func(*models.User) error) error {
	if err := ses.LoadSession(); err != nil {
		if _, ok := err.(AppError); ok || err == pgx.ErrNoRows {
			return ErrUnauthorized
		}
		// cannot run query properly
		return err
	}
	if ses.GetUser() == nil {
		return ErrUnauthorized
	}
	return f(ses.GetUser())
}

func (a App) UserActionCheckRead(ses *utils.Session, p Permission, f func(*models.User) error) error {
	if err := a.CheckUserRead(ses.GetRole(), p); err != nil {
		return err
	}
	return a.userAction(ses, p, f)
}
func (a App) UserActionCheckWrite(ses *utils.Session, p Permission, f func(*models.User) error) error {
	if err := a.CheckUserWrite(ses.GetRole(), p); err != nil {
		return err
	}
	return a.userAction(ses, p, f)
}

func (a App) CheckUserRead(r *models.Role, p Permission) error {
	if r == nil {
		return ErrUnauthorized
	}
	if !RoleHasPermissionRead(*r, p) {
		return ErrForbidden
	}
	return nil
}

func (a App) CheckUserWrite(r *models.Role, p Permission) error {
	if r == nil {
		return ErrUnauthorized
	}
	if !RoleHasPermissionWrite(*r, p) {
		return ErrForbidden
	}
	return nil
}

func (a App) CheckPayment(student *models.User, p Permission) error {
	if p == PermUnlimited || p == PermPlus {
		if student == nil {
			return ErrNotPaid
		}
		if student.Classrooms[0].TariffEndAt == nil || student.Classrooms[0].TariffEndAt != nil && time.Now().After(*student.Classrooms[0].TariffEndAt) {
			return ErrNotPaid
		}
		if student.Classrooms[0].TariffType == nil {
			return ErrNotPaid
		}
		if *student.Classrooms[0].TariffType == string(models.PaymentUnlimited) && (p == PermPlus || p == PermUnlimited) {
			return nil
		}
		if *student.Classrooms[0].TariffType == string(models.PaymentPlus) && (p == PermPlus) {
			return nil
		}
		return ErrNotPaid
	}
	return nil
}

func RoleHasPermissionWrite(r models.Role, p Permission) bool {
	if config.Conf.AppIsReadonly != nil && *config.Conf.AppIsReadonly {
		return false
	}
	if p == PermUser {
		return true
	}
	ps := PermissionsWriteByRole(r)
	if ps != nil {
		if i := slices.Index(ps, p); i >= 0 {
			return true
		}
	}
	return false
}

func RoleHasPermissionRead(r models.Role, p Permission) bool {
	if p == PermUser {
		return true
	}
	ps := PermissionsReadByRole(r)
	if ps != nil {
		if i := slices.Index(ps, p); i >= 0 {
			return true
		}
	}
	return false
}

func PermissionsWriteByRole(r models.Role) []Permission {
	if ps, ok := rolePermissions[r]; ok {
		return ps
	}
	return []Permission{}
}

func PermissionsReadByRole(r models.Role) []Permission {
	if ps, ok := rolePermissionsRead[r]; ok {
		return append(ps, PermissionsWriteByRole(r)...)
	}
	return []Permission{}
}

func (a App) GetPermissions(ses *utils.Session) ([]Permission, []Permission, error) {
	r := ses.GetRole()
	if r == nil {
		return nil, nil, ErrUnauthorized
	}
	psr := PermissionsReadByRole(*r)
	if psr == nil {
		return nil, nil, ErrForbidden
	}
	psw := PermissionsWriteByRole(*r)
	if psw == nil {
		return nil, nil, ErrForbidden
	}
	if config.Conf.AppIsReadonly != nil && *config.Conf.AppIsReadonly {
		psw = []Permission{}
	}
	return psr, psw, nil
}
