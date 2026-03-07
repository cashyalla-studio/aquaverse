import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

class WaterParamsScreen extends ConsumerStatefulWidget {
  final int tankId;
  const WaterParamsScreen({super.key, required this.tankId});

  @override
  ConsumerState<WaterParamsScreen> createState() => _WaterParamsScreenState();
}

class _WaterParamsScreenState extends ConsumerState<WaterParamsScreen> {
  final _tempC = TextEditingController();
  final _ph = TextEditingController();
  final _ammonia = TextEditingController();
  final _nitrite = TextEditingController();
  final _nitrate = TextEditingController();
  bool _loading = false;

  double? _parseDouble(String s) => s.isEmpty ? null : double.tryParse(s);

  Future<void> _submit() async {
    setState(() => _loading = true);
    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.post('/tanks/${widget.tankId}/water-params', data: {
        if (_parseDouble(_tempC.text) != null) 'temp_c': _parseDouble(_tempC.text),
        if (_parseDouble(_ph.text) != null) 'ph': _parseDouble(_ph.text),
        if (_parseDouble(_ammonia.text) != null) 'ammonia_ppm': _parseDouble(_ammonia.text),
        if (_parseDouble(_nitrite.text) != null) 'nitrite_ppm': _parseDouble(_nitrite.text),
        if (_parseDouble(_nitrate.text) != null) 'nitrate_ppm': _parseDouble(_nitrate.text),
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('수질이 기록되었습니다')));
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('오류: $e'), backgroundColor: Colors.red));
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  void dispose() {
    _tempC.dispose();
    _ph.dispose();
    _ammonia.dispose();
    _nitrite.dispose();
    _nitrate.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('수질 기록')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _field(_tempC, '수온 (°C)', '예: 26.0'),
          _field(_ph, 'pH', '예: 7.2'),
          _field(_ammonia, '암모니아 (ppm)', '예: 0.0'),
          _field(_nitrite, '아질산 (ppm)', '예: 0.0'),
          _field(_nitrate, '질산 (ppm)', '예: 10.0'),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: _loading ? null : _submit,
            child: _loading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('저장'),
          ),
        ],
      ),
    );
  }

  Widget _field(TextEditingController ctrl, String label, String hint) =>
      Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: TextField(
          controller: ctrl,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          decoration: InputDecoration(
            labelText: label,
            hintText: hint,
            border: const OutlineInputBorder(),
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
        ),
      );
}
