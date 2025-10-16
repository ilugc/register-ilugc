package register

import (
	"bytes"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

type QrBuffer struct {
	bytes.Buffer
}

func CreateQrBuffer() *QrBuffer {
	self := &QrBuffer{}
	return self
}

func (self *QrBuffer) Close() error {
	return nil
}

type Qr struct {
}

func CreateQr() *Qr {
	self := &Qr{}
	return self
}

func (self *Qr) Init() error {
	return nil
}

func (self *Qr) Gen(url string) (*QrBuffer, error) {
	qrcode, err := qrcode.New(url)
	if err != nil {
		G.logger.Println(err)
		return nil, err
	}

	qrbuffer := CreateQrBuffer()
	writer := standard.NewWithWriter(qrbuffer,
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
		standard.WithQRWidth(5))

	if err := qrcode.Save(writer); err != nil {
		G.logger.Println(err)
		return nil, err
	}
	return qrbuffer, nil
}
