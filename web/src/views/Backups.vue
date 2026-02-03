<template>
  <div class="backups-page">
    <h2 class="page-title">备份管理</h2>

    <el-card>
      <div class="toolbar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="主机">
            <el-input v-model="filters.host_name" placeholder="请输入主机名" clearable style="width: 150px" />
          </el-form-item>
          <el-form-item label="任务">
            <el-input v-model="filters.task_name" placeholder="请输入任务名" clearable style="width: 150px" />
          </el-form-item>
          <el-form-item label="存储类型">
            <el-select v-model="filters.storage_type" placeholder="全部" clearable style="width: 120px">
              <el-option label="本地存储" value="local" />
              <el-option label="SSH存储" value="ssh" />
              <el-option label="S3存储" value="s3" />
              <el-option label="OSS存储" value="oss" />
              <el-option label="NAS存储" value="nas" />
            </el-select>
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
            <el-button type="primary" @click="loadBackups">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
            <el-checkbox v-model="groupByHost" style="margin-left: 10px">按主机分组</el-checkbox>
          </el-form-item>
        </el-form>
        <div class="batch-actions">
          <el-button
            type="danger"
            :disabled="selectedBackups.length === 0"
            @click="handleBatchDelete"
          >
            批量删除 ({{ selectedBackups.length }})
          </el-button>
          <el-button @click="loadBackups">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>

      <el-table
        v-if="!groupByHost"
        :data="backups"
        stripe
        v-loading="loading"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="task_name" label="任务名称" width="150" />
        <el-table-column prop="host_name" label="主机" width="150" />
        <el-table-column prop="databases" label="数据库" width="150" show-overflow-tooltip />
        <el-table-column prop="backup_type" label="备份类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.backup_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="storage_type" label="存储类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ getStorageTypeLabel(row.storage_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="storage_name" label="存储介质" width="150" show-overflow-tooltip />
        <el-table-column prop="file_size" label="文件大小" width="120">
          <template #default="{ row }">
            {{ formatSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="备份时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="file_path" label="文件路径" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleDownload(row)"
            >
              下载
            </el-button>
            <el-button
              type="info"
              size="small"
              @click="showDetail(row)"
            >
              详情
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

      <!-- 分组视图 -->
      <div v-if="groupByHost" class="grouped-view">
        <el-collapse v-model="activeGroups">
          <el-collapse-item
            v-for="(group, hostName) in groupedBackups"
            :key="hostName"
            :name="hostName"
          >
            <template #title>
              <div class="group-header">
                <el-tag type="primary" size="large">{{ hostName }}</el-tag>
                <span class="group-count">{{ group.length }} 个备份</span>
              </div>
            </template>
            <el-table :data="group" stripe>
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="task_name" label="任务名称" width="150" />
              <el-table-column prop="databases" label="数据库" width="150" show-overflow-tooltip />
              <el-table-column prop="backup_type" label="备份类型" width="120">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.backup_type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="storage_type" label="存储类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small" type="info">{{ getStorageTypeLabel(row.storage_type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="storage_name" label="存储介质" width="150" show-overflow-tooltip />
              <el-table-column prop="file_size" label="文件大小" width="120">
                <template #default="{ row }">
                  {{ formatSize(row.file_size) }}
                </template>
              </el-table-column>
              <el-table-column prop="start_time" label="备份时间" width="180">
                <template #default="{ row }">
                  {{ formatTime(row.start_time) }}
                </template>
              </el-table-column>
              <el-table-column prop="file_path" label="文件路径" min-width="200" show-overflow-tooltip />
              <el-table-column label="操作" width="200" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" size="small" @click="handleDownload(row)">下载</el-button>
                  <el-button type="info" size="small" @click="showDetail(row)">详情</el-button>
                  <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-collapse-item>
        </el-collapse>
      </div>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadBackups"
          @current-change="loadBackups"
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
        <el-descriptions-item label="ID">{{ currentBackup.id }}</el-descriptions-item>
        <el-descriptions-item label="任务名称">{{ currentBackup.task_name }}</el-descriptions-item>
        <el-descriptions-item label="主机">{{ currentBackup.host_name }}</el-descriptions-item>
        <el-descriptions-item label="数据库">{{ currentBackup.databases }}</el-descriptions-item>
        <el-descriptions-item label="备份类型">{{ currentBackup.backup_type }}</el-descriptions-item>
        <el-descriptions-item label="存储类型">{{ getStorageTypeLabel(currentBackup.storage_type) }}</el-descriptions-item>
        <el-descriptions-item label="存储介质" :span="2">{{ currentBackup.storage_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(currentBackup.start_time) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(currentBackup.end_time) }}</el-descriptions-item>
        <el-descriptions-item label="总耗时">{{ currentBackup.duration }}秒</el-descriptions-item>
        <el-descriptions-item label="备份耗时">{{ currentBackup.backup_time || 0 }}秒</el-descriptions-item>
        <el-descriptions-item label="传输耗时">{{ currentBackup.transfer_time || 0 }}秒</el-descriptions-item>
        <el-descriptions-item label="文件大小">{{ formatSize(currentBackup.file_size) }}</el-descriptions-item>
        <el-descriptions-item label="文件路径" :span="2">
          <el-input v-model="currentBackup.file_path" readonly />
        </el-descriptions-item>
        <el-descriptions-item label="备份命令" :span="2">
          <el-input
            v-model="currentBackup.command"
            type="textarea"
            :rows="3"
            readonly
          />
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { logAPI, backupAPI } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'

const backups = ref([])
const loading = ref(false)
const detailDialogVisible = ref(false)
const currentBackup = ref({})
const selectedBackups = ref([])
const groupByHost = ref(false)
const activeGroups = ref([])

const filters = reactive({
  host_name: '',
  task_name: '',
  storage_type: '',
  date_range: null
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

// 按主机分组的备份数据
const groupedBackups = computed(() => {
  const groups = {}
  backups.value.forEach(backup => {
    const hostName = backup.host_name || '未知主机'
    if (!groups[hostName]) {
      groups[hostName] = []
    }
    groups[hostName].push(backup)
  })
  return groups
})

const loadBackups = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      status: 'success' // 只显示成功的备份
    }
    if (filters.host_name) {
      params.host_name = filters.host_name
    }
    if (filters.task_name) {
      params.task_name = filters.task_name
    }
    if (filters.storage_type) {
      params.storage_type = filters.storage_type
    }
    if (filters.date_range && filters.date_range.length === 2) {
      params.start_time = filters.date_range[0].toISOString()
      params.end_time = filters.date_range[1].toISOString()
    }

    const data = await logAPI.list(params)
    backups.value = data.logs || []
    pagination.total = data.total || 0
  } catch (error) {
    ElMessage.error('加载备份列表失败')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  filters.host_name = ''
  filters.task_name = ''
  filters.storage_type = ''
  filters.date_range = null
  pagination.page = 1
  loadBackups()
}

const handleSelectionChange = (selection) => {
  selectedBackups.value = selection
}

const handleDownload = async (row) => {
  try {
    const response = await backupAPI.download(row.id)
    // 创建下载链接
    const url = window.URL.createObjectURL(new Blob([response]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', row.file_path.split('/').pop())
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除备份文件 "${row.file_path}" 吗？此操作不可恢复！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await backupAPI.delete(row.id)
    ElMessage.success('删除成功')
    loadBackups()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedBackups.value.length} 个备份文件吗？此操作不可恢复！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const deletePromises = selectedBackups.value.map(backup => backupAPI.delete(backup.id))
    await Promise.all(deletePromises)

    ElMessage.success('批量删除成功')
    selectedBackups.value = []
    loadBackups()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

const showDetail = (row) => {
  currentBackup.value = row
  detailDialogVisible.value = true
}

const getStorageTypeLabel = (type) => {
  const labels = {
    local: '本地',
    ssh: 'SSH',
    s3: 'S3',
    oss: 'OSS',
    nas: 'NAS'
  }
  return labels[type] || type
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
  loadBackups()
})
</script>

<style scoped>
.backups-page {
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

.batch-actions {
  margin-top: 10px;
  display: flex;
  gap: 10px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.grouped-view {
  margin-top: 20px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}

.group-count {
  color: #909399;
  font-size: 14px;
}
</style>
