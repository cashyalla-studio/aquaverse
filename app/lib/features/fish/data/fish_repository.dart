import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../domain/fish_model.dart';

class FishRepository {
  final Dio _dio;
  FishRepository(this._dio);

  Future<FishListResult> listFish({
    String? query,
    String? family,
    String? careLevel,
    int page = 1,
    int limit = 24,
  }) async {
    final resp = await _dio.get('/fish', queryParameters: {
      if (query != null && query.isNotEmpty) 'q': query,
      if (family != null && family.isNotEmpty) 'family': family,
      if (careLevel != null && careLevel.isNotEmpty) 'care_level': careLevel,
      'page': page,
      'limit': limit,
    });
    return FishListResult.fromJson(resp.data);
  }

  Future<FishDetail> getFish(int id) async {
    final resp = await _dio.get('/fish/$id');
    return FishDetail.fromJson(resp.data);
  }

  Future<List<FishListItem>> searchFish(String query) async {
    final resp = await _dio.get('/fish/search', queryParameters: {'q': query});
    return (resp.data as List).map((e) => FishListItem.fromJson(e)).toList();
  }

  Future<List<String>> listFamilies() async {
    final resp = await _dio.get('/fish/families');
    return List<String>.from(resp.data);
  }
}

// Providers
final fishRepositoryProvider = Provider.family<FishRepository, String>((ref, locale) {
  final dio = ref.watch(dioProvider(locale));
  return FishRepository(dio);
});
