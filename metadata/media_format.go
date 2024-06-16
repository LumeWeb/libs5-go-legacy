package metadata

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.lumeweb.com/libs5-go/encoding"
)

var (
	_ msgpack.CustomDecoder = (*MediaFormat)(nil)
	_ msgpack.CustomEncoder = (*MediaFormat)(nil)
)

type MediaFormat struct {
	Subtype       string
	Role          string
	Ext           string
	Cid           *encoding.CID
	Height        int
	Width         int
	Languages     []string
	Asr           int
	Fps           int
	Bitrate       int
	AudioChannels int
	Vcodec        string
	Acodec        string
	Container     string
	DynamicRange  string
	Charset       string
	Value         []byte
	Duration      int
	Rows          int
	Columns       int
	Index         int
	InitRange     string
	IndexRange    string
	Caption       string
}

func (mmd *MediaFormat) EncodeMsgpack(encoder *msgpack.Encoder) error {
	return errors.New("Not implemented")
}

func NewMediaFormat(subtype string, role, ext, vcodec, acodec, container, dynamicRange, charset, initRange, indexRange, caption string, cid *encoding.CID, height, width, asr, fps, bitrate, audioChannels, duration, rows, columns, index int, languages []string, value []byte) *MediaFormat {
	return &MediaFormat{
		Subtype:       subtype,
		Role:          role,
		Ext:           ext,
		Cid:           cid,
		Height:        height,
		Width:         width,
		Languages:     languages,
		Asr:           asr,
		Fps:           fps,
		Bitrate:       bitrate,
		AudioChannels: audioChannels,
		Vcodec:        vcodec,
		Acodec:        acodec,
		Container:     container,
		DynamicRange:  dynamicRange,
		Charset:       charset,
		Value:         value,
		Duration:      duration,
		Rows:          rows,
		Columns:       columns,
		Index:         index,
		InitRange:     initRange,
		IndexRange:    indexRange,
		Caption:       caption,
	}
}
func (mmd *MediaFormat) DecodeMsgpack(dec *msgpack.Decoder) error {
	intMap, err := decodeIntMap(dec)
	if err != nil {
		return err
	}

	for key, value := range intMap {
		switch key {
		case 1:
			mmd.Cid, err = encoding.CIDFromBytes(value.([]byte))
			if err != nil {
				return err
			}

		case 2:
			mmd.Subtype = value.(string)
		case 3:
			mmd.Role = value.(string)
		case 4:
			mmd.Ext = value.(string)
		case 10:
			mmd.Height = intParse(value)
		case 11:
			mmd.Width = intParse(value)
		case 12:
			vals := value.([]interface{})

			for _, val := range vals {
				mmd.Languages = append(mmd.Languages, val.(string))
			}
		case 13:
			mmd.Asr = intParse(value)
		case 14:
			mmd.Fps = intParse(value)
		case 15:
			mmd.Bitrate = intParse(value)
		case 18:
			mmd.AudioChannels = intParse(value)
		case 19:
			mmd.Vcodec = value.(string)
		case 20:
			mmd.Acodec = value.(string)
		case 21:
			mmd.Container = value.(string)
		case 22:
			mmd.DynamicRange = value.(string)
		case 23:
			mmd.Charset = value.(string)
		case 24:
			mmd.Value = value.([]byte)
		case 25:
			mmd.Duration = intParse(value)
		case 26:
			mmd.Rows = intParse(value)
		case 27:
			mmd.Columns = intParse(value)
		case 28:
			mmd.Index = intParse(value)
		case 29:
			mmd.InitRange = value.(string)
		case 30:
			mmd.IndexRange = value.(string)
		case 31:
			mmd.Caption = value.(string)
		}
	}

	return nil
}

func intParse(value interface{}) int {
	switch value.(type) {
	case int:
		return value.(int)
	case uint:
		return int(value.(uint))
	case uint16:
		return int(value.(uint16))
	case uint32:
		return int(value.(uint32))
	case int16:
		return int(value.(int16))
	case int8:
		return int(value.(int8))
	}
	return 0
}
