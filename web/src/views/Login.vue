<template>
  <div class="login-container">
    <!-- 激励性标题 - 放在页面顶部，彩虹色 -->
    <div class="login-motto-top">
      <p class="login-motto-cn">财富自由，从这里开始</p>
      <p class="login-motto-en">Financial Freedom Starts Here</p>
    </div>

    <div class="login-wrapper">
      <div class="login-content">
        <template v-if="activeTab === 'password'">
          <form @submit.prevent="handlePasswordLogin" class="login-form">
            <Input
              v-model="passwordForm.username"
              label="用户名"
              placeholder="请输入用户名"
              :error="errors.username"
            />
            <Input
              v-model="passwordForm.password"
              type="password"
              label="密码"
              placeholder="请输入密码"
              :error="errors.password"
              @keyup.enter="handlePasswordLogin"
            />
            <Button
              type="primary"
              size="large"
              :loading="loading"
              block
              @click="handlePasswordLogin"
            >
              登录
            </Button>
          </form>
          <div class="login-switch">
            <button class="login-switch-link" @click="activeTab = 'oauth2'">
              使用第三方登录
            </button>
          </div>
        </template>

        <template v-else-if="activeTab === 'oauth2'">
          <div class="oauth-buttons">
            <Button
              type="default"
              size="large"
              :loading="oauthLoading"
              block
              @click="handleOAuth2Login('google')"
              class="oauth-button google-button"
            >
              <template #icon>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                  <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                  <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                  <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                  <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
                </svg>
              </template>
              使用 Google 登录
            </Button>
            <Button
              type="default"
              size="large"
              :loading="oauthLoading"
              block
              @click="handleOAuth2Login('github')"
              class="oauth-button github-button"
            >
              <template #icon>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"></path>
                </svg>
              </template>
              使用 GitHub 登录
            </Button>
          </div>
          <div class="login-switch">
            <button class="login-switch-link" @click="activeTab = 'password'">
              使用密码登录
            </button>
          </div>
        </template>
      </div>
    </div>
    
    <!-- 错误提示 Dialog - 居中弹窗，背景模糊 -->
    <Dialog
      v-model="errorDialogVisible"
      title="登录失败"
      :closeOnClickOverlay="true"
    >
      <p style="margin: 0; color: var(--color-text-regular); line-height: 1.6;">{{ errorMessage }}</p>
      <template #footer>
        <Button
          type="primary"
          @click="errorDialogVisible = false"
        >
          确定
        </Button>
      </template>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authAPI } from '@/api/auth'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import Dialog from '@/components/ui/Dialog.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const activeTab = ref('oauth2')
const loading = ref(false)
const oauthLoading = ref(false)
const errorDialogVisible = ref(false)
const errorMessage = ref('')

const passwordForm = reactive({
  username: '',
  password: '',
})

const errors = reactive({
  username: '',
  password: '',
})


const validateForm = () => {
  errors.username = ''
  errors.password = ''
  
  if (!passwordForm.username.trim()) {
    errors.username = '请输入用户名'
    return false
  }
  
  if (!passwordForm.password) {
    errors.password = '请输入密码'
    return false
  }
  
  return true
}

const handlePasswordLogin = async () => {
  if (!validateForm()) return

  loading.value = true
  try {
    const response = await authStore.login(passwordForm.username, passwordForm.password)

    if (response.status === 'require_mfa') {
      if ((window as any).$toast) {
        ;(window as any).$toast.info('需要 MFA 验证')
      }
      // TODO: 处理 MFA 流程
    } else if (response.status === 'success') {
      if ((window as any).$toast) {
        ;(window as any).$toast.success('登录成功')
      }
      const redirect = (route.query.redirect as string) || '/dashboard'
      router.push(redirect)
    }
  } catch (error: any) {
    // 详细错误信息输出到控制台
    console.error('登录失败:', error)
    if (error.response) {
      console.error('响应数据:', error.response.data)
      console.error('状态码:', error.response.status)
    }
    
    // 显示用户友好的错误消息（使用 Dialog）
    const userMessage = error.response?.data?.message || '登录失败，请稍后再试'
    errorMessage.value = userMessage
    errorDialogVisible.value = true
  } finally {
    loading.value = false
  }
}

const handleOAuth2Login = async (provider: string) => {
  oauthLoading.value = true
  try {
    const response = await authAPI.oauth2Init(provider)
    window.location.href = response.redirect_url
  } catch (error: any) {
    oauthLoading.value = false
    // 详细错误信息输出到控制台
    console.error(`OAuth2 登录失败 (${provider}):`, error)
    if (error.response) {
      console.error('响应数据:', error.response.data)
      console.error('状态码:', error.response.status)
    }
    
    // 显示用户友好的错误消息（使用 Dialog）
    errorMessage.value = 'OAuth2 登录失败，请稍后再试'
    errorDialogVisible.value = true
  }
}

</script>

<style scoped>
/* Apple/Google 风格登录页面 - 极简清爽设计 */
.login-container {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  align-items: center;
  min-height: 100vh;
  padding: 60px 24px 40px;
  background: #ffffff;
  position: relative;
  overflow: hidden;
  animation: fadeIn 0.4s ease-out;
}

/* 极简背景光效层 */
.login-container::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: 
    radial-gradient(circle at 20% 30%, rgba(0, 122, 255, 0.02) 0%, transparent 50%),
    radial-gradient(circle at 80% 70%, rgba(0, 122, 255, 0.015) 0%, transparent 50%);
  animation: float 30s ease-in-out infinite;
  pointer-events: none;
  z-index: 0;
}

/* 登录容器 */
.login-wrapper {
  width: 100%;
  max-width: 420px;
  position: relative;
  z-index: 1;
  animation: fadeInUp 0.4s ease-out 0.1s both;
  padding: 0;
  margin-bottom: auto;
}

/* 登录方式切换链接 - 小字，不显眼 */
.login-switch {
  text-align: center;
  margin-top: 20px;
}

.login-switch-link {
  font-size: 13px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
  color: #86868b;
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  transition: color 0.2s ease;
  font-weight: 400;
  text-decoration: none;
}

.login-switch-link:hover {
  color: #007AFF;
}

/* 激励性标题 - 放在页面顶部，彩虹色 */
.login-motto-top {
  text-align: center;
  margin-bottom: 40px;
  z-index: 2;
  position: relative;
}

.login-motto-cn {
  margin: 0 0 8px 0;
  font-size: 20px;
  font-weight: 500;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
  letter-spacing: 0.5px;
  line-height: 1.4;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 25%, #f093fb 50%, #4facfe 75%, #00f2fe 100%);
  background-size: 200% 200%;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: gradientShift 3s ease infinite;
}

.login-motto-en {
  margin: 0;
  font-size: 14px;
  font-weight: 400;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
  letter-spacing: 1px;
  line-height: 1.5;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 25%, #f093fb 50%, #4facfe 75%, #00f2fe 100%);
  background-size: 200% 200%;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: gradientShift 3s ease infinite;
  text-transform: uppercase;
}

@keyframes gradientShift {
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}

.login-content {
  width: 100%;
  margin-top: 0;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.oauth-buttons {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.oauth-button :deep(.ui-button) {
  width: 100% !important;
  min-height: 52px !important;
  font-size: 15px !important;
  font-weight: 500 !important;
  border-radius: 8px !important;
  transition: all 0.2s ease !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 12px !important;
}

.oauth-button :deep(.ui-button__icon) {
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  margin: 0 !important;
}

.oauth-button :deep(.ui-button__text) {
  flex: none !important;
}

/* Google 风格按钮 - 使用 :deep() 覆盖 Button 组件样式 */
.google-button :deep(.ui-button) {
  background-color: #ffffff !important;
  background: #ffffff !important;
  border: 1px solid #dadce0 !important;
  color: #3c4043 !important;
  box-shadow: 0 1px 2px 0 rgba(60, 64, 67, 0.1) !important;
}

.google-button :deep(.ui-button):hover:not(.ui-button--disabled) {
  background-color: #f8f9fa !important;
  background: #f8f9fa !important;
  box-shadow: 0 1px 3px 0 rgba(60, 64, 67, 0.2) !important;
}

.google-button :deep(.ui-button):active:not(.ui-button--disabled) {
  background-color: #f1f3f4 !important;
  background: #f1f3f4 !important;
  box-shadow: 0 1px 2px 0 rgba(60, 64, 67, 0.1) !important;
}

/* GitHub 风格按钮 - 使用 :deep() 覆盖 Button 组件样式 */
.github-button :deep(.ui-button) {
  background-color: #24292e !important;
  background: #24292e !important;
  border: none !important;
  color: #ffffff !important;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1) !important;
}

.github-button :deep(.ui-button):hover:not(.ui-button--disabled) {
  background-color: #1a1e22 !important;
  background: #1a1e22 !important;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.2) !important;
}

.github-button :deep(.ui-button):active:not(.ui-button--disabled) {
  background-color: #0d1117 !important;
  background: #0d1117 !important;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1) !important;
}


/* iOS 26 动画效果 */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(15px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) rotate(0deg);
  }
  33% {
    transform: translate(30px, -30px) rotate(120deg);
  }
  66% {
    transform: translate(-20px, 20px) rotate(240deg);
  }
}

@media (max-width: 767px) {
  .login-container {
    padding: 20px;
  }
  
  .login-wrapper {
    max-width: 100%;
  }
  
  .login-switch-link {
    font-size: 12px;
  }
  
  .login-container {
    padding: 40px 20px 32px;
  }
  
  .login-motto-top {
    margin-bottom: 32px;
  }
  
  .login-motto-cn {
    font-size: 18px;
  }
  
  .login-motto-en {
    font-size: 12px;
  }
}
</style>
