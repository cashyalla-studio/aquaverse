import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

class CareStreak {
  final int currentStreak;
  final int longestStreak;
  final String lastCheckInDate;

  const CareStreak({
    required this.currentStreak,
    required this.longestStreak,
    required this.lastCheckInDate,
  });

  factory CareStreak.fromJson(Map<String, dynamic> j) => CareStreak(
    currentStreak: j['current_streak'] as int? ?? 0,
    longestStreak: j['longest_streak'] as int? ?? 0,
    lastCheckInDate: j['last_check_in_date'] as String? ?? '',
  );
}

class CareTask {
  final int id;
  final String type;
  final String title;
  final String dueTime;
  final bool isDone;

  const CareTask({
    required this.id,
    required this.type,
    required this.title,
    required this.dueTime,
    required this.isDone,
  });

  factory CareTask.fromJson(Map<String, dynamic> j) => CareTask(
    id: j['id'] as int? ?? 0,
    type: j['type'] as String? ?? 'other',
    title: j['title'] as String? ?? '',
    dueTime: j['due_time'] as String? ?? '',
    isDone: j['is_done'] as bool? ?? false,
  );

  CareTask copyWith({bool? isDone}) => CareTask(
    id: id,
    type: type,
    title: title,
    dueTime: dueTime,
    isDone: isDone ?? this.isDone,
  );
}

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

final careStreakProvider = FutureProvider<CareStreak>((ref) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/users/me/streak');
  return CareStreak.fromJson(resp.data as Map<String, dynamic>);
});

final careTodayProvider =
    FutureProvider<List<CareTask>>((ref) async {
  final dio = ref.read(dioProvider('ko'));
  final resp = await dio.get('/users/me/care-today');
  final list =
      (resp.data as Map<String, dynamic>)['tasks'] as List<dynamic>? ?? [];
  return list
      .map((e) => CareTask.fromJson(e as Map<String, dynamic>))
      .toList();
});

// ---------------------------------------------------------------------------
// Screen
// ---------------------------------------------------------------------------

class CareHubScreen extends ConsumerStatefulWidget {
  const CareHubScreen({super.key});

  @override
  ConsumerState<CareHubScreen> createState() => _CareHubScreenState();
}

class _CareHubScreenState extends ConsumerState<CareHubScreen> {
  // Local mutable copy of tasks so we can toggle done state immediately
  List<CareTask>? _tasks;
  bool _tasksLoaded = false;

  @override
  void initState() {
    super.initState();
    // Tasks will be loaded via the provider; we mirror them locally once loaded.
  }

  IconData _taskIcon(String type) {
    return switch (type) {
      'feeding' => Icons.restaurant_outlined,
      'water_change' => Icons.water_drop_outlined,
      'filter' => Icons.filter_alt_outlined,
      'medicine' => Icons.medication_outlined,
      'cleaning' => Icons.cleaning_services_outlined,
      _ => Icons.task_alt_outlined,
    };
  }

  Color _taskColor(String type) {
    return switch (type) {
      'feeding' => const Color(0xFFF97316),
      'water_change' => const Color(0xFF0EA5E9),
      'filter' => const Color(0xFF8B5CF6),
      'medicine' => const Color(0xFFEF4444),
      'cleaning' => const Color(0xFF10B981),
      _ => const Color(0xFF6B7280),
    };
  }

  Future<void> _markDone(int index) async {
    if (_tasks == null) return;
    final task = _tasks![index];
    if (task.isDone) return;

    HapticFeedback.lightImpact();

    setState(() {
      _tasks![index] = task.copyWith(isDone: true);
    });

    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.post('/users/me/care-tasks/${task.id}/complete');
    } catch (e) {
      // Revert on failure
      if (mounted) {
        setState(() => _tasks![index] = task.copyWith(isDone: false));
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('완료 처리 실패: $e')),
        );
      }
    }
  }

  void _showAddTaskSheet() {
    final titleController = TextEditingController();
    String selectedType = 'feeding';
    DateTime selectedDate = DateTime.now();

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(
          left: 24,
          right: 24,
          top: 24,
          bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
        ),
        child: StatefulBuilder(
          builder: (ctx, setSheetState) => Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                '케어 일정 추가',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                value: selectedType,
                decoration: const InputDecoration(
                  labelText: '종류',
                  border: OutlineInputBorder(),
                ),
                items: const [
                  DropdownMenuItem(value: 'feeding', child: Text('먹이주기')),
                  DropdownMenuItem(value: 'water_change', child: Text('환수')),
                  DropdownMenuItem(value: 'filter', child: Text('여과기 청소')),
                  DropdownMenuItem(value: 'medicine', child: Text('약품 투여')),
                  DropdownMenuItem(value: 'cleaning', child: Text('수조 청소')),
                  DropdownMenuItem(value: 'other', child: Text('기타')),
                ],
                onChanged: (v) => setSheetState(() => selectedType = v ?? 'feeding'),
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: titleController,
                decoration: const InputDecoration(
                  labelText: '제목',
                  hintText: '예: 아침 먹이주기',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              InkWell(
                onTap: () async {
                  final picked = await showDatePicker(
                    context: ctx,
                    initialDate: selectedDate,
                    firstDate: DateTime.now(),
                    lastDate: DateTime.now().add(const Duration(days: 365)),
                  );
                  if (picked != null) {
                    setSheetState(() => selectedDate = picked);
                  }
                },
                child: InputDecorator(
                  decoration: const InputDecoration(
                    labelText: '날짜',
                    border: OutlineInputBorder(),
                    suffixIcon: Icon(Icons.calendar_today_outlined),
                  ),
                  child: Text(
                    '${selectedDate.year}-${selectedDate.month.toString().padLeft(2, '0')}-${selectedDate.day.toString().padLeft(2, '0')}',
                  ),
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () async {
                    final title = titleController.text.trim();
                    if (title.isEmpty) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('제목을 입력해주세요')),
                      );
                      return;
                    }
                    try {
                      final dio = ref.read(dioProvider('ko'));
                      await dio.post('/users/me/care-tasks', data: {
                        'type': selectedType,
                        'title': title,
                        'due_date': selectedDate.toIso8601String().split('T').first,
                      });
                      if (ctx.mounted) Navigator.pop(ctx);
                      ref.invalidate(careTodayProvider);
                      setState(() => _tasksLoaded = false);
                    } catch (e) {
                      if (ctx.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text('추가 실패: $e')),
                        );
                      }
                    }
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF10B981),
                    foregroundColor: Colors.white,
                    minimumSize: const Size.fromHeight(48),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: const Text(
                    '추가',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final streakAsync = ref.watch(careStreakProvider);
    final tasksAsync = ref.watch(careTodayProvider);

    // Mirror loaded tasks into local state once
    tasksAsync.whenData((tasks) {
      if (!_tasksLoaded) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) {
            setState(() {
              _tasks = List<CareTask>.from(tasks);
              _tasksLoaded = true;
            });
          }
        });
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: const Text('Life Care Hub'),
        backgroundColor: const Color(0xFFEA580C),
        foregroundColor: Colors.white,
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showAddTaskSheet,
        backgroundColor: const Color(0xFFEA580C),
        foregroundColor: Colors.white,
        child: const Icon(Icons.add),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(careStreakProvider);
          ref.invalidate(careTodayProvider);
          setState(() => _tasksLoaded = false);
        },
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // --- Streak banner ---
            streakAsync.when(
              loading: () => const _StreakBannerSkeleton(),
              error: (e, _) => _ErrorBanner(message: '스트릭 로드 실패: $e'),
              data: (streak) => _StreakBanner(streak: streak),
            ),
            const SizedBox(height: 20),

            // --- Today tasks ---
            Row(
              children: [
                const Text(
                  '오늘 할 일',
                  style: TextStyle(fontSize: 17, fontWeight: FontWeight.bold),
                ),
                const Spacer(),
                if (_tasks != null)
                  Text(
                    '${_tasks!.where((t) => t.isDone).length}/${_tasks!.length} 완료',
                    style: const TextStyle(color: Colors.grey, fontSize: 13),
                  ),
              ],
            ),
            const SizedBox(height: 10),

            tasksAsync.when(
              loading: () => const Center(
                child: Padding(
                  padding: EdgeInsets.all(32),
                  child: CircularProgressIndicator(),
                ),
              ),
              error: (e, _) => _ErrorBanner(message: '할 일 로드 실패: $e'),
              data: (_) {
                final tasks = _tasks;
                if (tasks == null || tasks.isEmpty) {
                  return const _EmptyTasks();
                }
                return ListView.separated(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  itemCount: tasks.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 8),
                  itemBuilder: (context, i) => _TaskCard(
                    task: tasks[i],
                    taskColor: _taskColor(tasks[i].type),
                    taskIcon: _taskIcon(tasks[i].type),
                    onComplete: () => _markDone(i),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Sub-widgets
// ---------------------------------------------------------------------------

class _StreakBanner extends StatelessWidget {
  final CareStreak streak;
  const _StreakBanner({required this.streak});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFFEA580C), Color(0xFFDC2626)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFFEA580C).withOpacity(0.35),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                '나의 케어 스트릭',
                style: TextStyle(color: Colors.white70, fontSize: 13),
              ),
              const SizedBox(height: 4),
              Row(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  const Text('🔥', style: TextStyle(fontSize: 36)),
                  const SizedBox(width: 6),
                  Text(
                    '${streak.currentStreak}일 연속',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                '최장 기록: ${streak.longestStreak}일',
                style: const TextStyle(color: Colors.white70, fontSize: 13),
              ),
            ],
          ),
          const Spacer(),
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.15),
              shape: BoxShape.circle,
            ),
            child: const Icon(Icons.local_fire_department,
                color: Colors.white, size: 36),
          ),
        ],
      ),
    );
  }
}

class _StreakBannerSkeleton extends StatelessWidget {
  const _StreakBannerSkeleton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 110,
      decoration: BoxDecoration(
        color: Colors.orange.shade100,
        borderRadius: BorderRadius.circular(20),
      ),
      child: const Center(child: CircularProgressIndicator()),
    );
  }
}

class _TaskCard extends StatelessWidget {
  final CareTask task;
  final Color taskColor;
  final IconData taskIcon;
  final VoidCallback onComplete;

  const _TaskCard({
    required this.task,
    required this.taskColor,
    required this.taskIcon,
    required this.onComplete,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: task.isDone ? 0 : 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
      color: task.isDone ? Colors.grey.shade50 : Colors.white,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        child: Row(
          children: [
            CircleAvatar(
              backgroundColor: task.isDone
                  ? Colors.grey.shade200
                  : taskColor.withOpacity(0.12),
              child: Icon(
                taskIcon,
                color: task.isDone ? Colors.grey : taskColor,
                size: 20,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    task.title,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 15,
                      decoration: task.isDone
                          ? TextDecoration.lineThrough
                          : TextDecoration.none,
                      color: task.isDone ? Colors.grey : Colors.black87,
                    ),
                  ),
                  if (task.dueTime.isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(
                      task.dueTime,
                      style: const TextStyle(color: Colors.grey, fontSize: 12),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(width: 8),
            task.isDone
                ? const Icon(Icons.check_circle, color: Color(0xFF10B981), size: 28)
                : ElevatedButton(
                    onPressed: onComplete,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF10B981),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(
                          horizontal: 14, vertical: 8),
                      minimumSize: Size.zero,
                      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                    child: const Text('완료', style: TextStyle(fontSize: 13)),
                  ),
          ],
        ),
      ),
    );
  }
}

class _EmptyTasks extends StatelessWidget {
  const _EmptyTasks();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(32),
      alignment: Alignment.center,
      child: const Column(
        children: [
          Text('🐠', style: TextStyle(fontSize: 48)),
          SizedBox(height: 12),
          Text(
            '오늘 케어 할 일이 없습니다',
            style: TextStyle(color: Colors.grey, fontSize: 15),
          ),
          SizedBox(height: 4),
          Text(
            'FAB을 눌러 일정을 추가하세요',
            style: TextStyle(color: Colors.grey, fontSize: 13),
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
