'use client';

import React, { useState, useMemo } from 'react';
import Link from 'next/link';
import { PageDetail, User } from '../lib/types';
import { CommentsSection } from './comments-section';

interface DocViewerProps {
  page: PageDetail;
  currentUser: User | null;
  onEdit: () => void;
  onOpenHistory: () => void;
  onDeletePage?: () => void;
}

interface TocItem {
  id: string;
  text: string;
  level: number;
}

export function DocViewer({
  page,
  currentUser,
  onEdit,
  onOpenHistory,
  onDeletePage,
}: DocViewerProps) {
  // Extract Table of Contents items from markdown headings
  const toc = useMemo(() => {
    const lines = page.content.split('\n');
    const items: TocItem[] = [];
    lines.forEach((line) => {
      const match = line.match(/^(#{1,3})\s+(.+)$/);
      if (match) {
        const level = match[1].length;
        const text = match[2].trim();
        const id = text.toLowerCase().replace(/[^a-z0-9]+/g, '-');
        items.push({ id, text, level });
      }
    });
    return items;
  }, [page.content]);

  return (
    <div className="flex-1 flex overflow-hidden">
      {/* Main Reading Column */}
      <div className="flex-1 overflow-y-auto px-8 py-6">
        <div className="max-w-4xl mx-auto">
          {/* Breadcrumbs Navigation */}
          <nav className="flex items-center space-x-1.5 text-xs text-slate-500 mb-4" aria-label="Breadcrumb">
            <Link href={`/spaces/${page.space_slug}`} className="hover:text-slate-800 dark:hover:text-slate-200">
              {page.space_name}
            </Link>
            {page.breadcrumbs.map((crumb) => (
              <React.Fragment key={crumb.id}>
                <span>/</span>
                <Link
                  href={`/spaces/${page.space_slug}/pages/${crumb.id}`}
                  className="hover:text-slate-800 dark:hover:text-slate-200 truncate max-w-[150px]"
                >
                  {crumb.title}
                </Link>
              </React.Fragment>
            ))}
          </nav>

          {/* Page Header */}
          <div className="flex items-start justify-between pb-4 border-b border-slate-200 dark:border-slate-800 mb-6">
            <div>
              <h1 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-100 mb-2">
                {page.title}
              </h1>

              {/* Metadata bar */}
              <div className="flex items-center space-x-3 text-xs text-slate-500">
                <span>
                  Updated by <strong className="text-slate-700 dark:text-slate-300">{page.updater_name}</strong>
                </span>
                <span>•</span>
                <span>{new Date(page.updated_at).toLocaleDateString(undefined, { dateStyle: 'medium' })}</span>
                <span>•</span>
                <button
                  type="button"
                  onClick={onOpenHistory}
                  className="text-blue-600 dark:text-blue-400 hover:underline inline-flex items-center space-x-1"
                >
                  <span>v{page.version_count}</span>
                  <span className="text-[10px] text-slate-400">(history)</span>
                </button>
              </div>
            </div>

            {/* Action Buttons */}
            {currentUser?.role !== 'viewer' && (
              <div className="flex items-center space-x-2">
                <button
                  type="button"
                  onClick={onEdit}
                  className="flex items-center space-x-1 px-3 py-1.5 text-xs font-medium text-slate-700 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md hover:bg-slate-50 dark:hover:bg-slate-700 shadow-sm transition-colors"
                >
                  <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                  <span>Edit</span>
                </button>
                {currentUser?.role === 'admin' && onDeletePage && (
                  <button
                    type="button"
                    onClick={onDeletePage}
                    className="p-1.5 text-slate-400 hover:text-red-600 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                    title="Delete page"
                  >
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Documentation Content Body */}
          <article className="prose prose-slate dark:prose-invert max-w-none text-sm leading-relaxed mb-12">
            <div className="whitespace-pre-wrap font-sans">{page.content}</div>
          </article>

          {/* Comments Section */}
          <CommentsSection pageId={page.id} currentUser={currentUser} />
        </div>
      </div>

      {/* Right Rail: Table of Contents (TOC) */}
      {toc.length > 0 && (
        <aside className="hidden xl:block w-56 flex-shrink-0 border-l border-slate-200 dark:border-slate-800 p-4 text-xs overflow-y-auto">
          <h4 className="font-semibold text-slate-900 dark:text-slate-100 uppercase tracking-wider text-[11px] mb-3">
            On this page
          </h4>
          <ul className="space-y-2 text-slate-500">
            {toc.map((item) => (
              <li
                key={item.id}
                style={{ paddingLeft: `${(item.level - 1) * 0.75}rem` }}
                className="hover:text-blue-600 dark:hover:text-blue-400 transition-colors truncate"
              >
                <a href={`#${item.id}`}>{item.text}</a>
              </li>
            ))}
          </ul>
        </aside>
      )}
    </div>
  );
}
