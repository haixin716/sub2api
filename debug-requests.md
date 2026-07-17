# 调试"请求记录"功能

## 问题：页面一直显示"加载中"

### 1. 检查浏览器控制台

打开浏览器开发者工具（F12），检查以下内容：

#### Console 标签
查看是否有 JavaScript 错误，特别是：
- `Failed to load requests: xxx`
- Import 相关错误
- API client 相关错误

#### Network 标签
查找 `/api/v1/requests` 请求：
- **如果没有这个请求**：说明前端代码有问题，API 调用没有执行
- **如果请求状态是 401**：认证问题（检查登录状态）
- **如果请求状态是 404**：后端路由没有注册
- **如果请求状态是 500**：后端代码错误
- **如果请求 Pending**：请求被阻塞或超时

### 2. 验证前端构建

```bash
# 检查最新构建时间
ls -lh /Users/mmt/codeRepository/remoteGit/github/sub2api/backend/internal/web/dist/index.html

# 重新构建前端
cd /Users/mmt/codeRepository/remoteGit/github/sub2api/frontend
npm run build

# 检查构建是否成功
echo $?  # 应该输出 0
```

### 3. 验证后端服务

```bash
# 检查后端进程
ps aux | grep sub2api | grep -v grep

# 检查后端日志（如果有的话）
tail -f backend.log

# 测试 API 端点（需要真实的 token）
curl "http://localhost:8080/api/v1/requests?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 4. 常见问题排查

#### 问题 A：API Client 导入错误
**症状**：Console 显示 `apiClient is undefined` 或类似错误

**已修复**：
- ✅ `frontend/src/api/request.ts` 改为 `import { apiClient }`
- ✅ `frontend/src/api/admin/request.ts` 改为 `import { apiClient }`

#### 问题 B：前端代码未更新
**症状**：修改代码后没有效果

**解决方案**：
```bash
# 1. 清除浏览器缓存（Ctrl+Shift+Delete）
# 2. 硬刷新页面（Ctrl+Shift+R 或 Cmd+Shift+R）
# 3. 重新构建前端
cd frontend && npm run build
```

#### 问题 C：后端未重新编译
**症状**：后端还在使用旧代码

**解决方案**：
```bash
cd backend
go build -o sub2api ./cmd/server
CONFIG_PATH=config.yaml ./sub2api
```

### 5. 快速验证步骤

在浏览器 Console 中运行：

```javascript
// 检查 API client
import('/src/api/request.ts').then(module => {
  console.log('API module:', module);
  console.log('Has list function:', typeof module.list === 'function');
});

// 手动调用 API
fetch('/api/v1/requests?page=1&page_size=20', {
  headers: {
    'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
  }
})
.then(r => r.json())
.then(data => console.log('API Response:', data))
.catch(err => console.error('API Error:', err));
```

### 6. 下一步

如果以上都检查过了还是不行，请提供：
1. 浏览器 Console 的完整错误信息
2. Network 标签中 `/api/v1/requests` 请求的详细信息（状态码、响应内容）
3. 后端日志中的相关错误
