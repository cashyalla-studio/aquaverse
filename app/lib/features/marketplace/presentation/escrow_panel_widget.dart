import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';

enum EscrowStatus { pending, funded, released, refunded, disputed }

EscrowStatus _parseStatus(String s) {
  switch (s) {
    case 'FUNDED': return EscrowStatus.funded;
    case 'RELEASED': return EscrowStatus.released;
    case 'REFUNDED': return EscrowStatus.refunded;
    case 'DISPUTED': return EscrowStatus.disputed;
    default: return EscrowStatus.pending;
  }
}

// 에스크로 상태 Provider
final escrowStatusProvider = FutureProvider.family<EscrowStatus?, int>(
  (ref, tradeId) async {
    try {
      final dio = ref.read(dioProvider('ko'));
      final res = await dio.get<Map<String, dynamic>>('/trades/$tradeId/escrow');
      return _parseStatus(res.data?['escrow_status'] as String? ?? '');
    } on DioException {
      return null;
    }
  },
);

class EscrowPanelWidget extends ConsumerStatefulWidget {
  final int tradeId;
  final bool isBuyer;
  final double amount;
  final String currency;

  const EscrowPanelWidget({
    super.key,
    required this.tradeId,
    required this.isBuyer,
    required this.amount,
    this.currency = 'KRW',
  });

  @override
  ConsumerState<EscrowPanelWidget> createState() => _EscrowPanelWidgetState();
}

class _EscrowPanelWidgetState extends ConsumerState<EscrowPanelWidget> {
  bool _loading = false;
  String? _error;

  Future<void> _doAction(String action) async {
    setState(() { _loading = true; _error = null; });
    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.post<void>('/trades/${widget.tradeId}/escrow/$action');
      ref.invalidate(escrowStatusProvider(widget.tradeId));
    } on DioException catch (e) {
      setState(() {
        _error = e.response?.data?['message'] as String? ?? '처리 중 오류가 발생했습니다';
      });
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final statusAsync = ref.watch(escrowStatusProvider(widget.tradeId));

    return statusAsync.when(
      data: (status) {
        if (status == null) return const SizedBox.shrink();
        return _buildPanel(context, colorScheme, status);
      },
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
    );
  }

  Widget _buildPanel(BuildContext context, ColorScheme colorScheme, EscrowStatus status) {
    final configs = {
      EscrowStatus.pending:  (icon: Icons.hourglass_empty, label: '에스크로 대기',      color: colorScheme.surfaceContainerHigh),
      EscrowStatus.funded:   (icon: Icons.lock_rounded,    label: '에스크로 보관 중',   color: colorScheme.primaryContainer),
      EscrowStatus.released: (icon: Icons.check_circle,    label: '대금 지급 완료',     color: const Color(0xFFE8F5E9)),
      EscrowStatus.refunded: (icon: Icons.undo_rounded,    label: '환불 완료',          color: const Color(0xFFFFF3E0)),
      EscrowStatus.disputed: (icon: Icons.warning_rounded, label: '분쟁 중',            color: colorScheme.errorContainer),
    };
    final cfg = configs[status]!;

    final amountStr = '₩${widget.amount.toStringAsFixed(0).replaceAllMapped(
      RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'),
      (m) => '${m[1]},',
    )}';

    return Container(
      margin: const EdgeInsets.symmetric(vertical: 8),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: cfg.color,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(cfg.icon, size: 20),
              const SizedBox(width: 8),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('안전결제 (에스크로)', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                  Text(cfg.label, style: const TextStyle(fontSize: 12)),
                ],
              ),
              const Spacer(),
              Text(amountStr, style: const TextStyle(fontWeight: FontWeight.bold)),
            ],
          ),
          const SizedBox(height: 8),
          const Text(
            '플랫폼이 대금을 안전하게 보관하고 거래 완료 후 판매자에게 지급합니다.',
            style: TextStyle(fontSize: 12),
          ),
          if (_error != null) ...[
            const SizedBox(height: 8),
            Text(_error!, style: TextStyle(color: Theme.of(context).colorScheme.error, fontSize: 12)),
          ],
          if (widget.isBuyer && status == EscrowStatus.pending) ...[
            const SizedBox(height: 12),
            FilledButton(
              onPressed: _loading ? null : () => _doAction('fund'),
              style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(44)),
              child: _loading
                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : const Text('에스크로 입금'),
            ),
          ],
          if (widget.isBuyer && status == EscrowStatus.funded) ...[
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: FilledButton(
                    onPressed: _loading ? null : () => _doAction('release'),
                    child: const Text('수령 확인'),
                  ),
                ),
                const SizedBox(width: 8),
                OutlinedButton(
                  onPressed: _loading ? null : () => _doAction('refund'),
                  child: const Text('환불 요청'),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}
