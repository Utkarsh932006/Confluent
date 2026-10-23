'use client';

import React, { useState, useEffect } from 'react';
import { Comment, User } from '../lib/types';
import { api } from '../lib/api';

interface CommentsSectionProps {
  pageId: number;
  currentUser: User | null;
}

export function CommentsSection({ pageId, currentUser }: CommentsSectionProps) {
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    loadComments();
  }, [pageId]);

  const loadComments = async () => {
    try {
      const data = await api.getComments(pageId);
      setComments(data);
    } catch (err) {
      console.error('Failed to load comments', err);
    } finally {
      setLoading(false);
    }
  };

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newComment.trim() || submitting) return;

    setSubmitting(true);
    try {
      const added = await api.addComment(pageId, newComment.trim());
      setComments((prev) => [...prev, added]);
      setNewComment('');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to post comment');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteComment = async (id: number) => {
    if (!confirm('Are you sure you want to delete this comment?')) return;
    try {
      await api.deleteComment(id);
      setComments((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete comment');
    }
  };

  return (
    <section className="pt-8 border-t border-slate-200 dark:border-slate-800">
      <h3 className="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4 flex items-center space-x-2">
        <span>Comments & Discussion</span>
        <span className="text-xs font-normal text-slate-500 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-full">
          {comments.length}
        </span>
      </h3>

      {/* New Comment Box */}
      {currentUser && (
        <form onSubmit={handleAddComment} className="mb-6">
          <div className="border border-slate-200 dark:border-slate-700 rounded-lg p-3 bg-white dark:bg-slate-900 shadow-sm focus-within:ring-1 focus-within:ring-blue-500">
            <textarea
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              placeholder="Leave a comment or question about this documentation..."
              rows={3}
              className="w-full text-xs text-slate-800 dark:text-slate-200 placeholder-slate-400 bg-transparent border-none outline-none resize-none focus:ring-0"
            />
            <div className="flex justify-between items-center pt-2 border-t border-slate-100 dark:border-slate-800">
              <span className="text-[11px] text-slate-400">Supports basic feedback discussion</span>
              <button
                type="submit"
                disabled={submitting || !newComment.trim()}
                className="px-3 py-1.5 text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 rounded-md transition-colors"
              >
                {submitting ? 'Posting...' : 'Comment'}
              </button>
            </div>
          </div>
        </form>
      )}

      {/* Comments List */}
      {loading ? (
        <div className="text-xs text-slate-400 italic">Loading comments...</div>
      ) : comments.length === 0 ? (
        <div className="text-xs text-slate-400 italic py-3">No comments yet. Start the conversation!</div>
      ) : (
        <div className="space-y-4">
          {comments.map((comment) => (
            <div
              key={comment.id}
              className="group flex items-start space-x-3 p-3 rounded-lg bg-slate-50/60 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-800"
            >
              <div className="w-7 h-7 rounded-full bg-blue-100 dark:bg-blue-900/60 text-blue-700 dark:text-blue-300 font-bold flex items-center justify-center text-xs flex-shrink-0">
                {comment.author_name ? comment.author_name.charAt(0).toUpperCase() : 'U'}
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between mb-1">
                  <div className="flex items-center space-x-2">
                    <span className="text-xs font-semibold text-slate-900 dark:text-slate-200">
                      {comment.author_name}
                    </span>
                    <span className="text-[11px] text-slate-400">
                      {new Date(comment.created_at).toLocaleDateString(undefined, {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                  </div>

                  {(currentUser?.role === 'admin' || currentUser?.id === comment.user_id) && (
                    <button
                      type="button"
                      onClick={() => handleDeleteComment(comment.id)}
                      className="opacity-0 group-hover:opacity-100 text-slate-400 hover:text-red-500 text-xs transition-opacity"
                      title="Delete comment"
                    >
                      Delete
                    </button>
                  )}
                </div>

                <p className="text-xs text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">
                  {comment.content}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
