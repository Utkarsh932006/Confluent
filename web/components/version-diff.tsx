'use client';

import React, { useState, useEffect } from 'react';
import { PageVersion } from '../lib/types';
import { api } from '../lib/api';

interface VersionHistoryProps {
  pageId: number;
  onClose: () => void;
}

export function VersionHistory({ pageId, onClose }: VersionHistoryProps) {
  const [versions, setVersions] = useState<PageVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<PageVersion | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadVersions();
  }, [pageId]);

  const loadVersions = async () => {
    try {
      const data = await api.getVersions(pageId);
      setVersions(data);
      if (data.length > 0) {
        loadVersionDetail(data[0].version_num);
      }
    } catch (err) {
      console.error('Failed to load version history', err);
    } finally {
      setLoading(false);
    }
  };

  const loadVersionDetail = async (versionNum: number) => {
    try {
      const detail = await api.getVersionDetail(pageId, versionNum);
      setSelectedVersion(detail);
    } catch (err) {
      console.error('Failed to load revision detail', err);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm p-6">
      <div className="w-full max-w-4xl h-[80vh] bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900">
          <h2 className="text-base font-semibold text-slate-800 dark:text-slate-100 flex items-center space-x-2">
            <span>Version History & Edits</span>
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-sm"
          >
            ✕ Close
          </button>
        </div>

        {/* Content Body: Timeline left, Revision View right */}
        <div className="flex-1 flex overflow-hidden">
          {/* Revisions List */}
          <div className="w-72 border-r border-slate-200 dark:border-slate-800 overflow-y-auto p-3 space-y-1">
            {loading ? (
              <div className="p-4 text-xs text-slate-400 italic">Loading revisions...</div>
            ) : (
              versions.map((v) => (
                <button
                  key={v.id}
                  type="button"
                  onClick={() => loadVersionDetail(v.version_num)}
                  className={`w-full text-left p-2.5 rounded-lg text-xs transition-colors ${
                    selectedVersion?.version_num === v.version_num
                      ? 'bg-blue-50 text-blue-900 font-medium dark:bg-blue-950/60 dark:text-blue-200'
                      : 'hover:bg-slate-50 dark:hover:bg-slate-800/60 text-slate-700 dark:text-slate-300'
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-semibold">Revision v{v.version_num}</span>
                    <span className="text-[10px] text-slate-400">
                      {new Date(v.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-500 truncate mb-1">
                    {v.change_summary || 'No summary'}
                  </p>
                  <p className="text-[10px] text-slate-400">By {v.editor_name}</p>
                </button>
              ))
            )}
          </div>

          {/* Historical Snapshot Content */}
          <div className="flex-1 overflow-y-auto p-6 bg-slate-50/30 dark:bg-slate-900/30">
            {selectedVersion ? (
              <div>
                <div className="mb-4 pb-3 border-b border-slate-200 dark:border-slate-800">
                  <h3 className="text-xl font-bold text-slate-900 dark:text-slate-100">
                    {selectedVersion.title}
                  </h3>
                  <div className="flex items-center space-x-2 text-xs text-slate-500 mt-1">
                    <span>Revision v{selectedVersion.version_num}</span>
                    <span>•</span>
                    <span>Edited by {selectedVersion.editor_name}</span>
                    <span>•</span>
                    <span>{new Date(selectedVersion.created_at).toLocaleString()}</span>
                  </div>
                  {selectedVersion.change_summary && (
                    <div className="mt-2 text-xs bg-amber-50 dark:bg-amber-950/40 text-amber-800 dark:text-amber-300 p-2 rounded border border-amber-200 dark:border-amber-900">
                      <strong>Change log:</strong> {selectedVersion.change_summary}
                    </div>
                  )}
                </div>

                <div className="prose prose-slate dark:prose-invert max-w-none text-xs leading-relaxed whitespace-pre-wrap font-mono">
                  {selectedVersion.content}
                </div>
              </div>
            ) : (
              <div className="text-xs text-slate-400 italic p-6">Select a revision from the left to view.</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
