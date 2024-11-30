package app

import (
	"fmt"

	"github.com/mekdep/server/internal/models"
	"github.com/xuri/excelize/v2"
)

type ResetPasswordRequest struct {
	SchoolId uint          `form:"school_id"`
	Roles    []models.Role `form:"roles[]"`
}

func ToolResetPassword() (string, error) {
	path := "web/uploads/tmp/" + RandStringBytes(16) + ".xlsx"
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	f.SetCellValue("Sheet2", "A2", "Hello world.")
	err := f.SaveAs(path)
	return path, err
}
