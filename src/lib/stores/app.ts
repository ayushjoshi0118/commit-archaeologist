import { writable, derived } from 'svelte/store';
import type { Repository, OverviewStats, Commit } from '../types';
import { api } from '../api/client';

export type ViewType = 'overview' | 'timeline' | 'hotspots' | 'dependencies' | 'commits' | 'contributors' | 'settings';

export const currentView = writable<ViewType>('overview');
export const repositories = writable<Repository[]>([]);
export const selectedRepo = writable<Repository | null>(null);
export const overviewStats = writable<OverviewStats | null>(null);

export const isLoading = writable<boolean>(false);
export const loadingMessage = writable<string>('');
export const errorMessage = writable<string | null>(null);

export const searchQuery = writable<string>('');
export const authorFilter = writable<string>('');
export const fileFilter = writable<string>('');
export const selectedCommit = writable<Commit | null>(null);
export const selectedHotspotFile = writable<string | null>(null);

export async function initApp() {
  try {
    isLoading.set(true);
    loadingMessage.set('Connecting to Commit Archaeologist engine...');
    const repos = await api.getRepositories();
    repositories.set(repos);

    if (repos.length > 0) {
      await loadRepoStats(repos[0].id);
    }
  } catch (err: any) {
    errorMessage.set(err.message || 'Failed to initialize application');
  } finally {
    isLoading.set(false);
  }
}

export async function loadRepoStats(repoId: string) {
  try {
    isLoading.set(true);
    loadingMessage.set('Loading repository intelligence...');
    errorMessage.set(null);

    const stats = await api.getOverview(repoId);
    overviewStats.set(stats);
    selectedRepo.set(stats.repository);
  } catch (err: any) {
    errorMessage.set(err.message || 'Failed to load repository details');
  } finally {
    isLoading.set(false);
  }
}

export async function analyzeNewRepository(path: string) {
  try {
    isLoading.set(true);
    loadingMessage.set(`Analyzing Git history & AST graph for ${path}...`);
    errorMessage.set(null);

    const stats = await api.analyzeRepository(path);
    overviewStats.set(stats);
    selectedRepo.set(stats.repository);

    // Refresh repos list
    const repos = await api.getRepositories();
    repositories.set(repos);
    currentView.set('overview');
  } catch (err: any) {
    errorMessage.set(err.message || 'Failed to analyze repository');
    throw err;
  } finally {
    isLoading.set(false);
  }
}

export async function launchDemoMode() {
  try {
    isLoading.set(true);
    loadingMessage.set('Generating synthetic ChronoLog repository & parsing ASTs...');
    errorMessage.set(null);

    const stats = await api.launchDemo();
    overviewStats.set(stats);
    selectedRepo.set(stats.repository);

    const repos = await api.getRepositories();
    repositories.set(repos);
    currentView.set('overview');
  } catch (err: any) {
    errorMessage.set(err.message || 'Failed to launch demo mode');
  } finally {
    isLoading.set(false);
  }
}
