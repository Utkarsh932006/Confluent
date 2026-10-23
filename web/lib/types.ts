export type Role = 'admin' | 'editor' | 'viewer';

export interface User {
  id: number;
  email: string;
  full_name: string;
  role: Role;
  created_at: string;
}

export interface Space {
  id: number;
  slug: string;
  name: string;
  description: string;
  icon: string;
  page_count?: number;
  created_at: string;
}

export interface PageTreeNode {
  id: number;
  space_id: number;
  parent_id: number | null;
  slug: string;
  title: string;
  order_index: number;
  children: PageTreeNode[];
}

export interface BreadcrumbItem {
  id: number;
  slug: string;
  title: string;
}

export interface PageDetail {
  id: number;
  space_id: number;
  parent_id: number | null;
  slug: string;
  title: string;
  content: string;
  order_index: number;
  created_by: number;
  updated_by: number;
  created_at: string;
  updated_at: string;
  space_slug: string;
  space_name: string;
  author_name: string;
  updater_name: string;
  breadcrumbs: BreadcrumbItem[];
  version_count: number;
}

export interface PageVersion {
  id: number;
  page_id: number;
  version_num: number;
  title: string;
  content?: string;
  change_summary: string;
  edited_by: number;
  editor_name: string;
  created_at: string;
}

export interface Comment {
  id: number;
  page_id: number;
  user_id: number;
  author_name: string;
  content: string;
  created_at: string;
}

export interface SearchResult {
  page_id: number;
  space_id: number;
  space_name: string;
  space_slug: string;
  title: string;
  snippet: string;
}
