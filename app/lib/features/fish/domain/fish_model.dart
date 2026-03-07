class FishListItem {
  final int id;
  final String scientificName;
  final String commonName;
  final String family;
  final String? careLevel;
  final String? temperament;
  final double? maxSizeCm;
  final int? minTankSizeLiters;
  final String? primaryImageUrl;
  final double qualityScore;

  const FishListItem({
    required this.id,
    required this.scientificName,
    required this.commonName,
    required this.family,
    this.careLevel,
    this.temperament,
    this.maxSizeCm,
    this.minTankSizeLiters,
    this.primaryImageUrl,
    required this.qualityScore,
  });

  factory FishListItem.fromJson(Map<String, dynamic> json) => FishListItem(
    id: json['id'],
    scientificName: json['scientific_name'] ?? '',
    commonName: json['common_name'] ?? '',
    family: json['family'] ?? '',
    careLevel: json['care_level'],
    temperament: json['temperament'],
    maxSizeCm: (json['max_size_cm'] as num?)?.toDouble(),
    minTankSizeLiters: json['min_tank_size_liters'],
    primaryImageUrl: json['primary_image_url'],
    qualityScore: (json['quality_score'] as num?)?.toDouble() ?? 0,
  );
}

class FishListResult {
  final List<FishListItem> items;
  final int totalCount;
  final int page;
  final int limit;

  const FishListResult({
    required this.items,
    required this.totalCount,
    required this.page,
    required this.limit,
  });

  factory FishListResult.fromJson(Map<String, dynamic> json) => FishListResult(
    items: (json['items'] as List? ?? []).map((e) => FishListItem.fromJson(e)).toList(),
    totalCount: json['total_count'] ?? 0,
    page: json['page'] ?? 1,
    limit: json['limit'] ?? 24,
  );
}

class FishDetail {
  final int id;
  final String scientificName;
  final String primaryCommonName;
  final String? family;
  final String? careLevel;
  final String? temperament;
  final double? maxSizeCm;
  final double? lifespanYears;
  final double? phMin;
  final double? phMax;
  final double? tempMinC;
  final double? tempMaxC;
  final int? minTankSizeLiters;
  final String? dietType;
  final String? dietNotes;
  final String? breedingNotes;
  final String? careNotes;
  final String? primaryImageUrl;
  final String? license;
  final String? attribution;
  final Map<String, String?>? translation;

  const FishDetail({
    required this.id,
    required this.scientificName,
    required this.primaryCommonName,
    this.family,
    this.careLevel,
    this.temperament,
    this.maxSizeCm,
    this.lifespanYears,
    this.phMin,
    this.phMax,
    this.tempMinC,
    this.tempMaxC,
    this.minTankSizeLiters,
    this.dietType,
    this.dietNotes,
    this.breedingNotes,
    this.careNotes,
    this.primaryImageUrl,
    this.license,
    this.attribution,
    this.translation,
  });

  factory FishDetail.fromJson(Map<String, dynamic> json) => FishDetail(
    id: json['id'],
    scientificName: json['scientific_name'] ?? '',
    primaryCommonName: json['primary_common_name'] ?? '',
    family: json['family'],
    careLevel: json['care_level'],
    temperament: json['temperament'],
    maxSizeCm: (json['max_size_cm'] as num?)?.toDouble(),
    lifespanYears: (json['lifespan_years'] as num?)?.toDouble(),
    phMin: (json['ph_min'] as num?)?.toDouble(),
    phMax: (json['ph_max'] as num?)?.toDouble(),
    tempMinC: (json['temp_min_c'] as num?)?.toDouble(),
    tempMaxC: (json['temp_max_c'] as num?)?.toDouble(),
    minTankSizeLiters: json['min_tank_size_liters'],
    dietType: json['diet_type'],
    dietNotes: json['diet_notes'],
    breedingNotes: json['breeding_notes'],
    careNotes: json['care_notes'],
    primaryImageUrl: json['primary_image_url'],
    license: json['license'],
    attribution: json['attribution'],
    translation: json['translation'] != null
        ? Map<String, String?>.from(json['translation'])
        : null,
  );

  String get localizedCommonName =>
    translation?['common_name'] ?? primaryCommonName;
  String? get localizedCareNotes =>
    translation?['care_notes'] ?? careNotes;
  String? get localizedBreedingNotes =>
    translation?['breeding_notes'] ?? breedingNotes;
  String? get localizedDietNotes =>
    translation?['diet_notes'] ?? dietNotes;
}
