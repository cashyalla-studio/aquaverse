import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../domain/community_model.dart';

class CommunityRepository {
  final Dio _dio;
  CommunityRepository(this._dio);

  Future<List<Board>> listBoards() async {
    final resp = await _dio.get('/boards');
    return (resp.data as List? ?? []).map((e) => Board.fromJson(e)).toList();
  }

  Future<PostListResult> listPosts(int boardId, {int page = 1}) async {
    final resp = await _dio.get('/boards/$boardId/posts', queryParameters: {
      'page': page,
      'limit': 20,
    });
    return PostListResult.fromJson(resp.data);
  }

  Future<PostDetail> getPost(int boardId, int postId) async {
    final resp = await _dio.get('/boards/$boardId/posts/$postId');
    return PostDetail.fromJson(resp.data);
  }

  Future<void> createPost(int boardId, String title, String body) async {
    await _dio.post('/boards/$boardId/posts', data: {
      'title': title,
      'body': body,
    });
  }

  Future<void> likePost(int boardId, int postId) async {
    await _dio.post('/boards/$boardId/posts/$postId/like');
  }

  Future<void> createComment(int boardId, int postId, String body) async {
    await _dio.post('/boards/$boardId/posts/$postId/comments', data: {
      'body': body,
    });
  }
}

final communityRepositoryProvider = Provider.family<CommunityRepository, String>(
  (ref, locale) => CommunityRepository(ref.watch(dioProvider(locale))),
);
