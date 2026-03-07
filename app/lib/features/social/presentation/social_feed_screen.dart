import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

class SocialFeedScreen extends ConsumerStatefulWidget {
  const SocialFeedScreen({super.key});

  @override
  ConsumerState<SocialFeedScreen> createState() => _SocialFeedScreenState();
}

class _SocialFeedScreenState extends ConsumerState<SocialFeedScreen> {
  List<Map<String, dynamic>> _feed = [];
  List<Map<String, dynamic>> _suggestions = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final dio = ref.read(dioProvider('ko'));
      final results = await Future.wait([
        dio.get('/social/feed'),
        dio.get('/social/suggestions'),
      ]);
      if (mounted) {
        setState(() {
          _feed = List<Map<String, dynamic>>.from(results[0].data['feed'] ?? []);
          _suggestions = List<Map<String, dynamic>>.from(results[1].data['suggestions'] ?? []);
          _loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _follow(String userId) async {
    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.post('/social/users/$userId/follow');
      setState(() => _suggestions.removeWhere((s) => s['user_id'] == userId));
    } catch (_) {}
  }

  String _verbLabel(String verb) {
    const labels = {
      'LISTED': '새 분양글을 올렸습니다',
      'SOLD': '분양을 완료했습니다',
      'REVIEWED': '리뷰를 남겼습니다',
      'JOINED': '가입했습니다',
    };
    return labels[verb] ?? verb;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('팔로잉 피드')),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _feed.isEmpty
              ? _buildEmpty()
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView.builder(
                    itemCount: _feed.length,
                    itemBuilder: (ctx, i) => _buildFeedItem(_feed[i]),
                  ),
                ),
    );
  }

  Widget _buildEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Text('🐠', style: TextStyle(fontSize: 48)),
          const SizedBox(height: 8),
          const Text('팔로우한 사용자의 활동이 없습니다', style: TextStyle(color: Colors.grey)),
          if (_suggestions.isNotEmpty) ...[
            const SizedBox(height: 16),
            const Text('추천 팔로우', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            ..._suggestions.take(3).map((s) => ListTile(
                  title: Text(s['username'] ?? ''),
                  subtitle: Text(
                    s['common_fish'] != null && s['common_fish'] != ''
                        ? '🐟 ${s['common_fish']}'
                        : '신뢰도 ${s['trust_score']}',
                  ),
                  trailing: TextButton(
                    onPressed: () => _follow(s['user_id']),
                    child: const Text('팔로우'),
                  ),
                )),
          ],
        ],
      ),
    );
  }

  Widget _buildFeedItem(Map<String, dynamic> item) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            CircleAvatar(
              child: Text(
                (item['actor_name'] as String? ?? '?')[0].toUpperCase(),
                style: const TextStyle(fontWeight: FontWeight.bold),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  RichText(
                    text: TextSpan(
                      style: DefaultTextStyle.of(context).style,
                      children: [
                        TextSpan(
                          text: item['actor_name'] ?? '',
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                        TextSpan(text: ' ${_verbLabel(item['verb'] ?? '')}'),
                      ],
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    item['created_at'] ?? '',
                    style: const TextStyle(color: Colors.grey, fontSize: 12),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
