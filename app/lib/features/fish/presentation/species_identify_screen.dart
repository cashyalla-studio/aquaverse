import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../../../core/api/api_client.dart';

class SpeciesIdentifyScreen extends ConsumerStatefulWidget {
  const SpeciesIdentifyScreen({super.key});

  @override
  ConsumerState<SpeciesIdentifyScreen> createState() => _SpeciesIdentifyScreenState();
}

class _SpeciesIdentifyScreenState extends ConsumerState<SpeciesIdentifyScreen> {
  File? _image;
  List<Map<String, dynamic>> _candidates = [];
  bool _loading = false;
  String _error = '';
  int _processingMs = 0;

  Future<void> _pickImage(ImageSource source) async {
    final picker = ImagePicker();
    final picked = await picker.pickImage(source: source, maxWidth: 1280, imageQuality: 85);
    if (picked == null) return;
    setState(() {
      _image = File(picked.path);
      _candidates = [];
      _error = '';
    });
  }

  Future<void> _identify() async {
    if (_image == null) return;
    setState(() { _loading = true; _error = ''; });
    try {
      final bytes = await _image!.readAsBytes();
      final base64Image = base64Encode(bytes);
      final ext = _image!.path.split('.').last.toLowerCase();
      final mediaType = ext == 'png' ? 'image/png' : 'image/jpeg';

      final dio = ref.read(dioProvider('ko'));
      final r = await dio.post('/species/identify', data: {
        'image_base64': base64Image,
        'media_type': mediaType,
      });
      if (mounted) {
        setState(() {
          _candidates = List<Map<String, dynamic>>.from(r.data['candidates'] ?? []);
          _processingMs = r.data['processing_ms'] ?? 0;
        });
      }
    } catch (e) {
      if (mounted) setState(() => _error = '식별에 실패했습니다. 다시 시도해주세요.');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Color _confidenceColor(double c) {
    if (c >= 0.8) return Colors.green;
    if (c >= 0.5) return Colors.orange;
    return Colors.grey;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('AI 어종 식별')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            // 이미지 선택 영역
            GestureDetector(
              onTap: () => _showPickerMenu(),
              child: Container(
                width: double.infinity,
                height: 220,
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.blue.shade200, width: 2),
                  borderRadius: BorderRadius.circular(16),
                  color: Colors.blue.shade50,
                ),
                child: _image != null
                    ? ClipRRect(
                        borderRadius: BorderRadius.circular(14),
                        child: Image.file(_image!, fit: BoxFit.contain),
                      )
                    : const Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text('🐠', style: TextStyle(fontSize: 48)),
                          SizedBox(height: 8),
                          Text('탭하여 사진 선택', style: TextStyle(color: Colors.blue)),
                          Text('카메라 또는 갤러리', style: TextStyle(color: Colors.grey, fontSize: 12)),
                        ],
                      ),
              ),
            ),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _image == null || _loading ? null : _identify,
                icon: _loading
                    ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                    : const Icon(Icons.search),
                label: Text(_loading ? 'AI 분석 중...' : 'AI로 어종 식별'),
              ),
            ),
            if (_error.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(_error, style: const TextStyle(color: Colors.red)),
            ],
            if (_candidates.isNotEmpty) ...[
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('식별 결과', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                  Text('${_processingMs}ms', style: const TextStyle(color: Colors.grey, fontSize: 12)),
                ],
              ),
              const SizedBox(height: 8),
              ..._candidates.asMap().entries.map((entry) {
                final i = entry.key;
                final c = entry.value;
                final confidence = (c['confidence'] as num?)?.toDouble() ?? 0.0;
                return Card(
                  margin: const EdgeInsets.only(bottom: 8),
                  color: i == 0 ? Colors.blue.shade50 : null,
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                if (i == 0)
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                    decoration: BoxDecoration(color: Colors.blue, borderRadius: BorderRadius.circular(4)),
                                    child: const Text('최유력', style: TextStyle(color: Colors.white, fontSize: 11)),
                                  ),
                                Text(c['name'] ?? '', style: const TextStyle(fontWeight: FontWeight.bold)),
                                Text(c['scientific_name'] ?? '', style: const TextStyle(color: Colors.grey, fontSize: 12, fontStyle: FontStyle.italic)),
                              ],
                            ),
                            Text(
                              '${(confidence * 100).round()}%',
                              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: _confidenceColor(confidence)),
                            ),
                          ],
                        ),
                        const SizedBox(height: 4),
                        Text(c['description'] ?? '', style: const TextStyle(fontSize: 13, color: Colors.black87)),
                      ],
                    ),
                  ),
                );
              }),
            ],
          ],
        ),
      ),
    );
  }

  void _showPickerMenu() {
    showModalBottomSheet(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.camera_alt),
              title: const Text('카메라'),
              onTap: () { Navigator.pop(context); _pickImage(ImageSource.camera); },
            ),
            ListTile(
              leading: const Icon(Icons.photo_library),
              title: const Text('갤러리'),
              onTap: () { Navigator.pop(context); _pickImage(ImageSource.gallery); },
            ),
          ],
        ),
      ),
    );
  }
}
