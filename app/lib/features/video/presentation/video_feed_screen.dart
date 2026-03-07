import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

class VideoPost {
  final int id;
  final String title, username, videoUrl, thumbnailUrl;
  final int viewCount, likeCount;
  final bool isLiked;

  VideoPost({required this.id, required this.title, required this.username,
    required this.videoUrl, required this.thumbnailUrl,
    required this.viewCount, required this.likeCount, required this.isLiked});

  factory VideoPost.fromJson(Map<String, dynamic> j) => VideoPost(
    id: j['id'] as int,
    title: j['title'] as String? ?? '',
    username: j['username'] as String? ?? '',
    videoUrl: j['video_url'] as String? ?? '',
    thumbnailUrl: j['thumbnail_url'] as String? ?? '',
    viewCount: j['view_count'] as int? ?? 0,
    likeCount: j['like_count'] as int? ?? 0,
    isLiked: j['is_liked'] as bool? ?? false,
  );
}

final videoFeedProvider = FutureProvider<List<VideoPost>>((ref) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/videos', queryParameters: {'limit': 10, 'offset': 0});
  final list = (resp.data as Map<String, dynamic>)['videos'] as List<dynamic>? ?? [];
  return list.map((e) => VideoPost.fromJson(e as Map<String, dynamic>)).toList();
});

class VideoFeedScreen extends ConsumerWidget {
  const VideoFeedScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final feedAsync = ref.watch(videoFeedProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('수조 영상')),
      body: feedAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('오류: $e')),
        data: (videos) => videos.isEmpty
            ? const Center(
                child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                  Text('🎬', style: TextStyle(fontSize: 48)),
                  SizedBox(height: 12),
                  Text('아직 영상이 없습니다', style: TextStyle(color: Colors.grey)),
                ]),
              )
            : ListView.builder(
                itemCount: videos.length,
                itemBuilder: (context, i) {
                  final v = videos[i];
                  return Container(
                    height: MediaQuery.of(context).size.height * 0.65,
                    margin: const EdgeInsets.symmetric(vertical: 8, horizontal: 12),
                    decoration: BoxDecoration(
                      color: Colors.black,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Stack(children: [
                      // 썸네일 (video_player 패키지 없이 이미지로 대체)
                      ClipRRect(
                        borderRadius: BorderRadius.circular(16),
                        child: v.thumbnailUrl.isNotEmpty
                            ? Image.network(v.thumbnailUrl, width: double.infinity,
                                height: double.infinity, fit: BoxFit.cover,
                                errorBuilder: (_, __, ___) => Container(color: Colors.grey.shade800,
                                    child: const Icon(Icons.play_circle_outline, size: 64, color: Colors.white)))
                            : Container(color: Colors.grey.shade800,
                                child: const Icon(Icons.play_circle_outline, size: 64, color: Colors.white)),
                      ),
                      // 정보 오버레이
                      Positioned(
                        bottom: 0, left: 0, right: 0,
                        child: Container(
                          padding: const EdgeInsets.all(16),
                          decoration: const BoxDecoration(
                            borderRadius: BorderRadius.vertical(bottom: Radius.circular(16)),
                            gradient: LinearGradient(
                              begin: Alignment.topCenter, end: Alignment.bottomCenter,
                              colors: [Colors.transparent, Colors.black87],
                            ),
                          ),
                          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            Text('@${v.username}', style: const TextStyle(color: Colors.white70, fontSize: 13)),
                            Text(v.title, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                            Row(children: [
                              const Icon(Icons.favorite_border, color: Colors.white, size: 16),
                              Text(' ${v.likeCount}', style: const TextStyle(color: Colors.white, fontSize: 12)),
                              const SizedBox(width: 12),
                              const Icon(Icons.remove_red_eye_outlined, color: Colors.white, size: 16),
                              Text(' ${v.viewCount}', style: const TextStyle(color: Colors.white, fontSize: 12)),
                            ]),
                          ]),
                        ),
                      ),
                    ]),
                  );
                },
              ),
      ),
    );
  }
}
