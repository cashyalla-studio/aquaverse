import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../l10n/app_localizations.dart';
import '../data/fish_repository.dart';
import '../domain/fish_model.dart';

// 상태 프로바이더
final fishListProvider = FutureProvider.family<FishListResult, Map<String, dynamic>>((ref, params) {
  final locale = params['locale'] as String? ?? 'en-US';
  final repo = ref.watch(fishRepositoryProvider(locale));
  return repo.listFish(
    query: params['query'] as String?,
    careLevel: params['care_level'] as String?,
    page: params['page'] as int? ?? 1,
  );
});

class FishListScreen extends ConsumerStatefulWidget {
  const FishListScreen({super.key});

  @override
  ConsumerState<FishListScreen> createState() => _FishListScreenState();
}

class _FishListScreenState extends ConsumerState<FishListScreen> {
  final _searchController = TextEditingController();
  String _query = '';
  String _careLevel = '';
  int _page = 1;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final params = {
      'locale': locale,
      'query': _query,
      'care_level': _careLevel,
      'page': _page,
    };
    final fishAsync = ref.watch(fishListProvider(params));

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.navEncyclopedia),
        elevation: 0,
        backgroundColor: const Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
      ),
      body: Column(
        children: [
          // 검색바
          Container(
            color: const Color(0xFF0EA5E9),
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
            child: TextField(
              controller: _searchController,
              onChanged: (v) => setState(() { _query = v; _page = 1; }),
              decoration: InputDecoration(
                hintText: l10n.commonSearch,
                prefixIcon: const Icon(Icons.search, color: Colors.white70),
                filled: true,
                fillColor: Colors.white.withOpacity(0.2),
                hintStyle: const TextStyle(color: Colors.white70),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide.none,
                ),
                contentPadding: const EdgeInsets.symmetric(vertical: 0),
              ),
              style: const TextStyle(color: Colors.white),
            ),
          ),

          // 필터 칩
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Row(
              children: [
                _FilterChip(
                  label: 'All',
                  selected: _careLevel.isEmpty,
                  onTap: () => setState(() { _careLevel = ''; _page = 1; }),
                ),
                const SizedBox(width: 8),
                for (final level in ['BEGINNER', 'INTERMEDIATE', 'EXPERT'])
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: _FilterChip(
                      label: level == 'BEGINNER' ? l10n.careLevelBeginner
                           : level == 'INTERMEDIATE' ? l10n.careLevelIntermediate
                           : l10n.careLevelExpert,
                      selected: _careLevel == level,
                      onTap: () => setState(() { _careLevel = level; _page = 1; }),
                    ),
                  ),
              ],
            ),
          ),

          // 목록
          Expanded(
            child: fishAsync.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(child: Text(l10n.commonError)),
              data: (result) => result.items.isEmpty
                ? Center(child: Text(l10n.commonNoResults))
                : GridView.builder(
                    padding: const EdgeInsets.all(12),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 2,
                      childAspectRatio: 0.78,
                      crossAxisSpacing: 10,
                      mainAxisSpacing: 10,
                    ),
                    itemCount: result.items.length,
                    itemBuilder: (context, index) => FishCard(fish: result.items[index]),
                  ),
            ),
          ),
        ],
      ),
    );
  }
}

class FishCard extends StatelessWidget {
  final FishListItem fish;
  const FishCard({super.key, required this.fish});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    Color careLevelColor = Colors.green;
    if (fish.careLevel == 'INTERMEDIATE') careLevelColor = Colors.orange;
    if (fish.careLevel == 'EXPERT') careLevelColor = Colors.red;

    return GestureDetector(
      onTap: () => context.push('/fish/${fish.id}'),
      child: Card(
        elevation: 2,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        clipBehavior: Clip.antiAlias,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 이미지
            Expanded(
              child: Stack(
                fit: StackFit.expand,
                children: [
                  fish.primaryImageUrl != null
                    ? CachedNetworkImage(
                        imageUrl: fish.primaryImageUrl!,
                        fit: BoxFit.cover,
                        placeholder: (_, __) => Container(color: const Color(0xFFE0F2FE)),
                        errorWidget: (_, __, ___) => const _FishPlaceholder(),
                      )
                    : const _FishPlaceholder(),
                  if (fish.careLevel != null)
                    Positioned(
                      top: 8, right: 8,
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                        decoration: BoxDecoration(
                          color: careLevelColor,
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Text(
                          fish.careLevel == 'BEGINNER' ? l10n.careLevelBeginner
                            : fish.careLevel == 'INTERMEDIATE' ? l10n.careLevelIntermediate
                            : l10n.careLevelExpert,
                          style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.w600),
                        ),
                      ),
                    ),
                ],
              ),
            ),

            // 정보
            Padding(
              padding: const EdgeInsets.all(10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    fish.commonName,
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Text(
                    fish.scientificName,
                    style: const TextStyle(fontSize: 11, color: Colors.grey, fontStyle: FontStyle.italic),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  if (fish.maxSizeCm != null || fish.minTankSizeLiters != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Row(
                        children: [
                          if (fish.maxSizeCm != null) ...[
                            const Icon(Icons.straighten, size: 12, color: Colors.grey),
                            const SizedBox(width: 2),
                            Text('${fish.maxSizeCm}cm', style: const TextStyle(fontSize: 10, color: Colors.grey)),
                            const SizedBox(width: 8),
                          ],
                          if (fish.minTankSizeLiters != null) ...[
                            const Icon(Icons.water, size: 12, color: Colors.blue),
                            const SizedBox(width: 2),
                            Text('${fish.minTankSizeLiters}L', style: const TextStyle(fontSize: 10, color: Colors.grey)),
                          ],
                        ],
                      ),
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

class _FishPlaceholder extends StatelessWidget {
  const _FishPlaceholder();
  @override
  Widget build(BuildContext context) => Container(
    color: const Color(0xFFE0F2FE),
    child: const Center(child: Text('🐠', style: TextStyle(fontSize: 40))),
  );
}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback onTap;
  const _FilterChip({required this.label, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) => GestureDetector(
    onTap: onTap,
    child: Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFF0EA5E9) : Colors.grey.shade100,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w600,
          color: selected ? Colors.white : Colors.grey.shade700,
        ),
      ),
    ),
  );
}
