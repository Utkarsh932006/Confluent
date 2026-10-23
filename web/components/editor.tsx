'use client';

import React, { useState, useEffect } from 'react';
import { PageTreeNode } from '../lib/types';

interface EditorProps {
  initialTitle?: string;
  initialContent?: string;
  initialParentId?: number | null;
  pageTree: PageTreeNode[];
  currentPageId?: number;
  onSave: (data: { title: string; content: string; parent_id: number | null; change_summary: string }) => Promise<void>;
  onCancel: () => void;
  isSaving?: boolean;
}

export function PageEditor({
  initialTitle = '',
  initialContent = '',
  initialParentId = null,
  pageTree,
  currentPageId,
  onSave,
  onCancel,
  isSaving = false,
}: EditorProps) {
  const [title, setTitle] = useState(initialTitle);
  const [content, setContent] = useState(initialContent);
  const [parentId, setParentId] = useState<number | null>(initialParentId);
  const [changeSummary, setChangeSummary] = useState('');
  const [mode, setMode] = useState<'write' | 'preview'>('write');

  // Flatten pageTree for parent dropdown selection (excluding current page and its subtree)
  const flattenPages = (nodes: PageTreeNode[], depth = 0): { id: number; title: string; depth: number }[] => {
    let list: { id: number; title: string; depth: number }[] = [];
    for (const node of nodes) {
      if (node.id === currentPageId) continue; // prevent circular parent
      list.push({ id: node.id, title: node.title, depth });
      if (node.children) {
        list = list.concat(flattenPages(node.children, depth + 1));
      }
    }
    return list;
  };

  const parentOptions = flattenPages(pageTree);

  // Keyboard shortcut Ctrl+S / Cmd+S
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        handleSubmit();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  });

  const insertSyntax = (before: string, after: string = '') => {
    const textarea = document.getElementById('doc-editor-textarea') as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selected = content.substring(start, end);
    const replacement = `${before}${selected || 'text'}${after}`;

    const newContent = content.substring(0, start) + replacement + content.substring(end);
    setContent(newContent);

    setTimeout(() => {
      textarea.focus();
      textarea.setSelectionRange(start + before.length, start + before.length + (selected ? selected.length : 4));
    }, 0);
  };

  const handleSubmit = async () => {
    if (!title.trim()) {
      alert('Please enter a page title');
      return;
    }
    await onSave({
      title: title.trim(),
      content,
      parent_id: parentId,
      change_summary: changeSummary.trim(),
    });
  };

  return (
    <div className="flex flex-col h-full bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-800">
      {/* Top Toolbar: Parent Selector & Save / Cancel Actions */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <div className="flex items-center space-x-3">
          <label className="text-xs font-medium text-slate-500 dark:text-slate-400">
            Parent:
          </label>
          <select
            value={parentId === null ? '' : parentId}
            onChange={(e) => setParentId(e.target.value === '' ? null : Number(e.target.value))}
            className="text-xs bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded px-2 py-1 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            <option value="">(Top-level Page)</option>
            {parentOptions.map((opt) => (
              <option key={opt.id} value={opt.id}>
                {'\u00A0'.repeat(opt.depth * 3)}↳ {opt.title}
              </option>
            ))}
          </select>
        </div>

        <div className="flex items-center space-x-2">
          <button
            type="button"
            onClick={onCancel}
            disabled={isSaving}
            className="px-3 py-1.5 text-xs font-medium text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100 transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSubmit}
            disabled={isSaving}
            className="flex items-center space-x-1 px-3.5 py-1.5 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 rounded-md shadow-sm transition-colors"
          >
            {isSaving ? <span>Saving...</span> : <span>Save & Publish</span>}
          </button>
        </div>
      </div>

      {/* Editor Body */}
      <div className="flex-1 flex flex-col max-w-4xl w-full mx-auto p-6 overflow-y-auto">
        {/* Title Input */}
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Untitled Page"
          className="text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-100 placeholder-slate-300 dark:placeholder-slate-700 border-none outline-none focus:ring-0 mb-4 bg-transparent"
        />

        {/* Change Summary (for version tracking) */}
        {currentPageId && (
          <div className="mb-4">
            <input
              type="text"
              value={changeSummary}
              onChange={(e) => setChangeSummary(e.target.value)}
              placeholder="Summary of changes (optional, e.g., 'Updated API endpoints and rate limit section')"
              className="w-full text-xs bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 rounded px-2.5 py-1.5 text-slate-700 dark:text-slate-300 placeholder-slate-400 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
        )}

        {/* Formatting Toolbar */}
        <div className="flex items-center justify-between border-y border-slate-200 dark:border-slate-800 py-1.5 mb-4 text-xs text-slate-600 dark:text-slate-400">
          <div className="flex items-center space-x-1">
            <button
              type="button"
              onClick={() => insertSyntax('# ', '')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded font-semibold"
              title="Heading 1"
            >
              H1
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('## ', '')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded font-semibold"
              title="Heading 2"
            >
              H2
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('### ', '')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded font-semibold"
              title="Heading 3"
            >
              H3
            </button>
            <span className="text-slate-300 dark:text-slate-700">|</span>
            <button
              type="button"
              onClick={() => insertSyntax('**', '**')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded font-bold"
              title="Bold"
            >
              B
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('*', '*')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded italic"
              title="Italic"
            >
              I
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('```\n', '\n```')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded font-mono"
              title="Code Block"
            >
              &lt;/&gt;
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('- ', '')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded"
              title="Bulleted List"
            >
              • List
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('1. ', '')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded"
              title="Numbered List"
            >
              1. List
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('[', '](https://)')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded"
              title="Link"
            >
              🔗 Link
            </button>
            <button
              type="button"
              onClick={() => insertSyntax('> ', '')}
              className="px-2 py-1 hover:bg-slate-100 dark:hover:bg-slate-800 rounded"
              title="Quote"
            >
              Quote
            </button>
          </div>

          <div className="flex items-center space-x-1 border border-slate-200 dark:border-slate-700 rounded p-0.5">
            <button
              type="button"
              onClick={() => setMode('write')}
              className={`px-2 py-0.5 rounded text-[11px] font-medium transition-colors ${
                mode === 'write'
                  ? 'bg-blue-600 text-white'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
              }`}
            >
              Write
            </button>
            <button
              type="button"
              onClick={() => setMode('preview')}
              className={`px-2 py-0.5 rounded text-[11px] font-medium transition-colors ${
                mode === 'preview'
                  ? 'bg-blue-600 text-white'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900'
              }`}
            >
              Preview
            </button>
          </div>
        </div>

        {/* Markdown Input Area or Preview */}
        {mode === 'write' ? (
          <textarea
            id="doc-editor-textarea"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Write documentation using Markdown (headings, code blocks, lists, quotes)..."
            className="flex-1 w-full min-h-[450px] p-2 bg-transparent text-slate-800 dark:text-slate-200 font-mono text-sm leading-relaxed border-none outline-none resize-none focus:ring-0"
          />
        ) : (
          <div className="flex-1 min-h-[450px] p-2 prose prose-slate dark:prose-invert max-w-none">
            {content ? (
              <pre className="whitespace-pre-wrap font-sans text-sm leading-relaxed">{content}</pre>
            ) : (
              <p className="italic text-slate-400">Nothing to preview.</p>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
