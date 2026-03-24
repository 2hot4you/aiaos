export interface User {
  id: string;
  username: string;
  display_name: string;
  role: 'admin' | 'user';
  enabled: boolean;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  name: string;
  description?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Season {
  id: string;
  project_id: string;
  season_number: number;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface Episode {
  id: string;
  season_id: string;
  episode_number: number;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface ModelConfig {
  id: string;
  name: string;
  model_type: 'text' | 'image' | 'video';
  provider: string;
  api_endpoint: string;
  model_identifier: string;
  is_default: boolean;
  enabled: boolean;
  created_at: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
