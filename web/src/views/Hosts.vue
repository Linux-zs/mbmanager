<template>
  <div class="hosts-page">
    <h2 class="page-title">主机管理</h2>

    <el-card>
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          添加主机
        </el-button>
        <div class="filter-group">
          <el-select v-model="selectedGroup" placeholder="选择分组" clearable style="width: 200px" @change="filterByGroup">
            <el-option label="全部分组" value="" />
            <el-option
              v-for="group in groups"
              :key="group"
              :label="group"
              :value="group"
            />
          </el-select>
        </div>
      </div>

      <el-table :data="filteredHosts" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="group" label="分组" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.group" size="small" type="info">{{ row.group }}</el-tag>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="host" label="主机地址" />
        <el-table-column prop="port" label="端口" width="100" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="mysql_version" label="MySQL版本" width="150">
          <template #default="{ row }">
            <span v-if="row.mysql_version" style="font-size: 12px">{{ row.mysql_version }}</span>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
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
          <el-input v-model="form.name" placeholder="请输入主机名称" />
        </el-form-item>

        <el-form-item label="主机地址" prop="host">
          <el-input v-model="form.host" placeholder="请输入主机地址" />
        </el-form-item>

        <el-form-item label="端口" prop="port">
          <el-input-number
            v-model="form.port"
            :min="1"
            :max="65535"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>

        <el-form-item label="分组">
          <el-select
            v-model="form.group"
            placeholder="选择或输入分组"
            filterable
            allow-create
            style="width: 100%"
          >
            <el-option
              v-for="group in groups"
              :key="group"
              :label="group"
              :value="group"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
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
import { ref, computed, onMounted } from 'vue'
import { hostAPI } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const hosts = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加主机')
const formRef = ref(null)
const submitting = ref(false)
const testingId = ref(null)
const selectedGroup = ref('')

const form = ref({
  name: '',
  host: '',
  port: 3306,
  username: '',
  password: '',
  group: '',
  description: '',
  status: 1
})

const rules = {
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  host: [{ required: true, message: '请输入主机地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

// 计算所有分组
const groups = computed(() => {
  const groupSet = new Set()
  hosts.value.forEach(host => {
    if (host.group) {
      groupSet.add(host.group)
    }
  })
  return Array.from(groupSet).sort()
})

// 过滤后的主机列表
const filteredHosts = computed(() => {
  if (!selectedGroup.value) {
    return hosts.value
  }
  return hosts.value.filter(host => host.group === selectedGroup.value)
})

const filterByGroup = () => {
  // 分组过滤已通过computed实现
}

const loadHosts = async () => {
  loading.value = true
  try {
    hosts.value = await hostAPI.list()
  } catch (error) {
    ElMessage.error('加载主机列表失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  dialogTitle.value = '添加主机'
  form.value = {
    name: '',
    host: '',
    port: 3306,
    username: '',
    password: '',
    group: '',
    description: '',
    status: 1
  }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑主机'
  form.value = { ...row }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (form.value.id) {
        await hostAPI.update(form.value.id, form.value)
        ElMessage.success('更新成功')
      } else {
        await hostAPI.create(form.value)
        ElMessage.success('添加成功')
      }
      dialogVisible.value = false
      loadHosts()
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定要删除该主机吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })

  try {
    await hostAPI.delete(row.id)
    ElMessage.success('删除成功')
    loadHosts()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const handleTest = async (row) => {
  testingId.value = row.id
  try {
    const result = await hostAPI.test(row.id)
    if (result.success) {
      ElMessage.success(`连接成功！找到 ${result.databases?.length || 0} 个数据库`)
    } else {
      ElMessage.error('连接失败：' + result.error)
    }
  } catch (error) {
    ElMessage.error('测试连接失败')
  } finally {
    testingId.value = null
  }
}

onMounted(() => {
  loadHosts()
})
</script>

<style scoped>
.hosts-page {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 24px;
  color: #303133;
}

.toolbar {
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-group {
  display: flex;
  gap: 10px;
}

.toolbar {
  margin-bottom: 20px;
}
</style>

