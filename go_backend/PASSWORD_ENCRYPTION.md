# 密码加密说明文档

## 当前使用的加密方法

系统当前使用 **bcrypt** 进行密码加密，这是业界推荐的安全密码加密方法。

### bcrypt 的优势

1. **安全性高**：bcrypt 是专门为密码哈希设计的算法，具有内置的盐值（salt）机制
2. **抗暴力破解**：通过 cost 参数可以调整计算成本，增加破解难度
3. **行业标准**：被广泛使用，经过充分的安全验证
4. **自动加盐**：每次加密都会生成不同的哈希值，即使密码相同

### 当前配置

- **加密库**：`golang.org/x/crypto/bcrypt`
- **Cost 值**：`bcrypt.DefaultCost` (默认值为 10)
- **位置**：`go_backend/internal/utils/password.go`

## 如何更换加密方法

如果您需要更换加密方法（例如使用 Argon2、scrypt 等），请按以下步骤操作：

### 步骤 1：修改密码工具函数

编辑 `go_backend/internal/utils/password.go` 文件：

```go
package utils

import (
    // 导入新的加密库，例如：
    // "golang.org/x/crypto/argon2"
    // "golang.org/x/crypto/scrypt"
)

// HashPassword 使用新的加密方法
func HashPassword(password string) (string, error) {
    // 替换为新的加密实现
    // 例如使用 Argon2:
    // salt := generateRandomSalt(16)
    // hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
    // return base64.EncodeToString(salt) + ":" + base64.EncodeToString(hash), nil
}

// CheckPasswordHash 验证密码
func CheckPasswordHash(password, hash string) bool {
    // 替换为新的验证实现
    // 例如使用 Argon2:
    // parts := strings.Split(hash, ":")
    // salt, _ := base64.DecodeString(parts[0])
    // expectedHash, _ := base64.DecodeString(parts[1])
    // computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
    // return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1
}
```

### 步骤 2：更新数据库中的现有密码

如果数据库中已有使用旧加密方法的密码，需要：

1. **迁移脚本**：创建一个迁移脚本，将所有旧密码重新加密
2. **或手动更新**：通过修改密码功能让用户重新设置密码

### 步骤 3：更新依赖

如果使用新的加密库，需要在 `go.mod` 中添加依赖：

```bash
go get golang.org/x/crypto/argon2
# 或
go get golang.org/x/crypto/scrypt
```

### 步骤 4：测试

确保以下功能正常工作：
- 用户登录
- 修改密码
- 新用户注册（如果有）

## 其他加密方法推荐

### Argon2
- **优势**：2015年密码哈希竞赛的获胜者，被认为是最安全的密码哈希算法
- **适用场景**：对安全性要求极高的场景
- **Go 库**：`golang.org/x/crypto/argon2`

### scrypt
- **优势**：由 Colin Percival 设计，内存密集型，抗硬件攻击
- **适用场景**：需要抗硬件攻击的场景
- **Go 库**：`golang.org/x/crypto/scrypt`

### PBKDF2
- **优势**：简单、成熟、广泛支持
- **适用场景**：需要兼容性好的场景
- **Go 库**：标准库 `golang.org/x/crypto/pbkdf2`

## 修改密码 API 使用说明

### 端点
```
PUT /api/mini/admin/password
```

### 请求头
```
Authorization: Bearer <token>
Content-Type: application/json
```

### 请求体
```json
{
  "old_password": "旧密码",
  "new_password": "新密码（至少6位）"
}
```

### 响应示例

**成功**：
```json
{
  "code": 200,
  "message": "密码修改成功"
}
```

**失败**：
```json
{
  "code": 401,
  "message": "原密码错误"
}
```

## 注意事项

1. **不要使用 MD5、SHA1、SHA256 等普通哈希算法**：这些算法速度太快，容易被暴力破解
2. **保持 cost 值合理**：cost 值太高会影响性能，太低会降低安全性
3. **定期更新密码**：建议用户定期更换密码
4. **密码强度要求**：建议前端和后端都验证密码强度（长度、复杂度等）

## 当前实现位置

- **加密函数**：`go_backend/internal/utils/password.go`
- **登录验证**：`go_backend/internal/api/handlers.go` - `AdminLogin()`
- **修改密码**：`go_backend/internal/api/handlers.go` - `ChangePassword()`
- **数据库模型**：`go_backend/internal/model/admin.go` - `UpdateAdminPassword()`

