<template>
  <div class="rich-content-container">
    <div class="header">
      <h2>富文本内容管理</h2>
      <el-button type="primary" @click="showEditDialog()">创建富文本</el-button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-select v-model="filters.content_type" placeholder="选择内容类型" clearable style="width: 150px" @change="fetchList">
        <el-option label="全部" value="" />
        <el-option label="通知" value="notice" />
        <el-option label="活动" value="activity" />
        <el-option label="其他" value="other" />
      </el-select>
      <el-select v-model="filters.status" placeholder="选择状态" clearable style="width: 150px; margin-left: 10px" @change="fetchList">
        <el-option label="全部" value="" />
        <el-option label="草稿" value="draft" />
        <el-option label="已发布" value="published" />
        <el-option label="已归档" value="archived" />
      </el-select>
    </div>

    <!-- 列表 -->
    <el-table :data="tableData" border style="width: 100%; margin-top: 20px" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column prop="content_type" label="内容类型" width="120">
        <template #default="{ row }">
          <el-tag v-if="row.content_type === 'notice'" type="info">通知</el-tag>
          <el-tag v-else-if="row.content_type === 'activity'" type="warning">活动</el-tag>
          <el-tag v-else type="success">其他</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'draft'" type="info">草稿</el-tag>
          <el-tag v-else-if="row.status === 'published'" type="success">已发布</el-tag>
          <el-tag v-else type="warning">已归档</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="view_count" label="浏览次数" width="100" />
      <el-table-column prop="created_by" label="创建人" width="120" />
      <el-table-column prop="published_at" label="发布时间" width="180">
        <template #default="{ row }">
          {{ row.published_at ? formatDate(row.published_at) : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="showEditDialog(row)">编辑</el-button>
          <el-button link type="success" size="small" v-if="row.status === 'draft'" @click="handlePublish(row.id)">发布</el-button>
          <el-button link type="warning" size="small" v-if="row.status === 'published'" @click="handleArchive(row.id)">归档</el-button>
          <el-button link type="primary" size="small" @click="copyMiniAppLink(row.id)">复制小程序链接</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </div>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="90%"
      :close-on-click-modal="false"
      @close="resetForm"
    >
      <el-form :model="form" label-width="100px">
        <el-form-item label="标题" required>
          <el-input v-model="form.title" placeholder="请输入标题" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item label="内容类型" required>
          <el-select v-model="form.content_type" placeholder="请选择内容类型" style="width: 200px">
            <el-option label="通知" value="notice" />
            <el-option label="活动" value="activity" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容" required>
          <div class="editor-wrapper">
            <!-- 工具栏 -->
            <div class="editor-toolbar">
              <!-- 文本格式 -->
              <div class="toolbar-group">
                <button 
                  @click.prevent="editor?.chain().focus().toggleBold().run()" 
                  :class="{ 'is-active': editor?.isActive('bold') }" 
                  class="toolbar-btn"
                  title="加粗 (Ctrl+B)"
                >
                  <strong>B</strong>
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().toggleItalic().run()" 
                  :class="{ 'is-active': editor?.isActive('italic') }" 
                  class="toolbar-btn"
                  title="斜体 (Ctrl+I)"
                >
                  <em>I</em>
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().toggleUnderline().run()" 
                  :class="{ 'is-active': editor?.isActive('underline') }" 
                  class="toolbar-btn"
                  title="下划线 (Ctrl+U)"
                >
                  <u>U</u>
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().toggleStrike().run()" 
                  :class="{ 'is-active': editor?.isActive('strike') }" 
                  class="toolbar-btn"
                  title="删除线"
                >
                  <s>S</s>
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().toggleHighlight().run()" 
                  :class="{ 'is-active': editor?.isActive('highlight') }" 
                  class="toolbar-btn"
                  title="高亮"
                >
                  <span style="background: yellow;">高</span>
                </button>
              </div>

              <div class="toolbar-divider"></div>

              <!-- 标题 -->
              <div class="toolbar-group">
                <el-select 
                  v-model="headingLevel" 
                  placeholder="标题" 
                  style="width: 100px"
                  @change="handleHeadingChange"
                  clearable
                >
                  <el-option label="正文" value="" />
                  <el-option label="标题 1" value="1" />
                  <el-option label="标题 2" value="2" />
                  <el-option label="标题 3" value="3" />
                  <el-option label="标题 4" value="4" />
                </el-select>
              </div>

              <div class="toolbar-divider"></div>

              <!-- 列表 -->
              <div class="toolbar-group">
                <button 
                  @click.prevent="editor?.chain().focus().toggleBulletList().run()" 
                  :class="{ 'is-active': editor?.isActive('bulletList') }" 
                  class="toolbar-btn"
                  title="无序列表"
                >
                  <span>●</span> 列表
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().toggleOrderedList().run()" 
                  :class="{ 'is-active': editor?.isActive('orderedList') }" 
                  class="toolbar-btn"
                  title="有序列表"
                >
                  <span>1.</span> 列表
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().toggleBlockquote().run()" 
                  :class="{ 'is-active': editor?.isActive('blockquote') }" 
                  class="toolbar-btn"
                  title="引用"
                >
                  <span>"</span> 引用
                </button>
              </div>

              <div class="toolbar-divider"></div>

              <!-- 文本对齐 -->
              <div class="toolbar-group">
                <button 
                  @click.prevent="editor?.chain().focus().setTextAlign('left').run()" 
                  :class="{ 'is-active': editor?.isActive({ textAlign: 'left' }) }" 
                  class="toolbar-btn"
                  title="左对齐"
                >
                  ⬅ 左
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().setTextAlign('center').run()" 
                  :class="{ 'is-active': editor?.isActive({ textAlign: 'center' }) }" 
                  class="toolbar-btn"
                  title="居中"
                >
                  ⬌ 中
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().setTextAlign('right').run()" 
                  :class="{ 'is-active': editor?.isActive({ textAlign: 'right' }) }" 
                  class="toolbar-btn"
                  title="右对齐"
                >
                  ➡ 右
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().setTextAlign('justify').run()" 
                  :class="{ 'is-active': editor?.isActive({ textAlign: 'justify' }) }" 
                  class="toolbar-btn"
                  title="两端对齐"
                >
                  ⬌⬌ 两端
                </button>
              </div>

              <div class="toolbar-divider"></div>

              <!-- 链接和图片 -->
              <div class="toolbar-group">
                <button 
                  @click.prevent="showLinkDialog" 
                  :class="{ 'is-active': editor?.isActive('link') }" 
                  class="toolbar-btn"
                  title="插入链接"
                >
                  🔗 链接
                </button>
                <label class="toolbar-btn" title="上传图片" style="cursor: pointer; margin: 0;">
                  🖼 图片
                  <input 
                    type="file" 
                    ref="imageInput"
                    @change="handleImageUpload"
                    accept="image/*"
                    style="display: none;"
                  />
                </label>
                <button 
                  @click.prevent="editor?.chain().focus().setHorizontalRule().run()" 
                  class="toolbar-btn"
                  title="分割线"
                >
                  ─ 分割线
                </button>
              </div>

              <div class="toolbar-divider"></div>

              <!-- 撤销重做 -->
              <div class="toolbar-group">
                <button 
                  @click.prevent="editor?.chain().focus().undo().run()" 
                  :disabled="!editor?.can().undo()"
                  class="toolbar-btn"
                  title="撤销 (Ctrl+Z)"
                >
                  ↶ 撤销
                </button>
                <button 
                  @click.prevent="editor?.chain().focus().redo().run()" 
                  :disabled="!editor?.can().redo()"
                  class="toolbar-btn"
                  title="重做 (Ctrl+Y)"
                >
                  ↷ 重做
                </button>
              </div>
            </div>

            <!-- 编辑器内容区 -->
            <editor-content :editor="editor" class="editor-content" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 链接对话框 -->
    <el-dialog v-model="linkDialogVisible" title="插入链接" width="400px">
      <el-form>
        <el-form-item label="链接地址">
          <el-input v-model="linkUrl" placeholder="https://example.com" />
        </el-form-item>
        <el-form-item label="链接文本">
          <el-input v-model="linkText" placeholder="链接文本（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="linkDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="insertLink">确定</el-button>
        <el-button v-if="editor?.isActive('link')" type="danger" @click="removeLink">移除链接</el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Link from '@tiptap/extension-link'
import TextAlign from '@tiptap/extension-text-align'
import Color from '@tiptap/extension-color'
import Highlight from '@tiptap/extension-highlight'
import Image from '@tiptap/extension-image'
import {
  getRichContentList,
  getRichContent,
  createRichContent,
  updateRichContent,
  publishRichContent,
  archiveRichContent,
  deleteRichContent,
  uploadImage
} from '../api/richContent'

// 数据
const tableData = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const saving = ref(false)
const editingId = ref(null)

// 筛选条件
const filters = reactive({
  content_type: '',
  status: ''
})

// 分页
const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

// 表单
const form = reactive({
  title: '',
  content: '',
  content_type: 'notice'
})

// 链接对话框
const linkDialogVisible = ref(false)
const linkUrl = ref('')
const linkText = ref('')

// 图片上传
const imageInput = ref(null)
const uploadingImage = ref(false)

// 标题级别
const headingLevel = ref('')

// Tiptap 编辑器
const editor = useEditor({
  extensions: [
    StarterKit.configure({
      heading: {
        levels: [1, 2, 3, 4]
      }
    }),
    Underline,
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        class: 'editor-link'
      }
    }),
    TextAlign.configure({
      types: ['heading', 'paragraph']
    }),
    Color,
    Highlight.configure({
      multicolor: true
    }),
    Image.configure({
      inline: true,
      allowBase64: true,
      HTMLAttributes: {
        class: 'editor-image'
      }
    })
  ],
  content: '',
  editorProps: {
    attributes: {
      class: 'prose prose-sm sm:prose lg:prose-lg xl:prose-2xl mx-auto focus:outline-none'
    }
  },
  onUpdate: ({ editor }) => {
    form.content = editor.getHTML()
    // 更新标题级别
    if (editor.isActive('heading')) {
      const level = editor.getAttributes('heading').level
      headingLevel.value = level ? String(level) : ''
    } else {
      headingLevel.value = ''
    }
  }
})

// 监听标题级别变化
const handleHeadingChange = (value) => {
  if (!editor.value) return
  
  if (value === '') {
    editor.value.chain().focus().setParagraph().run()
  } else {
    editor.value.chain().focus().toggleHeading({ level: parseInt(value) }).run()
  }
}

// 显示链接对话框
const showLinkDialog = () => {
  if (editor.value?.isActive('link')) {
    const attrs = editor.value.getAttributes('link')
    linkUrl.value = attrs.href || ''
    linkText.value = editor.value.getText() || ''
  } else {
    linkUrl.value = ''
    linkText.value = editor.value?.getText() || ''
  }
  linkDialogVisible.value = true
}

// 插入链接
const insertLink = () => {
  if (!linkUrl.value) {
    ElMessage.warning('请输入链接地址')
    return
  }
  
  if (editor.value?.isActive('link')) {
    // 更新现有链接
    editor.value.chain().focus().extendMarkRange('link').setLink({ href: linkUrl.value }).run()
  } else {
    // 插入新链接
    if (linkText.value) {
      editor.value?.chain().focus().insertContent(`<a href="${linkUrl.value}">${linkText.value}</a>`).run()
    } else {
      editor.value?.chain().focus().setLink({ href: linkUrl.value }).run()
    }
  }
  
  linkDialogVisible.value = false
  linkUrl.value = ''
  linkText.value = ''
}

// 移除链接
const removeLink = () => {
  editor.value?.chain().focus().unsetLink().run()
  linkDialogVisible.value = false
}

// 处理图片上传
const handleImageUpload = async (event) => {
  const file = event.target.files?.[0]
  if (!file) return

  // 文件类型校验
  const isImage = /\.(jpg|jpeg|png|gif|webp)$/i.test(file.name)
  if (!isImage) {
    ElMessage.error('请上传JPG、PNG、GIF或WEBP格式的图片')
    event.target.value = '' // 清空选择
    return
  }

  // 文件大小校验（5MB）
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过5MB')
    event.target.value = '' // 清空选择
    return
  }

  uploadingImage.value = true
  ElMessage({ message: '图片上传中...', type: 'info', duration: 0 })

  try {
    // 创建FormData
    const formData = new FormData()
    formData.append('file', file)

    // 上传图片
    const response = await uploadImage(formData)

    // 处理上传结果
    if (response.code === 200 && response.data && response.data.imageUrl) {
      const imageUrl = response.data.imageUrl
      
      // 插入图片到编辑器
      editor.value?.chain().focus().setImage({ 
        src: imageUrl,
        alt: file.name
      }).run()
      
      ElMessage.closeAll()
      ElMessage.success('图片上传成功')
    } else {
      ElMessage.closeAll()
      ElMessage.error('上传失败：' + (response.message || '未知错误'))
    }
  } catch (error) {
    ElMessage.closeAll()
    console.error('图片上传失败:', error)
    ElMessage.error('图片上传失败，请稍后重试')
  } finally {
    uploadingImage.value = false
    // 清空文件选择，以便可以重复选择同一文件
    if (event.target) {
      event.target.value = ''
    }
  }
}

// 获取列表
const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      ...filters
    }
    const res = await getRichContentList(params)
    tableData.value = res.data
    pagination.total = res.pagination.total
  } catch (error) {
    ElMessage.error('获取列表失败')
  } finally {
    loading.value = false
  }
}

// 显示编辑对话框
const showEditDialog = async (row) => {
  if (row) {
    // 编辑模式
    dialogTitle.value = '编辑富文本内容'
    editingId.value = row.id
    try {
      const res = await getRichContent(row.id)
      form.title = res.data.title
      form.content = res.data.content
      form.content_type = res.data.content_type
      editor.value?.commands.setContent(res.data.content)
    } catch (error) {
      ElMessage.error('获取详情失败')
      return
    }
  } else {
    // 创建模式
    dialogTitle.value = '创建富文本内容'
    editingId.value = null
    form.title = ''
    form.content = ''
    form.content_type = 'notice'
    editor.value?.commands.setContent('')
  }
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  form.title = ''
  form.content = ''
  form.content_type = 'notice'
  editingId.value = null
  editor.value?.commands.setContent('')
  headingLevel.value = ''
}

// 保存
const handleSave = async () => {
  if (!form.title) {
    ElMessage.warning('请输入标题')
    return
  }
  if (!form.content || form.content === '<p></p>') {
    ElMessage.warning('请输入内容')
    return
  }
  
  saving.value = true
  try {
    const data = {
      title: form.title,
      content: form.content,
      content_type: form.content_type
    }
    
    if (editingId.value) {
      await updateRichContent(editingId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createRichContent(data)
      ElMessage.success('创建成功')
    }
    
    dialogVisible.value = false
    fetchList()
  } catch (error) {
    ElMessage.error(editingId.value ? '更新失败' : '创建失败')
  } finally {
    saving.value = false
  }
}

// 发布
const handlePublish = async (id) => {
  try {
    await ElMessageBox.confirm('确定要发布该内容吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await publishRichContent(id)
    ElMessage.success('发布成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('发布失败')
    }
  }
}

// 归档
const handleArchive = async (id) => {
  try {
    await ElMessageBox.confirm('确定要归档该内容吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await archiveRichContent(id)
    ElMessage.success('归档成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('归档失败')
    }
  }
}

// 删除
const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该内容吗？此操作不可恢复！', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteRichContent(id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 复制小程序链接
const copyMiniAppLink = (id) => {
  const link = `/pages/rich-content/rich-content?id=${id}`
  navigator.clipboard.writeText(link).then(() => {
    ElMessage.success('小程序路径已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败，请手动复制：' + link)
  })
}

// 格式化日期
const formatDate = (dateStr) => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  fetchList()
})

onBeforeUnmount(() => {
  editor.value?.destroy()
})
</script>

<style scoped>
.rich-content-container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filter-bar {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.editor-wrapper {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  width: 100%;
  background: #fff;
}

.editor-toolbar {
  border-bottom: 1px solid #dcdfe6;
  padding: 8px 12px;
  background-color: #fafafa;
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-divider {
  width: 1px;
  height: 24px;
  background-color: #e0e0e0;
  margin: 0 4px;
}

.toolbar-btn {
  padding: 6px 12px;
  border: 1px solid #e0e0e0;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
  color: #333;
}

.toolbar-btn:hover:not(:disabled) {
  background-color: #f0f0f0;
  border-color: #409eff;
}

.toolbar-btn.is-active {
  background-color: #409eff;
  color: white;
  border-color: #409eff;
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.toolbar-group label.toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin: 0;
}

.editor-content {
  min-height: 400px;
  padding: 20px;
}

.editor-content :deep(.ProseMirror) {
  min-height: 400px;
  outline: none;
  line-height: 1.6;
}

.editor-content :deep(.ProseMirror p) {
  margin: 0.5em 0;
}

.editor-content :deep(.ProseMirror h1) {
  font-size: 2em;
  font-weight: bold;
  margin: 0.67em 0;
}

.editor-content :deep(.ProseMirror h2) {
  font-size: 1.5em;
  font-weight: bold;
  margin: 0.75em 0;
}

.editor-content :deep(.ProseMirror h3) {
  font-size: 1.17em;
  font-weight: bold;
  margin: 0.83em 0;
}

.editor-content :deep(.ProseMirror h4) {
  font-size: 1em;
  font-weight: bold;
  margin: 0.83em 0;
}

.editor-content :deep(.ProseMirror ul),
.editor-content :deep(.ProseMirror ol) {
  padding-left: 2em;
  margin: 0.5em 0;
}

.editor-content :deep(.ProseMirror li) {
  margin: 0.3em 0;
}

.editor-content :deep(.ProseMirror blockquote) {
  border-left: 3px solid #dcdfe6;
  padding-left: 1em;
  margin-left: 0;
  color: #666;
  font-style: italic;
}

.editor-content :deep(.ProseMirror hr) {
  border: none;
  border-top: 2px solid #dcdfe6;
  margin: 1em 0;
}

.editor-content :deep(.ProseMirror strong) {
  font-weight: bold;
}

.editor-content :deep(.ProseMirror em) {
  font-style: italic;
}

.editor-content :deep(.ProseMirror u) {
  text-decoration: underline;
}

.editor-content :deep(.ProseMirror s) {
  text-decoration: line-through;
}

.editor-content :deep(.ProseMirror mark) {
  background-color: #fef08a;
  padding: 2px 4px;
  border-radius: 2px;
}

.editor-content :deep(.editor-link) {
  color: #409eff;
  text-decoration: underline;
  cursor: pointer;
}

.editor-content :deep(.editor-link:hover) {
  color: #66b1ff;
}

.editor-content :deep(.editor-image) {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 10px 0;
  border-radius: 4px;
}

.editor-content :deep(.ProseMirror[style*="text-align: left"]) {
  text-align: left;
}

.editor-content :deep(.ProseMirror[style*="text-align: center"]) {
  text-align: center;
}

.editor-content :deep(.ProseMirror[style*="text-align: right"]) {
  text-align: right;
}

.editor-content :deep(.ProseMirror[style*="text-align: justify"]) {
  text-align: justify;
}
</style>
