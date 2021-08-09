package candihelper

import (
	"fmt"
)

// TryCatch model
type TryCatch struct {
	Try   func()
	Catch func(error)
}

// Do run TryCatch
func (t TryCatch) Do() {
	defer func() {
		if r := recover(); r != nil {
			var err error
			switch e := r.(type) {
			case error:
				err = e
			default:
				err = fmt.Errorf("%v", r)
			}

			if t.Catch != nil {
				t.Catch(err)
			}
		}
	}()
	t.Try()
}
