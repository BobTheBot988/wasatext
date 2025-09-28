package model

import "errors"

func AddError(err1 error, err2 error) error {
	return errors.New(err1.Error() + ":\n" + err2.Error())
}

func AddErrorString(err1 string, err2 string) error {
	return errors.New(err1 + ":\n" + err2)
}
