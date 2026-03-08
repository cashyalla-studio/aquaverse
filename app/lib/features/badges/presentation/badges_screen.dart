import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

class Badge {
  final String id;
  final String emoji;
  final String name;
  final String? earnedAt;
  final bool isEarned;

  const Badge({
    required this.id,
    required this.emoji,
    required this.name,
    this.earnedAt,
    required this.isEarned,
  });

  factory Badge.fromJson(Map<String, dynamic> j) => Badge(
    id: j['id'] as String? ?? '',
    emoji: j['emoji'] as String? ?? '🏅',
    name: j['name'] as String? ?? '',
    earnedAt: j['earned_at'] as String?,
    isEarned: j['is_earned'] as bool? ?? false,
  );
}

class Challenge {
  final String id;
  final String title;
  final String description;
  final String deadline;
  final double progress;
  final bool isJoined;

  const Challenge({
    required this.id,
    required this.title,
    required this.description,
    required this.deadline,
    required this.progress,
    required this.isJoined,
  });

  factory Challenge.fromJson(Map<String, dynamic> j) => Challenge(
    id: j['id'] as String? ?? '',
    title: j['title'] as String? ?? '',
    description: j['description'] as String? ?? '',
    deadline: j['deadline'] as String? ?? '',
    progress: (j['progress'] as num?)?.toDouble() ?? 0.0,
    isJoined: j['is_joined'] as bool? ?? false,
  );
}

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

final badgesProvider = FutureProvider<List<Badge>>((ref) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/users/me/badges');
  final list =
      (resp.data as Map<String, dynamic>)['badges'] as List<dynamic>? ?? [];
  return list
      .map((e) => Badge.fromJson(e as Map<String, dynamic>))
      .toList();
});

final challengesProvider = FutureProvider<List<Challenge>>((ref) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/challenges');
  final list =
      (resp.data as Map<String, dynamic>)['challenges'] as List<dynamic>? ?? [];
  return list
      .map((e) => Challenge.fromJson(e as Map<String, dynamic>))
      .toList();
});

// ---------------------------------------------------------------------------
// Screen
// ---------------------------------------------------------------------------

class BadgesScreen extends ConsumerStatefulWidget {
  const BadgesScreen({super.key});

  @override
  ConsumerState<BadgesScreen> createState() => _BadgesScreenState();
}

class _BadgesScreenState extends ConsumerState<BadgesScreen> {
  // Local mutable list so we can update isJoined optimistically
  List<Challenge>? _challenges;
  bool _challengesLoaded = false;

  Future<void> _joinChallenge(String challengeId) async {
    if (_challenges == null) return;
    final idx = _challenges!.indexWhere((c) => c.id == challengeId);
    if (idx < 0) return;
    final original = _challenges![idx];

    setState(() {
      _challenges![idx] = Challenge(
        id: original.id,
        title: original.title,
        description: original.description,
        deadline: original.deadline,
        progress: original.progress,
        isJoined: true,
      );
    });

    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.post('/challenges/$challengeId/join');
    } catch (e) {
      if (mounted) {
        setState(() => _challenges![idx] = original);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('참가 실패: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final badgesAsync = ref.watch(badgesProvider);
    final challengesAsync = ref.watch(challengesProvider);

    challengesAsync.whenData((challenges) {
      if (!_challengesLoaded) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) {
            setState(() {
              _challenges = List<Challenge>.from(challenges);
              _challengesLoaded = true;
            });
          }
        });
      }
    });

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(badgesProvider);
          ref.invalidate(challengesProvider);
          setState(() => _challengesLoaded = false);
        },
        child: CustomScrollView(
          slivers: [
            // --- Gradient AppBar ---
            SliverAppBar(
              expandedHeight: 120,
              pinned: true,
              flexibleSpace: FlexibleSpaceBar(
                title: const Text(
                  '나의 뱃지',
                  style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                ),
                background: Container(
                  decoration: const BoxDecoration(
                    gradient: LinearGradient(
                      colors: [Color(0xFF7C3AED), Color(0xFF4F46E5)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                  ),
                ),
              ),
              backgroundColor: const Color(0xFF7C3AED),
              foregroundColor: Colors.white,
            ),

            // --- Badges section ---
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 20, 16, 4),
                child: const Text(
                  '획득한 뱃지',
                  style: TextStyle(fontSize: 17, fontWeight: FontWeight.bold),
                ),
              ),
            ),

            SliverToBoxAdapter(
              child: badgesAsync.when(
                loading: () => const Padding(
                  padding: EdgeInsets.all(32),
                  child: Center(child: CircularProgressIndicator()),
                ),
                error: (e, _) => _ErrorBanner(message: '뱃지 로드 실패: $e'),
                data: (badges) {
                  if (badges.isEmpty) {
                    return const _EmptyBadges();
                  }
                  return Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    child: Wrap(
                      spacing: 12,
                      runSpacing: 12,
                      children: badges
                          .map((b) => _BadgeCard(badge: b))
                          .toList(),
                    ),
                  );
                },
              ),
            ),

            // --- Challenges section ---
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 28, 16, 8),
                child: Row(
                  children: [
                    const Text(
                      '진행 중인 챌린지',
                      style:
                          TextStyle(fontSize: 17, fontWeight: FontWeight.bold),
                    ),
                    const Spacer(),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 10, vertical: 3),
                      decoration: BoxDecoration(
                        color: const Color(0xFF7C3AED).withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        _challenges != null
                            ? '${_challenges!.length}개'
                            : '',
                        style: const TextStyle(
                          color: Color(0xFF7C3AED),
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

            challengesAsync.when(
              loading: () => const SliverToBoxAdapter(
                child: Padding(
                  padding: EdgeInsets.all(32),
                  child: Center(child: CircularProgressIndicator()),
                ),
              ),
              error: (e, _) => SliverToBoxAdapter(
                child: _ErrorBanner(message: '챌린지 로드 실패: $e'),
              ),
              data: (_) {
                final challenges = _challenges;
                if (challenges == null || challenges.isEmpty) {
                  return const SliverToBoxAdapter(child: _EmptyChallenges());
                }
                return SliverList(
                  delegate: SliverChildBuilderDelegate(
                    (ctx, i) => Padding(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                      child: _ChallengeCard(
                        challenge: challenges[i],
                        onJoin: () => _joinChallenge(challenges[i].id),
                      ),
                    ),
                    childCount: challenges.length,
                  ),
                );
              },
            ),

            // Bottom padding
            const SliverToBoxAdapter(child: SizedBox(height: 24)),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Sub-widgets
// ---------------------------------------------------------------------------

class _BadgeCard extends StatelessWidget {
  final Badge badge;
  const _BadgeCard({required this.badge});

  @override
  Widget build(BuildContext context) {
    final card = Card(
      shape:
          RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      elevation: badge.isEarned ? 3 : 0,
      color: badge.isEarned ? Colors.white : Colors.grey.shade100,
      child: SizedBox(
        width: 96,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 8),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(badge.emoji, style: const TextStyle(fontSize: 38)),
              const SizedBox(height: 6),
              Text(
                badge.name,
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: badge.isEarned ? Colors.black87 : Colors.grey,
                ),
                textAlign: TextAlign.center,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
              if (badge.isEarned && badge.earnedAt != null) ...[
                const SizedBox(height: 3),
                Text(
                  _formatDate(badge.earnedAt!),
                  style: const TextStyle(fontSize: 10, color: Colors.grey),
                  textAlign: TextAlign.center,
                ),
              ],
            ],
          ),
        ),
      ),
    );

    if (badge.isEarned) return card;

    // Greyscale for unearned badges
    return ColorFiltered(
      colorFilter: const ColorFilter.matrix([
        0.2126, 0.7152, 0.0722, 0, 0,
        0.2126, 0.7152, 0.0722, 0, 0,
        0.2126, 0.7152, 0.0722, 0, 0,
        0,      0,      0,      1, 0,
      ]),
      child: card,
    );
  }

  String _formatDate(String iso) {
    try {
      final dt = DateTime.parse(iso);
      return '${dt.year}.${dt.month.toString().padLeft(2, '0')}.${dt.day.toString().padLeft(2, '0')}';
    } catch (_) {
      return iso;
    }
  }
}

class _ChallengeCard extends StatelessWidget {
  final Challenge challenge;
  final VoidCallback onJoin;
  const _ChallengeCard({required this.challenge, required this.onJoin});

  @override
  Widget build(BuildContext context) {
    final progressPct = (challenge.progress * 100).toInt();

    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    challenge.title,
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                if (challenge.isJoined)
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 10, vertical: 3),
                    decoration: BoxDecoration(
                      color: const Color(0xFF10B981).withOpacity(0.12),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Text(
                      '참가 중',
                      style: TextStyle(
                        color: Color(0xFF10B981),
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 6),
            Text(
              challenge.description,
              style: const TextStyle(color: Color(0xFF6B7280), fontSize: 13),
            ),
            const SizedBox(height: 10),
            // Progress bar
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: challenge.progress.clamp(0.0, 1.0),
                backgroundColor: Colors.grey.shade200,
                valueColor: const AlwaysStoppedAnimation<Color>(
                  Color(0xFF7C3AED),
                ),
                minHeight: 8,
              ),
            ),
            const SizedBox(height: 4),
            Row(
              children: [
                Text(
                  '$progressPct% 달성',
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF7C3AED),
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const Spacer(),
                if (challenge.deadline.isNotEmpty)
                  Text(
                    '마감: ${challenge.deadline}',
                    style: const TextStyle(
                      fontSize: 11,
                      color: Colors.grey,
                    ),
                  ),
              ],
            ),
            if (!challenge.isJoined) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: onJoin,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF7C3AED),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                  child: const Text('참가하기'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _EmptyBadges extends StatelessWidget {
  const _EmptyBadges();

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.symmetric(vertical: 24, horizontal: 16),
      child: Column(
        children: [
          Text('🏅', style: TextStyle(fontSize: 48)),
          SizedBox(height: 10),
          Text(
            '아직 획득한 뱃지가 없습니다',
            style: TextStyle(color: Colors.grey, fontSize: 15),
          ),
          SizedBox(height: 4),
          Text(
            '챌린지에 참가하고 첫 뱃지를 받아보세요!',
            style: TextStyle(color: Colors.grey, fontSize: 13),
          ),
        ],
      ),
    );
  }
}

class _EmptyChallenges extends StatelessWidget {
  const _EmptyChallenges();

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.symmetric(vertical: 24, horizontal: 16),
      child: Column(
        children: [
          Text('🎯', style: TextStyle(fontSize: 48)),
          SizedBox(height: 10),
          Text(
            '진행 중인 챌린지가 없습니다',
            style: TextStyle(color: Colors.grey, fontSize: 15),
          ),
        ],
      ),
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  final String message;
  const _ErrorBanner({required this.message});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.red.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.red.shade200),
      ),
      child: Row(
        children: [
          const Icon(Icons.error_outline, color: Colors.red),
          const SizedBox(width: 8),
          Expanded(
            child: Text(message, style: const TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
  }
}
