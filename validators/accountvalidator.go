package validators

import "github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/models"
import er "errors"
import "strconv"

func AccountCreateValidation(acc *models.AccountCreate) error {
	if len(acc.Aadhar) != 12 {
		return er.New("Aadhar Length Should be 12")
	}
	if _, err := strconv.ParseInt(acc.Aadhar, 10, 64); err != nil {
		return er.New("Aadhar Should contain Digits Only")
	}
	if val, err := strconv.ParseInt(acc.Mobile, 10, 64); true {
		if err != nil {
			return er.New("Mobile Number Should contain Digits Only")
		}

		if val < 6000000000 || val > 9999999999 {
			return er.New("Mobile Number Not Valid")
		}
	}
	return nil
}