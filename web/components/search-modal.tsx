'use client';

import React, { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { SearchResult } from '../lib/types';
import { api } from '../lib/api';

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
  spaceSlug?: string;
}

export function SearchModal({ isOpen, onClose, spaceSlug }: SearchModalProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 50);
    } else {
      setQuery('');
      setResults([]);
    }
  }, [isOpen]);

  // Debounced search
  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const hits = await api.search(query.trim(), spaceSlug);
        setResults(hits);
        setSelectedIndex(0);
      } catch (err) {
        console.error('Search error', err);
      } finally {
        setLoading(false);
      }
    }, 150);

    return () => clearTimeout(timer);
  }, [query, spaceSlug]);

  // Key navigation
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev < results.length - 1 ? prev + 1 : 0));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : results.length - 1));
    } else if (e.key === 'Enter' && results[selectedIndex]) {
      const target = results[selectedIndex];
      window.location.href = `/spaces/${target.space_slug}/pages/${target.page_id}`;
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-20 bg-slate-900/40 backdrop-blur-sm p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl overflow-hidden"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Search Input Bar */}
        <div className="flex items-center px-4 py-3 border-b border-slate-200 dark:border-slate-800">
          <svg className="w-5 h-5 text-slate-400 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search documentation titles and content..."
            className="flex-1 bg-transparent text-sm text-slate-800 dark:text-slate-100 placeholder-slate-400 outline-none"
          />
          {loading && (
            <span className="text-xs text-slate-400 animate-pulse">Searching...</span>
          )}
          <kbd className="hidden sm:inline-block text-[10px] bg-slate-100 dark:bg-slate-800 text-slate-500 px-1.5 py-0.5 rounded ml-2">
            ESC
          </kbd>
        </div>

        {/* Results List */}
        <div className="max-h-96 overflow-y-auto p-2">
          {query.trim() === '' ? (
            <div className="p-8 text-center text-xs text-slate-400">
              Type keywords to search across all spaces and pages.
            </div>
          ) : results.length === 0 && !loading ? (
            <div className="p-8 text-center text-xs text-slate-400">
              No matching pages found for "{query}".
            </div>
          ) : (
            <div className="space-y-1">
              {results.map((item, idx) => (
                <Link
                  key={item.page_id}
                  href={`/spaces/${item.space_slug}/pages/${item.page_id}`}
                  onClick={onClose}
                  className={`block p-2.5 rounded-lg text-left transition-colors ${
                    idx === selectedIndex
                      ? 'bg-blue-50 text-blue-900 dark:bg-blue-950/60 dark:text-blue-200'
                      : 'hover:bg-slate-50 dark:hover:bg-slate-800/60 text-slate-700 dark:text-slate-300'
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm font-semibold truncate">{item.title}</span>
                    <span className="text-[10px] text-slate-400 font-medium px-1.5 py-0.5 bg-slate-100 dark:bg-slate-800 rounded">
                      {item.space_name}
                    </span>
                  </div>
                  <p
                    className="text-xs text-slate-500 dark:text-slate-400 line-clamp-2 leading-relaxed"
                    dangerouslySetInnerHTML={{ __html: item.snippet }}
                  />
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
