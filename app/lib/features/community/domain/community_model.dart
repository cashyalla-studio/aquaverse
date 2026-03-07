class Board {
  final int id;
  final String name;
  final String? description;
  final String locale;
  final String category;
  final bool isRtl;
  final int postCount;

  const Board({
    required this.id,
    required this.name,
    this.description,
    required this.locale,
    required this.category,
    required this.isRtl,
    required this.postCount,
  });

  factory Board.fromJson(Map<String, dynamic> json) => Board(
    id: json['id'],
    name: json['name'] ?? '',
    description: json['description'],
    locale: json['locale'] ?? '',
    category: json['category'] ?? '',
    isRtl: json['is_rtl'] ?? false,
    postCount: json['post_count'] ?? 0,
  );
}

class PostListItem {
  final int id;
  final int boardId;
  final String title;
  final String authorNickname;
  final String? authorAvatarUrl;
  final DateTime createdAt;
  final int viewCount;
  final int likeCount;
  final int commentCount;
  final bool isPinned;

  const PostListItem({
    required this.id,
    required this.boardId,
    required this.title,
    required this.authorNickname,
    this.authorAvatarUrl,
    required this.createdAt,
    required this.viewCount,
    required this.likeCount,
    required this.commentCount,
    required this.isPinned,
  });

  factory PostListItem.fromJson(Map<String, dynamic> json) => PostListItem(
    id: json['id'],
    boardId: json['board_id'],
    title: json['title'] ?? '',
    authorNickname: json['author_nickname'] ?? '',
    authorAvatarUrl: json['author_avatar_url'],
    createdAt: DateTime.tryParse(json['created_at'] ?? '') ?? DateTime.now(),
    viewCount: json['view_count'] ?? 0,
    likeCount: json['like_count'] ?? 0,
    commentCount: json['comment_count'] ?? 0,
    isPinned: json['is_pinned'] ?? false,
  );
}

class PostDetail {
  final int id;
  final int boardId;
  final String title;
  final String body;
  final String authorNickname;
  final String? authorAvatarUrl;
  final DateTime createdAt;
  final int viewCount;
  final int likeCount;
  final bool isLiked;
  final List<Comment> comments;

  const PostDetail({
    required this.id,
    required this.boardId,
    required this.title,
    required this.body,
    required this.authorNickname,
    this.authorAvatarUrl,
    required this.createdAt,
    required this.viewCount,
    required this.likeCount,
    required this.isLiked,
    required this.comments,
  });

  factory PostDetail.fromJson(Map<String, dynamic> json) => PostDetail(
    id: json['id'],
    boardId: json['board_id'],
    title: json['title'] ?? '',
    body: json['body'] ?? '',
    authorNickname: json['author_nickname'] ?? '',
    authorAvatarUrl: json['author_avatar_url'],
    createdAt: DateTime.tryParse(json['created_at'] ?? '') ?? DateTime.now(),
    viewCount: json['view_count'] ?? 0,
    likeCount: json['like_count'] ?? 0,
    isLiked: json['is_liked'] ?? false,
    comments: (json['comments'] as List? ?? [])
        .map((e) => Comment.fromJson(e))
        .toList(),
  );
}

class Comment {
  final int id;
  final String body;
  final String authorNickname;
  final DateTime createdAt;

  const Comment({
    required this.id,
    required this.body,
    required this.authorNickname,
    required this.createdAt,
  });

  factory Comment.fromJson(Map<String, dynamic> json) => Comment(
    id: json['id'],
    body: json['body'] ?? '',
    authorNickname: json['author_nickname'] ?? '',
    createdAt: DateTime.tryParse(json['created_at'] ?? '') ?? DateTime.now(),
  );
}

class PostListResult {
  final List<PostListItem> items;
  final int totalCount;
  final int page;

  const PostListResult({
    required this.items,
    required this.totalCount,
    required this.page,
  });

  factory PostListResult.fromJson(Map<String, dynamic> json) => PostListResult(
    items: (json['items'] as List? ?? [])
        .map((e) => PostListItem.fromJson(e))
        .toList(),
    totalCount: json['total_count'] ?? 0,
    page: json['page'] ?? 1,
  );
}
