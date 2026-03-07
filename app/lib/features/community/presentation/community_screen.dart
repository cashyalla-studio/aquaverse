import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../l10n/app_localizations.dart';
import '../data/community_repository.dart';
import '../domain/community_model.dart';

// 게시판 목록 프로바이더 (로케일별)
final boardListProvider = FutureProvider.family<List<Board>, String>((ref, locale) {
  return ref.watch(communityRepositoryProvider(locale)).listBoards();
});

// 게시글 목록 프로바이더
final postListProvider = FutureProvider.family<PostListResult, Map<String, dynamic>>((ref, params) {
  final boardId = params['board_id'] as int;
  final locale = params['locale'] as String;
  final page = params['page'] as int? ?? 1;
  return ref.watch(communityRepositoryProvider(locale)).listPosts(boardId, page: page);
});

// 게시글 상세 프로바이더
final postDetailProvider = FutureProvider.family<PostDetail, Map<String, dynamic>>((ref, params) {
  final boardId = params['board_id'] as int;
  final postId = params['post_id'] as int;
  final locale = params['locale'] as String;
  return ref.watch(communityRepositoryProvider(locale)).getPost(boardId, postId);
});

// 카테고리 아이콘 매핑
IconData _categoryIcon(String category) {
  switch (category) {
    case 'GENERAL': return Icons.chat_bubble_outline;
    case 'QNA': return Icons.help_outline;
    case 'SHOWCASE': return Icons.photo_library_outlined;
    case 'TRADING': return Icons.store_outlined;
    case 'DISEASE': return Icons.healing_outlined;
    case 'BEGINNER': return Icons.school_outlined;
    case 'BREEDING': return Icons.favorite_border;
    default: return Icons.forum_outlined;
  }
}

Color _categoryColor(String category) {
  switch (category) {
    case 'GENERAL': return const Color(0xFF0EA5E9);
    case 'QNA': return const Color(0xFF7C3AED);
    case 'SHOWCASE': return const Color(0xFFDB2777);
    case 'TRADING': return const Color(0xFF16A34A);
    case 'DISEASE': return const Color(0xFFDC2626);
    case 'BEGINNER': return const Color(0xFFD97706);
    case 'BREEDING': return const Color(0xFFEC4899);
    default: return const Color(0xFF6B7280);
  }
}

// 커뮤니티 홈 (게시판 목록)
class CommunityScreen extends ConsumerWidget {
  const CommunityScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final boardsAsync = ref.watch(boardListProvider(locale));

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.communityBoards),
        backgroundColor: const Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
      ),
      body: boardsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => _LocaleErrorWidget(locale: locale),
        data: (boards) {
          if (boards.isEmpty) {
            return _LocaleErrorWidget(locale: locale);
          }
          return ListView.separated(
            padding: const EdgeInsets.all(12),
            itemCount: boards.length,
            separatorBuilder: (_, __) => const SizedBox(height: 8),
            itemBuilder: (context, i) => _BoardCard(board: boards[i]),
          );
        },
      ),
    );
  }
}

class _LocaleErrorWidget extends StatelessWidget {
  final String locale;
  const _LocaleErrorWidget({required this.locale});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.language, size: 64, color: Color(0xFF9CA3AF)),
            const SizedBox(height: 16),
            Text(
              'Community boards are not available in your current language ($locale).',
              textAlign: TextAlign.center,
              style: const TextStyle(color: Color(0xFF6B7280), fontSize: 14),
            ),
          ],
        ),
      ),
    );
  }
}

class _BoardCard extends StatelessWidget {
  final Board board;
  const _BoardCard({required this.board});

  @override
  Widget build(BuildContext context) {
    final color = _categoryColor(board.category);
    return InkWell(
      onTap: () => context.push('/community/${board.id}', extra: board),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 8, offset: const Offset(0, 2)),
          ],
        ),
        child: Row(
          children: [
            Container(
              width: 48, height: 48,
              decoration: BoxDecoration(
                color: color.withOpacity(0.12),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(_categoryIcon(board.category), color: color, size: 24),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(board.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                  if (board.description != null)
                    Text(board.description!, style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
                ],
              ),
            ),
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  '${board.postCount}',
                  style: TextStyle(fontWeight: FontWeight.bold, color: color, fontSize: 14),
                ),
                const Text('posts', style: TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
              ],
            ),
            const SizedBox(width: 8),
            const Icon(Icons.chevron_right, color: Color(0xFF9CA3AF)),
          ],
        ),
      ),
    );
  }
}

// 게시글 목록 화면
class BoardScreen extends ConsumerStatefulWidget {
  final int boardId;
  final Board? board;
  const BoardScreen({super.key, required this.boardId, this.board});

  @override
  ConsumerState<BoardScreen> createState() => _BoardScreenState();
}

class _BoardScreenState extends ConsumerState<BoardScreen> {
  int _page = 1;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final params = {'board_id': widget.boardId, 'locale': locale, 'page': _page};
    final postsAsync = ref.watch(postListProvider(params));

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.board?.name ?? l10n.navCommunity),
        backgroundColor: const Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/community/${widget.boardId}/create'),
        backgroundColor: const Color(0xFF0EA5E9),
        icon: const Icon(Icons.edit),
        label: Text(l10n.communityNewPost),
      ),
      body: postsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text(l10n.commonError)),
        data: (result) {
          if (result.items.isEmpty) {
            return Center(child: Text(l10n.commonNoResults));
          }
          return ListView.separated(
            padding: const EdgeInsets.all(12),
            itemCount: result.items.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, i) => _PostListTile(
              post: result.items[i],
              boardId: widget.boardId,
            ),
          );
        },
      ),
    );
  }
}

class _PostListTile extends StatelessWidget {
  final PostListItem post;
  final int boardId;
  const _PostListTile({required this.post, required this.boardId});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () => context.push('/community/$boardId/posts/${post.id}'),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (post.isPinned)
              const Padding(
                padding: EdgeInsets.only(bottom: 4),
                child: Row(
                  children: [
                    Icon(Icons.push_pin, size: 12, color: Color(0xFF0EA5E9)),
                    SizedBox(width: 4),
                    Text('Pinned', style: TextStyle(fontSize: 11, color: Color(0xFF0EA5E9), fontWeight: FontWeight.w600)),
                  ],
                ),
              ),
            Text(post.title, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
            const SizedBox(height: 6),
            Row(
              children: [
                Text(post.authorNickname, style: const TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
                const SizedBox(width: 8),
                Text(_formatDate(post.createdAt), style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
                const Spacer(),
                _StatChip(icon: Icons.visibility_outlined, count: post.viewCount),
                const SizedBox(width: 8),
                _StatChip(icon: Icons.thumb_up_outlined, count: post.likeCount),
                const SizedBox(width: 8),
                _StatChip(icon: Icons.comment_outlined, count: post.commentCount),
              ],
            ),
          ],
        ),
      ),
    );
  }

  String _formatDate(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    return '${dt.month}/${dt.day}';
  }
}

class _StatChip extends StatelessWidget {
  final IconData icon;
  final int count;
  const _StatChip({required this.icon, required this.count});

  @override
  Widget build(BuildContext context) => Row(
    children: [
      Icon(icon, size: 12, color: const Color(0xFF9CA3AF)),
      const SizedBox(width: 2),
      Text('$count', style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
    ],
  );
}

// 게시글 상세 화면
class PostDetailScreen extends ConsumerStatefulWidget {
  final int boardId;
  final int postId;
  const PostDetailScreen({super.key, required this.boardId, required this.postId});

  @override
  ConsumerState<PostDetailScreen> createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends ConsumerState<PostDetailScreen> {
  final _commentController = TextEditingController();

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final params = {'board_id': widget.boardId, 'post_id': widget.postId, 'locale': locale};
    final postAsync = ref.watch(postDetailProvider(params));

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 0.5,
        actions: [
          IconButton(icon: const Icon(Icons.share_outlined), onPressed: () {}),
          IconButton(icon: const Icon(Icons.more_vert), onPressed: () {}),
        ],
      ),
      body: postAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text(l10n.commonError)),
        data: (post) => Column(
          children: [
            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(post.title, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        CircleAvatar(
                          radius: 16,
                          backgroundColor: const Color(0xFFE0F2FE),
                          child: Text(
                            post.authorNickname.isNotEmpty ? post.authorNickname[0].toUpperCase() : '?',
                            style: const TextStyle(fontSize: 13, color: Color(0xFF0EA5E9), fontWeight: FontWeight.bold),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(post.authorNickname, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
                            Text(
                              _formatDateTime(post.createdAt),
                              style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF)),
                            ),
                          ],
                        ),
                        const Spacer(),
                        Text('${post.viewCount} views', style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
                      ],
                    ),
                    const Divider(height: 24),
                    Text(post.body, style: const TextStyle(fontSize: 14, height: 1.6)),
                    const SizedBox(height: 16),
                    // 좋아요 버튼
                    Row(
                      children: [
                        OutlinedButton.icon(
                          onPressed: () async {
                            final repo = ref.read(communityRepositoryProvider(locale));
                            try {
                              await repo.likePost(widget.boardId, widget.postId);
                              ref.invalidate(postDetailProvider(params));
                            } catch (_) {}
                          },
                          icon: Icon(
                            post.isLiked ? Icons.thumb_up : Icons.thumb_up_outlined,
                            size: 16,
                            color: post.isLiked ? const Color(0xFF0EA5E9) : null,
                          ),
                          label: Text('${post.likeCount} ${l10n.communityLike}'),
                          style: OutlinedButton.styleFrom(
                            foregroundColor: post.isLiked ? const Color(0xFF0EA5E9) : Colors.grey,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),
                    // 댓글 섹션
                    Text(
                      '${l10n.communityComment} (${post.comments.length})',
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                    ),
                    const SizedBox(height: 12),
                    ...post.comments.map((c) => _CommentTile(comment: c)),
                  ],
                ),
              ),
            ),
            // 댓글 입력창
            Container(
              padding: EdgeInsets.only(
                left: 12, right: 12, top: 8,
                bottom: MediaQuery.of(context).viewInsets.bottom + 8,
              ),
              decoration: const BoxDecoration(
                color: Colors.white,
                border: Border(top: BorderSide(color: Color(0xFFE5E7EB))),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _commentController,
                      decoration: InputDecoration(
                        hintText: l10n.communityComment,
                        border: OutlineInputBorder(borderRadius: BorderRadius.circular(24)),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                        isDense: true,
                      ),
                      maxLines: null,
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton(
                    icon: const Icon(Icons.send, color: Color(0xFF0EA5E9)),
                    onPressed: () async {
                      final text = _commentController.text.trim();
                      if (text.isEmpty) return;
                      final repo = ref.read(communityRepositoryProvider(locale));
                      try {
                        await repo.createComment(widget.boardId, widget.postId, text);
                        _commentController.clear();
                        ref.invalidate(postDetailProvider(params));
                      } catch (_) {}
                    },
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatDateTime(DateTime dt) {
    return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')} '
        '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }
}

class _CommentTile extends StatelessWidget {
  final Comment comment;
  const _CommentTile({required this.comment});

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CircleAvatar(
          radius: 14,
          backgroundColor: const Color(0xFFF3F4F6),
          child: Text(
            comment.authorNickname.isNotEmpty ? comment.authorNickname[0].toUpperCase() : '?',
            style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280)),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Text(comment.authorNickname, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 12)),
                  const SizedBox(width: 8),
                  Text(
                    '${comment.createdAt.month}/${comment.createdAt.day}',
                    style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF)),
                  ),
                ],
              ),
              const SizedBox(height: 2),
              Text(comment.body, style: const TextStyle(fontSize: 13)),
            ],
          ),
        ),
      ],
    ),
  );
}

// 게시글 작성 화면
class CreatePostScreen extends ConsumerStatefulWidget {
  final int boardId;
  const CreatePostScreen({super.key, required this.boardId});

  @override
  ConsumerState<CreatePostScreen> createState() => _CreatePostScreenState();
}

class _CreatePostScreenState extends ConsumerState<CreatePostScreen> {
  final _titleController = TextEditingController();
  final _bodyController = TextEditingController();
  bool _submitting = false;

  @override
  void dispose() {
    _titleController.dispose();
    _bodyController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.communityNewPost),
        backgroundColor: const Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
        actions: [
          TextButton(
            onPressed: _submitting ? null : () async {
              final title = _titleController.text.trim();
              final body = _bodyController.text.trim();
              if (title.isEmpty || body.isEmpty) return;
              setState(() => _submitting = true);
              try {
                final repo = ref.read(communityRepositoryProvider(locale));
                await repo.createPost(widget.boardId, title, body);
                if (context.mounted) context.pop();
              } catch (_) {
                setState(() => _submitting = false);
              }
            },
            child: Text(l10n.commonSave, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            TextField(
              controller: _titleController,
              decoration: InputDecoration(
                hintText: 'Title',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              ),
            ),
            const SizedBox(height: 12),
            Expanded(
              child: TextField(
                controller: _bodyController,
                decoration: InputDecoration(
                  hintText: 'Write your post...',
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                  alignLabelWithHint: true,
                ),
                maxLines: null,
                expands: true,
                textAlignVertical: TextAlignVertical.top,
              ),
            ),
            if (_submitting) const Padding(
              padding: EdgeInsets.only(top: 8),
              child: LinearProgressIndicator(),
            ),
          ],
        ),
      ),
    );
  }
}
