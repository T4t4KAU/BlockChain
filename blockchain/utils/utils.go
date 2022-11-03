package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
)

const LENGTH = 12

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

// JsonToSlice 标准JSON格式转字符串切片
func JsonToSlice(jsonString string) []string {
	var strSlice []string
	if err := json.Unmarshal([]byte(jsonString), &strSlice); err != nil {
		log.Panicf("json to []string failed:%v\n", err)
	}
	return strSlice
}

// Reverse 反转切片
func Reverse(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func GobEncoder(data interface{}) []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(data)
	if err != nil {
		log.Panicf("encode the data failed: %v\n", err)
	}
	return result.Bytes()
}

// CommandToBytes 命令转化为字节数组
func CommandToBytes(command string) []byte {
	var commandBytes [LENGTH]byte
	for i, c := range command {
		commandBytes[i] = byte(c)
	}
	return commandBytes[:]
}

// BytesToCommand 解析请求中命令
func BytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x00 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}
