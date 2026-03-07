import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:go_router/go_router.dart';
import '../../../l10n/app_localizations.dart';
import '../data/fish_repository.dart';
import '../domain/fish_model.dart';

final fishDetailProvider = FutureProvider.family<FishDetail, Map<String, dynamic>>((ref, params) {
  final id = params['id'] as int;
  final locale = params['locale'] as String? ?? 'en-US';
  return ref.watch(fishRepositoryProvider(locale)).getFish(id);
});

class FishDetailScreen extends ConsumerWidget {
  final int fishId;
  const FishDetailScreen({super.key, required this.fishId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final params = {'id': fishId, 'locale': locale};
    final fishAsync = ref.watch(fishDetailProvider(params));

    return fishAsync.when(
      loading: () => Scaffold(
        appBar: AppBar(backgroundColor: const Color(0xFF0EA5E9)),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (e, _) => Scaffold(
        appBar: AppBar(title: const Text('Error')),
        body: Center(child: Text(l10n.commonError)),
      ),
      data: (fish) => _FishDetailView(fish: fish, l10n: l10n),
    );
  }
}

class _FishDetailView extends StatelessWidget {
  final FishDetail fish;
  final AppLocalizations l10n;
  const _FishDetailView({required this.fish, required this.l10n});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: CustomScrollView(
        slivers: [
          // 히어로 이미지 앱바
          SliverAppBar(
            expandedHeight: 280,
            pinned: true,
            backgroundColor: const Color(0xFF0EA5E9),
            foregroundColor: Colors.white,
            flexibleSpace: FlexibleSpaceBar(
              background: fish.primaryImageUrl != null
                ? CachedNetworkImage(
                    imageUrl: fish.primaryImageUrl!,
                    fit: BoxFit.cover,
                    placeholder: (_, __) => Container(color: const Color(0xFF0EA5E9)),
                    errorWidget: (_, __, ___) => _ImagePlaceholder(),
                  )
                : _ImagePlaceholder(),
            ),
            actions: [
              IconButton(
                icon: const Icon(Icons.store_outlined),
                onPressed: () => context.push('/marketplace?q=${fish.primaryCommonName}'),
                tooltip: 'Find in Marketplace',
              ),
            ],
          ),

          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 이름 섹션
                  Text(
                    fish.localizedCommonName,
                    style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                  Text(
                    fish.scientificName,
                    style: const TextStyle(fontSize: 15, color: Colors.grey, fontStyle: FontStyle.italic),
                  ),
                  if (fish.family != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Text(
                        'Family: ${fish.family}',
                        style: const TextStyle(fontSize: 13, color: Color(0xFF6B7280)),
                      ),
                    ),

                  const SizedBox(height: 16),

                  // 배지 행
                  Wrap(
                    spacing: 8, runSpacing: 6,
                    children: [
                      if (fish.careLevel != null) _CareLevelBadge(level: fish.careLevel!, l10n: l10n),
                      if (fish.temperament != null) _TemperamentBadge(temperament: fish.temperament!),
                      if (fish.dietType != null) _DietBadge(diet: fish.dietType!),
                    ],
                  ),

                  const SizedBox(height: 20),

                  // 파라미터 카드
                  _ParameterGrid(fish: fish, l10n: l10n),

                  // 케어 노트
                  if (fish.localizedCareNotes != null) ...[
                    const SizedBox(height: 20),
                    _SectionCard(
                      title: l10n.fishCareNotes,
                      icon: Icons.spa_outlined,
                      color: const Color(0xFF16A34A),
                      child: Text(fish.localizedCareNotes!, style: const TextStyle(fontSize: 13, height: 1.6)),
                    ),
                  ],

                  // 식이 노트
                  if (fish.localizedDietNotes != null) ...[
                    const SizedBox(height: 12),
                    _SectionCard(
                      title: l10n.fishDiet,
                      icon: Icons.restaurant_outlined,
                      color: const Color(0xFFD97706),
                      child: Text(fish.localizedDietNotes!, style: const TextStyle(fontSize: 13, height: 1.6)),
                    ),
                  ],

                  // 번식 노트
                  if (fish.localizedBreedingNotes != null) ...[
                    const SizedBox(height: 12),
                    _SectionCard(
                      title: l10n.fishBreeding,
                      icon: Icons.favorite_border,
                      color: const Color(0xFFEC4899),
                      child: Text(fish.localizedBreedingNotes!, style: const TextStyle(fontSize: 13, height: 1.6)),
                    ),
                  ],

                  // 저작권 표시
                  if (fish.attribution != null) ...[
                    const SizedBox(height: 20),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF3F4F6),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(Icons.info_outline, size: 14, color: Color(0xFF6B7280)),
                              const SizedBox(width: 4),
                              Text(
                                fish.license ?? 'Attribution',
                                style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: Color(0xFF6B7280)),
                              ),
                            ],
                          ),
                          const SizedBox(height: 4),
                          Text(
                            fish.attribution!,
                            style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF)),
                          ),
                        ],
                      ),
                    ),
                  ],

                  const SizedBox(height: 80),
                ],
              ),
            ),
          ),
        ],
      ),

      // 마켓플레이스 이동 FAB
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/marketplace?q=${Uri.encodeComponent(fish.primaryCommonName)}'),
        backgroundColor: const Color(0xFF0EA5E9),
        icon: const Icon(Icons.store_outlined),
        label: const Text('Find for Adoption'),
      ),
    );
  }
}

class _ImagePlaceholder extends StatelessWidget {
  @override
  Widget build(BuildContext context) => Container(
    color: const Color(0xFF0EA5E9),
    child: const Center(child: Text('🐠', style: TextStyle(fontSize: 80))),
  );
}

class _CareLevelBadge extends StatelessWidget {
  final String level;
  final AppLocalizations l10n;
  const _CareLevelBadge({required this.level, required this.l10n});

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (level) {
      'BEGINNER' => (l10n.careLevelBeginner, const Color(0xFF16A34A)),
      'INTERMEDIATE' => (l10n.careLevelIntermediate, const Color(0xFFD97706)),
      _ => (l10n.careLevelExpert, const Color(0xFFDC2626)),
    };
    return _Badge(label: label, color: color);
  }
}

class _TemperamentBadge extends StatelessWidget {
  final String temperament;
  const _TemperamentBadge({required this.temperament});

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (temperament) {
      'PEACEFUL' => ('Peaceful', const Color(0xFF2563EB)),
      'SEMI_AGGRESSIVE' => ('Semi-Aggressive', const Color(0xFFD97706)),
      _ => ('Aggressive', const Color(0xFFDC2626)),
    };
    return _Badge(label: label, color: color);
  }
}

class _DietBadge extends StatelessWidget {
  final String diet;
  const _DietBadge({required this.diet});

  @override
  Widget build(BuildContext context) {
    final label = diet.replaceAll('_', ' ').toLowerCase()
        .split(' ').map((w) => w.isEmpty ? '' : '${w[0].toUpperCase()}${w.substring(1)}').join(' ');
    return _Badge(label: '🍽 $label', color: const Color(0xFF7C3AED));
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
      color: color.withOpacity(0.12),
      borderRadius: BorderRadius.circular(20),
      border: Border.all(color: color.withOpacity(0.3)),
    ),
    child: Text(label, style: TextStyle(fontSize: 12, color: color, fontWeight: FontWeight.w600)),
  );
}

class _ParameterGrid extends StatelessWidget {
  final FishDetail fish;
  final AppLocalizations l10n;
  const _ParameterGrid({required this.fish, required this.l10n});

  @override
  Widget build(BuildContext context) {
    final items = <(String, String, IconData, Color)>[];
    if (fish.maxSizeCm != null) items.add((l10n.fishMaxSize, '${fish.maxSizeCm}cm', Icons.straighten, const Color(0xFF0EA5E9)));
    if (fish.minTankSizeLiters != null) items.add((l10n.fishTankSize, '${fish.minTankSizeLiters}L', Icons.water, const Color(0xFF2563EB)));
    if (fish.phMin != null && fish.phMax != null) items.add((l10n.fishPhRange, '${fish.phMin} – ${fish.phMax}', Icons.science_outlined, const Color(0xFF7C3AED)));
    if (fish.tempMinC != null && fish.tempMaxC != null) items.add((l10n.fishTempRange, '${fish.tempMinC}°–${fish.tempMaxC}°C', Icons.thermostat_outlined, const Color(0xFFDC2626)));
    if (fish.lifespanYears != null) items.add((l10n.fishLifespan, '${fish.lifespanYears} yrs', Icons.timeline, const Color(0xFF16A34A)));

    if (items.isEmpty) return const SizedBox.shrink();

    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      crossAxisSpacing: 10,
      mainAxisSpacing: 10,
      childAspectRatio: 2.4,
      children: items.map((item) => _ParamCard(
        label: item.$1,
        value: item.$2,
        icon: item.$3,
        color: item.$4,
      )).toList(),
    );
  }
}

class _ParamCard extends StatelessWidget {
  final String label;
  final String value;
  final IconData icon;
  final Color color;
  const _ParamCard({required this.label, required this.value, required this.icon, required this.color});

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
    decoration: BoxDecoration(
      color: color.withOpacity(0.06),
      borderRadius: BorderRadius.circular(12),
      border: Border.all(color: color.withOpacity(0.15)),
    ),
    child: Row(
      children: [
        Icon(icon, size: 20, color: color),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF6B7280))),
              Text(value, style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: color)),
            ],
          ),
        ),
      ],
    ),
  );
}

class _SectionCard extends StatelessWidget {
  final String title;
  final IconData icon;
  final Color color;
  final Widget child;
  const _SectionCard({required this.title, required this.icon, required this.color, required this.child});

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.all(14),
    decoration: BoxDecoration(
      color: Colors.white,
      borderRadius: BorderRadius.circular(14),
      boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 8)],
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 16, color: color),
            const SizedBox(width: 6),
            Text(title, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: color)),
          ],
        ),
        const SizedBox(height: 8),
        child,
      ],
    ),
  );
}
