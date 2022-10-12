package block

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
)

// IntToHex int类型转字节数组
func IntToHex(data int64) []byte {
	buffer := new(bytes.Buffer)
	// 按大端序将数据写入缓冲区
	err := binary.Write(buffer, binary.BigEndian, data)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

// JsonToSlice 标准JSON格式转切片
func JsonToSlice(jsonString string) []string {
	var strSlice []string
	if err := json.Unmarshal([]byte(jsonString), &strSlice); err != nil {
		log.Panicf("json to []string failed:%v\n", err)
	}
	return strSlice
}
