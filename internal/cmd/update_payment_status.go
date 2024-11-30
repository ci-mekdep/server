package cmd

import (
	"errors"
	"log"
	"os"
	"strconv"

	apiutils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/app"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
	"github.com/spf13/cobra"
)

func UpdatePaymentStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update-payment-status",
		Run: UpdatePaymentStatusRun,
	}
}
func UpdatePaymentStatusRun(cmd *cobra.Command, args []string) {
	err := UpdatePaymentStatus(&apiutils.Session{})
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
func UpdatePaymentStatus(ses *apiutils.Session) error {
	args := models.PaymentTransactionFilterRequest{}
	limit := 5000
	args.Limit = &limit
	status := string(models.PaymentStatusProcess)
	args.Status = &status
	payments, total, _, _, err := store.Store().PaymentTransactionsFindBy(ses.Context(), args)
	if total > limit {
		err := errors.New("More errors " + strconv.Itoa(total))
		utils.LoggerDesc("in UpdatePaymentStatus more payments thank expected 5000").Error(err)
	}
	if err != nil {
		utils.LoggerDesc("in UpdatePaymentStatus error ocurred").Error(err)
	}

	log.Println("Found payments with status processing: " + strconv.Itoa(total))
	log.Println("Updating...")
	for _, payment := range payments {
		ok, err := app.PaymentHandleUpdate(payment, 1)
		log.Println("Payment by number ", payment.OrderNumber, " response is: ", ok, " status is: ", payment.Status, err)
		if err != nil {
			utils.LoggerDesc("in UpdatePaymentStatus error ocurred").Error(err)
		}
	}

	log.Println("Success.")

	return nil
}
