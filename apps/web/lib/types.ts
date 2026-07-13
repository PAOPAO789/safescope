export type Role = "admin" | "analyst" | "viewer";

export interface User {
  id: string;
  email: string;
  name: string;
  role: Role;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  status: "active" | "archived";
  owner_id: string;
  asset_count: number;
  created_at: string;
  updated_at: string;
}

export interface Asset {
  id: string;
  project_id: string;
  type: "domain" | "ip" | "url" | "service";
  value: string;
  status: "unknown" | "alive" | "down";
  tags: string[];
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Dashboard {
  projects: number;
  assets: number;
  alive_assets: number;
  archived_projects: number;
  recent_assets: Asset[];
  assets_by_type: Array<{ label: string; value: number }>;
}
