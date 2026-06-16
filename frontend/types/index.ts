// Users
export interface User {
  id: string;
  email: string;
  role: "admin" | "viewer";
  created_at: string;
  archived_at?: string | null;
}

// Services
export type ServiceStatus = "operational" | "degraded" | "outage" | "maintenance";

export interface Service {
  id: string;
  name: string;
  description?: string | null;
  status: ServiceStatus;
  created_at: string;
  archived_at?: string | null;
}

// Monitors
export type MonitorType = "http" | "tcp" | "ssl" | "keyword" | "dns";
export type MonitorStatus = "up" | "down" | "degraded";

export interface Monitor {
  id: string;
  service_id: string;
  name: string;
  type: MonitorType;
  url?: string | null;
  host?: string | null;
  port?: number | null;
  interval_seconds: number;
  timeout_ms: number;
  retry_count: number;
  consecutive_failures: number;
  degraded_threshold_ms?: number | null;
  http_method?: string | null;
  http_expected_status?: number | null;
  enabled: boolean;
  next_check_at?: string | null;
  created_at: string;
  updated_at: string;
  archived_at?: string | null;
}

// Check Results
export interface CheckResult {
  id: string;
  monitor_id: string;
  status: MonitorStatus;
  response_time_ms?: number | null;
  error?: string | null;
  checked_at: string;
}

// Incidents
export type IncidentSeverity = "low" | "medium" | "high";
export type IncidentStatus = "open" | "resolved";

export interface Incident {
  id: string;
  service_id: string;
  monitor_id?: string | null;
  title: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  resolved_at?: string | null;
  created_at: string;
  updated_at: string;
}

// API Keys
export interface APIKey {
  id: string;
  user_id: string;
  name: string;
  prefix: string;
  created_at: string;
  last_used?: string | null;
  archived_at?: string | null;
}

export interface CreatedAPIKey extends APIKey {
  key: string; // full key shown once
}

// Notification Channels
export type NotificationType = "discord" | "email";

export interface NotificationChannel {
  id: string;
  name: string;
  type: NotificationType;
  enabled: boolean;
  created_at: string;
  archived_at?: string | null;
}

// Webhooks
export interface Webhook {
  id: string;
  name: string;
  events: string[];
  enabled: boolean;
  created_at: string;
  archived_at?: string | null;
}

export interface WebhookLog {
  id: string;
  webhook_id: string;
  event: string;
  status_code?: number | null;
  error?: string | null;
  duration_ms?: number | null;
  delivered_at: string;
}

// Settings
export interface Settings {
  id: number;
  site_title: string;
  logo_url: string;
  public_status_enabled: boolean;
  updated_at: string;
}

export interface PublicSettings {
  site_title: string;
  logo_url: string;
}

// Maintenance Windows
export interface MaintenanceWindow {
  id: string;
  service_id: string;
  title: string;
  description?: string | null;
  starts_at: string;
  ends_at: string;
  created_at: string;
  archived_at?: string | null;
}

// Incident Updates (timeline)
export type IncidentUpdateStatus = "investigating" | "identified" | "monitoring" | "resolved";

export interface IncidentUpdate {
  id: string;
  incident_id: string;
  status: IncidentUpdateStatus;
  message: string;
  created_at: string;
}

// Uptime
export interface DailyUptimeStatus {
  date: string;
  status: "up" | "degraded" | "down" | "no_data";
  uptime_percent: number;
}

export interface MonitorUptime {
  monitor_id: string;
  monitor_name: string;
  days: DailyUptimeStatus[];
}

export interface ServiceUptime {
  service_id: string;
  monitors: MonitorUptime[];
}

// Latency sparkline
export interface ResponseTimePoint {
  checked_at: string;
  response_time_ms: number;
  status: string;
}

// Audit log
export interface AuditLogEntry {
  id: string;
  user_id?: string | null;
  action: string;
  resource: string;
  resource_id?: string | null;
  ip?: string | null;
  user_agent?: string | null;
  diff?: Record<string, unknown> | null;
  created_at: string;
}

export interface AuditLogResponse {
  entries: AuditLogEntry[];
  total: number;
  limit: number;
  offset: number;
}

// Extended types with new fields
export interface ServiceExtended extends Service {
  public_visible: boolean;
  show_uptime: boolean;
  dedicated_page_enabled: boolean;
  slug?: string | null;
  custom_domain?: string | null;
  uptime_range_days: number;
  status_override?: string | null;
}

export interface MonitorExtended extends Monitor {
  ssl_expiry_threshold_days?: number;
  keyword_match?: string | null;
  keyword_should_exist?: boolean;
  dns_record_type?: string | null;
  dns_expected_value?: string | null;
}

// Public API types
export interface PublicService extends Service {
  monitors: PublicMonitor[];
}

export interface PublicMonitor {
  id: string;
  name: string;
  type: MonitorType;
  status: MonitorStatus;
  response_time_ms?: number | null;
}

// API responses
export interface APIError {
  error: string;
  field?: string;
}
