CREATE TABLE IF NOT EXISTS repositories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    default_branch TEXT NOT NULL DEFAULT 'main',
    last_analyzed_at DATETIME,
    total_commits INTEGER NOT NULL DEFAULT 0,
    total_files INTEGER NOT NULL DEFAULT 0,
    total_authors INTEGER NOT NULL DEFAULT 0,
    is_demo BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS commits (
    hash TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL,
    short_hash TEXT NOT NULL,
    parent_hashes TEXT NOT NULL DEFAULT '',
    author_name TEXT NOT NULL,
    author_email TEXT NOT NULL,
    author_date DATETIME NOT NULL,
    committer_date DATETIME NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT 'main',
    insertions INTEGER NOT NULL DEFAULT 0,
    deletions INTEGER NOT NULL DEFAULT 0,
    files_changed_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(repo_id) REFERENCES repositories(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_commits_repo_date ON commits(repo_id, author_date DESC);
CREATE INDEX IF NOT EXISTS idx_commits_author ON commits(repo_id, author_email);

CREATE TABLE IF NOT EXISTS commit_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    commit_hash TEXT NOT NULL,
    repo_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    change_type TEXT NOT NULL,
    old_path TEXT NOT NULL DEFAULT '',
    insertions INTEGER NOT NULL DEFAULT 0,
    deletions INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(commit_hash) REFERENCES commits(hash) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_commit_files_commit ON commit_files(commit_hash);
CREATE INDEX IF NOT EXISTS idx_commit_files_path ON commit_files(repo_id, file_path);

CREATE TABLE IF NOT EXISTS file_hotspots (
    repo_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    extension TEXT NOT NULL,
    commit_count INTEGER NOT NULL DEFAULT 0,
    author_count INTEGER NOT NULL DEFAULT 0,
    total_insertions INTEGER NOT NULL DEFAULT 0,
    total_deletions INTEGER NOT NULL DEFAULT 0,
    lines_of_code INTEGER NOT NULL DEFAULT 0,
    last_modified DATETIME NOT NULL,
    churn_score REAL NOT NULL DEFAULT 0.0,
    PRIMARY KEY(repo_id, file_path),
    FOREIGN KEY(repo_id) REFERENCES repositories(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_hotspots_churn ON file_hotspots(repo_id, churn_score DESC);

CREATE TABLE IF NOT EXISTS dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo_id TEXT NOT NULL,
    source_file TEXT NOT NULL,
    target_file TEXT NOT NULL,
    import_statement TEXT NOT NULL,
    language TEXT NOT NULL,
    line_number INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(repo_id) REFERENCES repositories(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_deps_repo ON dependencies(repo_id);
CREATE INDEX IF NOT EXISTS idx_deps_source ON dependencies(repo_id, source_file);
CREATE INDEX IF NOT EXISTS idx_deps_target ON dependencies(repo_id, target_file);

CREATE TABLE IF NOT EXISTS contributors (
    repo_id TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    total_commits INTEGER NOT NULL DEFAULT 0,
    first_commit_date DATETIME NOT NULL,
    last_commit_date DATETIME NOT NULL,
    insertions INTEGER NOT NULL DEFAULT 0,
    deletions INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(repo_id, email),
    FOREIGN KEY(repo_id) REFERENCES repositories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS contributor_activity (
    repo_id TEXT NOT NULL,
    email TEXT NOT NULL,
    day_of_week INTEGER NOT NULL,
    hour_of_day INTEGER NOT NULL,
    commit_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(repo_id, email, day_of_week, hour_of_day),
    FOREIGN KEY(repo_id) REFERENCES repositories(id) ON DELETE CASCADE
);
