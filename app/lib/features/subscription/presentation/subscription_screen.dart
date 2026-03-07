import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

class Plan {
  final String id, name;
  final int priceKrw;
  final List<String> features;
  Plan({required this.id, required this.name, required this.priceKrw, required this.features});
  factory Plan.fromJson(Map<String, dynamic> j) => Plan(
    id: j['id'] as String,
    name: j['name'] as String,
    priceKrw: j['price_krw'] as int? ?? 0,
    features: (j['features'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? [],
  );
}

final subscriptionPlansProvider = FutureProvider<List<Plan>>((ref) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/subscriptions/plans');
  final list = (resp.data as Map<String, dynamic>)['plans'] as List<dynamic>? ?? [];
  return list.map((e) => Plan.fromJson(e as Map<String, dynamic>)).toList();
});

class SubscriptionScreen extends ConsumerWidget {
  const SubscriptionScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final plansAsync = ref.watch(subscriptionPlansProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('구독 플랜')),
      body: plansAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('오류: $e')),
        data: (plans) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            const Text('Finara 구독 플랜', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
            const SizedBox(height: 16),
            ...plans.map((plan) => Card(
              margin: const EdgeInsets.only(bottom: 12),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(
                  color: plan.id == 'PRO' ? Colors.blue.shade300 : Colors.grey.shade200,
                  width: plan.id == 'PRO' ? 2 : 1,
                ),
              ),
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                  Row(children: [
                    Text(plan.name, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
                    if (plan.id == 'PRO') ...[
                      const SizedBox(width: 8),
                      Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(color: Colors.blue, borderRadius: BorderRadius.circular(12)),
                        child: const Text('추천', style: TextStyle(color: Colors.white, fontSize: 11))),
                    ],
                  ]),
                  const SizedBox(height: 4),
                  Text(
                    plan.priceKrw == 0 ? '무료' : '₩${plan.priceKrw.toString().replaceAllMapped(RegExp(r'(\d{3})(?=\d)'), (m) => '${m[1]},')}',
                    style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold),
                  ),
                  if (plan.priceKrw > 0)
                    const Text('/월', style: TextStyle(color: Colors.grey)),
                  const SizedBox(height: 12),
                  ...plan.features.map((f) => Padding(
                    padding: const EdgeInsets.symmetric(vertical: 3),
                    child: Row(children: [
                      const Icon(Icons.check_circle, color: Colors.green, size: 16),
                      const SizedBox(width: 8),
                      Expanded(child: Text(f, style: const TextStyle(fontSize: 14))),
                    ]),
                  )),
                  if (plan.id == 'PRO') ...[
                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.blue.shade600,
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        ),
                        onPressed: () {
                          // TODO: 구독 처리
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('준비 중입니다'))
                          );
                        },
                        child: const Text('1개월 무료 체험', style: TextStyle(fontSize: 16, color: Colors.white)),
                      ),
                    ),
                  ],
                ]),
              ),
            )),
          ],
        ),
      ),
    );
  }
}
