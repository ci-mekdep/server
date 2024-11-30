package store

import (
	"context"
	"time"

	"github.com/mekdep/server/internal/models"
)

type IStore interface {
	ConfirmCodeGenerate(ctx context.Context, m *models.User) (string, error)
	ConfirmCodeClear(ctx context.Context, m *models.User) error
	ConfirmCodeDelete(ctx context.Context, id string) error
	CheckConfirmCode(ctx context.Context, m *models.User, code string) (string, error)

	UsersFindByIds(ctx context.Context, ids []string) ([]*models.User, error)
	UsersFindById(ctx context.Context, id string) (*models.User, error)
	UsersFindByUsername(ctx context.Context, username string, schoolId *string, onlyAdmin bool) (models.User, error)
	UsersFindBy(ctx context.Context, f models.UserFilterRequest) ([]*models.User, int, error)
	UserUpdate(ctx context.Context, model *models.User) (*models.User, error)
	UserUpdateRelations(ctx context.Context, model *models.User) (*models.User, error)
	UserCreate(ctx context.Context, model *models.User) (*models.User, error)
	UserDelete(ctx context.Context, l []*models.User) ([]*models.User, error)
	UserDeleteSchoolRole(ctx context.Context, userIds []string, schoolIds []string, roles []string) (int, error)
	UserDeleteSchool(ctx context.Context, l []*models.User, schoolIds []string) ([]*models.User, error)
	UserDeleteFromClassroom(ctx context.Context, userId string, classroomIds []string) error
	UserDeleteAllRelations(ctx context.Context, l *[]*models.User) error
	UsersLoadRelations(ctx context.Context, l *[]*models.User, isDetail bool) error
	UsersLoadRelationsParents(ctx context.Context, l *[]*models.User) error
	UsersLoadRelationsChildren(ctx context.Context, l *[]*models.User) error
	UsersLoadRelationsClassrooms(ctx context.Context, l *[]*models.User) error
	UsersLoadRelationsClassroomsAll(ctx context.Context, l *[]*models.User) error
	UsersLoadRelationsTeacherClassroom(ctx context.Context, l *[]*models.User) error
	UsersLoadCount(ctx context.Context, schoolIds []string) (models.DashboardUsersCount, error)
	UsersLoadCountBySchool(ctx context.Context, schoolIds []string) ([]models.DashboardUsersCount, error)
	UsersLoadCountByClassroom(ctx context.Context, schoolIds []string) ([]models.DashboardUsersCountByClassroom, error)
	UsersOnlineCount(ctx context.Context, schoolId *string) (int, error)
	GetTeacherIdByName(
		ctx context.Context,
		dto models.GetTeacherIdByNameQueryDto,
	) (*string, error)
	UserClassroomGet(ctx context.Context, uid string, classroomId string) (*models.UserClassroom, error)
	UpdateUserPayment(ctx context.Context, uid string, expireAt time.Time, classroomId string) (*models.User, error)
	GetDateUserPayment(ctx context.Context, userUid string, classroomId string) (time.Time, error)
	UpdateUserPaymentClassroom(ctx context.Context, userUid string, classroomId string) (*models.User, error)
	UserChangeSchoolAndClassroom(ctx context.Context, studentId, schoolId, classroomId *string) error

	// TODO: replace l *[]*models.Model -> []*models.Model (without pointer) on all list models
	ClassroomsFindByIds(ctx context.Context, ids []string) ([]*models.Classroom, error)
	ClassroomsFindById(ctx context.Context, id string) (*models.Classroom, error)
	ClassroomsFindBy(ctx context.Context, f models.ClassroomFilterRequest) ([]*models.Classroom, int, error)
	ClassroomsUpdate(ctx context.Context, data *models.Classroom) (*models.Classroom, error)
	ClassroomsCreate(ctx context.Context, m *models.Classroom) (*models.Classroom, error)
	ClassroomsDelete(ctx context.Context, l []*models.Classroom) ([]*models.Classroom, error)
	ClassroomsDeleteStudent(ctx context.Context, userIds []string) error
	ClassroomsUpdateRelations(ctx context.Context, data *models.Classroom, model *models.Classroom) error
	ClassroomsLoadRelations(ctx context.Context, l *[]*models.Classroom, isDetail bool) error
	ClassroomsLoadSchool(ctx context.Context, l *[]*models.Classroom) error
	GetClassroomIdByName(
		ctx context.Context,
		dto models.GetClassroomIdByNameQueryDto,
	) (*string, error)
	ClassroomStudentsCountBySchool(ctx context.Context) ([]models.SchoolStudentsCount, error)

	UserLogsFindByIds(ctx context.Context, ids []string) ([]*models.UserLog, error)
	UserLogsFindById(ctx context.Context, id string) (*models.UserLog, error)
	UserLogsFindBy(ctx context.Context, f models.UserLogFilterRequest) ([]*models.UserLog, int, error)
	UserLogsUpdate(ctx context.Context, data models.UserLog) (*models.UserLog, error)
	UserLogsCreate(ctx context.Context, m models.UserLog) (*models.UserLog, error)
	UserLogsDelete(ctx context.Context, l []*models.UserLog) ([]*models.UserLog, error)
	UserLogsLoadRelations(ctx context.Context, l *[]*models.UserLog) error

	PeriodsFindByIds(ctx context.Context, ids []string) ([]*models.Period, error)
	PeriodsFindById(ctx context.Context, id string) (*models.Period, error)
	PeriodsListFilters(ctx context.Context, f models.PeriodFilterRequest) ([]*models.Period, int, error)
	PeriodsUpdate(ctx context.Context, data *models.Period) (*models.Period, error)
	PeriodsCreate(ctx context.Context, m *models.Period) (*models.Period, error)
	PeriodsDelete(ctx context.Context, l []*models.Period) ([]*models.Period, error)
	PeriodsUpdateRelations(ctx context.Context, data *models.Period, model *models.Period) error
	PeriodsLoadRelations(ctx context.Context, l *[]*models.Period) error

	PeriodGradesFindByIds(ctx context.Context, ids []string) ([]models.PeriodGrade, error)
	PeriodGradesFindById(ctx context.Context, id string) (models.PeriodGrade, error)
	PeriodGradesFindBy(ctx context.Context, f models.PeriodGradeFilterRequest) ([]*models.PeriodGrade, int, error)
	PeriodGradesUpdate(ctx context.Context, data *models.PeriodGrade) (models.PeriodGrade, error)
	PeriodGradesUpdateBatch(ctx context.Context, data []models.PeriodGrade) error

	PeriodGradesCreate(ctx context.Context, m *models.PeriodGrade) (models.PeriodGrade, error)
	PeriodGradesDelete(ctx context.Context, l []*models.PeriodGrade) ([]*models.PeriodGrade, error)
	PeriodGradesFindOrCreate(ctx context.Context, data *models.PeriodGrade) (models.PeriodGrade, error)
	PeriodGradesUpdateOrCreate(ctx context.Context, data *models.PeriodGrade) (models.PeriodGrade, error)
	PeriodGradesUpdateValues(ctx context.Context, data models.PeriodGrade) (*models.PeriodGrade, error)
	PeriodGradesLoadRelations(ctx context.Context, l *[]*models.PeriodGrade) error
	PeriodGradeByStudent(ctx context.Context, student_id string) ([]*models.PeriodGrade, error)
	DeletePeriodGradeByStudentAndSubjects(ctx context.Context, student_id string, subjectIds []string) error

	SchoolsFindByIds(ctx context.Context, ids []string) ([]*models.School, error)
	SchoolsFindByCode(ctx context.Context, codes []string) ([]*models.School, error)
	SchoolsFindById(ctx context.Context, id string) (*models.School, error)
	SchoolsFindBy(ctx context.Context, f models.SchoolFilterRequest) ([]*models.School, int, error)
	SchoolUpdate(ctx context.Context, data *models.School) (*models.School, error)
	SchoolCreate(ctx context.Context, m *models.School) (*models.School, error)
	SchoolDelete(ctx context.Context, l []*models.School) ([]*models.School, error)
	SchoolUpdateRelations(ctx context.Context, data *models.School, model *models.School) error
	SchoolsLoadRelations(ctx context.Context, l *[]*models.School) error
	SchoolsLoadParents(ctx context.Context, l *[]*models.School) error

	SchoolSettingsGet(ctx context.Context, schoolIds []string) ([]models.SchoolSetting, error)
	SchoolSettingsUpdate(ctx context.Context, schoolId string, values []models.SchoolSettingRequest) error
	SchoolSettingsUpdateQuery(ctx context.Context, m []models.SchoolSettingRequest) (string, []interface{})

	SubjectsFindByIds(ctx context.Context, ids []string) ([]*models.Subject, error)
	SubjectsFindById(ctx context.Context, id string) (*models.Subject, error)
	SubjectsListFilters(ctx context.Context, f *models.SubjectFilterRequest) ([]*models.Subject, int, error)
	SubjectsUpdate(ctx context.Context, data *models.Subject) (*models.Subject, error)
	SubjectsCreate(ctx context.Context, m *models.Subject) (*models.Subject, error)
	SubjectsDelete(ctx context.Context, l []*models.Subject) ([]*models.Subject, error)
	SubjectsUpdateRelations(ctx context.Context, data *models.Subject, model *models.Subject)
	SubjectsLoadRelations(ctx context.Context, l *[]*models.Subject, isDetail bool) error
	SubjectsRatingByStudentWithPrev(ctx context.Context, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectRating, error)
	SubjectsRatingByStudent(ctx context.Context, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectRating, error)
	SubjectsPercentByStudent(ctx context.Context, studentId string, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectPercent, error)
	SubjectsPercentByStudentWithPrev(ctx context.Context, studentId string, classroomId string, startDate time.Time, endDate time.Time) ([]models.SubjectPercent, error)
	SubjectsPercents(ctx context.Context, subjectIds []string, startDate time.Time, endDate time.Time) ([]models.DashboardSubjectsPercent, error)
	SubjectsPercentsBySchool(ctx context.Context, schoolIds []string, startDate time.Time, endDate time.Time) ([]models.DashboardSubjectsPercentBySchool, error)
	SubjectsPeriodGradeFinished(ctx context.Context, schoolIds []string, periodNumber int) ([]models.SubjectPeriodGradeFinished, error)
	SubjectsGradeStrike(ctx context.Context, classroomId, studentId string) ([]models.SubjectLessonGrades, error)
	SubjectGrades(ctx context.Context, studentId string, startDate time.Time, endDate time.Time) ([]models.SubjectGrade, error)
	StudentRatingBySchool(ctx context.Context, schoolId string, startDate time.Time, endDate time.Time) ([]*models.User, []int, error)
	MapOldSubjectsToNewSubjectsInPeriodGrade(ctx context.Context, student_id string, periodGrades []*models.PeriodGrade, oldSubjects []*models.Subject, newSubjects []*models.Subject) error

	SubjectExamFindByIds(ctx context.Context, ids []string) ([]*models.SubjectExam, error)
	SubjectExamFindById(ctx context.Context, id string) (*models.SubjectExam, error)
	SubjectExamsFindBy(ctx context.Context, f *models.SubjectExamFilterRequest) (l []*models.SubjectExam, total int, err error)
	SubjectExamUpdate(ctx context.Context, m *models.SubjectExam) (*models.SubjectExam, error)
	SubjectExamCreate(ctx context.Context, m *models.SubjectExam) (*models.SubjectExam, error)
	SubjectExamDelete(ctx context.Context, model []*models.SubjectExam) ([]*models.SubjectExam, error)
	SubjectExamLoadRelations(ctx context.Context, l *[]*models.SubjectExam) error
	SubjectsFindByClassroomId(ctx context.Context, classroomId string) ([]*models.Subject, error)

	TimetablesFindByIds(ctx context.Context, ids []string) ([]*models.Timetable, error)
	TimetableFindById(ctx context.Context, id string) (*models.Timetable, error)
	TimetablesFindBy(ctx context.Context, f models.TimetableFilterRequest) ([]*models.Timetable, int, error)
	TimetableUpdate(ctx context.Context, data *models.Timetable) (*models.Timetable, error)
	TimetableCreate(ctx context.Context, m *models.Timetable) (*models.Timetable, error)
	TimetablesDelete(ctx context.Context, l []*models.Timetable) ([]*models.Timetable, error)
	TimetableUpdateRelations(ctx context.Context, data *models.Timetable, model *models.Timetable)
	TimetablesLoadRelations(ctx context.Context, l *[]*models.Timetable) error

	ShiftsFindByIds(ctx context.Context, ids []string) ([]*models.Shift, error)
	ShiftsFindById(ctx context.Context, Id string) (*models.Shift, error)
	ShiftsFindBy(ctx context.Context, f models.ShiftFilterRequest) (shifts []*models.Shift, total int, err error)
	UpdateShift(ctx context.Context, model *models.Shift) (*models.Shift, error)
	CreateShift(ctx context.Context, m *models.Shift) (*models.Shift, error)
	DeleteShifts(ctx context.Context, items []*models.Shift) ([]*models.Shift, error)
	ShiftUpdateRelations(ctx context.Context, data *models.Shift, model *models.Shift)
	ShiftLoadRelations(ctx context.Context, l *[]*models.Shift) error

	LessonsFindByIds(ctx context.Context, ids []string) ([]models.Lesson, error)
	LessonsFindById(ctx context.Context, id string) (models.Lesson, error)
	LessonsFindBy(ctx context.Context, data models.LessonFilterRequest) ([]*models.Lesson, int, error)
	LessonsUpdateBy(ctx context.Context, data models.LessonFilterRequest, sets map[string]interface{}) (int, error)
	LessonsUpdate(ctx context.Context, model models.Lesson) (models.Lesson, error)
	LessonsCreate(ctx context.Context, model models.Lesson) (models.Lesson, error)
	LessonsDelete(ctx context.Context, l []*models.Lesson) ([]*models.Lesson, error)
	LessonsLoadRelations(ctx context.Context, l *[]*models.Lesson) error
	LessonsLoadSubject(ctx context.Context, l *[]*models.Lesson) error
	LessonsUpdateBatch(ctx context.Context, l []models.Lesson) error
	LessonsCreateBatch(ctx context.Context, l []models.Lesson) error
	LessonsDeleteBatch(ctx context.Context, l []string) error

	LessonsLikes(ctx context.Context, lessonID string, userID string) error
	LessonsLikesByUser(ctx context.Context, lessonID string, userID string) (bool, error)
	LessonsLikesThenUnlike(ctx context.Context, lessonID string, userID string) error
	LessonLikesLoadRelations(ctx context.Context, l *[]*models.LessonLikes) error
	LessonLikesCount(ctx context.Context, teacherId string, month time.Time) (int, error)

	AssignmentFindOrCreate(ctx context.Context, data models.Assignment) (models.Assignment, error)
	AssignmentFindById(ctx context.Context, id string) (models.Assignment, error)
	AssignmentUpdate(ctx context.Context, model models.Assignment) (models.Assignment, error)
	AssignmentUpdateOrCreate(ctx context.Context, data models.Assignment) (models.Assignment, error)
	AssignmentsFindBy(ctx context.Context, f models.AssignmentFilterRequest) ([]models.Assignment, int, error)
	AssignmentsFindByIds(ctx context.Context, ids []string) ([]models.Assignment, error)

	GradesFindByIds(ctx context.Context, ids []string) ([]models.Grade, error)
	GradesFindById(ctx context.Context, id string) (models.Grade, error)
	GradesFindBy(ctx context.Context, f models.GradeFilterRequest) ([]*models.Grade, int, error)
	GradesUpdate(ctx context.Context, data models.Grade) (models.Grade, error)
	GradesCreateOrUpdate(ctx context.Context, data models.Grade) (models.Grade, error)
	GradesCreate(ctx context.Context, model models.Grade) (models.Grade, error)
	GradesDelete(ctx context.Context, l []*models.Grade) ([]*models.Grade, error)
	GradesLoadRelations(ctx context.Context, l []*models.Grade) error
	GradesLoadRelationLessons(ctx context.Context, l []*models.Grade) error

	AbsentsFindByIds(ctx context.Context, ids []string) ([]models.Absent, error)
	AbsentsFindById(ctx context.Context, id string) (models.Absent, error)
	AbsentsFindBy(ctx context.Context, f models.AbsentFilterRequest) ([]*models.Absent, int, error)
	AbsentsUpdate(ctx context.Context, model models.Absent) (models.Absent, error)
	AbsentsCreateOrUpdate(ctx context.Context, data models.Absent) (models.Absent, error)
	AbsentsCreate(ctx context.Context, model models.Absent) (models.Absent, error)
	AbsentsDelete(ctx context.Context, l []*models.Absent) ([]*models.Absent, error)
	AbsentsLoadRelations(ctx context.Context, l []*models.Absent) error
	AbsentsLoadRelationsLessons(ctx context.Context, l []*models.Absent) error

	StudentNotesFindOrCreate(ctx context.Context, data *models.StudentNote) (models.StudentNote, error)
	StudentNotesUpdateOrCreate(ctx context.Context, data *models.StudentNote) (models.StudentNote, error)
	StudentNotesFindByIds(ctx context.Context, ids []string) ([]*models.StudentNote, error)
	StudentNoteFindById(ctx context.Context, id string) (*models.StudentNote, error)
	StudentNoteUpdate(ctx context.Context, model *models.StudentNote) (models.StudentNote, error)
	StudentNotesFindBy(ctx context.Context, f models.StudentNoteFilterRequest) ([]*models.StudentNote, int, error)

	SessionsSelect(ctx context.Context, f models.SessionFilter) ([]models.Session, error)
	SessionsClear(ctx context.Context, now time.Time) error
	SessionsCreate(ctx context.Context, m models.Session) (models.Session, error)
	SessionsDelete(ctx context.Context, f models.SessionFilter) error

	UserNotificationsFindBy(ctx context.Context, f models.UserNotificationFilterRequest) (userNotifications []*models.UserNotification, total int, err error)
	UserNotificationFindById(ctx context.Context, ID string) (*models.UserNotification, error)
	UserNotificationFindByIds(ctx context.Context, Ids []string) ([]*models.UserNotification, error)
	UserNotificationsUpdate(ctx context.Context, model models.UserNotification) (*models.UserNotification, error)
	UserNotificationsUpdateRead(ctx context.Context, ids []string) error
	UserNotificationsSelectTotalUnread(ctx context.Context, userId string, role string) (int, error)
	UserNotificationsLoadRelations(ctx context.Context, l *[]*models.UserNotification) error
	UserNotificationsCreateBatch(ctx context.Context, l []models.UserNotification) error
	UserNotificationsLoadRelationUser(l *[]*models.UserNotification) error

	NotificationsFindBy(ctx context.Context, f models.NotificationsFilterRequest) (notifications []*models.Notifications, total int, err error)
	NotificationFindById(ctx context.Context, ID string) (*models.Notifications, error)
	NotificationFindByIds(ctx context.Context, Ids []string) ([]*models.Notifications, error)
	NotificationCreate(ctx context.Context, model *models.Notifications) (*models.Notifications, error)
	NotificationsLoadRelations(ctx context.Context, l *[]*models.Notifications) error
	NotificationUpdate(ctx context.Context, model *models.Notifications) (*models.Notifications, error)
	NotificationDelete(ctx context.Context, ID string) error

	PaymentTransactionsFindByIds(ctx context.Context, ids []string) ([]*models.PaymentTransaction, error)
	PaymentTransactionsFindById(ctx context.Context, id string) (*models.PaymentTransaction, error)
	PaymentTransactionsFindBy(ctx context.Context, f models.PaymentTransactionFilterRequest) ([]*models.PaymentTransaction, int, map[string]int, map[string]int, error)
	PaymentTransactionUpdate(ctx context.Context, data *models.PaymentTransaction) (*models.PaymentTransaction, error)
	PaymentTransactionCreate(ctx context.Context, m *models.PaymentTransaction) (*models.PaymentTransaction, error)
	PaymentTransactionDelete(ctx context.Context, l []*models.PaymentTransaction) ([]*models.PaymentTransaction, error)
	PaymentTransactionsLoadRelations(ctx context.Context, l *[]*models.PaymentTransaction) error
	PaymentsTransactionsCountBySchool(ctx context.Context, f models.PaymentTransactionFilterRequest) ([]models.PaymentTransactionsCount, error)

	TopicsFindBy(ctx context.Context, f models.TopicsFilterRequest) (topics []*models.Topics, total int, err error)
	TopicsFindById(ctx context.Context, ID string) (*models.Topics, error)
	TopicsFindByIds(ctx context.Context, Ids []string) ([]*models.Topics, error)
	TopicsCreate(ctx context.Context, model *models.Topics) (*models.Topics, error)
	TopicsUpdate(ctx context.Context, model *models.Topics) (*models.Topics, error)
	TopicsDelete(ctx context.Context, items []*models.Topics) ([]*models.Topics, error)
	TopicsLoadRelations(ctx context.Context, l *[]*models.Topics) error

	BookFindBy(ctx context.Context, f models.BookFilterRequest) (books []*models.Book, total int, err error)
	BookFindById(ctx context.Context, ID string) (*models.Book, error)
	BookFindByIds(ctx context.Context, Ids []string) ([]*models.Book, error)
	BookGetAuthors(ctx context.Context) ([]string, error)
	BookCreate(ctx context.Context, model *models.Book) (*models.Book, error)
	BookUpdate(ctx context.Context, model *models.Book) (*models.Book, error)
	BookDelete(ctx context.Context, items []*models.Book) ([]*models.Book, error)

	BaseSubjectsFindBy(ctx context.Context, f models.BaseSubjectsFilterRequest) (baseSubjects []*models.BaseSubjects, total int, err error)
	BaseSubjectsFindById(ctx context.Context, ID string) (*models.BaseSubjects, error)
	BaseSubjectsFindByIds(ctx context.Context, Ids []string) ([]*models.BaseSubjects, error)
	BaseSubjectsCreate(ctx context.Context, model *models.BaseSubjects) (*models.BaseSubjects, error)
	BaseSubjectsUpdate(ctx context.Context, model *models.BaseSubjects) (*models.BaseSubjects, error)
	BaseSubjectsDelete(ctx context.Context, items []*models.BaseSubjects) ([]*models.BaseSubjects, error)
	BaseSubjectsLoadRelations(ctx context.Context, l *[]*models.BaseSubjects) error

	SmsSendersFindBy(ctx context.Context, f models.SmsSenderFilterRequest) (smsSenders []*models.SmsSender, total int, err error)
	SmsSendersFindById(ctx context.Context, ID string) (*models.SmsSender, error)
	SmsSendersFindByIds(ctx context.Context, IDs []string) ([]*models.SmsSender, error)
	SmsSenderCreate(ctx context.Context, model *models.SmsSender) (*models.SmsSender, error)

	ContactItemsFindBy(ctx context.Context, f models.ContactItemsFilterRequest) (contactItems []*models.ContactItems, total int, err error)
	ContactItemsFindById(ctx context.Context, Id string) (*models.ContactItems, error)
	ContactItemsFindByIds(ctx context.Context, Ids []string) ([]*models.ContactItems, error)
	ContactItemUpdate(ctx context.Context, model *models.ContactItems) (*models.ContactItems, error)
	ContactItemCreate(ctx context.Context, model *models.ContactItems) (*models.ContactItems, error)
	ContactItemsDelete(ctx context.Context, items []*models.ContactItems) ([]*models.ContactItems, error)
	ContactItemLoadRelations(ctx context.Context, l *[]*models.ContactItems, isDetail bool) error
	ContactItemsCountByType(ctx context.Context, f models.ContactItemsFilterRequest) ([]models.ContactItemsCount, error)

	MessageGroupsFindById(ctx context.Context, id string) (models.MessageGroup, error)
	MessageGroupsFindBy(ctx context.Context, f models.GetMessageGroupsRequest) ([]*models.MessageGroup, int, error)
	CreateMessageGroupCommand(ctx context.Context, m models.MessageGroup) (models.MessageGroup, error)

	GetMessageReadsQuery(ctx context.Context, dto models.GetMessageReadsQueryDto) (map[string]int, error)
	GetMessagesQuery(ctx context.Context, dto models.GetMessagesQueryDto) ([]*models.Message, error)
	CreateMessageCommand(ctx context.Context, message models.Message) (models.Message, error)
	CreateMessageReadsCommand(ctx context.Context, messageReads []models.MessageRead) error
	LoadMessagesWithParents(ctx context.Context, l *[]*models.Message) error

	ReportsFindBy(ctx context.Context, f models.ReportsFilterRequest) (reports []*models.Reports, total int, err error)
	ReportsFindById(ctx context.Context, Id string) (*models.Reports, error)
	ReportsFindByIds(ctx context.Context, Ids []string) ([]*models.Reports, error)
	ReportsCreate(ctx context.Context, model *models.Reports) (*models.Reports, error)
	ReportsDelete(ctx context.Context, items []*models.Reports) ([]*models.Reports, error)
	ReportsUpdate(ctx context.Context, model *models.Reports) (*models.Reports, error)

	ReportItemsFindBy(ctx context.Context, f models.ReportItemsFilterRequest) (reportItems []*models.ReportItems, total int, err error)
	ReportItemsFindById(ctx context.Context, ID string) (*models.ReportItems, error)
	ReportItemsFindByIds(ctx context.Context, Ids []string) ([]*models.ReportItems, error)
	ReportItemsCreateBatch(ctx context.Context, l []models.ReportItems) error
	ReportItemsCreate(ctx context.Context, model models.ReportItems) (*models.ReportItems, error)
	ReportItemsLoadRelations(ctx context.Context, l *[]*models.ReportItems) error
	ReportItemsUpdate(ctx context.Context, model models.ReportItems) (*models.ReportItems, error)

	SettingsFindById(ctx context.Context, id string) (model *models.Settings, err error)
	SettingsUpsert(ctx context.Context, data *models.Settings) (model *models.Settings, err error)
	SettingsFindBy(ctx context.Context, opts map[string]interface{}) (list []*models.Settings, err error)

	TeacherExcusesFindById(ctx context.Context, id string, loadRelations bool) (model *models.TeacherExcuse, err error)
	TeacherExcuseInsert(ctx context.Context, data *models.TeacherExcuse) (model *models.TeacherExcuse, err error)
	TeacherExcuseUpdate(ctx context.Context, data *models.TeacherExcuse) (model *models.TeacherExcuse, err error)
	TeacherExcusesFindBy(ctx context.Context, opts map[string]interface{}) (list *models.TeacherExcuses, err error)
	TeacherExcusesDelete(ctx context.Context, ids []string) (list *models.TeacherExcuses, err error)

	SchoolTransfersFindById(ctx context.Context, id string) (model *models.SchoolTransfer, err error)
	SchoolTransfersFindBy(ctx context.Context, opts map[string]interface{}) (list *models.SchoolTransfers, err error)
	SchoolTransfersUpdate(ctx context.Context, data *models.SchoolTransfer) (model *models.SchoolTransfer, err error)
	SchoolTransfersInsert(ctx context.Context, data *models.SchoolTransfer) (model *models.SchoolTransfer, err error)
	SchoolTransfersDelete(ctx context.Context, ids []string) (list *models.SchoolTransfers, err error)
	SchoolTransfersLoadRelations(ctx context.Context, list *models.SchoolTransfers) error
}
