<template>
  <div class="notifications-page">
    <h2 class="page-title">通知管理</h2>

    <el-card>
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          添加通知
        </el-button>
      </div>

      <el-table :data="notifications" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.type }}</el-tag>
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
              测试通知
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
          <el-input v-model="form.name" placeholder="请输入通知名称" />
        </el-form-item>

        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" style="width: 100%" @change="handleTypeChange">
            <el-option label="邮件" value="email" />
            <el-option label="Webhook" value="webhook" />
          </el-select>
        </el-form-item>

        <!-- 邮件配置 -->
        <template v-if="form.type === 'email'">
          <el-form-item label="邮箱提供商">
            <el-select v-model="emailProvider" placeholder="选择邮箱提供商快速配置" style="width: 100%" @change="handleProviderChange" clearable>
              <el-option label="163邮箱" value="163" />
              <el-option label="QQ邮箱" value="qq" />
              <el-option label="Gmail" value="gmail" />
              <el-option label="Outlook" value="outlook" />
              <el-option label="自定义" value="custom" />
            </el-select>
          </el-form-item>
          <el-form-item label="SMTP主机">
            <el-input v-model="configForm.smtp_host" placeholder="smtp.gmail.com" />
          </el-form-item>
          <el-form-item label="SMTP端口">
            <el-input-number v-model="configForm.smtp_port" :min="1" :max="65535" style="width: 100%" />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="configForm.username" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="configForm.password" type="password" show-password />
          </el-form-item>
          <el-form-item label="发件人">
            <el-input v-model="configForm.from" />
          </el-form-item>
          <el-form-item label="收件人">
            <el-input v-model="toInput" placeholder="多个邮箱用逗号分隔" />
          </el-form-item>
          <el-form-item label="使用SSL">
            <el-switch v-model="configForm.use_ssl" />
          </el-form-item>
        </template>

        <!-- Webhook配置 -->
        <template v-if="form.type === 'webhook'">
          <el-form-item label="Webhook类型">
            <el-select v-model="configForm.webhook_type" style="width: 100%">
              <el-option label="钉钉" value="dingtalk" />
              <el-option label="企业微信" value="wecom" />
              <el-option label="飞书" value="feishu" />
              <el-option label="Slack" value="slack" />
              <el-option label="自定义" value="custom" />
            </el-select>
          </el-form-item>
          <el-form-item label="Webhook URL">
            <el-input v-model="configForm.webhook_url" type="textarea" :rows="2" placeholder="https://..." />
          </el-form-item>
          <el-form-item label="Secret" v-if="configForm.webhook_type === 'dingtalk'">
            <el-input v-model="configForm.secret" placeholder="钉钉机器人签名密钥（可选）" />
          </el-form-item>
          <el-form-item label="消息模板">
            <el-tabs v-model="activeTemplate">
              <el-tab-pane label="成功通知" name="success">
                <el-input
                  v-model="configForm.template_success"
                  type="textarea"
                  :rows="6"
                  placeholder="备份成功通知模板"
                />
                <div class="template-help">
                  可用变量: {{task_name}}, {{host_name}}, {{databases}}, {{file_size}}, {{duration}}, {{start_time}}, {{end_time}}
                </div>
              </el-tab-pane>
              <el-tab-pane label="失败通知" name="failure">
                <el-input
                  v-model="configForm.template_failure"
                  type="textarea"
                  :rows="6"
                  placeholder="备份失败通知模板"
                />
                <div class="template-help">
                  可用变量: {{task_name}}, {{host_name}}, {{databases}}, {{error_message}}, {{start_time}}
                </div>
              </el-tab-pane>
            </el-tabs>
          </el-form-item>
        </template>

        <!-- 钉钉配置 (兼容旧版) -->
        <template v-if="form.type === 'dingtalk'">
          <el-form-item label="Webhook URL">
            <el-input v-model="configForm.webhook_url" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item label="Secret">
            <el-input v-model="configForm.secret" placeholder="可选" />
          </el-form-item>
        </template>

        <!-- 企业微信配置 (兼容旧版) -->
        <template v-if="form.type === 'wecom'">
          <el-form-item label="Webhook URL">
            <el-input v-model="configForm.webhook_url" type="textarea" :rows="2" />
          </el-form-item>
        </template>

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
import { notificationAPI } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const notifications = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('添加通知')
const formRef = ref(null)
const submitting = ref(false)
const testingId = ref(null)
const toInput = ref('')
const activeTemplate = ref('success')
const emailProvider = ref('')

const form = ref({
  name: '',
  type: 'email',
  config: '{}',
  status: 1
})

const configForm = reactive({
  smtp_host: '',
  smtp_port: 465,
  username: '',
  password: '',
  from: '',
  to: [],
  use_ssl: true,
  webhook_url: '',
  webhook_type: 'dingtalk',
  secret: '',
  template_success: '✅ 备份成功\n\n任务: {{task_name}}\n主机: {{host_name}}\n数据库: {{databases}}\n文件大小: {{file_size}}\n耗时: {{duration}}秒\n开始时间: {{start_time}}\n结束时间: {{end_time}}',
  template_failure: '❌ 备份失败\n\n任务: {{task_name}}\n主机: {{host_name}}\n数据库: {{databases}}\n错误信息: {{error_message}}\n开始时间: {{start_time}}'
})

const rules = {
  name: [{ required: true, message: '请输入通知名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择通知类型', trigger: 'change' }]
}

const loadNotifications = async () => {
  loading.value = true
  try {
    notifications.value = await notificationAPI.list()
  } catch (error) {
    ElMessage.error('加载通知列表失败')
  } finally {
    loading.value = false
  }
}

const handleTypeChange = () => {
  Object.keys(configForm).forEach(key => {
    if (key === 'smtp_port') {
      configForm[key] = 465
    } else if (key === 'use_ssl') {
      configForm[key] = true
    } else if (key === 'to') {
      configForm[key] = []
    } else if (key === 'webhook_type') {
      configForm[key] = 'dingtalk'
    } else if (key === 'template_success') {
      configForm[key] = '✅ 备份成功\n\n任务: {{task_name}}\n主机: {{host_name}}\n数据库: {{databases}}\n文件大小: {{file_size}}\n耗时: {{duration}}秒\n开始时间: {{start_time}}\n结束时间: {{end_time}}'
    } else if (key === 'template_failure') {
      configForm[key] = '❌ 备份失败\n\n任务: {{task_name}}\n主机: {{host_name}}\n数据库: {{databases}}\n错误信息: {{error_message}}\n开始时间: {{start_time}}'
    } else {
      configForm[key] = ''
    }
  })
  emailProvider.value = ''
}

const handleProviderChange = (provider) => {
  const providers = {
    '163': {
      smtp_host: 'smtp.163.com',
      smtp_port: 465,
      use_ssl: true
    },
    'qq': {
      smtp_host: 'smtp.qq.com',
      smtp_port: 465,
      use_ssl: true
    },
    'gmail': {
      smtp_host: 'smtp.gmail.com',
      smtp_port: 587,
      use_ssl: true
    },
    'outlook': {
      smtp_host: 'smtp.office365.com',
      smtp_port: 587,
      use_ssl: true
    }
  }

  if (provider && provider !== 'custom' && providers[provider]) {
    const config = providers[provider]
    configForm.smtp_host = config.smtp_host
    configForm.smtp_port = config.smtp_port
    configForm.use_ssl = config.use_ssl
  }
}

const handleAdd = () => {
  dialogTitle.value = '添加通知'
  form.value = {
    name: '',
    type: 'email',
    config: '{}',
    status: 1
  }
  handleTypeChange()
  toInput.value = ''
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑通知'
  form.value = { ...row }

  try {
    const config = JSON.parse(row.config || '{}')
    Object.keys(configForm).forEach(key => {
      if (key === 'to' && Array.isArray(config[key])) {
        configForm[key] = config[key]
        toInput.value = config[key].join(',')
      } else {
        configForm[key] = config[key] !== undefined ? config[key] : configForm[key]
      }
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
      if (key === 'to' && toInput.value) {
        config[key] = toInput.value.split(',').map(email => email.trim()).filter(email => email)
      } else if (configForm[key] !== '' && configForm[key] !== undefined) {
        config[key] = configForm[key]
      }
    })
    form.value.config = JSON.stringify(config)

    submitting.value = true
    try {
      if (form.value.id) {
        await notificationAPI.update(form.value.id, form.value)
        ElMessage.success('更新成功')
      } else {
        await notificationAPI.create(form.value)
        ElMessage.success('添加成功')
      }
      dialogVisible.value = false
      loadNotifications()
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定要删除该通知吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })

  try {
    await notificationAPI.delete(row.id)
    ElMessage.success('删除成功')
    loadNotifications()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const handleTest = async (row) => {
  testingId.value = row.id
  try {
    const result = await notificationAPI.test(row.id)
    if (result.success) {
      ElMessage.success('测试通知发送成功')
    } else {
      ElMessage.error('发送失败：' + result.error)
    }
  } catch (error) {
    ElMessage.error('测试通知失败')
  } finally {
    testingId.value = null
  }
}

onMounted(() => {
  loadNotifications()
})
</script>

<style scoped>
.notifications-page {
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

.template-help {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}
</style>
