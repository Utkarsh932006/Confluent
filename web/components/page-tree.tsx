'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { PageTreeNode } from '../lib/types';

interface PageTreeProps {
  nodes: PageTreeNode[];
  spaceSlug: string;
  activePageId?: number;
  onAddSubpage?: (parentId: number) => void;
}

export function PageTree({ nodes, spaceSlug, activePageId, onAddSubpage }: PageTreeProps) {
  if (!nodes || nodes.length === 0) {
    return (
      <div className="px-3 py-4 text-xs text-slate-400 italic">
        No pages yet in this space. Click "New Page" to create one.
      </div>
    );
  }

  return (
    <nav className="space-y-0.5 text-sm" aria-label="Documentation Page Tree">
      {nodes.map((node) => (
        <TreeNodeItem
          key={node.id}
          node={node}
          spaceSlug={spaceSlug}
          activePageId={activePageId}
          onAddSubpage={onAddSubpage}
          level={0}
        />
      ))}
    </nav>
  );
}

interface TreeNodeItemProps {
  node: PageTreeNode;
  spaceSlug: string;
  activePageId?: number;
  onAddSubpage?: (parentId: number) => void;
  level: number;
}

function TreeNodeItem({ node, spaceSlug, activePageId, onAddSubpage, level }: TreeNodeItemProps) {
  const [isOpen, setIsOpen] = useState(true);
  const hasChildren = node.children && node.children.length > 0;
  const isActive = activePageId === node.id;

  return (
    <div>
      <div
        className={`group flex items-center justify-between py-1.5 px-2 rounded-md transition-colors ${
          isActive
            ? 'bg-blue-50 text-blue-700 font-medium dark:bg-blue-950/50 dark:text-blue-300'
            : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800/60 dark:hover:text-slate-200'
        }`}
        style={{ paddingLeft: `${Math.max(0.5, level * 0.75 + 0.5)}rem` }}
      >
        <div className="flex items-center min-w-0 flex-1 space-x-1.5">
          {hasChildren ? (
            <button
              type="button"
              onClick={() => setIsOpen(!isOpen)}
              className="p-0.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-transform"
              aria-label={isOpen ? 'Collapse folder' : 'Expand folder'}
            >
              <svg
                className={`w-3.5 h-3.5 transform transition-transform ${isOpen ? 'rotate-90' : ''}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>
          ) : (
            <span className="w-3.5 h-3.5 inline-block text-slate-300 dark:text-slate-600 text-center">·</span>
          )}

          <Link
            href={`/spaces/${spaceSlug}/pages/${node.id}`}
            className="truncate flex-1 text-xs sm:text-sm"
          >
            {node.title}
          </Link>
        </div>

        {onAddSubpage && (
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault();
              onAddSubpage(node.id);
            }}
            className="opacity-0 group-hover:opacity-100 p-1 text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition-opacity"
            title="Add subpage"
          >
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          </button>
        )}
      </div>

      {hasChildren && isOpen && (
        <div className="relative border-l border-slate-200 dark:border-slate-800 ml-3.5 my-0.5">
          {node.children.map((child) => (
            <TreeNodeItem
              key={child.id}
              node={child}
              spaceSlug={spaceSlug}
              activePageId={activePageId}
              onAddSubpage={onAddSubpage}
              level={level + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
}
