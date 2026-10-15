import type {
  Repository,
  Commit,
  OverviewStats,
  FileHotspot,
  DependencyGraph,
  Contributor,
  ContributorActivity,
  TimelinePoint,
  CommitDiffResponse,
  AppSettings
} from '../types';

const API_BASE = '/api';

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${endpoint}`;
  try {
    const res = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(options?.headers || {})
      }
    });

    if (!res.ok) {
      let errorMsg = `HTTP ${res.status}: ${res.statusText}`;
      try {
        const errJson = await res.json();
        if (errJson && errJson.error) {
          errorMsg = errJson.error;
        }
      } catch {
        // Ignore json parse error on error response
      }
      throw new Error(errorMsg);
    }

    return await res.json() as T;
  } catch (err: any) {
    console.error(`API Error on ${endpoint}:`, err);
    throw err;
  }
}

export const api = {
  getHealth: () => request<{ status: string }>('/health'),

  getSettings: () => request<AppSettings>('/settings'),

  getRepositories: () => request<Repository[]>('/repositories'),

  analyzeRepository: (path: string) =>
    request<OverviewStats>('/analyze', {
      method: 'POST',
      body: JSON.stringify({ path })
    }),

  launchDemo: () =>
    request<OverviewStats>('/demo', {
      method: 'POST',
      body: JSON.stringify({})
    }),

  getOverview: (repoId: string) =>
    request<OverviewStats>(`/overview?repoId=${encodeURIComponent(repoId)}`),

  getCommits: (params: {
    repoId: string;
    author?: string;
    search?: string;
    file?: string;
    startDate?: string;
    endDate?: string;
    branch?: string;
    limit?: number;
    offset?: number;
  }) => {
    const q = new URLSearchParams();
    q.set('repoId', params.repoId);
    if (params.author) q.set('author', params.author);
    if (params.search) q.set('search', params.search);
    if (params.file) q.set('file', params.file);
    if (params.startDate) q.set('startDate', params.startDate);
    if (params.endDate) q.set('endDate', params.endDate);
    if (params.branch) q.set('branch', params.branch);
    if (params.limit) q.set('limit', params.limit.toString());
    if (params.offset) q.set('offset', params.offset.toString());
    return request<Commit[]>(`/commits?${q.toString()}`);
  },

  getCommitDetail: (hash: string) =>
    request<Commit>(`/commit-detail?hash=${encodeURIComponent(hash)}`),

  getDiff: (repoId: string, hash: string, file?: string) => {
    const q = new URLSearchParams({ repoId, hash });
    if (file) q.set('file', file);
    return request<CommitDiffResponse>(`/diff?${q.toString()}`);
  },

  getFileContent: (repoId: string, hash: string, file: string) =>
    request<{ filePath: string; content: string }>(
      `/file-content?repoId=${encodeURIComponent(repoId)}&hash=${encodeURIComponent(hash)}&file=${encodeURIComponent(file)}`
    ),

  getHotspots: (repoId: string, limit = 100) =>
    request<FileHotspot[]>(`/hotspots?repoId=${encodeURIComponent(repoId)}&limit=${limit}`),

  getDependencies: (repoId: string) =>
    request<DependencyGraph>(`/dependencies?repoId=${encodeURIComponent(repoId)}`),

  getContributors: (repoId: string) =>
    request<Contributor[]>(`/contributors?repoId=${encodeURIComponent(repoId)}`),

  getActivity: (repoId: string, email?: string) => {
    const q = new URLSearchParams({ repoId });
    if (email) q.set('email', email);
    return request<ContributorActivity[]>(`/activity?${q.toString()}`);
  },

  getTimeline: (repoId: string) =>
    request<TimelinePoint[]>(`/timeline?repoId=${encodeURIComponent(repoId)}`)
};
