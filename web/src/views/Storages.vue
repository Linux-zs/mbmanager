<template>
  <div class="storages-page">
    <h2 class="page-title">存储管理</h2>

    <el-card>
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          添加存储
        </el-button>
      </div>

      <el-table :data="storages" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_default" label="默认" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_default === 1 ? 'success' : 'info'" size="small">
              {{ row.is_default === 1 ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="磁盘空间" width="200">
          <template #default="{ row }">
            <div v-if="(row.type === 'local' || row.type === 'ssh') && row.diskSpace">
              <el-progress
                :percentage="Math.round(row.diskSpace.percentage)"
                :status="row.diskSpace.percentage > 90 ? 'exception' : row.diskSpace.percentage > 75 ? 'warning' : 'success'"
                :stroke-width="16"
              />
              <div style="font-size: 12px; color: #909399; margin-top: 4px">
                {{ formatSize(row.diskSpace.used) }} / {{ formatSize(row.diskSpace.total) }}
              </div>
            </div>
            <span v-else-if="row.type === 'local' || row.type === 'ssh'" style="color: #909399">加载中...</span>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleTest(row)"
              :loading="testingId === row.id"
            >
              测试连接
            </el-button>
            <el-button
              type="warning"
              size="small"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入存储名称" />
        </el-form-item>

        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" style="width: 100%" @change="handleTypeChange">
            <el-option label="本地存储" value="local" />
            <el-option label="SSH远程存储" value="ssh" />
            <el-option label="AWS S3" value="s3" />
            <el-option label="阿里云OSS" value="oss" />
            <el-option label="NAS" value="nas" />
          </el-select>
        </el-form-item>

        <!-- 本地存储配置 -->
        <template v-if="form.type === 'local' || form.type === 'nas'">
          <el-form-item label="基础路径">
            <el-input v-model="configForm.base_path" placeholder="/data/backups" />
          </el-form-item>
        </template>

        <!-- SSH存储配置 -->
        <template v-if="form.type === 'ssh'">
          <el-form-item label="主机地址">
            <el-input v-model="configForm.host" placeholder="192.168.1.100" />
          </el-form-item>
          <el-form-item label="端口">
            <el-input-number v-model="configForm.port" :min="1" :max="65535" style="width: 100%" />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="configForm.username" placeholder="root" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="configForm.password" type="password" show-password placeholder="SSH密码（与私钥二选一）" />
          </el-form-item>
          <el-form-item label="私钥">
            <el-input
              v-model="configForm.private_key"
              type="textarea"
              :rows="4"
              placeholder="SSH私钥内容（与密码二选一）"
            />
          </el-form-item>
          <el-form-item label="基础路径">
            <el-input v-model="configForm.base_path" placeholder="/data/backups" />
          </el-form-item>
        </template>

        <!-- S3配置 -->
        <template v-if="form.type === 's3'">
          <el-form-item label="Endpoint">
            <el-input v-model="configForm.endpoint" placeholder="s3.amazonaws.com" />
          </el-form-item>
          <el-form-item label="Access Key">
            <el-input v-model="configForm.access_key_id" />
          </el-form-item>
          <el-form-item label="Secret Key">
            <el-input v-model="configForm.secret_access_key" type="password" show-password />
          </el-form-item>
          <el-form-item label="Bucket">
            <el-input v-model="configForm.bucket" />
          </el-form-item>
          <el-form-item label="Region">
            <el-input v-model="configForm.region" placeholder="us-east-1" />
          </el-form-item>
        </template>

        <!-- OSS配置 -->
        <template v-if="form.type === 'oss'">
          <el-form-item label="Endpoint">
            <el-input v-model="configForm.endpoint" placeholder="oss-cn-hangzhou.aliyuncs.com" />
          </el-form-item>
          <el-form-item label="Access Key">
            <el-input v-model="configForm.access_key_id" />
          </el-form-item>
          <el-form-item label="Secret Key">
            <el-input v-model="configForm.access_key_secret" type="password" show-password />
          </el-form-item>
          <el-form-item label="Bucket">
            <el-input v-model="configForm.bucket" />
          </el-form-item>
        </template>

        <el-form-item label="默认存储">
          <el-switch
            v-model="form.is_default"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>

        <el-form-item label="状态">
          <el-switch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { storageAPI } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const storages = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加存储')
const formRef = ref(null)
const submitting = ref(false)
const testingId = ref(null)

const form = ref({
  name: '',
  type: 'local',
  config: '{}',
  is_default: 0,
  status: 1
})

const configForm = reactive({
  base_path: '/data/backups',
  host: '',
  port: 22,
  username: '',
  password: '',
  private_key: '',
  endpoint: '',
  access_key_id: '',
  secret_access_key: '',
  access_key_secret: '',
  bucket: '',
  region: ''
})

const rules = {
  name: [{ required: true, message: '请输入存储名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择存储类型', trigger: 'change' }]
}

const loadStorages = async () => {
  loading.value = true
  try {
    storages.value = await storageAPI.list()
    // 为本地存储和SSH存储加载磁盘空间信息
    for (const storage of storages.value) {
      if (storage.type === 'local' || storage.type === 'ssh') {
        try {
          const diskSpace = await storageAPI.getDiskSpace(storage.id)
          storage.diskSpace = diskSpace
        } catch (error) {
          console.error(`Failed to load disk space for storage ${storage.id}:`, error)
        }
      }
    }
  } catch (error) {
    ElMessage.error('加载存储列表失败')
  } finally {
    loading.value = false
  }
}

const handleTypeChange = () => {
  Object.keys(configForm).forEach(key => {
    configForm[key] = ''
  })
  if (form.value.type === 'local' || form.value.type === 'nas') {
    configForm.base_path = '/data/backups'
  } else if (form.value.type === 'ssh') {
    configForm.base_path = '/data/backups'
    configForm.port = 22
  }
}

const handleAdd = () => {
  dialogTitle.value = '添加存储'
  form.value = {
    name: '',
    type: 'local',
    config: '{}',
    is_default: 0,
    status: 1
  }
  handleTypeChange()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑存储'
  form.value = { ...row }

  try {
    const config = JSON.parse(row.config || '{}')
    Object.keys(configForm).forEach(key => {
      configForm[key] = config[key] || ''
    })
  } catch (e) {
    console.error('解析配置失败', e)
  }

  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    const config = {}
    Object.keys(configForm).forEach(key => {
      if (configForm[key]) {
        config[key] = configForm[key]
      }
    })
    form.value.config = JSON.stringify(config)

    submitting.value = true
    try {
      if (form.value.id) {
        await storageAPI.update(form.value.id, form.value)
        ElMessage.success('更新成功')
      } else {
        await storageAPI.create(form.value)
        ElMessage.success('添加成功')
      }
      dialogVisible.value = false
      loadStorages()
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定要删除该存储吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })

  try {
    await storageAPI.delete(row.id)
    ElMessage.success('删除成功')
    loadStorages()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const handleTest = async (row) => {
  testingId.value = row.id
  try {
    const result = await storageAPI.test(row.id)
    if (result.success) {
      ElMessage.success('连接测试成功')
    } else {
      ElMessage.error('连接失败：' + result.error)
    }
  } catch (error) {
    ElMessage.error('测试连接失败')
  } finally {
    testingId.value = null
  }
}

// 格式化文件大小
const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

onMounted(() => {
  loadStorages()
})
</script>

<style scoped>
.storages-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 24px;
  color: #303133;
}

.toolbar {
  margin-bottom: 20px;
}
</style>
