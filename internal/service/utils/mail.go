package utils

import (
	"bytes"
	"html/template"
)

func verificationEmail(code string) ([]byte, error) {

	tmp, err := template.ParseFiles("./templates/verification.gohtml")
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)

	err = tmp.Execute(b, code)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func forgetPassEmail(url string) ([]byte, error) {

	tmp, err := template.ParseFiles("./templates/forgetPass.gohtml")
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)

	err = tmp.Execute(b, url)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil

}
