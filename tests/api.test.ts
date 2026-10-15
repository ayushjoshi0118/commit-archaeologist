import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api } from '../src/lib/api/client';

describe('API Client Suite', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('correctly constructs query string for commits filtering', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => [{ hash: '123', subject: 'test commit' }]
    });
    global.fetch = mockFetch;

    const res = await api.getCommits({
      repoId: 'repo-abc',
      search: 'scaffold',
      author: 'elena',
      limit: 25
    });

    expect(mockFetch).toHaveBeenCalledOnce();
    const calledUrl = mockFetch.mock.calls[0][0];
    expect(calledUrl).toContain('/api/commits?');
    expect(calledUrl).toContain('repoId=repo-abc');
    expect(calledUrl).toContain('search=scaffold');
    expect(calledUrl).toContain('author=elena');
    expect(calledUrl).toContain('limit=25');
    expect(res).toHaveLength(1);
  });

  it('handles backend error responses with informative messages', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      statusText: 'Bad Request',
      json: async () => ({ error: 'invalid repository path: directory does not exist or lacks .git' })
    });
    global.fetch = mockFetch;

    await expect(api.analyzeRepository('/invalid/path')).rejects.toThrow(
      'invalid repository path: directory does not exist or lacks .git'
    );
  });

  it('fetches application settings and verifies AI is disabled by default', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        appName: 'Commit Archaeologist',
        version: '1.0.0',
        aiEnabled: false,
        localOnly: true,
        supportedParsers: ['Go', 'TypeScript', 'JavaScript', 'Python']
      })
    });
    global.fetch = mockFetch;

    const settings = await api.getSettings();
    expect(settings.aiEnabled).toBe(false);
    expect(settings.localOnly).toBe(true);
    expect(settings.supportedParsers).toContain('Go');
    expect(settings.supportedParsers).toContain('TypeScript');
  });
});
