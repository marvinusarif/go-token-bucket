package bucket

import (
	"bytes"
	"fmt"
)

type DefaultToken struct {
	id   int    //unexported field
	Name string //exported field
}

//this default function should be used
func (v DefaultToken) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintln(&b, v.id, v.Name)
	return b.Bytes(), nil
}

func (v *DefaultToken) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, &v.id, &v.Name)
	return err
}
