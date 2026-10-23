import { Space, PageTreeNode, PageDetail, PageVersion, Comment, SearchResult, User } from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

class ApiClient {
  private getToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('confluent_token');
  }

  setToken(token: string) {
    if (typeof window !== 'undefined') {
      localStorage.setItem('confluent_token', token);
    }
  }

  clearToken() {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('confluent_token');
    }
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ error: 'An unexpected error occurred' }));
      throw new Error(errBody.error || `Request failed with status ${res.status}`);
    }

    return res.json();
  }

  // Auth
  async login(email: string, password: string): Promise<{ token: string; user: User }> {
    const data = await this.request<{ token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    this.setToken(data.token);
    return data;
  }

  async register(email: string, password: string, fullName: string): Promise<{ token: string; user: User }> {
    const data = await this.request<{ token: string; user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, full_name: fullName }),
    });
    this.setToken(data.token);
    return data;
  }

  async getMe(): Promise<User> {
    return this.request<User>('/auth/me');
  }

  // Spaces
  async listSpaces(): Promise<Space[]> {
    return this.request<Space[]>('/spaces');
  }

  async getSpace(slug: string): Promise<Space> {
    return this.request<Space>(`/spaces/${slug}`);
  }

  async createSpace(data: { name: string; slug?: string; description?: string; icon?: string }): Promise<Space> {
    return this.request<Space>('/spaces', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // Pages & Hierarchy Tree
  async getPageTree(spaceSlug: string): Promise<PageTreeNode[]> {
    return this.request<PageTreeNode[]>(`/spaces/${spaceSlug}/pages`);
  }

  async getPage(id: number): Promise<PageDetail> {
    return this.request<PageDetail>(`/pages/${id}`);
  }

  async createPage(spaceSlug: string, data: { title: string; content: string; parent_id?: number | null }): Promise<{ id: number; slug: string }> {
    return this.request<{ id: number; slug: string }>(`/spaces/${spaceSlug}/pages`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updatePage(id: number, data: { title: string; content: string; parent_id?: number | null; change_summary?: string }): Promise<void> {
    return this.request<void>(`/pages/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deletePage(id: number): Promise<void> {
    return this.request<void>(`/pages/${id}`, { method: 'DELETE' });
  }

  // Version History
  async getVersions(pageId: number): Promise<PageVersion[]> {
    return this.request<PageVersion[]>(`/pages/${pageId}/versions`);
  }

  async getVersionDetail(pageId: number, versionNum: number): Promise<PageVersion> {
    return this.request<PageVersion>(`/pages/${pageId}/versions/${versionNum}`);
  }

  // Comments
  async getComments(pageId: number): Promise<Comment[]> {
    return this.request<Comment[]>(`/pages/${pageId}/comments`);
  }

  async addComment(pageId: number, content: string): Promise<Comment> {
    return this.request<Comment>(`/pages/${pageId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    });
  }

  async deleteComment(commentId: number): Promise<void> {
    return this.request<void>(`/comments/${commentId}`, { method: 'DELETE' });
  }

  // Search
  async search(query: string, spaceSlug?: string): Promise<SearchResult[]> {
    const params = new URLSearchParams({ q: query });
    if (spaceSlug) params.append('space', spaceSlug);
    return this.request<SearchResult[]>(`/search?${params.toString()}`);
  }
}

export const api = new ApiClient();
