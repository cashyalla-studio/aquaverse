import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';
import '../domain/listing_model.dart';

class ListingRepository {
  final Dio _dio;
  ListingRepository(this._dio);

  Future<ListingDetail> getListing(int id) async {
    final resp = await _dio.get('/listings/$id');
    return ListingDetail.fromJson(resp.data);
  }

  Future<int> createListing(CreateListingRequest req) async {
    final resp = await _dio.post('/listings', data: req.toJson());
    return resp.data['id'];
  }

  Future<void> watchFish(int fishDataId, {String? location, double? lat, double? lng}) async {
    await _dio.post('/watches', data: {
      'fish_data_id': fishDataId,
      if (location != null) 'location': location,
      if (lat != null) 'latitude': lat,
      if (lng != null) 'longitude': lng,
    });
  }

  Future<void> reportFraud(int listingId, String reason) async {
    await _dio.post('/fraud-reports', data: {
      'listing_id': listingId,
      'reason': reason,
    });
  }

  Future<int?> initiateTrade(int listingId) async {
    final resp = await _dio.post('/trades', data: {'listing_id': listingId});
    return resp.data?['id'] as int?;
  }
}

final listingRepositoryProvider = Provider.family<ListingRepository, String>(
  (ref, locale) => ListingRepository(ref.watch(dioProvider(locale))),
);
