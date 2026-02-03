<template>
  <div class="dashboard">
    <h2 class="page-title">仪表盘</h2>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card" @click="$router.push('/hosts')">
          <div class="stat-content">
            <div class="stat-icon host">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.host_count }}</div>
              <div class="stat-label">主机数量</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" @click="$router.push('/tasks')">
          <div class="stat-content">
            <div class="stat-icon task">
              <el-icon><Calendar /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.task_count }}</div>
              <div class="stat-label">备份任务</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" @click="goToLogs('success')">
          <div class="stat-content">
            <div class="stat-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.success_count }}</div>
              <div class="stat-label">成功备份</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" @click="goToLogs('failed')">
          <div class="stat-content">
            <div class="stat-icon failed">
              <el-icon><CircleCloseFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.failed_count }}</div>
              <div class="stat-label">失败备份</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近备份日志 -->
    <el-card class="recent-logs">
      <template #header>
        <div class="card-header">
          <span>最近备份日志</span>
          <el-button
            type="primary"
            size="small"
            @click="$router.push('/logs')"
          >
            查看全部
          </el-button>
        </div>
      </template>

      <el-table :data="recentLogs" stripe>
        <el-table-column prop="task_name" label="任务名称" />
        <el-table-column prop="host_name" label="主机" />
        <el-table-column prop="backup_type" label="备份类型" />
        <el-table-column prop="status" label="状态">
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
        <el-table-column prop="duration" label="耗时">
          <template #default="{ row }">
            {{ row.duration }}秒
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { dashboardAPI, logAPI } from '../api'
import { ElMessage } from 'element-plus'

const router = useRouter()

const stats = ref({
  host_count: 0,
  task_count: 0,
  success_count: 0,
  failed_count: 0
})

const recentLogs = ref([])

const loadStats = async () => {
  try {
    const data = await dashboardAPI.stats()
    stats.value = data
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  }
}

const loadRecentLogs = async () => {
  try {
    const data = await logAPI.list({ page: 1, page_size: 10 })
    recentLogs.value = data.logs || []
  } catch (error) {
    ElMessage.error('加载日志失败')
  }
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const goToLogs = (status) => {
  // 跳转到日志页面并传递状态参数
  router.push({ path: '/logs', query: { status } })
}

onMounted(() => {
  loadStats()
  loadRecentLogs()
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 24px;
  color: #303133;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  cursor: pointer;
  transition: transform 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
}

.stat-icon.host {
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
}

.stat-icon.task {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.success {
  background: linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%);
}

.stat-icon.failed {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.recent-logs {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

