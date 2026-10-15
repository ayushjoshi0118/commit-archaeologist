import { describe, it, expect } from 'vitest';

describe('Commit and File Metrics Formatting', () => {
  it('computes churn score based on commit count, authors, and insertions/deletions', () => {
    const commits = 15;
    const authors = 3;
    const insertions = 120;
    const deletions = 30;

    const totalChurn = insertions + deletions;
    const score = (commits * 1.5) + (authors * 2.0) + (Math.log2(1.0 + totalChurn) * 1.2);

    expect(score).toBeGreaterThan(commits);
    expect(Number(score.toFixed(1))).toBeCloseTo(37.2, 1);
  });

  it('filters commits by search term across subject and hash', () => {
    const commits = [
      { hash: 'abc12345', shortHash: 'abc1234', subject: 'feat(ast): add parser', authorName: 'Alice' },
      { hash: 'def67890', shortHash: 'def6789', subject: 'fix: resolve race condition', authorName: 'Bob' },
      { hash: '98765432', shortHash: '9876543', subject: 'docs: update readme', authorName: 'Charlie' }
    ];

    const filterBySearch = (query: string) => {
      const q = query.toLowerCase();
      return commits.filter(c =>
        c.subject.toLowerCase().includes(q) ||
        c.hash.toLowerCase().includes(q) ||
        c.authorName.toLowerCase().includes(q)
      );
    };

    expect(filterBySearch('ast')).toHaveLength(1);
    expect(filterBySearch('race')).toHaveLength(1);
    expect(filterBySearch('Alice')).toHaveLength(1);
    expect(filterBySearch('98765432')).toHaveLength(1);
    expect(filterBySearch('nonexistent')).toHaveLength(0);
  });

  it('properly formats line counts and byte counts for display', () => {
    const formatLOC = (loc: number) => {
      return loc >= 1000 ? `${(loc / 1000).toFixed(1)}k` : loc.toString();
    };

    expect(formatLOC(150)).toBe('150');
    expect(formatLOC(1250)).toBe('1.3k');
    expect(formatLOC(25000)).toBe('25.0k');
  });
});
