export interface Repository {
  id: string;
  name: string;
  path: string;
  defaultBranch: string;
  lastAnalyzedAt: string | null;
  totalCommits: number;
  totalFiles: number;
  totalAuthors: number;
  isDemo: boolean;
}

export interface CommitFile {
  id?: number;
  commitHash: string;
  repoId?: string;
  filePath: string;
  changeType: 'A' | 'M' | 'D' | 'R' | string;
  oldPath?: string;
  insertions: number;
  deletions: number;
}

export interface Commit {
  hash: string;
  repoId: string;
  shortHash: string;
  parentHashes: string;
  authorName: string;
  authorEmail: string;
  authorDate: string;
  committerDate: string;
  subject: string;
  body: string;
  branch: string;
  insertions: number;
  deletions: number;
  filesChangedCount: number;
  files?: CommitFile[];
}

export interface FileHotspot {
  repoId: string;
  filePath: string;
  extension: string;
  commitCount: number;
  authorCount: number;
  totalInsertions: number;
  totalDeletions: number;
  linesOfCode: number;
  lastModified: string;
  churnScore: number;
}

export interface DependencyNode {
  id: string;
  name: string;
  language: string;
  inDegree: number;
  outDegree: number;
  x?: number;
  y?: number;
  vx?: number;
  vy?: number;
}

export interface DependencyLink {
  source: string | DependencyNode;
  target: string | DependencyNode;
  statement: string;
  language: string;
}

export interface DependencyGraph {
  nodes: DependencyNode[];
  links: DependencyLink[];
}

export interface Contributor {
  repoId: string;
  name: string;
  email: string;
  totalCommits: number;
  firstCommitDate: string;
  lastCommitDate: string;
  insertions: number;
  deletions: number;
}

export interface ContributorActivity {
  repoId: string;
  email: string;
  dayOfWeek: number; // 0=Sunday, 6=Saturday
  hourOfDay: number; // 0-23
  commitCount: number;
}

export interface TimelinePoint {
  date: string;
  commitCount: number;
  insertions: number;
  deletions: number;
  authors: number;
}

export interface OverviewStats {
  repository: Repository;
  topHotspots: FileHotspot[];
  topContributors: Contributor[];
  recentCommits: Commit[];
  languageBreakdown: Record<string, number>;
  timeline: TimelinePoint[];
  totalDependencies: number;
}

export interface DiffLine {
  type: 'add' | 'del' | 'context' | 'header';
  content: string;
  oldLine?: number;
  newLine?: number;
}

export interface DiffHunk {
  header: string;
  lines: DiffLine[];
}

export interface FileDiff {
  filePath: string;
  oldPath?: string;
  changeType: string;
  hunks: DiffHunk[];
  raw: string;
}

export interface CommitDiffResponse {
  commitHash: string;
  files: FileDiff[];
}

export interface AppSettings {
  appName: string;
  version: string;
  aiEnabled: boolean;
  localOnly: boolean;
  supportedParsers: string[];
}
