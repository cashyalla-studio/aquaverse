import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';

class FishRec {
  final int fishId;
  final String fishName;
  final String? scientificName;
  final String? imageUrl;
  final String? reason;
  final double score;

  FishRec({
    required this.fishId,
    required this.fishName,
    this.scientificName,
    this.imageUrl,
    this.reason,
    required this.score,
  });

  factory FishRec.fromJson(Map<String, dynamic> j) => FishRec(
        fishId: j['fish_id'] as int,
        fishName: j['fish_name'] as String,
        scientificName: j['scientific_name'] as String?,
        imageUrl: j['image_url'] as String?,
        reason: j['reason'] as String?,
        score: (j['score'] as num?)?.toDouble() ?? 0,
      );
}

final tankRecommendProvider = FutureProvider.family<List<FishRec>, int>((ref, tankId) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/tanks/$tankId/recommend');
  final data = resp.data as Map<String, dynamic>;
  final list = data['recommendations'] as List<dynamic>? ?? [];
  return list.map((e) => FishRec.fromJson(e as Map<String, dynamic>)).toList();
});

class TankDetailScreen extends ConsumerWidget {
  final int tankId;
  final String tankName;
  const TankDetailScreen({super.key, required this.tankId, required this.tankName});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final recAsync = ref.watch(tankRecommendProvider(tankId));

    return Scaffold(
      appBar: AppBar(title: Text(tankName)),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Text('추천 합사 어종',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold)),
          ),
          Expanded(
            child: recAsync.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(child: Text('오류: $e')),
              data: (recs) => recs.isEmpty
                  ? const Center(child: Text('수조에 어종을 추가하면\n합사 추천이 시작됩니다', textAlign: TextAlign.center))
                  : GridView.builder(
                      padding: const EdgeInsets.all(12),
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        childAspectRatio: 0.85,
                        crossAxisSpacing: 12,
                        mainAxisSpacing: 12,
                      ),
                      itemCount: recs.length,
                      itemBuilder: (context, i) {
                        final r = recs[i];
                        return Card(
                          elevation: 2,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                r.imageUrl != null && r.imageUrl!.isNotEmpty
                                    ? ClipOval(
                                        child: Image.network(r.imageUrl!, width: 64, height: 64, fit: BoxFit.cover,
                                            errorBuilder: (_, __, ___) => const Icon(Icons.set_meal, size: 48)))
                                    : const Icon(Icons.set_meal, size: 48, color: Colors.blue),
                                const SizedBox(height: 8),
                                Text(r.fishName,
                                    style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                                    textAlign: TextAlign.center,
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis),
                                if (r.reason != null && r.reason!.isNotEmpty) ...[
                                  const SizedBox(height: 4),
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                    decoration: BoxDecoration(
                                      color: Colors.green.shade50,
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Text(r.reason!,
                                        style: TextStyle(fontSize: 10, color: Colors.green.shade700),
                                        textAlign: TextAlign.center),
                                  ),
                                ],
                              ],
                            ),
                          ),
                        );
                      },
                    ),
            ),
          ),
        ],
      ),
    );
  }
}
