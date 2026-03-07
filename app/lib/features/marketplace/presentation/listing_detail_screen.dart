import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../l10n/app_localizations.dart';
import '../data/listing_repository.dart';
import '../domain/listing_model.dart';
import 'escrow_panel_widget.dart';

final listingDetailProvider = FutureProvider.family<ListingDetail, Map<String, dynamic>>((ref, params) {
  final id = params['id'] as int;
  final locale = params['locale'] as String? ?? 'en-US';
  return ref.watch(listingRepositoryProvider(locale)).getListing(id);
});

class ListingDetailScreen extends ConsumerWidget {
  final int listingId;
  const ListingDetailScreen({super.key, required this.listingId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final params = {'id': listingId, 'locale': locale};
    final listingAsync = ref.watch(listingDetailProvider(params));

    return listingAsync.when(
      loading: () => Scaffold(
        appBar: AppBar(backgroundColor: const Color(0xFF0EA5E9)),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (e, _) => Scaffold(
        appBar: AppBar(title: const Text('Error')),
        body: Center(child: Text(l10n.commonError)),
      ),
      data: (listing) => _ListingDetailView(listing: listing, l10n: l10n, locale: locale),
    );
  }
}

class _ListingDetailView extends ConsumerStatefulWidget {
  final ListingDetail listing;
  final AppLocalizations l10n;
  final String locale;
  const _ListingDetailView({
    required this.listing,
    required this.l10n,
    required this.locale,
  });

  @override
  ConsumerState<_ListingDetailView> createState() => _ListingDetailViewState();
}

class _ListingDetailViewState extends ConsumerState<_ListingDetailView> {
  int? _tradeId;

  ListingDetail get listing => widget.listing;
  AppLocalizations get l10n => widget.l10n;
  String get locale => widget.locale;

  Color get _healthColor => switch (listing.healthStatus) {
    'EXCELLENT' => const Color(0xFF16A34A),
    'GOOD' => const Color(0xFF2563EB),
    'DISEASE_HISTORY' => const Color(0xFFCA8A04),
    _ => const Color(0xFFDC2626),
  };

  String get _healthLabel => switch (listing.healthStatus) {
    'EXCELLENT' => l10n.marketplaceHealthExcellent,
    'GOOD' => l10n.marketplaceHealthGood,
    'DISEASE_HISTORY' => l10n.marketplaceHealthDiseaseHistory,
    _ => l10n.marketplaceHealthUnderTreatment,
  };

  String get _tradeLabel => switch (listing.tradeType) {
    'DIRECT' => '🤝 ${l10n.marketplaceDirect}',
    'COURIER' => '📦 ${l10n.marketplaceCourier}',
    'AQUA_COURIER' => '🐟 ${l10n.marketplaceAquaCourier}',
    _ => '✅ All',
  };

  @override
  Widget build(BuildContext context) {
    final double listingPrice = double.tryParse(listing.priceKrw) ?? 0.0;

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          // 이미지 앱바
          SliverAppBar(
            expandedHeight: 260,
            pinned: true,
            backgroundColor: const Color(0xFF0EA5E9),
            foregroundColor: Colors.white,
            actions: [
              IconButton(
                icon: const Icon(Icons.notifications_outlined),
                onPressed: () async {
                  final repo = ref.read(listingRepositoryProvider(locale));
                  try {
                    await repo.watchFish(listing.id);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(l10n.marketplaceWatchAlert)),
                    );
                  } catch (_) {}
                },
              ),
              IconButton(
                icon: const Icon(Icons.flag_outlined),
                onPressed: () => _showReportDialog(context),
              ),
            ],
            flexibleSpace: FlexibleSpaceBar(
              background: listing.imageUrls.isEmpty
                ? Container(
                    color: const Color(0xFFE0F2FE),
                    child: const Center(child: Text('🐠', style: TextStyle(fontSize: 80))),
                  )
                : PageView.builder(
                    itemCount: listing.imageUrls.length,
                    itemBuilder: (_, i) => Image.network(
                      listing.imageUrls[i],
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        color: const Color(0xFFE0F2FE),
                        child: const Center(child: Text('🐠', style: TextStyle(fontSize: 80))),
                      ),
                    ),
                  ),
            ),
          ),

          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 제목 + 가격
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(listing.title, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
                            Text(
                              listing.fishCommonName,
                              style: const TextStyle(fontSize: 13, color: Color(0xFF6B7280)),
                            ),
                            Text(
                              listing.fishScientificName,
                              style: const TextStyle(fontSize: 12, color: Colors.grey, fontStyle: FontStyle.italic),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          listing.isFree
                            ? Text(l10n.marketplaceFree, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Color(0xFF16A34A)))
                            : Text('₩${listing.priceKrw}', style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold)),
                          if (listing.bredBySeller)
                            Container(
                              margin: const EdgeInsets.only(top: 4),
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                              decoration: BoxDecoration(
                                color: const Color(0xFFE0F2FE),
                                borderRadius: BorderRadius.circular(20),
                              ),
                              child: Text(l10n.marketplaceBredBySeller, style: const TextStyle(fontSize: 10, color: Color(0xFF0EA5E9))),
                            ),
                        ],
                      ),
                    ],
                  ),

                  const SizedBox(height: 12),

                  // 배지
                  Wrap(
                    spacing: 8, runSpacing: 6,
                    children: [
                      _Badge(label: _healthLabel, color: _healthColor),
                      _Badge(label: _tradeLabel, color: const Color(0xFF0EA5E9)),
                    ],
                  ),

                  const Divider(height: 24),

                  // 판매자 정보
                  Row(
                    children: [
                      CircleAvatar(
                        radius: 20,
                        backgroundColor: const Color(0xFFE0F2FE),
                        child: Text(
                          listing.sellerNickname.isNotEmpty ? listing.sellerNickname[0].toUpperCase() : '?',
                          style: const TextStyle(fontSize: 16, color: Color(0xFF0EA5E9), fontWeight: FontWeight.bold),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(listing.sellerNickname, style: const TextStyle(fontWeight: FontWeight.bold)),
                            Row(
                              children: [
                                const Icon(Icons.star, size: 14, color: Color(0xFFFBBF24)),
                                const SizedBox(width: 2),
                                Text(
                                  '${listing.sellerTrustScore.toStringAsFixed(1)} ${l10n.marketplaceTrustScore}',
                                  style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280)),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      Row(
                        children: [
                          const Icon(Icons.location_on, size: 14, color: Colors.grey),
                          Text(listing.location, style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
                        ],
                      ),
                    ],
                  ),

                  const Divider(height: 24),

                  // 상세 정보
                  _InfoGrid(listing: listing, l10n: l10n),

                  // 설명
                  if (listing.description != null) ...[
                    const SizedBox(height: 16),
                    Container(
                      padding: const EdgeInsets.all(14),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(14),
                        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 8)],
                      ),
                      child: Text(listing.description!, style: const TextStyle(fontSize: 13, height: 1.6)),
                    ),
                  ],

                  // 에스크로 패널 (tradeId가 있을 때만 표시)
                  if (_tradeId != null)
                    EscrowPanelWidget(
                      tradeId: _tradeId!,
                      isBuyer: true,
                      amount: listingPrice,
                    ),

                  // 여름 AquaCourier 경고
                  if (listing.tradeType == 'DIRECT' || listing.tradeType == 'COURIER')
                    Container(
                      margin: const EdgeInsets.only(top: 16),
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: const Color(0xFFFEF3C7),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: const Color(0xFFFCD34D)),
                      ),
                      child: const Row(
                        children: [
                          Icon(Icons.warning_amber, color: Color(0xFFD97706), size: 16),
                          SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              'Summer shipping: Consider AquaCourier for live fish safety.',
                              style: TextStyle(fontSize: 11, color: Color(0xFF92400E)),
                            ),
                          ),
                        ],
                      ),
                    ),

                  const SizedBox(height: 100),
                ],
              ),
            ),
          ),
        ],
      ),

      // 거래 시작 버튼
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: ElevatedButton(
            onPressed: listing.status == 'ACTIVE' ? () => _confirmTrade(context) : null,
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF0EA5E9),
              foregroundColor: Colors.white,
              minimumSize: const Size.fromHeight(52),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
            ),
            child: Text(
              listing.status == 'ACTIVE' ? 'Request Trade' : 'Unavailable',
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
          ),
        ),
      ),
    );
  }

  void _confirmTrade(BuildContext context) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Request Trade', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            Text('Do you want to contact ${listing.sellerNickname} to trade "${listing.title}"?'),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: OutlinedButton(
                    onPressed: () => Navigator.pop(ctx),
                    child: Text(l10n.commonCancel),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton(
                    onPressed: () async {
                      Navigator.pop(ctx);
                      final repo = ref.read(listingRepositoryProvider(locale));
                      try {
                        final newTradeId = await repo.initiateTrade(listing.id);
                        if (mounted) {
                          setState(() => _tradeId = newTradeId);
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Trade request sent!')),
                          );
                        }
                      } catch (_) {}
                    },
                    style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF0EA5E9)),
                    child: const Text('Confirm', style: TextStyle(color: Colors.white)),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  void _showReportDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) {
        String? selected;
        return StatefulBuilder(
          builder: (ctx, setState) => AlertDialog(
            title: Text(l10n.marketplaceReportFraud),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                for (final reason in ['Fake listing', 'Wrong info', 'Scam attempt', 'Dead fish sold'])
                  RadioListTile<String>(
                    title: Text(reason),
                    value: reason,
                    groupValue: selected,
                    onChanged: (v) => setState(() => selected = v),
                    dense: true,
                  ),
              ],
            ),
            actions: [
              TextButton(onPressed: () => Navigator.pop(ctx), child: Text(l10n.commonCancel)),
              ElevatedButton(
                onPressed: selected == null ? null : () async {
                  Navigator.pop(ctx);
                  final repo = ref.read(listingRepositoryProvider(locale));
                  try {
                    await repo.reportFraud(listing.id, selected!);
                    if (context.mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Report submitted. Thank you.')),
                      );
                    }
                  } catch (_) {}
                },
                child: const Text('Submit'),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _InfoGrid extends StatelessWidget {
  final ListingDetail listing;
  final AppLocalizations l10n;
  const _InfoGrid({required this.listing, required this.l10n});

  @override
  Widget build(BuildContext context) {
    final items = <(String, String)>[];
    if (listing.quantity != null) items.add((l10n.marketplaceQuantity, '${listing.quantity}'));
    if (listing.currentSizeCm != null) items.add((l10n.marketplaceSize, '${listing.currentSizeCm}cm'));
    if (listing.ageMonths != null) {
      final months = listing.ageMonths!;
      items.add((l10n.marketplaceAge, months >= 12 ? '${months ~/ 12}y ${months % 12}m' : '${months}m'));
    }
    if (listing.sex != null) items.add((l10n.marketplaceSex, listing.sex!));

    if (items.isEmpty) return const SizedBox.shrink();

    return Wrap(
      spacing: 10, runSpacing: 10,
      children: items.map((item) => Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: const Color(0xFFF9FAFB),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: const Color(0xFFE5E7EB)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(item.$1, style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
            Text(item.$2, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold)),
          ],
        ),
      )).toList(),
    );
  }
}

class _Badge extends StatelessWidget {
  final String label;
  final Color color;
  const _Badge({required this.label, required this.color});

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
    decoration: BoxDecoration(
      color: color.withOpacity(0.1),
      borderRadius: BorderRadius.circular(20),
    ),
    child: Text(label, style: TextStyle(fontSize: 12, color: color, fontWeight: FontWeight.w600)),
  );
}
