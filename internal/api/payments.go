package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/app/app_validation"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
)

func PaymentRoutes(api *gin.RouterGroup) {
	rs := api.Group("/payment")
	{
		rs.POST("/checkout", PaymentCheckout)
		rs.GET("/checkout", PaymentUpdate)
		rs.GET("/history", PaymentTransactions)
		rs.GET("/calc", PaymentCalculationGet)
		rs.POST("/calc", PaymentCalculationPost)
		rs.GET("", PaymentTransactionsList)
		rs.GET(":id", PaymentTransactionDetail)
	}
}

func paymentTransactionsListQuery(ses *utils.Session, data models.PaymentTransactionFilterRequest) ([]*models.PaymentTransactionResponse, int, map[string]int, map[string]int, error) {
	return app.PaymentTransactionsList(ses, data)
}

func paymentTransactionAvailableCheck(ses *utils.Session, data models.PaymentTransactionFilterRequest) (bool, error) {
	_, t, _, _, err := paymentTransactionsListQuery(ses, data)
	if err != nil {
		return false, app.ErrRequired.SetKey("id")
	}
	return t > 0, nil
}

func PaymentTransactionsList(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminPayments, func(user *models.User) (err error) {
		r := models.PaymentTransactionFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}
		res, total, totalAmount, totalTransactions, err := paymentTransactionsListQuery(&ses, r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"transactions":       res,
			"total":              total,
			"total_amount":       totalAmount,
			"total_transactions": totalTransactions,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PaymentTransactionDetail(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermAdminPayments, func(user *models.User) error {
		id := c.Param("id")
		if ok, err := paymentTransactionAvailableCheck(&ses, models.PaymentTransactionFilterRequest{ID: &id}); err != nil {
			return err
		} else if !ok {
			return app.ErrNotfound
		}
		m, err := app.PaymentTransactionDetail(&ses, id)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"transaction": m,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PaymentCheckout(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermPayments, func(user *models.User) (err error) {
		r := models.PaymentCheckoutRequest{}
		if err := BindAny(c, &r); err != nil {
			return err
		}
		err = app_validation.ValidatePaymentsCreate(&ses, r)
		if err != nil {
			return err
		}
		res, err := app.Ap().PaymentCheckout(&ses, user, r)
		if err != nil {
			return err
		}
		userLog(models.UserLog{
			SchoolId:          ses.GetSchoolId(),
			SessionId:         ses.GetSessionId(),
			UserId:            user.ID,
			SubjectId:         &res.ID,
			Subject:           models.LogSubjectPayment,
			SubjectAction:     models.LogActionCreate,
			SubjectProperties: r,
		})
		Success(c, gin.H{
			"transaction": res,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PaymentUpdate(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermPayments, func(user *models.User) error {
		tr, err := store.Store().PaymentTransactionsFindById(ses.Context(), c.Query("id"))
		if err != nil {
			return err
		}
		if tr.PayerId != user.ID {
			return app.ErrNotfound
		}
		err = store.Store().PaymentTransactionsLoadRelations(ses.Context(), &[]*models.PaymentTransaction{tr})
		if err != nil {
			return err
		}
		statusOk, err := app.PaymentHandleUpdate(tr, 2)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"status": statusOk,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PaymentTransactions(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermPayments, func(user *models.User) error {
		r := models.PaymentTransactionFilterRequest{}
		if errMsg, errKey := BindAndValidate(c, &r); errMsg != "" || errKey != "" {
			return app.NewAppError(errMsg, errKey, "")
		}

		r.PayerId = &user.ID

		res, total, err := app.Ap().PaymentTransactions(r)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"transactions": res,
			"total":        total,
		})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PaymentCalculationGet(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckRead(&ses, app.PermPayments, func(user *models.User) error {
		tariffs, err := app.PaymentCalculationGet(&ses)
		if err != nil {
			return err
		}
		Success(c, gin.H{"tariffs": tariffs})
		return nil
	})
	if err != nil {
		handleError(c, err)
		return
	}
}

func PaymentCalculationPost(c *gin.Context) {
	ses := utils.InitSession(c)
	err := app.Ap().UserActionCheckWrite(&ses, app.PermPayments, func(u *models.User) (err error) {
		req := models.PaymentCheckoutRequest{}
		if err := BindAny(c, &req); err != nil {
			return err
		}
		err = app_validation.ValidatePaymentsCreate(&ses, req)
		if err != nil {
			return err
		}
		res, err := app.PaymentCalculationPost(&ses, req)
		if err != nil {
			return err
		}
		Success(c, gin.H{
			"calc": res,
		})
		return
	})
	if err != nil {
		handleError(c, err)
		return
	}
}
