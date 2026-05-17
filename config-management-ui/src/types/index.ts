export interface ApiResponse<T> {
  code: number
  message: string
  data: T
  total?: number
}

export interface ConfigGroup {
  id: number
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface ConfigItem {
  id: number
  group_id: number
  code: string
  name: string
  description: string
  data_type: string
  current_value: string
  default_value: string
  min_value: string | null
  max_value: string | null
  allowed_values: string | null
  is_sensitive: boolean
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface ConfigVersion {
  id: number
  item_id: number
  value_before: string
  value_after: string
  change_reason: string
  changed_by: string
  created_at: string
}

export interface ConfigRelease {
  id: number
  title: string
  description: string
  status: string
  created_by: string
  approved_by: string | null
  created_at: string
  updated_at: string
  approved_at: string | null
  released_at: string | null
}

export interface ConfigReleaseItem {
  id: number
  release_id: number
  item_id: number
  value_before: string
  value_after: string
  change_reason: string
  created_at: string
}

export interface AuditLog {
  id: number
  operation_type: string
  operation_object: string
  operator: string
  operator_ip: string
  operation_result: string
  operation_details: string
  sm3_hash: string
  created_at: string
}

export interface Role {
  id: number
  name: string
  description: string
  created_at: string
}

export interface RolePermission {
  id: number
  role_id: number
  permission: string
  created_at: string
}

export interface UserRole {
  id: number
  user_id: string
  role_id: number
  created_at: string
}

export interface DashboardStats {
  total_config: number
  enabled_config: number
  pending_releases: number
  today_changes: number
  alert_count: number
}
