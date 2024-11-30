package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mekdep/server/config"
	apiUtils "github.com/mekdep/server/internal/api/utils"
	"github.com/mekdep/server/internal/models"
	"github.com/mekdep/server/internal/store"
	"github.com/mekdep/server/internal/utils"
)

func PaymentHandleBankApi(m *models.PaymentTransaction) error {
	err := PaymentHandleCheckout(m)
	if err == nil {
		go func() {
			mm := *m
			_, err1 := PaymentHandleUpdate(&mm, 5)
			if err1 != nil {
				utils.LoggerDesc("in worker payment update").Error(err1)
			}
		}()
	}
	return err
}

func PaymentHandleCheckout(m *models.PaymentTransaction) (err error) {
	if m.BankType == models.PaymentBankHalkBank {
		err := paymentBankCheckout(m, paymentHalkBankRequest)
		return err
	} else if m.BankType == models.PaymentBankRysgalBank {
		err := paymentBankCheckout(m, paymentRysgalbankRequest)
		return err
	} else if m.BankType == models.PaymentBankSenagatBank {
		err := paymentBankCheckout(m, paymentSenagatbankRequest)
		return err
	} else if m.BankType == models.PaymentBankTfebBank {
		err := paymentBankCheckout(m, paymentHalkBankRequest)
		return err
	}
	return ErrInvalid.SetKey("bank_type")
}

func PaymentHandleUpdate(m *models.PaymentTransaction, triesCount int) (bool, error) {
	if !m.IsStatusProcessing() {
		return false, ErrExpired.SetComment("Already processed").SetKey("id")
	}
	if m.BankType == models.PaymentBankHalkBank {
		return paymentBankUpdate(m.ID, triesCount, 0, paymentHalkBankRequest)
	} else if m.BankType == models.PaymentBankRysgalBank {
		return paymentBankUpdate(m.ID, triesCount, 0, paymentRysgalbankRequest)
	} else if m.BankType == models.PaymentBankSenagatBank {
		return paymentBankUpdate(m.ID, triesCount, 0, paymentSenagatbankRequest)
	} else if m.BankType == models.PaymentBankTfebBank {
		return paymentBankUpdate(m.ID, triesCount, 0, paymentHalkBankRequest)
	}
	return false, ErrInvalid.SetKey("bank_type")
}

func paymentBankCheckout(m *models.PaymentTransaction, f BankRequestFunc) error {
	v, err := f("register.do", url.Values{
		"orderNumber": []string{m.ID},
		"amount":      []string{string(models.MoneyFromFloat(m.Amount))},
		"currency":    []string{"934"},
		"language":    []string{"ru"},
		"returnUrl":   []string{config.Conf.AppUrl + "/share/payment?offset=1&method=finish.html"},
		"description": []string{"TBM-IMM. Toleg " + strconv.Itoa(int(m.OriginalAmount)) +
			". Mohlet " + strconv.Itoa(m.SchoolMonths) +
			". Cagalar " + strconv.Itoa(len(m.Students))},
	})
	if err != nil {
		return err
	}
	if v, ok := v["orderId"].(string); !ok || v == "" {
		dd, _ := json.Marshal(v)
		err = errors.New("error bank api checkout: no orderId: " + string(dd))
		utils.LoggerDesc("error bank api checkout").
			Error(err)
		return err
	}

	m.OrderNumber = new(string)
	m.OrderUrl = new(string)
	*m.OrderNumber = v["orderId"].(string)
	*m.OrderUrl = v["formUrl"].(string)
	return nil
}

var bankOrderStatusMsg map[int]string = map[int]string{
	0: "Заказ зарегистрирован, но не оплачен",
	1: "Предавторизованная сумма захолдирована (для двухстадийных платежей)",
	2: "Проведена полная авторизация суммы заказа",
	3: "Авторизация отменена",
	4: "По транзакции была проведена операция возврата",
	5: "Инициирована авторизация через ACS банка-эмитента",
	6: "Авторизация отклонена",
}

func paymentBankUpdate(tid string, leftTry int, tried int, f BankRequestFunc) (bool, error) {
	m, err := store.Store().PaymentTransactionsFindById(context.Background(), tid)
	if err != nil {
		return false, err
	}
	if isPaymentHandleSpecialTariff(m) || m.OrderNumber == nil {
		return true, nil
	}
	err = store.Store().PaymentTransactionsLoadRelations(context.Background(), &[]*models.PaymentTransaction{m})
	if err != nil {
		return false, err
	}
	if tried == 0 && leftTry <= 3 {
		time.Sleep(time.Second * 2)
	} else if tried == 1 && leftTry <= 3 {
		time.Sleep(time.Second * 5)
	} else if tried == 0 {
		time.Sleep(time.Minute * 15)
	} else {
		time.Sleep(time.Minute * 3)
	}
	log.Println("updating " + *m.OrderNumber)
	// fetch
	v, err := f("getOrderStatusExtended.do", url.Values{
		"orderId":  []string{*m.OrderNumber},
		"language": []string{"ru"},
	})
	if err != nil {
		return false, err
	}
	// log.Println(v)
	// map[actionCode:-100 actionCodeDescription:
	//     amount:1350 attributes:[map[name:mdOrder value:c146256e-bbe2-4948-9aba-be5fb6d2e610]] currency:934 date:1.70298161507e+12 errorCode:0 errorMessage:Успешно orderNumber:49
	//    orderStatus:0]
	orderStatus := int(v["orderStatus"].(float64))
	m.SystemComment = new(string)
	*m.SystemComment = strconv.Itoa(int(orderStatus)) + " " + bankOrderStatusMsg[orderStatus]
	if orderStatus == 2 {
		err = paymentSuccess(*m)
		if err != nil {
			return false, err
		}
		return true, nil
	} else if orderStatus > 2 {
		err = paymentFailed(*m)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	if leftTry > 1 {
		// still processing....
		return paymentBankUpdate(m.ID, leftTry-1, tried+1, f)
	}
	// if end then give up, as not ok
	return false, nil
}

var paymentChecked map[string]bool

func paymentSuccess(mm models.PaymentTransaction) error {
	log.Println("payment success: " + *mm.OrderNumber)
	if paymentChecked == nil {
		paymentChecked = map[string]bool{}
	}
	if paymentChecked[mm.ID] {
		return nil
	}
	paymentChecked[mm.ID] = true
	mm.Status = models.PaymentStatusCompleted
	_, err := store.Store().PaymentTransactionUpdate(context.Background(), &mm)
	if err != nil {
		return err
	}
	// upgrade
	for _, v := range mm.Students {
		err = UserTariffUpgrade(&apiUtils.Session{}, &mm, v, []*models.User{mm.Payer})
		if err != nil {
			return err
		}
	}
	return nil
}
func paymentFailed(mm models.PaymentTransaction) error {
	log.Println("payment failed: " + *mm.OrderNumber)
	mm.Status = models.PaymentStatusFailed
	_, err := store.Store().PaymentTransactionUpdate(context.Background(), &mm)
	if err != nil {
		return err
	}

	return nil
}

type BankRequestFunc func(path string, query url.Values) (map[string]interface{}, error)

func paymentHalkBankRequest(path string, query url.Values) (map[string]interface{}, error) {
	query["userName"] = []string{"103161020674"}
	query["password"] = []string{"Pgd734hdR21dERg"}
	if v, ok := query["orderNumber"]; ok && len(v) > 0 {
		cleanOrderNumber := strings.ReplaceAll(query["orderNumber"][0], "-", "")
		query["orderNumber"][0] = cleanOrderNumber
	}
	ur := "https://mpi.gov.tm/payment/rest/" + path + "?" + query.Encode()
	res, err := http.Get(ur)
	if err != nil {
		utils.LoggerDesc("error bank api checkout").Error(err)
		err = errors.New("error bank api checkout")
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	var v map[string]interface{}
	err = json.Unmarshal(resBody, &v)
	if err != nil {
		return nil, err
	}
	if ec, ok := v["errorCode"].(string); ok && ec != "0" {
		err = errors.New("error bank api checkout")
		utils.LoggerDesc("error bank api checkout").
			Error(errors.New("Order number " + query["orderNumber"][0] + "; Code " + ec + "; Msg " + v["errorMessage"].(string)))
		return nil, err
	}
	return v, nil
}

func paymentRysgalbankRequest(path string, query url.Values) (map[string]interface{}, error) {
	query["userName"] = []string{"mekdepAPI"}
	query["password"] = []string{"53430564toleg"}
	if v, ok := query["orderNumber"]; ok && len(v) > 0 {
		cleanOrderNumber := strings.ReplaceAll(query["orderNumber"][0], "-", "")
		query["orderNumber"][0] = cleanOrderNumber
	}
	ur := "https://epg.rysgalbank.tm/epg/rest/" + path + "?" + query.Encode()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(ur)
	if err != nil {
		utils.LoggerDesc("error bank api checkout").Error(err)
		err = errors.New("error bank api checkout")
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	var v map[string]interface{}
	err = json.Unmarshal(resBody, &v)
	if err != nil {
		return nil, err
	}
	if ec, ok := v["errorCode"].(string); ok && ec != "0" {
		err = errors.New("error bank api checkout")
		utils.LoggerDesc("error bank api checkout").
			Error(errors.New(ec + " : " + v["errorMessage"].(string)))
		return nil, err
	}
	return v, nil
}

func paymentSenagatbankRequest(path string, query url.Values) (map[string]interface{}, error) {
	query["userName"] = []string{"mekdep_edu"}
	query["password"] = []string{"mekdep_edu1"}
	if v, ok := query["orderNumber"]; ok && len(v) > 0 {
		cleanOrderNumber := strings.ReplaceAll(query["orderNumber"][0], "-", "")
		query["orderNumber"][0] = cleanOrderNumber
	}
	ur := "https://epg.senagatbank.com.tm/epg/rest/" + path + "?" + query.Encode()
	res, err := http.Get(ur)
	if err != nil {
		utils.LoggerDesc("error bank api checkout").Error(err)
		err = errors.New("error bank api checkout")
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	var v map[string]interface{}
	err = json.Unmarshal(resBody, &v)
	if err != nil {
		return nil, err
	}
	if ec, ok := v["errorCode"].(string); ok && ec != "0" {
		err = errors.New("error bank api checkout")
		utils.LoggerDesc("error bank api checkout").
			Error(errors.New(ec + " : " + v["errorMessage"].(string)))
		return nil, err
	}
	return v, nil
}
