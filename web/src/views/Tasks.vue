<template>
  <div class="tasks-page">
    <h2 class="page-title">任务管理</h2>

    <el-card>
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          添加任务
        </el-button>
      </div>

      <el-table :data="tasks" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="任务名称" />
        <el-table-column prop="backup_type" label="备份类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.backup_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="schedule_type" label="调度类型" width="100" />
        <el-table-column prop="next_run_at" label="下次执行" width="180">
          <template #default="{ row }">
            {{ formatTime(row.next_run_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="last_run_at" label="上次执行" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_run_at) }}
          </template>
        </el-table-column>
        <el-table-column label="最后备份" width="200">
          <template #default="{ row }">
            <div v-if="row.last_backup">
              <el-tag
                :type="row.last_backup.status === 'success' ? 'success' : row.last_backup.status === 'running' ? 'warning' : 'danger'"
                size="small"
                style="cursor: pointer"
                @click="handleStatusClick(row.last_backup.status, row.name)"
              >
                {{ row.last_backup.status === 'success' ? '成功' : row.last_backup.status === 'running' ? '运行中' : '失败' }}
              </el-tag>
              <div style="font-size: 12px; color: #909399; margin-top: 4px">
                {{ formatSize(row.last_backup.file_size) }}
              </div>
            </div>
            <span v-else style="color: #909399">-</span>
          </template>
        </el-table-column>
        <el-table-column label="存储介质" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.storage" size="small" type="info">
              {{ row.storage.name }}
            </el-tag>
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
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <el-button
              type="info"
              size="small"
              @click="handleViewLogs(row)"
            >
              备份历史
            </el-button>
            <el-button
              type="success"
              size="small"
              @click="handleRun(row)"
              :loading="runningId === row.id"
            >
              立即执行
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
      width="700px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称" />
        </el-form-item>

        <el-form-item label="主机" prop="host_id">
          <el-select v-model="form.host_id" placeholder="请选择主机" style="width: 100%">
            <el-option
              v-for="host in hosts"
              :key="host.id"
              :label="host.name"
              :value="host.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="数据库">
          <el-input
            v-model="databasesInput"
            placeholder='请输入数据库名称，多个用逗号分隔，如: db1,db2'
          />
          <div style="margin-top: 4px; font-size: 12px; color: #909399">
            💡 留空则备份所有数据库
          </div>
        </el-form-item>

        <el-form-item label="备份类型" prop="backup_type">
          <el-select v-model="form.backup_type" style="width: 100%" @change="handleBackupTypeChange">
            <el-option label="mysqldump" value="mysqldump" />
            <el-option label="mydumper" value="mydumper" />
            <el-option label="xtrabackup" value="xtrabackup" />
          </el-select>
        </el-form-item>

        <el-form-item label="压缩格式" prop="compression_type">
          <el-select v-model="form.compression_type" style="width: 100%">
            <el-option label="不压缩" value="none" />
            <el-option label="GZIP压缩" value="gzip" />
            <el-option label="ZIP压缩" value="zip" />
          </el-select>
          <div style="margin-top: 4px; font-size: 12px; color: #909399">
            💡 GZIP压缩率更高，ZIP兼容性更好，不压缩传输更快
          </div>
        </el-form-item>

        <el-form-item label="调度类型" prop="schedule_type">
          <el-select v-model="form.schedule_type" style="width: 100%">
            <el-option label="一次性" value="once" />
            <el-option label="每天" value="daily" />
            <el-option label="每周" value="weekly" />
            <el-option label="每月" value="monthly" />
            <el-option label="Cron表达式" value="cron" />
          </el-select>
        </el-form-item>

        <!-- 调度配置 -->
        <el-form-item
          v-if="form.schedule_type === 'daily'"
          label="执行时间"
        >
          <el-time-picker
            v-model="scheduleTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择时间"
          />
        </el-form-item>

        <el-form-item
          v-if="form.schedule_type === 'weekly'"
          label="星期"
        >
          <el-select v-model="scheduleWeekday" style="width: 200px">
            <el-option label="星期一" :value="1" />
            <el-option label="星期二" :value="2" />
            <el-option label="星期三" :value="3" />
            <el-option label="星期四" :value="4" />
            <el-option label="星期五" :value="5" />
            <el-option label="星期六" :value="6" />
            <el-option label="星期日" :value="0" />
          </el-select>
          <el-time-picker
            v-model="scheduleTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择时间"
            style="margin-left: 10px"
          />
        </el-form-item>

        <el-form-item
          v-if="form.schedule_type === 'monthly'"
          label="日期"
        >
          <el-input-number
            v-model="scheduleDay"
            :min="1"
            :max="31"
            style="width: 200px"
          />
          <el-time-picker
            v-model="scheduleTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择时间"
            style="margin-left: 10px"
          />
        </el-form-item>

        <el-form-item
          v-if="form.schedule_type === 'cron'"
          label="Cron表达式"
        >
          <el-input
            v-model="scheduleCron"
            placeholder="如: 0 2 * * *"
          />
        </el-form-item>

        <el-form-item label="存储" prop="storage_id">
          <el-select v-model="form.storage_id" placeholder="请选择存储" style="width: 100%">
            <el-option
              v-for="storage in storages"
              :key="storage.id"
              :label="storage.name"
              :value="storage.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="保留天数">
          <el-input-number
            v-model="form.retention_days"
            :min="0"
            :max="365"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399">0表示永久保留</span>
        </el-form-item>

        <el-form-item label="备份选项">
          <el-input
            v-model="backupOptionsInput"
            type="textarea"
            :rows="3"
            placeholder="请输入备份参数"
          />
          <div style="margin-top: 4px; font-size: 12px; color: #909399">
            💡 已自动填入默认参数，您可以直接修改或添加额外参数
          </div>
        </el-form-item>

        <el-form-item label="通知渠道">
          <el-select
            v-model="selectedNotifications"
            multiple
            placeholder="请选择通知渠道"
            style="width: 100%"
          >
            <el-option
              v-for="notif in notifications"
              :key="notif.id"
              :label="notif.name"
              :value="notif.id"
            />
          </el-select>
          <span style="margin-left: 10px; color: #909399">可选择多个通知渠道</span>
        </el-form-item>

        <el-form-item label="成功通知">
          <el-switch
            v-model="form.notify_on_success"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>

        <el-form-item label="失败通知">
          <el-switch
            v-model="form.notify_on_failure"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>

        <el-form-item label="是否启用">
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

    <!-- 备份历史对话框 -->
    <el-dialog
      v-model="logsDialogVisible"
      :title="`备份历史 - ${currentTaskName}`"
      width="900px"
    >
      <el-table :data="taskLogs" stripe v-loading="logsLoading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'success' ? 'success' : 'danger'"
              size="small"
            >
              {{ row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时" width="100">
          <template #default="{ row }">
            {{ row.duration }}秒
          </template>
        </el-table-column>
        <el-table-column prop="file_size" label="文件大小" width="120">
          <template #default="{ row }">
            {{ formatSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="file_path" label="文件路径" show-overflow-tooltip />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleDownload(row)"
              :disabled="row.status !== 'success'"
            >
              下载
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDeleteBackup(row)"
              :disabled="row.status !== 'success'"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination" style="margin-top: 20px">
        <el-pagination
          v-model:current-page="logsPagination.page"
          v-model:page-size="logsPagination.page_size"
          :total="logsPagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadTaskLogs"
          @current-change="loadTaskLogs"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { taskAPI, hostAPI, storageAPI, notificationAPI } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()

const tasks = ref([])
const hosts = ref([])
const storages = ref([])
const notifications = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加任务')
const formRef = ref(null)
const submitting = ref(false)
const selectedNotifications = ref([])
const runningId = ref(null)
const isEditMode = ref(false) // 标记是否为编辑模式

// 备份历史相关
const logsDialogVisible = ref(false)
const taskLogs = ref([])
const logsLoading = ref(false)
const currentTaskId = ref(null)
const currentTaskName = ref('')
const logsPagination = ref({
  page: 1,
  page_size: 20,
  total: 0
})

// 调度配置辅助变量
const scheduleTime = ref('02:00')
const scheduleWeekday = ref(1)
const scheduleDay = ref(1)
const scheduleCron = ref('0 2 * * *')
const databasesInput = ref('')
const backupOptionsInput = ref('')

const form = ref({
  name: '',
  host_id: null,
  databases: '[]',
  backup_type: 'mysqldump',
  compression_type: 'gzip',
  schedule_type: 'daily',
  schedule_config: '{}',
  storage_id: null,
  retention_days: 7,
  notify_on_success: 0,
  notify_on_failure: 1,
  backup_options: '',
  status: 1
})

const rules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  host_id: [{ required: true, message: '请选择主机', trigger: 'change' }],
  backup_type: [{ required: true, message: '请选择备份类型', trigger: 'change' }],
  schedule_type: [{ required: true, message: '请选择调度类型', trigger: 'change' }],
  storage_id: [{ required: true, message: '请选择存储', trigger: 'change' }]
}

const loadTasks = async () => {
  loading.value = true
  try {
    tasks.value = await taskAPI.list()
  } catch (error) {
    ElMessage.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

const loadHosts = async () => {
  try {
    hosts.value = await hostAPI.list()
  } catch (error) {
    console.error('加载主机列表失败')
  }
}

const loadStorages = async () => {
  try {
    storages.value = await storageAPI.list()
  } catch (error) {
    console.error('加载存储列表失败')
  }
}

const loadNotifications = async () => {
  try {
    notifications.value = await notificationAPI.list()
  } catch (error) {
    console.error('加载通知列表失败')
  }
}

const getBackupOptionsPlaceholder = () => {
  if (form.value.backup_type === 'mysqldump') {
    return '默认参数：--single-transaction --quick --lock-tables=false --routines --triggers --events\n可在此添加额外参数或覆盖默认参数'
  } else if (form.value.backup_type === 'mydumper') {
    return '默认参数：--threads 4（不使用--compress，最终会打包成tar.gz）\n可在此添加额外参数或覆盖默认参数'
  } else if (form.value.backup_type === 'xtrabackup') {
    return 'xtrabackup需要SSH配置，请填写JSON格式的SSH连接信息'
  }
  return '请输入备份参数'
}

// 获取默认备份参数
const getDefaultBackupOptions = (backupType) => {
  if (backupType === 'mysqldump') {
    return '--single-transaction --quick --lock-tables=false --routines --triggers --events'
  } else if (backupType === 'mydumper') {
    return '--threads 4'
  } else if (backupType === 'xtrabackup') {
    return JSON.stringify({
      ssh_config: {
        host: '',
        port: 22,
        username: '',
        password: '',
        xtrabackup_path: 'xtrabackup'
      }
    }, null, 2)
  }
  return ''
}

// 处理备份类型变化
const handleBackupTypeChange = (newType) => {
  // 无论新建还是编辑，都自动更新为对应的默认参数
  backupOptionsInput.value = getDefaultBackupOptions(newType)
}

const handleAdd = () => {
  dialogTitle.value = '添加任务'
  isEditMode.value = false
  form.value = {
    name: '',
    host_id: null,
    databases: '[]',
    backup_type: 'mysqldump',
    schedule_type: 'daily',
    schedule_config: '{}',
    storage_id: null,
    retention_days: 7,
    notification_ids: '[]',
    notify_on_success: 0,
    notify_on_failure: 1,
    backup_options: '',
    status: 1
  }
  scheduleTime.value = '02:00'
  scheduleWeekday.value = 1
  scheduleDay.value = 1
  scheduleCron.value = '0 2 * * *'
  databasesInput.value = ''
  backupOptionsInput.value = getDefaultBackupOptions('mysqldump')
  selectedNotifications.value = []
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑任务'
  isEditMode.value = true
  form.value = { ...row }

  // 解析数据库列表
  try {
    const dbs = JSON.parse(row.databases || '[]')
    databasesInput.value = dbs.join(',')
  } catch (e) {
    databasesInput.value = ''
  }

  // 解析调度配置
  try {
    const config = JSON.parse(row.schedule_config || '{}')
    scheduleTime.value = config.time || '02:00'
    scheduleWeekday.value = config.weekday || 1
    scheduleDay.value = config.day || 1
    scheduleCron.value = config.expression || '0 2 * * *'
  } catch (e) {
    scheduleTime.value = '02:00'
  }

  // 解析备份选项（直接显示字符串）
  const backupOpts = row.backup_options || ''
  // 如果是空的或者是'{}'，则使用默认参数
  if (backupOpts === '' || backupOpts === '{}') {
    backupOptionsInput.value = getDefaultBackupOptions(row.backup_type)
  } else {
    backupOptionsInput.value = backupOpts
  }

  // 解析通知ID列表
  try {
    const notifIds = JSON.parse(row.notification_ids || '[]')
    selectedNotifications.value = notifIds
  } catch (e) {
    selectedNotifications.value = []
  }

  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    // 构建数据库列表
    const dbList = databasesInput.value
      .split(',')
      .map(db => db.trim())
      .filter(db => db)
    form.value.databases = JSON.stringify(dbList)

    // 构建调度配置
    const scheduleConfig = {}
    if (form.value.schedule_type === 'daily') {
      scheduleConfig.time = scheduleTime.value
    } else if (form.value.schedule_type === 'weekly') {
      scheduleConfig.weekday = scheduleWeekday.value
      scheduleConfig.time = scheduleTime.value
    } else if (form.value.schedule_type === 'monthly') {
      scheduleConfig.day = scheduleDay.value
      scheduleConfig.time = scheduleTime.value
    } else if (form.value.schedule_type === 'cron') {
      scheduleConfig.expression = scheduleCron.value
    }
    form.value.schedule_config = JSON.stringify(scheduleConfig)

    // 构建备份选项（直接保存命令行参数字符串）
    form.value.backup_options = backupOptionsInput.value.trim() || ''

    // 构建通知ID列表
    form.value.notification_ids = JSON.stringify(selectedNotifications.value)

    submitting.value = true
    try {
      if (form.value.id) {
        await taskAPI.update(form.value.id, form.value)
        ElMessage.success('更新成功')
      } else {
        await taskAPI.create(form.value)
        ElMessage.success('添加成功')
      }
      dialogVisible.value = false
      loadTasks()
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定要删除该任务吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })

  try {
    await taskAPI.delete(row.id)
    ElMessage.success('删除成功')
    loadTasks()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const handleRun = async (row) => {
  runningId.value = row.id
  try {
    await taskAPI.run(row.id)
    ElMessage.success('任务已开始执行')
    setTimeout(() => loadTasks(), 2000)
  } catch (error) {
    ElMessage.error('执行失败')
  } finally {
    runningId.value = null
  }
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(2)} ${units[unitIndex]}`
}

const handleStatusClick = (status, taskName) => {
  // 跳转到日志管理页面并筛选对应状态和任务
  router.push({
    path: '/logs',
    query: {
      status: status,
      task_name: taskName
    }
  })
}

const handleViewLogs = (row) => {
  currentTaskId.value = row.id
  currentTaskName.value = row.name
  logsPagination.value.page = 1
  logsDialogVisible.value = true
  loadTaskLogs()
}

const loadTaskLogs = async () => {
  if (!currentTaskId.value) return

  logsLoading.value = true
  try {
    const data = await taskAPI.logs(currentTaskId.value, {
      page: logsPagination.value.page,
      page_size: logsPagination.value.page_size
    })
    taskLogs.value = data.logs || []
    logsPagination.value.total = data.total || 0
  } catch (error) {
    ElMessage.error('加载备份历史失败')
  } finally {
    logsLoading.value = false
  }
}

const handleDownload = async (row) => {
  try {
    // 使用fetch下载文件，这样会包含认证token
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/v1/backups/${row.id}/download`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) {
      throw new Error('下载失败')
    }

    // 获取文件blob
    const blob = await response.blob()

    // 创建下载链接
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = row.file_path.split('/').pop()
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    // 释放URL对象
    window.URL.revokeObjectURL(url)

    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const handleDeleteBackup = async (row) => {
  await ElMessageBox.confirm('确定要删除该备份文件吗？删除后无法恢复！', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })

  try {
    await taskAPI.deleteBackup(row.id)
    ElMessage.success('删除成功')
    loadTaskLogs()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadTasks()
  loadHosts()
  loadStorages()
  loadNotifications()
})
</script>

<style scoped>
.tasks-page {
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

