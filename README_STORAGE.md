# Ceph RBD StorageClass 配置指南

本目录包含了用于创建和管理 Ceph RBD StorageClass 的配置文件。

## 文件说明

### 1. sandbox-ceph-rbd.yml
完整的 Ceph RBD 配置，包含：
- Ceph 认证密钥 (Secret)
- Ceph RBD StorageClass
- 示例 PersistentVolume

### 2. create-storageclass.sh
交互式 StorageClass 创建脚本，支持：
- Ceph RBD StorageClass
- Local StorageClass  
- NFS StorageClass
- 自动设置默认 StorageClass

### 3. ceph-config-example.yml
配置示例文件，包含：
- 详细的配置说明
- 示例 PVC 配置
- 参数说明

## 使用方法

### 快速部署
```bash
# 1. 修改配置参数
vim sandbox-ceph-rbd.yml

# 2. 应用配置
kubectl apply -f sandbox-ceph-rbd.yml

# 3. 验证部署
kubectl get storageclass
kubectl get secret -n sandbox
```

### 使用交互式脚本
```bash
# 运行创建脚本
./create-storageclass.sh

# 选择选项 1 创建 Ceph RBD StorageClass
```

## 配置参数说明

### Ceph 连接参数
- `clusterID`: Ceph 集群标识符
- `pool`: Ceph 存储池名称 (默认: rbd)
- `monitors`: Ceph Monitor 地址列表
- `user`: Ceph 用户 (默认: admin)

### StorageClass 参数
- `provisioner`: CSI 驱动名称
- `reclaimPolicy`: 回收策略 (Delete/Retain)
- `allowVolumeExpansion`: 是否允许卷扩展
- `volumeBindingMode`: 卷绑定模式

## 故障排除

### 常见问题
1. **Secret 认证失败**
   - 检查 Ceph 密钥是否正确
   - 验证 base64 编码

2. **StorageClass 创建失败**
   - 检查 CSI 驱动是否安装
   - 验证 Ceph 集群连接

3. **PVC 绑定失败**
   - 检查存储池是否存在
   - 验证用户权限

### 调试命令
```bash
# 检查 StorageClass 状态
kubectl describe storageclass sandbox-ceph-rbd

# 检查 Secret 状态
kubectl describe secret ceph-secret -n sandbox

# 检查 PVC 事件
kubectl get events -n sandbox --sort-by='.lastTimestamp'
```

## 注意事项

1. **安全性**: Ceph 密钥包含敏感信息，请妥善保管
2. **网络**: 确保 Kubernetes 节点能访问 Ceph 集群
3. **权限**: 确保 Ceph 用户有足够的存储池权限
4. **版本**: 检查 CSI 驱动与 Ceph 版本兼容性
