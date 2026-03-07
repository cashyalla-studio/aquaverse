import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../l10n/app_localizations.dart';

class TanksScreen extends ConsumerWidget {
  const TanksScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.navMyTanks),
        backgroundColor: const Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showAddTankDialog(context),
        backgroundColor: const Color(0xFF0EA5E9),
        icon: const Icon(Icons.add),
        label: const Text('Add Tank'),
      ),
      body: const _TanksContent(),
    );
  }

  void _showAddTankDialog(BuildContext context) {
    final nameController = TextEditingController();
    String type = 'FRESHWATER';
    double? volumeL;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(
          left: 24, right: 24, top: 24,
          bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
        ),
        child: StatefulBuilder(
          builder: (ctx, setState) => Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Add New Tank', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
              const SizedBox(height: 16),
              TextField(
                controller: nameController,
                decoration: const InputDecoration(
                  labelText: 'Tank Name',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<String>(
                value: type,
                decoration: const InputDecoration(labelText: 'Type', border: OutlineInputBorder()),
                items: const [
                  DropdownMenuItem(value: 'FRESHWATER', child: Text('🪣 Freshwater')),
                  DropdownMenuItem(value: 'SALTWATER', child: Text('🌊 Saltwater')),
                  DropdownMenuItem(value: 'BRACKISH', child: Text('🏖 Brackish')),
                  DropdownMenuItem(value: 'PLANTED', child: Text('🌿 Planted')),
                ],
                onChanged: (v) => setState(() => type = v ?? 'FRESHWATER'),
              ),
              const SizedBox(height: 12),
              TextField(
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
                decoration: const InputDecoration(
                  labelText: 'Volume (Liters)',
                  border: OutlineInputBorder(),
                  suffixText: 'L',
                ),
                onChanged: (v) => volumeL = double.tryParse(v),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () {
                    // TODO: API 연동
                    Navigator.pop(ctx);
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF0EA5E9),
                    foregroundColor: Colors.white,
                    minimumSize: const Size.fromHeight(48),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  child: const Text('Add Tank', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _TanksContent extends StatelessWidget {
  const _TanksContent();

  // 더미 데이터 (실제: API 연동 예정)
  static const _dummyTanks = [
    _TankData(
      name: 'South American Community',
      type: 'FRESHWATER',
      volumeL: 120,
      fish: ['Discus x2', 'Cardinal Tetra x20', 'Corydoras x6'],
      params: {'pH': '6.8', 'Temp': '28°C', 'Ammonia': '0'},
    ),
    _TankData(
      name: 'Planted Nano',
      type: 'PLANTED',
      volumeL: 30,
      fish: ['Neon Tetra x10', 'Amano Shrimp x15'],
      params: {'pH': '7.0', 'Temp': '25°C', 'CO2': 'Injected'},
    ),
  ];

  @override
  Widget build(BuildContext context) {
    if (_dummyTanks.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.water, size: 80, color: Color(0xFFBAE6FD)),
            const SizedBox(height: 16),
            const Text('No tanks yet', style: TextStyle(fontSize: 16, color: Color(0xFF6B7280))),
            const SizedBox(height: 8),
            const Text('Add your first tank to track fish and parameters.', style: TextStyle(fontSize: 13, color: Color(0xFF9CA3AF))),
          ],
        ),
      );
    }

    return ListView.separated(
      padding: const EdgeInsets.all(12),
      itemCount: _dummyTanks.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (context, i) => _TankCard(tank: _dummyTanks[i]),
    );
  }
}

class _TankData {
  final String name;
  final String type;
  final double volumeL;
  final List<String> fish;
  final Map<String, String> params;
  const _TankData({
    required this.name, required this.type, required this.volumeL,
    required this.fish, required this.params,
  });
}

class _TankCard extends StatelessWidget {
  final _TankData tank;
  const _TankCard({required this.tank});

  String get _typeEmoji => switch (tank.type) {
    'SALTWATER' => '🌊',
    'BRACKISH' => '🏖',
    'PLANTED' => '🌿',
    _ => '🪣',
  };

  Color get _typeColor => switch (tank.type) {
    'SALTWATER' => const Color(0xFF2563EB),
    'BRACKISH' => const Color(0xFF92400E),
    'PLANTED' => const Color(0xFF16A34A),
    _ => const Color(0xFF0EA5E9),
  };

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 8, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 헤더
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: _typeColor.withOpacity(0.08),
              borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
            ),
            child: Row(
              children: [
                Text(_typeEmoji, style: const TextStyle(fontSize: 28)),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(tank.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
                      Text('${tank.volumeL.toInt()}L • ${tank.type[0] + tank.type.substring(1).toLowerCase()}',
                          style: TextStyle(fontSize: 12, color: _typeColor)),
                    ],
                  ),
                ),
                IconButton(icon: const Icon(Icons.more_vert, size: 20), onPressed: () {}),
              ],
            ),
          ),

          // 파라미터
          Padding(
            padding: const EdgeInsets.fromLTRB(14, 10, 14, 0),
            child: Wrap(
              spacing: 8, runSpacing: 6,
              children: tank.params.entries.map((e) => _ParamChip(label: e.key, value: e.value)).toList(),
            ),
          ),

          // 어종 목록
          Padding(
            padding: const EdgeInsets.all(14),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Inhabitants (${tank.fish.length})',
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Color(0xFF6B7280))),
                const SizedBox(height: 6),
                ...tank.fish.map((f) => Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Row(
                    children: [
                      const Text('🐠', style: TextStyle(fontSize: 14)),
                      const SizedBox(width: 6),
                      Text(f, style: const TextStyle(fontSize: 13)),
                    ],
                  ),
                )),
                const SizedBox(height: 6),
                TextButton.icon(
                  onPressed: () {},
                  icon: const Icon(Icons.add, size: 16),
                  label: const Text('Add Fish', style: TextStyle(fontSize: 12)),
                  style: TextButton.styleFrom(foregroundColor: const Color(0xFF0EA5E9)),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ParamChip extends StatelessWidget {
  final String label;
  final String value;
  const _ParamChip({required this.label, required this.value});

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
    decoration: BoxDecoration(
      color: const Color(0xFFF3F4F6),
      borderRadius: BorderRadius.circular(20),
    ),
    child: RichText(
      text: TextSpan(
        children: [
          TextSpan(text: '$label: ', style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
          TextSpan(text: value, style: const TextStyle(fontSize: 11, color: Color(0xFF111827), fontWeight: FontWeight.w600)),
        ],
      ),
    ),
  );
}
