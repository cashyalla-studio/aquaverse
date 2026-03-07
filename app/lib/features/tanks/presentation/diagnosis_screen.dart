import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/api/api_client.dart';

class FishStatus {
  final String fishName, status, issue, suggestion;
  FishStatus(
      {required this.fishName,
      required this.status,
      required this.issue,
      required this.suggestion});
  factory FishStatus.fromJson(Map<String, dynamic> j) => FishStatus(
        fishName: j['fish_name'] as String? ?? '',
        status: j['status'] as String? ?? 'good',
        issue: j['issue'] as String? ?? '',
        suggestion: j['suggestion'] as String? ?? '',
      );
}

class TankDiagnosis {
  final int tankId;
  final String summary, createdAt;
  final List<FishStatus> fishStates;
  final List<String> actions;
  TankDiagnosis(
      {required this.tankId,
      required this.summary,
      required this.createdAt,
      required this.fishStates,
      required this.actions});
  factory TankDiagnosis.fromJson(Map<String, dynamic> j) => TankDiagnosis(
        tankId: (j['tank_id'] as int?) ?? 0,
        summary: j['summary'] as String? ?? '',
        createdAt: j['created_at'] as String? ?? '',
        fishStates: (j['fish_states'] as List<dynamic>? ?? [])
            .map((e) => FishStatus.fromJson(e as Map<String, dynamic>))
            .toList(),
        actions: (j['actions'] as List<dynamic>? ?? [])
            .map((e) => e.toString())
            .toList(),
      );
}

final diagnosisProvider =
    FutureProvider.family<TankDiagnosis, int>((ref, tankId) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/tanks/$tankId/diagnosis');
  return TankDiagnosis.fromJson(resp.data as Map<String, dynamic>);
});

class DiagnosisScreen extends ConsumerWidget {
  final int tankId;
  const DiagnosisScreen({super.key, required this.tankId});

  Color _statusColor(String status) {
    switch (status) {
      case 'good':
        return Colors.green;
      case 'warning':
        return Colors.orange;
      case 'danger':
        return Colors.red;
      default:
        return Colors.grey;
    }
  }

  String _statusEmoji(String status) {
    switch (status) {
      case 'good':
        return '🟢';
      case 'warning':
        return '🟡';
      case 'danger':
        return '🔴';
      default:
        return '⚪';
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final diagAsync = ref.watch(diagnosisProvider(tankId));

    return Scaffold(
      appBar: AppBar(
        title: const Text('AI 진단 결과'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.invalidate(diagnosisProvider(tankId)),
          ),
        ],
      ),
      body: diagAsync.when(
        loading: () => const Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              CircularProgressIndicator(),
              SizedBox(height: 12),
              Text('Claude AI가 분석 중입니다...',
                  style: TextStyle(color: Colors.grey)),
            ],
          ),
        ),
        error: (e, _) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 48, color: Colors.red),
                const SizedBox(height: 12),
                Text('$e',
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: Colors.red)),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () =>
                      context.go('/tanks/$tankId/water-params'),
                  child: const Text('수질 먼저 기록하기'),
                ),
              ],
            ),
          ),
        ),
        data: (diag) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.blue.shade50,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(diag.summary,
                  style: const TextStyle(fontSize: 15)),
            ),
            const SizedBox(height: 16),
            if (diag.fishStates.isNotEmpty) ...[
              const Text('어종별 상태',
                  style: TextStyle(
                      fontWeight: FontWeight.bold, fontSize: 16)),
              const SizedBox(height: 8),
              ...diag.fishStates.map((fs) => Card(
                    margin: const EdgeInsets.only(bottom: 8),
                    child: ListTile(
                      leading: Text(_statusEmoji(fs.status),
                          style: const TextStyle(fontSize: 24)),
                      title: Text(fs.fishName,
                          style: const TextStyle(
                              fontWeight: FontWeight.w600)),
                      subtitle: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          if (fs.issue.isNotEmpty)
                            Text(fs.issue,
                                style: const TextStyle(fontSize: 12)),
                          if (fs.suggestion.isNotEmpty)
                            Text('→ ${fs.suggestion}',
                                style: TextStyle(
                                    fontSize: 12,
                                    color: Colors.blue.shade700)),
                        ],
                      ),
                      trailing: Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color:
                              _statusColor(fs.status).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(fs.status,
                            style: TextStyle(
                                color: _statusColor(fs.status),
                                fontSize: 12)),
                      ),
                    ),
                  )),
              const SizedBox(height: 8),
            ],
            if (diag.actions.isNotEmpty) ...[
              const Text('권장 조치',
                  style: TextStyle(
                      fontWeight: FontWeight.bold, fontSize: 16)),
              const SizedBox(height: 8),
              ...diag.actions.asMap().entries.map((e) => Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('${e.key + 1}. ',
                            style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Colors.blue)),
                        Expanded(
                            child: Text(e.value,
                                style: const TextStyle(fontSize: 14))),
                      ],
                    ),
                  )),
            ],
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () =>
                  context.go('/tanks/$tankId/water-params'),
              icon: const Icon(Icons.water_drop),
              label: const Text('수질 다시 기록하기'),
            ),
          ],
        ),
      ),
    );
  }
}
