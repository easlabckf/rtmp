package amf

import (
	"fmt"
	"io"
)

func (d *Decoder) DecodeBatch(r io.Reader, ver Version) (ret []interface{}, err error) {
	var v interface{}
	for {
		v, err = d.Decode(r, ver)
		if err != nil {
			break
		}
		ret = append(ret, v)
	}
	return
}

func (d *Decoder) Decode(r io.Reader, ver Version) (interface{}, error) {
	switch ver {
	case 0:
		return d.DecodeAmf0(r)
	case 3:
		return d.DecodeAmf3(r)
	}

	return nil, fmt.Errorf("decode amf: unsupported version %d", ver)
}
