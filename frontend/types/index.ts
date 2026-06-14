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
export type MonitorType = "http" | "tcp";
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
