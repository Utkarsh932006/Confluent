'use client';

import React from 'react';
import Link from 'next/link';
import { Space, PageTreeNode, User } from '../lib/types';
import { PageTree } from './page-tree';

interface SidebarProps {
  currentSpace: Space | null;
  spaces: Space[];
  pageTree: PageTreeNode[];
  activePageId?: number;
  currentUser: User | null;
  onOpenSearch: () => void;
  onNewPage: () => void;
  onLogout: () => void;
}

export function Sidebar({
  currentSpace,
  spaces,
  pageTree,
  activePageId,
  currentUser,
  onOpenSearch,
  onNewPage,
  onLogout,
}: SidebarProps) {
  return (
    <aside className="w-64 flex-shrink-0 flex flex-col h-screen border-r border-slate-200 bg-slate-50/50 dark:bg-slate-900/50 dark:border-slate-800 select-none">
      {/* Top Header: Logo & Space Switcher */}
      <div className="p-3 border-b border-slate-200 dark:border-slate-800">
        <div className="flex items-center justify-between mb-2">
          <Link href="/" className="flex items-center space-x-2">
            <div className="w-6 h-6 rounded bg-blue-600 text-white font-bold flex items-center justify-center text-xs">
              C
            </div>
            <span className="font-semibold text-sm tracking-tight text-slate-800 dark:text-slate-100">
              Confluent
            </span>
          </Link>
          {currentUser?.role === 'admin' && (
            <Link
              href="/admin"
              className="text-xs text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 px-1.5 py-0.5 rounded border border-slate-200 dark:border-slate-700"
            >
              Admin
            </Link>
          )}
        </div>

        {/* Space Selector Dropdown */}
        <div className="relative">
          <select
            value={currentSpace?.slug || ''}
            onChange={(e) => {
              if (e.target.value) {
                window.location.href = `/spaces/${e.target.value}`;
              }
            }}
            className="w-full text-xs font-medium bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md py-1.5 px-2.5 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            <option value="" disabled>
              Select a space...
            </option>
            {spaces.map((s) => (
              <option key={s.id} value={s.slug}>
                {s.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Global Search & New Page CTA */}
      <div className="p-2 space-y-1.5">
        <button
          type="button"
          onClick={onOpenSearch}
          className="w-full flex items-center justify-between px-2.5 py-1.5 text-xs text-slate-500 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md shadow-sm hover:border-slate-300 dark:hover:border-slate-600 transition-colors"
        >
          <span className="flex items-center space-x-1.5">
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <span>Search docs...</span>
          </span>
          <kbd className="text-[10px] bg-slate-100 dark:bg-slate-700 text-slate-400 px-1.5 py-0.5 rounded font-mono">
            ⌘K
          </kbd>
        </button>

        {currentUser?.role !== 'viewer' && currentSpace && (
          <button
            type="button"
            onClick={onNewPage}
            className="w-full flex items-center justify-center space-x-1.5 px-2.5 py-1.5 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md shadow-sm transition-colors"
          >
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            <span>New Page</span>
          </button>
        )}
      </div>

      {/* Page Navigation Tree */}
      <div className="flex-1 overflow-y-auto px-2 py-1 scrollbar-thin">
        {currentSpace ? (
          <>
            <div className="px-2 py-1 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
              {currentSpace.name} Docs
            </div>
            <PageTree
              nodes={pageTree}
              spaceSlug={currentSpace.slug}
              activePageId={activePageId}
              onAddSubpage={currentUser?.role !== 'viewer' ? () => onNewPage() : undefined}
            />
          </>
        ) : (
          <div className="p-3 text-xs text-slate-400 italic text-center">
            Select a space from the dropdown above to view documents.
          </div>
        )}
      </div>

      {/* Bottom User Profile Section */}
      {currentUser && (
        <div className="p-2.5 border-t border-slate-200 dark:border-slate-800 bg-white/50 dark:bg-slate-900/50 flex items-center justify-between">
          <div className="flex items-center space-x-2 min-w-0">
            <div className="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center text-xs font-semibold text-slate-700 dark:text-slate-200 uppercase">
              {currentUser.full_name.charAt(0)}
            </div>
            <div className="min-w-0">
              <p className="text-xs font-medium text-slate-800 dark:text-slate-200 truncate">
                {currentUser.full_name}
              </p>
              <span className="inline-block text-[10px] px-1.5 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 capitalize">
                {currentUser.role}
              </span>
            </div>
          </div>
          <button
            type="button"
            onClick={onLogout}
            title="Log out"
            className="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
          </button>
        </div>
      )}
    </aside>
  );
}
