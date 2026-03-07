import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:geolocator/geolocator.dart';
import '../../../core/api/api_client.dart';

class NearbyBusiness {
  final int id;
  final String storeName, city, address, phone;
  final bool isVerified;
  final double avgRating;
  final double? distanceKm;

  NearbyBusiness({required this.id, required this.storeName, required this.city,
    required this.address, required this.phone, required this.isVerified,
    required this.avgRating, this.distanceKm});

  factory NearbyBusiness.fromJson(Map<String, dynamic> j) => NearbyBusiness(
    id: j['id'] as int,
    storeName: j['store_name'] as String? ?? '',
    city: j['city'] as String? ?? '',
    address: j['address'] as String? ?? '',
    phone: j['phone'] as String? ?? '',
    isVerified: j['is_verified'] as bool? ?? false,
    avgRating: (j['avg_rating'] as num?)?.toDouble() ?? 0,
    distanceKm: (j['distance_km'] as num?)?.toDouble(),
  );
}

class MapScreen extends ConsumerStatefulWidget {
  const MapScreen({super.key});

  @override
  ConsumerState<MapScreen> createState() => _MapScreenState();
}

class _MapScreenState extends ConsumerState<MapScreen> {
  List<NearbyBusiness> _businesses = [];
  bool _loading = false;
  String _error = '';
  double _radius = 5.0;
  Position? _position;

  Future<void> _locate() async {
    setState(() { _loading = true; _error = ''; });
    try {
      LocationPermission perm = await Geolocator.checkPermission();
      if (perm == LocationPermission.denied) {
        perm = await Geolocator.requestPermission();
      }
      if (perm == LocationPermission.deniedForever) {
        setState(() { _error = '위치 권한이 거부됐습니다. 설정에서 변경하세요.'; _loading = false; });
        return;
      }
      final pos = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.medium);
      _position = pos;
      await _search();
    } catch (e) {
      setState(() { _error = '위치 오류: $e'; _loading = false; });
    }
  }

  Future<void> _search() async {
    if (_position == null) return;
    setState(() { _loading = true; _error = ''; });
    try {
      final dio = ref.read(dioProvider('ko'));
      final resp = await dio.get('/businesses/nearby', queryParameters: {
        'lat': _position!.latitude, 'lng': _position!.longitude,
        'radius': _radius, 'limit': 20,
      });
      final list = (resp.data as Map<String, dynamic>)['businesses'] as List<dynamic>? ?? [];
      setState(() {
        _businesses = list.map((e) => NearbyBusiness.fromJson(e as Map<String, dynamic>)).toList();
      });
    } catch (e) {
      setState(() { _error = '검색 오류: $e'; });
    } finally {
      setState(() { _loading = false; });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('주변 수족관 찾기')),
      body: Column(children: [
        // 반경 선택 + 검색 버튼
        Container(
          padding: const EdgeInsets.all(12),
          color: Colors.white,
          child: Column(children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [3.0, 5.0, 10.0, 20.0].map((r) => Padding(
                padding: const EdgeInsets.symmetric(horizontal: 4),
                child: FilterChip(
                  label: Text('${r.toInt()}km'),
                  selected: _radius == r,
                  onSelected: (_) => setState(() { _radius = r; if (_position != null) _search(); }),
                ),
              )).toList(),
            ),
            const SizedBox(height: 8),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _loading ? null : _locate,
                icon: const Icon(Icons.my_location),
                label: Text(_position == null ? '내 위치로 검색' : '다시 검색 (${_radius.toInt()}km)'),
              ),
            ),
            if (_error.isNotEmpty)
              Padding(padding: const EdgeInsets.only(top: 8),
                child: Text(_error, style: const TextStyle(color: Colors.red, fontSize: 12))),
          ]),
        ),

        // 결과 목록
        Expanded(
          child: _loading
              ? const Center(child: CircularProgressIndicator())
              : _businesses.isEmpty
                  ? Center(child: Text(_position == null
                      ? '위치 검색 버튼을 눌러 주변 업체를 찾으세요'
                      : '${_radius.toInt()}km 반경 내 등록된 업체가 없습니다'))
                  : ListView.builder(
                      padding: const EdgeInsets.all(12),
                      itemCount: _businesses.length,
                      itemBuilder: (context, i) {
                        final b = _businesses[i];
                        return Card(
                          margin: const EdgeInsets.only(bottom: 8),
                          child: ListTile(
                            leading: CircleAvatar(
                              backgroundColor: Colors.blue.shade50,
                              child: const Text('🐠'),
                            ),
                            title: Row(children: [
                              Expanded(child: Text(b.storeName, style: const TextStyle(fontWeight: FontWeight.w600))),
                              if (b.distanceKm != null)
                                Text('${b.distanceKm!.toStringAsFixed(1)}km',
                                    style: const TextStyle(fontSize: 12, color: Colors.grey)),
                            ]),
                            subtitle: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                              if (b.address.isNotEmpty) Text(b.address, style: const TextStyle(fontSize: 12)),
                              Row(children: [
                                Text('★' * b.avgRating.round(), style: const TextStyle(color: Colors.amber, fontSize: 12)),
                                if (b.isVerified) ...[
                                  const SizedBox(width: 4),
                                  Container(padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                                    decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(4)),
                                    child: Text('인증', style: TextStyle(fontSize: 9, color: Colors.blue.shade700))),
                                ],
                              ]),
                            ]),
                          ),
                        );
                      },
                    ),
        ),
      ]),
    );
  }
}
