import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../l10n/app_localizations.dart';

class MarketplaceScreen extends ConsumerStatefulWidget {
  const MarketplaceScreen({super.key});

  @override
  ConsumerState<MarketplaceScreen> createState() => _MarketplaceScreenState();
}

class _MarketplaceScreenState extends ConsumerState<MarketplaceScreen> {
  String _search = '';
  String _tradeType = '';

  static const _healthColors = {
    'EXCELLENT': Color(0xFF16A34A),
    'GOOD': Color(0xFF2563EB),
    'DISEASE_HISTORY': Color(0xFFCA8A04),
    'UNDER_TREATMENT': Color(0xFFDC2626),
  };

  static const _tradeEmoji = {
    'DIRECT': '🤝',
    'COURIER': '📦',
    'AQUA_COURIER': '🐟',
    'ALL': '✅',
  };

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.marketplaceTitle),
        backgroundColor: const Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
        actions: [
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () {/* 알림 구독 */},
            tooltip: l10n.marketplaceWatchAlert,
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/marketplace/create'),
        backgroundColor: const Color(0xFF0EA5E9),
        icon: const Icon(Icons.add),
        label: Text(l10n.marketplaceCreateListing),
      ),
      body: Column(
        children: [
          // 검색 + 필터
          Container(
            color: Colors.white,
            padding: const EdgeInsets.all(12),
            child: Column(
              children: [
                TextField(
                  onChanged: (v) => setState(() => _search = v),
                  decoration: InputDecoration(
                    hintText: l10n.commonSearch,
                    prefixIcon: const Icon(Icons.search),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                    contentPadding: const EdgeInsets.symmetric(vertical: 0),
                  ),
                ),
                const SizedBox(height: 8),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      _Chip(label: 'All', selected: _tradeType.isEmpty, onTap: () => setState(() => _tradeType = '')),
                      const SizedBox(width: 6),
                      _Chip(label: '🤝 ${l10n.marketplaceDirect}', selected: _tradeType == 'DIRECT', onTap: () => setState(() => _tradeType = 'DIRECT')),
                      const SizedBox(width: 6),
                      _Chip(label: '🐟 ${l10n.marketplaceAquaCourier}', selected: _tradeType == 'AQUA_COURIER', onTap: () => setState(() => _tradeType = 'AQUA_COURIER')),
                      const SizedBox(width: 6),
                      _Chip(label: '📦 ${l10n.marketplaceCourier}', selected: _tradeType == 'COURIER', onTap: () => setState(() => _tradeType = 'COURIER')),
                    ],
                  ),
                ),
              ],
            ),
          ),

          // 여름 경고 배너
          Container(
            margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: const Color(0xFFFEF3C7),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: const Color(0xFFFCD34D)),
            ),
            child: Row(
              children: [
                const Icon(Icons.warning_amber, color: Color(0xFFD97706), size: 18),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Summer shipping risk: Consider AquaCourier for live fish safety.',
                    style: const TextStyle(fontSize: 12, color: Color(0xFF92400E)),
                  ),
                ),
              ],
            ),
          ),

          // 목록 (더미 데이터 - 실제로는 API 연동)
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.all(12),
              itemCount: 5, // TODO: API 연동
              separatorBuilder: (_, __) => const SizedBox(height: 8),
              itemBuilder: (context, index) => _ListingCard(index: index, l10n: l10n),
            ),
          ),
        ],
      ),
    );
  }
}

class _ListingCard extends StatelessWidget {
  final int index;
  final AppLocalizations l10n;
  const _ListingCard({required this.index, required this.l10n});

  @override
  Widget build(BuildContext context) {
    // 더미 데이터
    final samples = [
      {'title': 'Discus Breeding Pair', 'fish': 'Symphysodon aequifasciatus', 'price': '150,000', 'health': 'EXCELLENT', 'trade': 'DIRECT', 'location': 'Seoul, Gangnam'},
      {'title': 'Neon Tetra x20', 'fish': 'Paracheirodon innesi', 'price': '0', 'health': 'GOOD', 'trade': 'ALL', 'location': 'Busan'},
      {'title': 'Platinum Arowana', 'fish': 'Scleropages formosus', 'price': '2,500,000', 'health': 'EXCELLENT', 'trade': 'AQUA_COURIER', 'location': 'Incheon'},
      {'title': 'Altum Angelfish x3', 'fish': 'Pterophyllum altum', 'price': '45,000', 'health': 'GOOD', 'trade': 'DIRECT', 'location': 'Daejeon'},
      {'title': 'Cardinal Tetra x50', 'fish': 'Paracheirodon axelrodi', 'price': '30,000', 'health': 'GOOD', 'trade': 'COURIER', 'location': 'Gwangju'},
    ];
    final s = samples[index % samples.length];
    final isFree = s['price'] == '0';

    return GestureDetector(
      onTap: () => context.push('/marketplace/$index'),
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 8, offset: const Offset(0, 2))],
        ),
        child: Row(
          children: [
            // 이미지
            Container(
              width: 90, height: 90,
              margin: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFE0F2FE),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Center(child: Text('🐠', style: TextStyle(fontSize: 36))),
            ),

            // 정보
            Expanded(
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 4),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(s['title']!, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                    Text(s['fish']!, style: const TextStyle(fontSize: 11, color: Colors.grey, fontStyle: FontStyle.italic)),
                    const SizedBox(height: 6),

                    // 배지
                    Wrap(
                      spacing: 4, runSpacing: 4,
                      children: [
                        _Badge(
                          label: s['health']!.replaceAll('_', ' '),
                          color: Color(0xFF16A34A), // EXCELLENT
                        ),
                        _Badge(
                          label: s['trade'] == 'AQUA_COURIER' ? '🐟 AquaCourier'
                            : s['trade'] == 'DIRECT' ? '🤝 Direct'
                            : s['trade'] == 'ALL' ? '✅ All'
                            : '📦 Courier',
                          color: Colors.grey,
                        ),
                      ],
                    ),

                    const SizedBox(height: 4),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Row(
                          children: [
                            const Icon(Icons.location_on, size: 12, color: Colors.grey),
                            Text(s['location']!, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                          ],
                        ),
                        isFree
                          ? Text(l10n.marketplaceFree, style: const TextStyle(color: Color(0xFF16A34A), fontWeight: FontWeight.bold, fontSize: 14))
                          : Text('₩${s['price']}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                      ],
                    ),
                  ],
                ),
              ),
            ),

            const Padding(
              padding: EdgeInsets.only(right: 12),
              child: Icon(Icons.chevron_right, color: Colors.grey),
            ),
          ],
        ),
      ),
    );
  }
}

class _Badge extends StatelessWidget {
  final String label;
  final Color color;
  const _Badge({required this.label, required this.color});

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
    decoration: BoxDecoration(
      color: color.withOpacity(0.1),
      borderRadius: BorderRadius.circular(20),
    ),
    child: Text(label, style: TextStyle(fontSize: 10, color: color, fontWeight: FontWeight.w600)),
  );
}

class _Chip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback onTap;
  const _Chip({required this.label, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) => GestureDetector(
    onTap: onTap,
    child: Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFF0EA5E9) : Colors.grey.shade100,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(label, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: selected ? Colors.white : Colors.grey.shade700)),
    ),
  );
}
