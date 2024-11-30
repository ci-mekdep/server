package pgx

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/utils"
	"go.elastic.co/apm/module/apmpgx/v2"
)

// FIELDS
const sqlPaymentTransactionNewFields = `pt.unit_price, pt.school_price, pt.center_price, pt.discount_price, pt.used_days, pt.used_days_price, pt.school_months, pt.center_months, pt.school_classroom_uids, pt.center_classroom_uids,`
const sqlPaymentTransactionFields = `pt.uid, pt.school_uid, pt.payer_uid, pt.user_uids, pt.tariff_type, pt.status, pt.amount, pt.original_amount, pt.bank_type, pt.card_name, pt.order_number, pt.order_url, pt.comment, pt.system_comment, ` + sqlPaymentTransactionNewFields + ` pt.updated_at, pt.created_at`

// CRUD
const sqlPaymentTransactionSelect = `select ` + sqlPaymentTransactionFields + ` from payment_transactions pt where pt.uid = ANY($1::uuid[])`

// TODO: need to do OVER()
const sqlPaymentTransactionSelectMany = `SELECT ` + sqlPaymentTransactionFields + `, 
    COUNT(*) OVER() AS total, 
    SUM(CASE WHEN pt.bank_type = 'rysgalbank' THEN pt.amount ELSE 0 END) OVER() AS total_amount_rysgalbank,
    SUM(CASE WHEN pt.bank_type = 'halkbank' THEN pt.amount ELSE 0 END) OVER() AS total_amount_halkbank,
    SUM(CASE WHEN pt.bank_type = 'senagatbank' THEN pt.amount ELSE 0 END) OVER() AS total_amount_senagatbank,
    SUM(CASE WHEN pt.bank_type = 'tfeb' THEN pt.amount ELSE 0 END) OVER() AS total_amount_tfeb,
	COUNT(CASE WHEN pt.bank_type = 'rysgalbank' THEN 1 END) OVER() AS total_transactions_rysgalbank,
	COUNT(CASE WHEN pt.bank_type = 'halkbank' THEN 1 END) OVER() AS total_transactions_halkbank,
	COUNT(CASE WHEN pt.bank_type = 'senagatbank' THEN 1 END) OVER() AS total_transactions_senagatbank,
	COUNT(CASE WHEN pt.bank_type = 'tfeb' THEN 1 END) OVER() AS total_transactions_tfeb
FROM payment_transactions pt 
WHERE pt.uid=pt.uid 
LIMIT $1 OFFSET $2`
const sqlPaymentTransactionInsert = `insert into payment_transactions`
const sqlPaymentTransactionUpdate = `update payment_transactions set uid=uid`
const sqlPaymentTransactionDelete = `delete from payment_transactions where pt.uid=ANY($1::uuid[])`

// JOINS
const sqlPaymentTransactionSchool = `select ` + sqlSchoolFields + `, pt.uid from payment_transactions pt
	right join schools s on (s.uid=pt.school_uid) where pt.uid = ANY($1::uuid[])`
const sqlPaymentTransactionPayer = `select ` + sqlUserFields + `, pt.uid from payment_transactions pt
	right join users u on (u.uid=pt.payer_uid) where pt.uid = ANY($1::uuid[])`
const sqlPaymentTransactionStudents = `select ` + sqlUserFields + `, pt.uid from payment_transactions pt
	right join users u on (u.uid=ANY(pt.user_uids::uuid[])) where pt.uid = ANY($1::uuid[])`

const sqlPaymentTransactionSchoolClassrooms = `select ` + sqlClassroomFields + `, pt.uid from payment_transactions pt
	right join classrooms c on (c.uid=ANY(pt.school_classroom_uids::uuid[])) where pt.uid = ANY($1::uuid[])`

const sqlPaymentTransactionCenterClassrooms = `select ` + sqlClassroomFields + `, pt.uid from payment_transactions pt
	right join classrooms c on (c.uid=ANY(pt.center_classroom_uids::uuid[])) where pt.uid = ANY($1::uuid[])`

// COUNTS
const sqlPaymentsCountBySchool = `SELECT
	s.code,
	COUNT(*) as total_count,
	SUM(CASE WHEN pt.bank_type = 'halkbank' THEN 1 ELSE 0 END) AS halkbank_count,
	SUM(CASE WHEN pt.bank_type = 'senagatbank' THEN 1 ELSE 0 END) AS senagatbank_count,
	SUM(CASE WHEN pt.bank_type = 'rysgalbank' THEN 1 ELSE 0 END) AS rysgalbank_count,
	SUM(CASE WHEN pt.bank_type = 'tfeb' THEN 1 ELSE 0 END) AS tfeb_count
FROM payment_transactions pt 
JOIN schools s ON pt.school_uid=s.uid 
WHERE pt.uid=pt.uid
GROUP BY s.code, pt.school_uid`

func scanPaymentTransaction(rows pgx.Row, m *models.PaymentTransaction, addColumns ...interface{}) (err error) {
	err = rows.Scan(parseColumnsForScan(m, addColumns...)...)
	return
}

func (d *PgxStore) PaymentsTransactionsCountBySchool(ctx context.Context, f models.PaymentTransactionFilterRequest) ([]models.PaymentTransactionsCount, error) {
	items := []models.PaymentTransactionsCount{}
	args := []interface{}{}
	qs, args := PaymentTransactionsListBuildQuery(f, args, sqlPaymentsCountBySchool)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, qs, args...)
		for rows.Next() {
			var item models.PaymentTransactionsCount
			if err := rows.Scan(&item.SchoolCode, &item.TotalCount, &item.HalkbankCount, &item.SenagatbankCount, &item.RysgalbankCount, &item.TfebCount); err != nil {
				return err
			}
			items = append(items, item)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return items, nil
}

func (d *PgxStore) PaymentTransactionsFindByIds(ctx context.Context, ids []string) ([]*models.PaymentTransaction, error) {
	l := []*models.PaymentTransaction{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPaymentTransactionSelect, (ids))
		for rows.Next() {
			u := models.PaymentTransaction{}
			err := scanPaymentTransaction(rows, &u)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			l = append(l, &u)
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func (d *PgxStore) PaymentTransactionsFindById(ctx context.Context, id string) (*models.PaymentTransaction, error) {
	l, err := d.PaymentTransactionsFindByIds(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	if len(l) < 1 {
		return nil, pgx.ErrNoRows
	}
	return l[0], err
}

func (d *PgxStore) PaymentTransactionsFindBy(ctx context.Context, f models.PaymentTransactionFilterRequest) ([]*models.PaymentTransaction, int, map[string]int, map[string]int, error) {
	if f.Limit == nil {
		f.Limit = new(int)
		*f.Limit = 12
	}
	if f.Offset == nil {
		f.Offset = new(int)
		*f.Offset = 0
	}
	args := []interface{}{f.Limit, f.Offset}
	qs, args := PaymentTransactionsListBuildQuery(f, args, sqlPaymentTransactionSelectMany)
	l := []*models.PaymentTransaction{}
	var total int
	totalAmount := make(map[string]int)
	totalTransactions := make(map[string]int)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		apmpgx.Instrument(tx.Conn().Config())

		rows, err := tx.Query(ctx, qs, args...)
		var rysgalBankAmount, halkBankAmount, senagatBankAmount, tfebAmount float64
		var rysgalBankTransactions, halkBankTransactions, senagatBankTransactions, tfebTransactions int
		for rows.Next() {
			sub := models.PaymentTransaction{}
			err = scanPaymentTransaction(rows, &sub, &total,
				&rysgalBankAmount, &halkBankAmount, &senagatBankAmount, &tfebAmount,
				&rysgalBankTransactions, &halkBankTransactions, &senagatBankTransactions, &tfebTransactions)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			l = append(l, &sub)
			totalAmount["rysgalbank"] = int(rysgalBankAmount)
			totalAmount["halkbank"] = int(halkBankAmount)
			totalAmount["senagatbank"] = int(senagatBankAmount)
			totalAmount["tfeb"] = int(tfebAmount)
		}

		totalTransactions["rysgalbank"] += rysgalBankTransactions
		totalTransactions["halkbank"] += halkBankTransactions
		totalTransactions["senagatbank"] += senagatBankTransactions
		totalTransactions["tfeb"] += tfebTransactions

		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, 0, nil, nil, err
	}
	return l, total, totalAmount, totalTransactions, nil
}

func (d *PgxStore) PaymentTransactionCreate(ctx context.Context, data *models.PaymentTransaction) (*models.PaymentTransaction, error) {
	qs, args := PaymentTransactionsCreateQuery(data)
	qs += " RETURNING uid"
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		err = tx.QueryRow(ctx, qs, args...).Scan(&data.ID)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.PaymentTransactionsFindById(ctx, data.ID)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return editModel, nil
}

func (d *PgxStore) PaymentTransactionUpdate(ctx context.Context, data *models.PaymentTransaction) (*models.PaymentTransaction, error) {
	// origModel := d.UsersFindById(strconv.Itoa(int(model.ID)))
	qs, args := PaymentTransactionsUpdateQuery(data)
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, qs, args...)
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	editModel, err := d.PaymentTransactionsFindById(ctx, data.ID)
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return editModel, nil
}

func (d *PgxStore) PaymentTransactionDelete(ctx context.Context, l []*models.PaymentTransaction) ([]*models.PaymentTransaction, error) {
	ids := []string{}
	for _, i := range l {
		ids = append(ids, i.ID)
	}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		_, err = tx.Query(ctx, sqlPaymentTransactionDelete, (ids))
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}
	return l, nil
}

func PaymentTransactionsCreateQuery(m *models.PaymentTransaction) (string, []interface{}) {
	args := []interface{}{}
	cols := ""
	vals := ""
	q := PaymentTransactionAtomicQuery(m, true)
	for k, v := range q {
		args = append(args, v)
		cols += ", " + k
		vals += ", $" + strconv.Itoa(len(args))
	}
	qs := sqlPaymentTransactionInsert + " (" + strings.Trim(cols, ", ") + ") VALUES (" + strings.Trim(vals, ", ") + ")"
	return qs, args
}

func PaymentTransactionsUpdateQuery(m *models.PaymentTransaction) (string, []interface{}) {
	args := []interface{}{}
	sets := ""
	q := PaymentTransactionAtomicQuery(m, false)
	for k, v := range q {
		args = append(args, v)
		sets += ", " + k + "=$" + strconv.Itoa(len(args))
	}
	args = append(args, m.ID)
	qs := strings.ReplaceAll(sqlPaymentTransactionUpdate, "set uid=uid", "set uid=uid "+sets+" ") + "where uid=$" + strconv.Itoa(len(args))
	return qs, args
}

func PaymentTransactionAtomicQuery(m *models.PaymentTransaction, isCreate bool) map[string]interface{} {
	q := map[string]interface{}{}
	if m.SchoolId != "" {
		q["school_uid"] = m.SchoolId
	}
	if m.PayerId != "" {
		q["payer_uid"] = m.PayerId
	}
	if m.UserIds != nil {
		q["user_uids"] = m.UserIds
	}
	if m.TariffType != "" {
		q["tariff_type"] = m.TariffType
	}
	if m.Status != "" {
		q["status"] = m.Status
	}
	q["amount"] = m.Amount
	q["original_amount"] = m.OriginalAmount
	if m.BankType != "" {
		q["bank_type"] = m.BankType
	}
	if m.CardName != nil {
		q["card_name"] = m.CardName
	}
	if m.OrderNumber != nil {
		q["order_number"] = m.OrderNumber
	}
	if m.OrderUrl != nil {
		q["order_url"] = m.OrderUrl
	}
	if m.Comment != nil {
		q["comment"] = m.Comment
	}
	if m.SystemComment != nil {
		q["system_comment"] = m.SystemComment
	}
	if m.UnitPrice != nil {
		q["unit_price"] = m.UnitPrice
	}
	if m.SchoolPrice != nil {
		q["school_price"] = m.SchoolPrice
	}
	if m.CenterPrice != nil {
		q["center_price"] = m.CenterPrice
	}
	if m.DiscountPrice != nil {
		q["discount_price"] = m.DiscountPrice
	}
	if m.UsedDays != nil {
		q["used_days"] = m.UsedDays
	}
	if m.UsedDaysPrice != nil {
		q["used_days_price"] = m.UsedDaysPrice
	}
	if m.SchoolMonths != 0 {
		q["school_months"] = m.SchoolMonths
	}
	if m.CenterMonths != 0 {
		q["center_months"] = m.CenterMonths
	}
	if m.SchoolClassroomIds != nil {
		q["school_classroom_uids"] = m.SchoolClassroomIds
	}
	if m.CenterClassroomIds != nil {
		q["center_classroom_uids"] = m.CenterClassroomIds
	}
	if isCreate {
		q["created_at"] = time.Now()
	}
	q["updated_at"] = time.Now()
	return q
}

func PaymentTransactionsListBuildQuery(f models.PaymentTransactionFilterRequest, args []interface{}, qs string) (string, []interface{}) {
	var wheres string = ""

	if f.ID != nil {
		args = append(args, *f.ID)
		wheres += " and pt.uid=$" + strconv.Itoa(len(args))
	}
	if f.Ids != nil {
		args = append(args, *f.Ids)
		wheres += " and pt.uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.SchoolId != nil && *f.SchoolId != "" {
		args = append(args, *f.SchoolId)
		wheres += " and pt.school_uid=$" + strconv.Itoa(len(args))
	}
	if f.SchoolIds != nil {
		args = append(args, *f.SchoolIds)
		wheres += " and pt.school_uid = ANY($" + strconv.Itoa(len(args)) + "::uuid[])"
	}
	if f.PayerId != nil {
		args = append(args, *f.PayerId)
		wheres += " and pt.payer_uid=$" + strconv.Itoa(len(args))
	}
	if f.StudentId != nil {
		args = append(args, *f.StudentId)
		wheres += " and $" + strconv.Itoa(len(args)) + "=ANY(pt.user_uids)"
	}
	if f.TariffType != nil {
		args = append(args, *f.TariffType)
		wheres += " and pt.tariff_type=$" + strconv.Itoa(len(args))
	}
	if f.BankType != nil {
		args = append(args, *f.BankType)
		wheres += " and pt.bank_type=$" + strconv.Itoa(len(args))
	}
	if f.Status != nil {
		args = append(args, *f.Status)
		wheres += " and pt.status=$" + strconv.Itoa(len(args))
	}
	if f.StartDate != nil {
		args = append(args, *f.StartDate)
		wheres += " and pt.created_at >= $" + strconv.Itoa(len(args))
	}
	if f.EndDate != nil {
		args = append(args, *f.EndDate)
		wheres += " and pt.created_at <= $" + strconv.Itoa(len(args))
	}
	if qs == sqlPaymentTransactionSelectMany {
		wheres += " group by pt.uid "
		wheres += " order by pt.created_at desc"
	}
	qs = strings.ReplaceAll(qs, "pt.uid=pt.uid", "pt.uid=pt.uid "+wheres+" ")
	return qs, args
}

func (d *PgxStore) PaymentTransactionsLoadRelations(ctx context.Context, l *[]*models.PaymentTransaction) error {
	ids := []string{}
	for _, m := range *l {
		ids = append(ids, m.ID)
	}
	if len(ids) < 1 {
		return nil
	}
	// load admin
	if rs, err := d.PaymentTransactionsLoadSchool(ctx, ids); err == nil {
		var schoolParents []*models.School
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					schoolParents = append(schoolParents, r.Relation)
					m.School = r.Relation
				}
			}
		}
		err = d.SchoolsLoadParents(ctx, &schoolParents)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	// load parent
	if rs, err := d.PaymentTransactionsLoadPayer(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID {
					m.Payer = r.Relation
				}
			}
		}
	} else {
		return err
	}
	// load school classrooms
	if rs, err := d.PaymentTransactionsLoadSchoolClassrooms(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if m.Classrooms == nil {
					m.Classrooms = []*models.Classroom{}
				}
				if r.ID == m.ID && r.Relation != nil {
					m.Classrooms = append(m.Classrooms, r.Relation)
				}
			}
		}
	} else {
		return err
	}
	// load center classrooms
	if rs, err := d.PaymentTransactionsLoadCenterClassrooms(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if m.Classrooms == nil {
					m.Classrooms = []*models.Classroom{}
				}
				if r.ID == m.ID && r.Relation != nil {
					m.Classrooms = append(m.Classrooms, r.Relation)
				}
			}
		}
	} else {
		return err
	}

	// load students
	for _, m := range *l {
		m.Students = []models.User{}
	}
	if rs, err := d.PaymentTransactionsLoadStudents(ctx, ids); err == nil {
		for _, r := range rs {
			for _, m := range *l {
				if r.ID == m.ID && r.Relation != nil {
					m.Students = append(m.Students, *r.Relation)
				}
			}
		}
	} else {
		return err
	}
	return nil
}

type PaymentTransactionsLoadSchoolClassroomsItem struct {
	ID       string
	Relation *models.Classroom
}

func (d *PgxStore) PaymentTransactionsLoadSchoolClassrooms(ctx context.Context, ids []string) ([]PaymentTransactionsLoadSchoolClassroomsItem, error) {
	res := []PaymentTransactionsLoadSchoolClassroomsItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPaymentTransactionSchoolClassrooms, ids)
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, PaymentTransactionsLoadSchoolClassroomsItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type PaymentTransactionsLoadCenterClassroomsItem struct {
	ID       string
	Relation *models.Classroom
}

func (d *PgxStore) PaymentTransactionsLoadCenterClassrooms(ctx context.Context, ids []string) ([]PaymentTransactionsLoadCenterClassroomsItem, error) {
	res := []PaymentTransactionsLoadCenterClassroomsItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPaymentTransactionCenterClassrooms, ids)
		for rows.Next() {
			sub := models.Classroom{}
			pid := ""
			err = scanClassroom(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, PaymentTransactionsLoadCenterClassroomsItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type PaymentTransactionsLoadStudentsItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) PaymentTransactionsLoadStudents(ctx context.Context, ids []string) ([]PaymentTransactionsLoadStudentsItem, error) {
	res := []PaymentTransactionsLoadStudentsItem{}

	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPaymentTransactionStudents, ids)
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				utils.LoggerDesc("Scan error").Error(err)
				return err
			}
			res = append(res, PaymentTransactionsLoadStudentsItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
		return nil, err
	}

	return res, nil
}

type PaymentTransactionsLoadSchoolItem struct {
	ID       string
	Relation *models.School
}

func (d *PgxStore) PaymentTransactionsLoadSchool(ctx context.Context, ids []string) ([]PaymentTransactionsLoadSchoolItem, error) {
	res := []PaymentTransactionsLoadSchoolItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPaymentTransactionSchool, (ids))
		for rows.Next() {
			sub := models.School{}
			pid := ""
			err = scanSchool(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, PaymentTransactionsLoadSchoolItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
	}
	return res, nil
}

type PaymentTransactionLoadPayerItem struct {
	ID       string
	Relation *models.User
}

func (d *PgxStore) PaymentTransactionsLoadPayer(ctx context.Context, ids []string) ([]PaymentTransactionLoadPayerItem, error) {
	res := []PaymentTransactionLoadPayerItem{}
	err := d.runQuery(ctx, func(tx *pgxpool.Conn) (err error) {
		rows, err := tx.Query(ctx, sqlPaymentTransactionPayer, (ids))
		for rows.Next() {
			sub := models.User{}
			pid := ""
			err = scanUser(rows, &sub, &pid)
			if err != nil {
				return err
			}
			res = append(res, PaymentTransactionLoadPayerItem{ID: pid, Relation: &sub})
		}
		return
	})
	if err != nil {
		utils.LoggerDesc("Query error").Error(err)
	}
	return res, nil
}
