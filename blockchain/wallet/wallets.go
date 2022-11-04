package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// 管理钱包集合

const walletFile = "wallets-%s.dat"

// Wallets 钱包集合基本结构
type Wallets struct {
	Wallets map[string]*Wallet // 关联地址和钱包
}

// NewWallets 初始化钱包集合
func NewWallets(nodeId string) *Wallets {
	// 从钱包文件中获取钱包信息
	name := fmt.Sprintf(walletFile, nodeId)
	if _, err := os.Stat(name); os.IsNotExist(err) {
		// 如果文件不存在 则返回空表
		wallets := &Wallets{}
		wallets.Wallets = make(map[string]*Wallet)
		return wallets
	}
	content, err := ioutil.ReadFile(name) // 读取文件内容
	if err != nil {
		log.Panicf("read the file content failed: %v\n", err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panicf("decode the file content failed: %v\n", err)
	}
	return &wallets
}

// CreateWallet 添加新钱包
func (wallets *Wallets) CreateWallet(nodeId string) {
	wallet := NewWallet()
	wallets.Wallets[string(wallet.GetAddress())] = wallet
	wallets.SaveWallets(nodeId) // 保存钱包
	fmt.Println("[" + string(wallet.GetAddress()) + "]")
}

// SaveWallets 持久化钱包
func (wallets *Wallets) SaveWallets(nodeId string) {
	var content bytes.Buffer // 钱包内容

	name := fmt.Sprintf(walletFile, nodeId)
	// 注册椭圆 在内部调用接口
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(&wallets) // 序列化
	if err != nil {
		log.Panicf("encode the struct of wallets failed: %v", err)
	}

	// 将数据存入文件
	err = ioutil.WriteFile(name, content.Bytes(), 0644)
	if err != nil {
		log.Panicf("save content of wallet into file[%s] failed: %v", name, err)
	}
}
