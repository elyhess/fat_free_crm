export interface PaginatedResult<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface Account {
  id: number;
  user_id: number;
  assigned_to: number;
  name: string;
  access: string;
  rating: number;
  category?: string;
  email?: string;
  website?: string;
  phone?: string;
  contacts_count: number;
  opportunities_count: number;
  created_at: string;
  updated_at: string;
}

export interface Contact {
  id: number;
  user_id: number;
  assigned_to: number;
  first_name: string;
  last_name: string;
  access: string;
  title?: string;
  department?: string;
  email?: string;
  phone?: string;
  mobile?: string;
  created_at: string;
  updated_at: string;
}

export interface Lead {
  id: number;
  user_id: number;
  assigned_to: number;
  first_name: string;
  last_name: string;
  access: string;
  company?: string;
  title?: string;
  status?: string;
  email?: string;
  phone?: string;
  rating: number;
  created_at: string;
  updated_at: string;
}

export interface Opportunity {
  id: number;
  user_id: number;
  assigned_to: number;
  name: string;
  access: string;
  stage?: string;
  probability?: number;
  amount?: number;
  discount?: number;
  closes_on?: string;
  created_at: string;
  updated_at: string;
}

export interface Campaign {
  id: number;
  user_id: number;
  assigned_to: number;
  name: string;
  access: string;
  status?: string;
  budget?: number;
  leads_count: number;
  opportunities_count: number;
  starts_on?: string;
  ends_on?: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: number;
  user_id: number;
  assigned_to: number;
  name: string;
  priority?: string;
  category?: string;
  bucket?: string;
  due_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface TaskBucket {
  bucket: string;
  count: number;
}

export interface TaskSummary {
  buckets: TaskBucket[];
  total_tasks: number;
}

export interface PipelineStage {
  stage: string;
  count: number;
  total_amount: number;
  weighted_sum: number;
}

export interface PipelineResponse {
  stages: PipelineStage[];
  total_count: number;
  total_amount: number;
  total_weighted: number;
}
