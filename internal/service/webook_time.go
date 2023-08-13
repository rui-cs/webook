package service

import (
	"fmt"
	"time"
)

type WebookTime time.Time

func (w WebookTime) MarshalJSON() ([]byte, error) {
	t := time.Time(w)

	var zero time.Time

	if t.Unix() == zero.Unix() {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprint(time.Time(w).Format("2006-01-02"))), nil
}

func (w *WebookTime) UnmarshalJSON(b []byte) error {
	t, err := time.ParseInLocation("\"2006-01-02\"", string(b), time.Local) // layout中包含引号
	if err != nil {
		return err
	}

	*w = WebookTime(t)
	return nil
}

//func (w *WebookTime) Scan(value interface{}) error {
//	t, ok := value.(time.Time)
//	if !ok {
//		return errors.New("value is not time.Time")
//	}
//
//	*w = WebookTime(t)
//
//	return nil
//}
//
//func (w WebookTime) Value() (driver.Value, error) {
//	t := time.Time(w)
//	var zero time.Time
//
//	if t.Unix() == zero.Unix() {
//		return nil, nil
//	}
//
//	return t, nil
//}
