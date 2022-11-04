package crypto

import (
	"blockchain/utils"
	"bytes"
	"math/big"
)

// Base58是Bitcoin中使用的编码方式 主要用于产生钱包地址
// 采用数字、大写字母、小写字母, 去除歧义字符 0、O、I、l 总计58个字符作为编码的字母表

// base58字母表
var base58Alphabet = []byte("" +
	"123456789" +
	"ABCDEFGHJKLMNPQRSTUVWXYZ" +
	"abcdefghijkmnopqrstuvwxyz")

// Base58Encode base58编码
func Base58Encode(input []byte) []byte {
	var result []byte // 编码结果
	// 字节数组转big.int
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(int64(len(base58Alphabet))) // 求余的基本长度
	zero := big.NewInt(0)
	mod := &big.Int{}
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		// 在字母表中进行索引
		result = append(result, base58Alphabet[mod.Int64()])
	}
	utils.Reverse(result) // 倒序得到结果
	result = append([]byte{base58Alphabet[0]}, result...)
	return result
}

// Base58Decode base58解码函数
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	zeroBytes := 1
	// 去除前缀 再查找input中指定数字/字符在基数表中出现的索引
	data := input[zeroBytes:]
	for _, b := range data {
		charIndex := bytes.IndexByte(base58Alphabet, b) // 返回字符在切片中第一次出现的索引
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}
	decoded := result.Bytes()
	return decoded
}
