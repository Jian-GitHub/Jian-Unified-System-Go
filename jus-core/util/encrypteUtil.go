package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cloudflare/circl/kem"
	"log"
	"os"

	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
)

// KeyPairConfig 密钥对结构体 (用于 JSON 序列化)
type KeyPairConfig struct {
	publicKey  string `json:"publicKey"`  // Base64 编码
	privateKey string `json:"privateKey"` // Base64 编码
}

// MLKEMKeyManager MLKEM 抗量子 加解密算法管理器
type MLKEMKeyManager struct {
	//keyPair KeyPairConfig
	publicKey  kem.PublicKey
	privateKey kem.PrivateKey
}

// getDefaultKeyPairConfig 默认算法密钥对配置 - JSON
func getDefaultKeyPairConfig() *KeyPairConfig {
	return &KeyPairConfig{
		publicKey:  "",
		privateKey: "",
	}
}

// DefaultMLKEMKeyManager 获取默认配置的 MLKEM 管理器
func DefaultMLKEMKeyManager() MLKEMKeyManager {
	// Load Key pair
	loadedPub, loadedPriv, err := getDefaultKeyPairConfig().loadKeys()
	if err != nil {
		log.Fatalf("加载密钥失败: %v", err)
	}
	fmt.Println("密钥已从变量加载")
	mlkem := MLKEMKeyManager{
		publicKey:  loadedPub,
		privateKey: loadedPriv,
	}
	return mlkem
}

// NewMLKEMKeyManager 使用指定密钥对配置生成 MLKEM 管理器
func NewMLKEMKeyManager(keyPair *KeyPairConfig) MLKEMKeyManager {
	// Load Key pair
	loadedPub, loadedPriv, err := keyPair.loadKeys()
	if err != nil {
		log.Fatalf("load MLKEM keys error: %v", err)
	}
	fmt.Println("MLKEM Keys loaded")
	mlkem := MLKEMKeyManager{
		publicKey:  loadedPub,
		privateKey: loadedPriv,
	}
	return mlkem
}

// 保存密钥对到 JSON 文件
func saveKeyPair(filename string, pub kem.PublicKey, priv kem.PrivateKey) error {
	pubRaw, _ := pub.MarshalBinary()
	privRaw, _ := priv.MarshalBinary()

	keyPair := KeyPairConfig{
		publicKey:  base64.StdEncoding.EncodeToString(pubRaw),
		privateKey: base64.StdEncoding.EncodeToString(privRaw),
	}

	data, err := json.MarshalIndent(keyPair, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}

// 从 JSON 文件加载密钥对
func loadKeyPair(filename string) (kem.PublicKey, kem.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	var keyPair KeyPairConfig
	if err := json.Unmarshal(data, &keyPair); err != nil {
		return nil, nil, err
	}

	pubRaw, err := base64.StdEncoding.DecodeString(keyPair.publicKey)
	if err != nil {
		return nil, nil, err
	}

	privRaw, err := base64.StdEncoding.DecodeString(keyPair.privateKey)
	if err != nil {
		return nil, nil, err
	}

	// 使用 KEM scheme 解析密钥
	scheme := mlkem768.Scheme()
	pub, err := scheme.UnmarshalBinaryPublicKey(pubRaw)
	if err != nil {
		return nil, nil, err
	}

	priv, err := scheme.UnmarshalBinaryPrivateKey(privRaw)
	if err != nil {
		return nil, nil, err
	}

	return pub, priv, nil
}

// loadKeys 从 keyPair 密钥对配置 加载密钥
func (p *KeyPairConfig) loadKeys() (kem.PublicKey, kem.PrivateKey, error) {
	// 1. 解码公钥
	pubRaw, err := base64.StdEncoding.DecodeString(p.publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("公钥解码失败: %v", err)
	}

	// 2. 解码私钥
	privRaw, err := base64.StdEncoding.DecodeString(p.privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("私钥解码失败: %v", err)
	}

	// 3. 使用 KEM Scheme 解析
	scheme := mlkem768.Scheme()
	pubKey, err := scheme.UnmarshalBinaryPublicKey(pubRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("公钥解析失败: %v", err)
	}

	privKey, err := scheme.UnmarshalBinaryPrivateKey(privRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("私钥解析失败: %v", err)
	}

	return pubKey, privKey, nil
}

// EncryptMessage 加密消息
func (m *MLKEMKeyManager) EncryptMessage(message string) (string, error) {
	// 1. 生成KEM共享密钥和密文
	ciphertext, sharedKey, err := m.publicKey.Scheme().Encapsulate(m.publicKey)
	if err != nil {
		return "", err
	}

	// 2. 使用共享密钥派生AES-GCM密钥（实际应用应使用HKDF）
	aesKey := sha256.Sum256(sharedKey) // 简单示例，实际用HKDF更安全
	block, _ := aes.NewCipher(aesKey[:])
	gcm, _ := cipher.NewGCM(block)

	// 3. 加密消息
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	encryptedMsg := gcm.Seal(nil, nonce, []byte(message), nil)

	// 4. 组合完整密文（Nonce + KEM密文 + AES密文）
	fullCiphertext := append(nonce, ciphertext...)
	fullCiphertext = append(fullCiphertext, encryptedMsg...)

	// 5. 返回Base64编码的完整密文和共享密钥（仅用于验证）
	return base64.StdEncoding.EncodeToString(fullCiphertext), nil
}

// DecryptMessage 解密消息
func (m *MLKEMKeyManager) DecryptMessage(ciphertextB64 string) (string, error) {
	// 1. Base64解码
	fullCiphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	// 2. 拆分Nonce、KEM密文、AES密文
	newCipher, err := aes.NewCipher(make([]byte, 32))
	if err != nil {
		return "", err
	}
	gcm, _ := cipher.NewGCM(newCipher) // 临时对象获取NonceSize
	nonceSize := gcm.NonceSize()
	kemCiphertextSize := mlkem768.CiphertextSize

	nonce := fullCiphertext[:nonceSize]
	kemCiphertext := fullCiphertext[nonceSize : nonceSize+kemCiphertextSize]
	encryptedMsg := fullCiphertext[nonceSize+kemCiphertextSize:]

	// 3. 解封装KEM获取共享密钥
	sharedKey, err := m.privateKey.Scheme().Decapsulate(m.privateKey, kemCiphertext)
	if err != nil {
		return "", err
	}

	// 4. 解密消息
	aesKey := sha256.Sum256(sharedKey)
	block, _ := aes.NewCipher(aesKey[:])
	gcm, _ = cipher.NewGCM(block)
	decryptedMsg, err := gcm.Open(nil, nonce, encryptedMsg, nil)
	if err != nil {
		return "", err
	}

	return string(decryptedMsg), nil
}

// BytesEqual 安全比较字节切片 - 判断加解密字符串是否一致
func (m *MLKEMKeyManager) BytesEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// Check 检查是否一致
func (m *MLKEMKeyManager) Check(candidate, decryptedMessage string) bool {
	decryptMessage, err := m.DecryptMessage(decryptedMessage)
	if err != nil {
		return false
	}

	if decryptMessage != candidate {
		return false
	} else {
		return true
	}
}
