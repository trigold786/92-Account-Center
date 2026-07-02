<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
      <div>
        <h2 style="margin:0;color:var(--text-primary)">财务后台</h2>
        <p style="margin:6px 0 0;color:var(--text-secondary)">订单、退款、发票与对账管理</p>
      </div>
      <el-button type="primary" @click="refreshAll">刷新</el-button>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="订单" name="orders" v-if="hasPermission('finance.order.view')">
        <el-card>
          <template #header>
            <div style="display:flex;gap:12px;flex-wrap:wrap">
              <el-select v-model="orderQuery.status" clearable placeholder="订单状态" aria-label="订单状态筛选" style="width:140px">
                <el-option label="待支付" value="pending" />
                <el-option label="已支付" value="paid" />
                <el-option label="已取消" value="cancelled" />
                <el-option label="已退款" value="refunded" />
              </el-select>
              <el-input v-model="orderQuery.user_id" placeholder="用户ID" aria-label="用户ID查询" style="width:160px" />
              <el-button @click="loadOrders">查询</el-button>
            </div>
          </template>
          <div style="overflow-x: auto; width: 100%">
          <el-table :data="orders" stripe v-loading="ordersLoading">
            <el-table-column prop="order_no" label="订单号" min-width="160" />
            <el-table-column prop="user_id" label="用户" width="90" />
            <el-table-column prop="product_type" label="类型" width="120" />
            <el-table-column prop="product_name" label="商品" min-width="160" />
            <el-table-column label="金额" width="110"><template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template></el-table-column>
            <el-table-column prop="payment_method" label="支付方式" width="120" />
            <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="orderStatusType(row.status)">{{ row.status }}</el-tag></template></el-table-column>
            <el-table-column prop="created_at" label="创建时间" min-width="180" />
          </el-table>
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="退款" name="refunds" v-if="hasPermission('finance.refund.approve')">
        <el-card>
          <template #header><el-button @click="loadRefunds">刷新退款</el-button></template>
          <el-table :data="refunds" stripe v-loading="refundsLoading">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column label="退款单号" min-width="160"><template #default="{ row }">{{ row.refund_no || '-' }}</template></el-table-column>
            <el-table-column prop="order_id" label="订单ID" width="100" />
            <el-table-column prop="user_id" label="用户" width="90" />
            <el-table-column label="金额" width="110"><template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template></el-table-column>
            <el-table-column prop="reason" label="原因" min-width="180" />
            <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="row.status === 'approved' ? 'success' : row.status === 'rejected' ? 'danger' : 'warning'">{{ row.status }}</el-tag></template></el-table-column>
            <el-table-column label="渠道" width="100"><template #default="{ row }">{{ row.provider || '-' }}</template></el-table-column>
            <el-table-column label="渠道退款号" min-width="160"><template #default="{ row }">{{ row.provider_refund_id || '-' }}</template></el-table-column>
            <el-table-column label="渠道状态" width="110"><template #default="{ row }">{{ row.provider_status || '-' }}</template></el-table-column>
            <el-table-column label="渠道错误" min-width="180"><template #default="{ row }">{{ row.provider_error || '-' }}</template></el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button size="small" type="success" :disabled="row.status !== 'pending'" @click="approve(row)">批准</el-button>
                <el-button size="small" type="danger" :disabled="row.status !== 'pending'" @click="reject(row)">驳回</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="发票" name="invoices" v-if="hasPermission('finance.invoice.manage')">
        <el-card>
          <template #header>
            <div style="display:flex;justify-content:space-between">
              <el-button @click="loadInvoices">刷新发票</el-button>
              <el-button type="primary" @click="invoiceDialog = true">创建发票</el-button>
            </div>
          </template>
          <el-table :data="invoices" stripe v-loading="invoicesLoading">
            <el-table-column prop="invoice_no" label="发票号" min-width="150" />
            <el-table-column prop="order_id" label="订单ID" width="100" />
            <el-table-column prop="title" label="抬头" min-width="160" />
            <el-table-column prop="email" label="邮箱" min-width="180" />
            <el-table-column label="金额" width="110"><template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template></el-table-column>
            <el-table-column prop="status" label="状态" width="100" />
            <el-table-column prop="created_at" label="创建时间" min-width="180" />
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="对账" name="reconcile" v-if="hasPermission('finance.order.view')">
        <el-card>
          <div style="display:flex;gap:12px;align-items:center;flex-wrap:wrap">
            <el-select v-model="reconcileProvider" placeholder="渠道" style="width:140px">
              <el-option label="微信" value="wechat" />
              <el-option label="支付宝" value="alipay" />
            </el-select>
            <el-date-picker v-model="reconcileDate" type="date" value-format="YYYY-MM-DD" placeholder="日期" />
            <el-button type="primary" @click="runReconcile">开始对账</el-button>
          </div>
          <el-descriptions v-if="reconcileReport" :column="2" border style="margin-top:16px">
            <el-descriptions-item label="报告ID">{{ reconcileReport.id }}</el-descriptions-item>
            <el-descriptions-item label="状态">{{ reconcileReport.status }}</el-descriptions-item>
            <el-descriptions-item label="订单总数">{{ reconcileReport.total_orders }}</el-descriptions-item>
            <el-descriptions-item label="匹配订单">{{ reconcileReport.matched_orders }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="invoiceDialog" title="创建发票" width="min(520px, 90vw)">
      <el-form :model="invoiceForm" label-width="90px">
        <el-form-item label="用户ID"><el-input-number v-model="invoiceForm.user_id" :min="1" style="width:100%" /></el-form-item>
        <el-form-item label="订单ID"><el-input-number v-model="invoiceForm.order_id" :min="1" style="width:100%" /></el-form-item>
        <el-form-item label="抬头"><el-input v-model="invoiceForm.title" /></el-form-item>
        <el-form-item label="税号"><el-input v-model="invoiceForm.tax_id" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="invoiceForm.email" /></el-form-item>
        <el-form-item label="金额"><el-input-number v-model="invoiceForm.amount" :min="0.01" :precision="2" style="width:100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="invoiceDialog = false">取消</el-button>
        <el-button type="primary" @click="submitInvoice">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { usePermissionStore } from '@/store/permission'
import { approveRefund, createInvoice, listInvoices, listOrders, listRefunds, reconcile, rejectRefund } from '@/api/finance'
import type { PaymentOrder, Refund } from '@/types/api'

const perm = usePermissionStore()
const hasPermission = (p: string) => perm.hasPermission(p)
const activeTab = ref('orders')

const orders = ref<PaymentOrder[]>([])
const refunds = ref<Refund[]>([])
const invoices = ref<PaymentOrder[]>([])
const ordersLoading = ref(false)
const refundsLoading = ref(false)
const invoicesLoading = ref(false)
const orderQuery = reactive<{ status?: string; user_id?: string }>({})

const invoiceDialog = ref(false)
const invoiceForm = reactive({ user_id: 1, order_id: 1, title: '', tax_id: '', email: '', amount: 0.01 })

const reconcileProvider = ref('wechat')
const reconcileDate = ref(new Date().toISOString().slice(0, 10))
interface ReconcileReport { id: string; status: string; total_orders: number; matched_orders: number }
const reconcileReport = ref<ReconcileReport | null>(null)

function unwrapList(res: { data?: { data?: unknown } }, keys: string[]): unknown[] {
  const data = (res.data?.data ?? res.data) as Record<string, unknown> | unknown[] | undefined
  if (!data) return []
  for (const key of keys) {
    if (typeof data === 'object' && data !== null && !Array.isArray(data) && Array.isArray((data as Record<string, unknown>)[key])) {
      return (data as Record<string, unknown>)[key] as unknown[]
    }
  }
  if (Array.isArray(data)) return data
  return []
}

function orderStatusType(status: string) {
  if (status === 'paid') return 'success'
  if (status === 'pending') return 'warning'
  if (status === 'refunded') return 'info'
  return 'danger'
}

async function loadOrders() {
  ordersLoading.value = true
  try { orders.value = unwrapList(await listOrders(orderQuery), ['orders']) as PaymentOrder[] } catch (e: any) { ElMessage.error(e.message) }
  finally { ordersLoading.value = false }
}

async function loadRefunds() {
  refundsLoading.value = true
  try { refunds.value = unwrapList(await listRefunds({ page_size: 50 }), ['refunds']) as Refund[] } catch (e: any) { ElMessage.error(e.message) }
  finally { refundsLoading.value = false }
}

async function loadInvoices() {
  invoicesLoading.value = true
  try { invoices.value = unwrapList(await listInvoices({ page_size: 50 }), ['invoices']) as PaymentOrder[] } catch (e: any) { ElMessage.error(e.message) }
  finally { invoicesLoading.value = false }
}

async function approve(row: Refund) {
  await ElMessageBox.confirm(`确认批准退款 #${row.id}？`, '批准退款')
  try { await approveRefund(row.id); ElMessage.success('退款已批准'); await loadRefunds(); await loadOrders() } catch (e: any) { ElMessage.error(e.message) }
}

async function reject(row: Refund) {
  const note = window.prompt('请输入驳回原因', '资料不完整') || ''
  if (!note) return
  try { await rejectRefund(row.id, note); ElMessage.success('退款已驳回'); await loadRefunds() } catch (e: any) { ElMessage.error(e.message) }
}

async function submitInvoice() {
  try { await createInvoice(invoiceForm); ElMessage.success('发票已创建'); invoiceDialog.value = false; await loadInvoices() } catch (e: any) { ElMessage.error(e.message) }
}

async function runReconcile() {
  try {
    const res = await reconcile(reconcileProvider.value, reconcileDate.value)
    reconcileReport.value = res.data?.data ?? res.data
    ElMessage.success('对账完成')
  } catch (e: any) { ElMessage.error(e.message) }
}

async function refreshAll() {
  await Promise.allSettled([loadOrders(), loadRefunds(), loadInvoices()])
}

onMounted(refreshAll)
</script>
