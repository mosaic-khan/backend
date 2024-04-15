package utils

import (
	"bytes"
	"html/template"
)

func verificationEmail(code string) (string, error) {

	tmp, err := template.ParseFiles("./templates/verification.gohtml")
	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)

	err = tmp.Execute(b, code)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func forgetPassEmail(url string) (string, error) {

	tmp, err := template.ParseFiles("./templates/forgetPass.gohtml")
	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)

	err = tmp.Execute(b, url)
	if err != nil {
		return "", err
	}

	return b.String(), nil

}
