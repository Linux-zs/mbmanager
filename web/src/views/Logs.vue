<template>
  <div class="logs-page">
    <h2 class="page-title">备份日志</h2>

    <el-card>
      <div class="toolbar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="状态">
            <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px">
              <el-option label="成功" value="success" />
              <el-option label="失败" value="failed" />
              <el-option label="运行中" value="running" />
            </el-select>
          </el-form-item>
          <el-form-item label="备份类型">
            <el-select v-model="filters.backup_type" placeholder="全部" clearable style="width: 140px">
              <el-option label="mysqldump" value="mysqldump" />
              <el-option label="mydumper" value="mydumper" />
              <el-option label="xtrabackup" value="xtrabackup" />
            </el-select>
          </el-form-item>
          <el-form-item label="任务名称">
            <el-input v-model="filters.task_name" placeholder="请输入" clearable style="width: 150px" />
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="filters.date_range"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              style="width: 240px"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadLogs">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="logs" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="task_name" label="任务名称" />
        <el-table-column prop="host_name" label="主机" />
        <el-table-column prop="backup_type" label="备份类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.backup_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'"
              size="small"
            >
              {{ row.status === 'success' ? '成功' : row.status === 'failed' ? '失败' : '运行中' }}
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
        <el-table-column prop="file_path" label="文件路径" width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="showDetail(row)"
            >
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadLogs"
          @current-change="loadLogs"
        />
      </div>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="备份详情"
      width="800px"
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="任务名称">{{ currentLog.task_name }}</el-descriptions-item>
        <el-descriptions-item label="主机">{{ currentLog.host_name }}</el-descriptions-item>
        <el-descriptions-item label="备份类型">{{ currentLog.backup_type }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag
            :type="currentLog.status === 'success' ? 'success' : currentLog.status === 'failed' ? 'danger' : 'warning'"
            size="small"
          >
            {{ currentLog.status === 'success' ? '成功' : currentLog.status === 'failed' ? '失败' : '运行中' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(currentLog.start_time) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(currentLog.end_time) }}</el-descriptions-item>
        <el-descriptions-item label="总耗时">{{ currentLog.duration }}秒</el-descriptions-item>
        <el-descriptions-item label="备份耗时">{{ currentLog.backup_time || 0 }}秒</el-descriptions-item>
        <el-descriptions-item label="传输耗时">{{ currentLog.transfer_time || 0 }}秒</el-descriptions-item>
        <el-descriptions-item label="文件大小">{{ formatSize(currentLog.file_size) }}</el-descriptions-item>
        <el-descriptions-item label="文件路径" :span="2">{{ currentLog.file_path }}</el-descriptions-item>
        <el-descriptions-item label="存储类型">{{ currentLog.storage_type }}</el-descriptions-item>
        <el-descriptions-item label="数据库">{{ currentLog.databases }}</el-descriptions-item>
        <el-descriptions-item label="备份命令" :span="2">
          <el-input
            v-model="currentLog.command"
            type="textarea"
            :rows="3"
            readonly
          />
        </el-descriptions-item>
        <el-descriptions-item v-if="currentLog.error_message" label="错误信息" :span="2">
          <el-alert
            :title="currentLog.error_message"
            type="error"
            :closable="false"
          />
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { logAPI } from '../api'
import { ElMessage } from 'element-plus'

const route = useRoute()

const logs = ref([])
const loading = ref(false)
const detailDialogVisible = ref(false)
const currentLog = ref({})

const filters = reactive({
  status: '',
  backup_type: '',
  task_name: '',
  date_range: null
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const loadLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.status) {
      params.status = filters.status
    }
    if (filters.backup_type) {
      params.backup_type = filters.backup_type
    }
    if (filters.task_name) {
      params.task_name = filters.task_name
    }
    if (filters.date_range && filters.date_range.length === 2) {
      params.start_time = filters.date_range[0].toISOString()
      params.end_time = filters.date_range[1].toISOString()
    }

    const data = await logAPI.list(params)
    logs.value = data.logs || []
    pagination.total = data.total || 0
  } catch (error) {
    ElMessage.error('加载日志失败')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  filters.status = ''
  filters.backup_type = ''
  filters.task_name = ''
  filters.date_range = null
  pagination.page = 1
  loadLogs()
}

const showDetail = (row) => {
  currentLog.value = row
  detailDialogVisible.value = true
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

onMounted(() => {
  // 从路由参数中获取status和task_name
  if (route.query.status) {
    filters.status = route.query.status
  }
  if (route.query.task_name) {
    filters.task_name = route.query.task_name
  }
  loadLogs()
})
</script>

<style scoped>
.logs-page {
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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
