import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

class BusinessProfile {
  final int id;
  final String storeName, city, description, logoUrl;
  final bool isVerified;
  final double avgRating;
  final int reviewCount;

  BusinessProfile({required this.id, required this.storeName, required this.city,
    required this.description, required this.logoUrl, required this.isVerified,
    required this.avgRating, required this.reviewCount});

  factory BusinessProfile.fromJson(Map<String, dynamic> j) => BusinessProfile(
    id: j['id'] as int,
    storeName: j['store_name'] as String? ?? '',
    city: j['city'] as String? ?? '',
    description: j['description'] as String? ?? '',
    logoUrl: j['logo_url'] as String? ?? '',
    isVerified: j['is_verified'] as bool? ?? false,
    avgRating: (j['avg_rating'] as num?)?.toDouble() ?? 0,
    reviewCount: j['review_count'] as int? ?? 0,
  );
}

final businessListProvider = FutureProvider.family<List<BusinessProfile>, String>((ref, city) async {
  final dio = ref.read(dioProvider('ko'));
  final params = city.isNotEmpty ? {'city': city, 'limit': 20} : {'limit': 20};
  final resp = await dio.get('/businesses', queryParameters: params);
  final list = (resp.data as Map<String, dynamic>)['businesses'] as List<dynamic>? ?? [];
  return list.map((e) => BusinessProfile.fromJson(e as Map<String, dynamic>)).toList();
});

class BusinessListScreen extends ConsumerStatefulWidget {
  const BusinessListScreen({super.key});

  @override
  ConsumerState<BusinessListScreen> createState() => _BusinessListScreenState();
}

class _BusinessListScreenState extends ConsumerState<BusinessListScreen> {
  final _cityCtrl = TextEditingController();
  String _city = '';

  @override
  void dispose() { _cityCtrl.dispose(); super.dispose(); }

  @override
  Widget build(BuildContext context) {
    final listAsync = ref.watch(businessListProvider(_city));

    return Scaffold(
      appBar: AppBar(title: const Text('수족관 업체 찾기')),
      body: Column(children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(children: [
            Expanded(
              child: TextField(
                controller: _cityCtrl,
                decoration: const InputDecoration(
                  hintText: '도시 검색 (예: 서울)',
                  border: OutlineInputBorder(),
                  contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                onSubmitted: (v) => setState(() => _city = v),
              ),
            ),
            const SizedBox(width: 8),
            ElevatedButton(
              onPressed: () => setState(() => _city = _cityCtrl.text),
              child: const Text('검색'),
            ),
          ]),
        ),
        Expanded(
          child: listAsync.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Center(child: Text('오류: $e')),
            data: (businesses) => businesses.isEmpty
                ? const Center(child: Text('등록된 업체가 없습니다'))
                : ListView.builder(
                    padding: const EdgeInsets.symmetric(horizontal: 12),
                    itemCount: businesses.length,
                    itemBuilder: (context, i) {
                      final b = businesses[i];
                      return Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          leading: CircleAvatar(
                            backgroundColor: Colors.blue.shade50,
                            backgroundImage: b.logoUrl.isNotEmpty ? NetworkImage(b.logoUrl) : null,
                            child: b.logoUrl.isEmpty ? const Text('🐠') : null,
                          ),
                          title: Row(children: [
                            Text(b.storeName, style: const TextStyle(fontWeight: FontWeight.w600)),
                            if (b.isVerified) ...[
                              const SizedBox(width: 4),
                              Container(padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(4)),
                                child: Text('인증', style: TextStyle(fontSize: 10, color: Colors.blue.shade700))),
                            ],
                          ]),
                          subtitle: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            if (b.city.isNotEmpty) Text(b.city, style: const TextStyle(fontSize: 12)),
                            Row(children: [
                              Text('★' * b.avgRating.round(), style: const TextStyle(color: Colors.amber, fontSize: 12)),
                              Text(' (${b.reviewCount})', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                            ]),
                          ]),
                          onTap: () {/* GoRouter로 상세 페이지 이동 */},
                        ),
                      );
                    },
                  ),
          ),
        ),
      ]),
    );
  }
}
