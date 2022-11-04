package wallet

import (
	"blockchain/crypto"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
)

// 比特币钱包 用于管理公钥-私钥的密钥对
// 私钥是随机生成的数字 并由椭圆曲线乘法获取公钥
// 基于公钥使用单向加密哈希函数生成比特币地址

// 非对称加密将私钥应用于交易的数字指纹以产生数字签名 该签名只能由知晓私钥的人生成
// 访问公钥和交易指纹的任何人都可以使用它们来验证签名

const (
	AddressCheckSumLen = 4 // 地址校验和长度
)

// Wallet 钱包基本结构
type Wallet struct {
	// 包含一个密钥对
	PrivateKey ecdsa.PrivateKey // 私钥
	PublicKey  []byte           // 公钥
}

// NewWallet 初始化钱包
func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair() // 生成密钥对
	return &Wallet{PrivateKey: privateKey, PublicKey: publicKey}
}

// 椭圆曲加密法 一种基于离散对数问题的非对称加密法
// 通过椭圆曲线乘法可以从私钥计算得到公钥 其反向运算(获取离散对数)是极其困难的
// 维基百科: https://en.bitcoin.it/wiki/Secp256k1

// 创建密钥对
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()                             // 获取到一个椭圆
	priKey, err := ecdsa.GenerateKey(curve, rand.Reader) // 通过椭圆生成密钥对
	if err != nil {
		log.Panicf("generate private key failed %v\n", err)
	}
	// 通过私钥生成公钥
	pubKey := append(priKey.PublicKey.X.Bytes(), priKey.PublicKey.Y.Bytes()...)

	return *priKey, pubKey
}

// CheckSum 生成校验和
func CheckSum(input []byte) []byte {
	firstHash := sha256.Sum256(input)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:AddressCheckSumLen]
}

// GetAddress 通过钱包(公钥)生成地址
func (w *Wallet) GetAddress() []byte {
	ripemd160Hash := crypto.Ripemd160Hash(w.PublicKey) // 取公钥哈希值
	checkSumBytes := CheckSum(ripemd160Hash)           // 计算上步哈希值校验和

	// 将校验和添加到哈希值尾部
	addressBytes := append(ripemd160Hash, checkSumBytes...)
	base58Bytes := crypto.Base58Encode(addressBytes) // 将上步结果base58编码
	return base58Bytes
}

// IsValidAddress 校验钱包地址有效性
func IsValidAddress(addressBytes []byte) bool {
	// 将地址base58解码
	pubkeyCheckSumByte := crypto.Base58Decode(addressBytes)
	checkSumBytes := pubkeyCheckSumByte[len(pubkeyCheckSumByte)-AddressCheckSumLen:] // 取出末端4字节数据
	ripemd160Hash := pubkeyCheckSumByte[:len(pubkeyCheckSumByte)-AddressCheckSumLen] // 取出原哈希字符串

	// 计算校验和
	checkedBytes := CheckSum(ripemd160Hash)
	if bytes.Compare(checkSumBytes, checkedBytes) == 0 {
		return true
	}
	return false
}

// StringToHash160 地址字符串转换为哈希编码
func StringToHash160(address string) []byte {
	if !IsValidAddress([]byte(address)) {
		log.Fatalln("\tinvalid address:", address)
	}
	pubKeyHash := crypto.Base58Decode([]byte(address))         // 对地址base58解码
	hash160 := pubKeyHash[:len(pubKeyHash)-AddressCheckSumLen] // 除去校验和 提取哈希值
	return hash160
}
