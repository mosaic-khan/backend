package utils

import (
	"bytes"
	"html/template"
)

func verificationEmail(code string) (string, error) {

	tmp, err := template.ParseFiles("./templates/verification_email.gohtml")
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
