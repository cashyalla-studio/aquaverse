import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';

class CitesCheckResult {
  final bool isBlocked;
  final bool hasWarning;
  final String? appendix;
  final String? message;
  final bool isInvasiveKr;

  const CitesCheckResult({
    required this.isBlocked,
    required this.hasWarning,
    this.appendix,
    this.message,
    required this.isInvasiveKr,
  });

  factory CitesCheckResult.fromJson(Map<String, dynamic> json) =>
      CitesCheckResult(
        isBlocked: json['is_blocked'] as bool? ?? false,
        hasWarning: json['has_warning'] as bool? ?? false,
        appendix: json['appendix'] as String?,
        message: json['message'] as String?,
        isInvasiveKr: json['is_invasive_kr'] as bool? ?? false,
      );
}

// CITES 체크 Provider (학명 기반 family)
final citesCheckProvider = FutureProvider.family<CitesCheckResult?, String>(
  (ref, scientificName) async {
    if (scientificName.trim().length < 3) return null;
    try {
      final dio = ref.read(dioProvider('ko'));
      final res = await dio.get<Map<String, dynamic>>(
        '/cites/check',
        queryParameters: {'scientific_name': scientificName},
      );
      final result = CitesCheckResult.fromJson(res.data!);
      return result.hasWarning ? result : null;
    } on DioException {
      return null;
    }
  },
);

/// CITES 경고 위젯 — 학명 입력 시 비동기로 체크
class CitesWarningWidget extends ConsumerWidget {
  final String scientificName;

  const CitesWarningWidget({super.key, required this.scientificName});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (scientificName.trim().length < 3) return const SizedBox.shrink();

    final check = ref.watch(citesCheckProvider(scientificName));

    return check.when(
      data: (result) {
        if (result == null) return const SizedBox.shrink();
        return _CitesWarningBanner(result: result);
      },
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
    );
  }
}

class _CitesWarningBanner extends StatelessWidget {
  final CitesCheckResult result;

  const _CitesWarningBanner({required this.result});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final isBlocked = result.isBlocked;

    final backgroundColor = isBlocked
        ? colorScheme.errorContainer
        : const Color(0xFFFFF8E1);
    final foregroundColor = isBlocked
        ? colorScheme.onErrorContainer
        : const Color(0xFF795548);
    final icon = isBlocked ? Icons.block_rounded : Icons.warning_amber_rounded;

    return AnimatedSize(
      duration: const Duration(milliseconds: 300),
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 8),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: backgroundColor,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isBlocked ? colorScheme.error.withOpacity(0.4) : Colors.amber.shade300,
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, color: isBlocked ? colorScheme.error : Colors.amber.shade700, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    isBlocked
                        ? 'CITES 보호종 — 거래 차단'
                        : 'CITES 부속서 ${result.appendix} 보호종',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 13,
                      color: foregroundColor,
                    ),
                  ),
                ),
              ],
            ),
            if (result.message != null) ...[
              const SizedBox(height: 4),
              Text(
                result.message!,
                style: TextStyle(fontSize: 12, color: foregroundColor.withOpacity(0.8)),
              ),
            ],
            if (result.isInvasiveKr) ...[
              const SizedBox(height: 4),
              Text(
                '한국 생태계 교란종 — 자연 방류 시 불법',
                style: TextStyle(
                  fontSize: 12,
                  color: Colors.orange.shade800,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
